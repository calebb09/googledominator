package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"googledominator-backend/config"
	"googledominator-backend/db"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
)

type CreateCheckoutSessionInput struct {
	PlanSlug     string `json:"plan_slug" binding:"required"`
	BillingCycle string `json:"billing_cycle"` // legacy: "monthly" | "yearly" | "annual"
	// BillingType drives the new dual-price logic:
	//   "monthly"           – one-time priceMonthly + monthly recurringPayment (DEFAULT)
	//   "recurring_only" – monthly recurringPayment only
	//   "annual"         – one-time priceAnnual only
	BillingType string `json:"billing_type"`
	Email       string `json:"email" binding:"required,email"`
	SuccessURL  string `json:"success_url"`
	CancelURL   string `json:"cancel_url"`
}

type OrderResponse struct {
	ID                    string    `json:"id"`
	Email                 string    `json:"email"`
	StripeSessionID       string    `json:"stripe_session_id"`
	StripePaymentIntentID string    `json:"stripe_payment_intent_id,omitempty"`
	PlanSlug              string    `json:"plan_slug,omitempty"`
	Amount                float64   `json:"amount"`
	Currency              string    `json:"currency"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
}

var (
	mockOrdersLock sync.RWMutex
	mockOrders     = []OrderResponse{
		{
			ID:              "order-sample-001",
			Email:           "jane@apextax.com",
			StripeSessionID: "cs_test_mock_existing-website_monthly",
			PlanSlug:        "existing-website",
			Amount:          197.00,
			Currency:        "usd",
			Status:          "completed",
			CreatedAt:       time.Now().Add(-48 * time.Hour),
		},
	}
)

func CreateCheckoutSession(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input CreateCheckoutSessionInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		stripe.Key = cfg.StripeSecretKey

		successURL := input.SuccessURL
		if successURL == "" {
			successURL = cfg.FrontendURL + "/checkout/success?session_id={CHECKOUT_SESSION_ID}"
		}

		cancelURL := input.CancelURL
		if cancelURL == "" {
			cancelURL = cfg.FrontendURL + "/checkout/cancel"
		}

		ctx := c.Request.Context()

		var amountInCents int64 = 19700
		var planName = "GoogleDominator Service"
		var stripePriceID string

		// Resolve billing type — new field takes priority, legacy billing_cycle maps as fallback
		billingType := strings.ToLower(strings.TrimSpace(input.BillingType))
		if billingType == "" {
			// Map legacy billing_cycle to billing_type
			cycle := strings.ToLower(strings.TrimSpace(input.BillingCycle))
			switch {
			case cycle == "yearly" || cycle == "annual" || cycle == "annually" || cycle == "year":
				billingType = "annual"
			case cycle == "recurring_only":
				billingType = "recurring_only"
			default:
				// Default: full dual-price checkout (one-time + recurring)
				billingType = "monthly"
			}
		}

		// Plan data resolved from DB or fallback defaults
		var oneTimeCents int64         // priceMonthly      – one-time fee
		var recurringCents int64       // recurringPayment  – charged every month
		var annualCents int64          // priceAnnual       – annual one-time payment
		var stripeAnnualPriceID string // StripePriceIdAnnual – pre-created Stripe Price for annual

		// Attempt dynamic plan lookup from DB
		if db.Instance != nil && db.Instance.IsConnected {
			plan, err := db.Instance.Prisma.PricingPlan.FindUnique(
				db.PricingPlan.Slug.Equals(input.PlanSlug),
			).Exec(ctx)

			if err == nil && plan != nil {
				planName = plan.Name
				oneTimeCents = int64(plan.PriceMonthly * 100)
				recurringCents = int64(plan.RecurringPayment) * 100
				annualCents = int64(plan.PriceAnnual * 100)
				if priceID, ok := plan.StripePriceIDMonthly(); ok {
					stripePriceID = priceID
				}
				if priceID, ok := plan.StripePriceIDAnnual(); ok {
					stripeAnnualPriceID = priceID
				}
			}
		}

		// Fallback to hardcoded default plans if DB didn't return a result
		if planName == "GoogleDominator Service" {
			for _, dp := range defaultPricingPlans {
				if dp.Slug == input.PlanSlug {
					planName = dp.Name
					oneTimeCents = int64(dp.PriceMonthly * 100)
					recurringCents = int64(dp.RecurringPayment) * 100
					annualCents = int64(dp.PriceAnnual * 100)
					stripePriceID = dp.StripePriceIdMonthly
					stripeAnnualPriceID = dp.StripePriceIdAnnual
					break
				}
			}
		}

		// Last-resort hardcoded values for unknown slugs
		if planName == "GoogleDominator Service" {
			switch input.PlanSlug {
			case "existing-website":
				planName = "Existing Website Optimization"
				oneTimeCents = 19700
				recurringCents = 19700
				annualCents = 149700
			case "new-website":
				planName = "New Turnkey Tax Website"
				oneTimeCents = 29700
				recurringCents = 29700
				annualCents = 199700
			case "enterprise-tax":
				planName = "Enterprise Tax Practice Dominator"
				oneTimeCents = 49700
				recurringCents = 49700
				annualCents = 399700
			default:
				planName = "GoogleDominator Plan (" + input.PlanSlug + ")"
				oneTimeCents = 19700
				recurringCents = 19700
				annualCents = 149700
			}
		}

		// Build line items and select session mode based on billing type
		var lineItems []*stripe.CheckoutSessionLineItemParams
		var mode string
		var trialDays int64 // 0 = no trial; >0 = free trial period before recurring charge begins

		switch billingType {
		case "monthly":
			// Dual-price: one-time payment on first invoice + monthly recurring subscription
			// Stripe allows a one-time price_data item alongside a recurring item in subscription mode.
			// The one-time item is charged on the first invoice only and never again.
			lineItems = []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("usd"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String(planName + " – One-Time Payment"),
							Description: stripe.String("One-time payment charged today only"),
						},
						UnitAmount: stripe.Int64(oneTimeCents),
						// No Recurring block = one-time charge on the first invoice only
					},
					Quantity: stripe.Int64(1),
				},
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("usd"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String(planName + " – Monthly Subscription"),
							Description: stripe.String("GoogleDominator Local SEO monthly recurring charge"),
						},
						UnitAmount: stripe.Int64(recurringCents),
						Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
							Interval: stripe.String("month"),
						},
					},
					Quantity: stripe.Int64(1),
				},
			}
			mode = string(stripe.CheckoutSessionModeSubscription)
			amountInCents = oneTimeCents + recurringCents // for DB recording / mock response
			trialDays = 30                                // 1-month free trial before recurring begins

		case "recurring_only":
			// Subscription only — no setup fee
			lineItems = []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("usd"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String(planName + " – Monthly Service"),
							Description: stripe.String("GoogleDominator Local SEO monthly subscription"),
						},
						UnitAmount: stripe.Int64(recurringCents),
						Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
							Interval: stripe.String("month"),
						},
					},
					Quantity: stripe.Int64(1),
				},
			}
			mode = string(stripe.CheckoutSessionModeSubscription)
			amountInCents = recurringCents

		default: // "annual" — one-time annual payment
			if stripeAnnualPriceID != "" {
				// Use the pre-created Stripe Price ID for the annual plan
				lineItems = []*stripe.CheckoutSessionLineItemParams{
					{
						Price:    stripe.String(stripeAnnualPriceID),
						Quantity: stripe.Int64(1),
					},
				}
				mode = string(stripe.CheckoutSessionModePayment)
			} else {
				// Fall back to dynamic pricing using the priceAnnual amount
				lineItems = []*stripe.CheckoutSessionLineItemParams{
					{
						PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
							Currency: stripe.String("usd"),
							ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
								Name:        stripe.String(planName + " (Annual)"),
								Description: stripe.String("GoogleDominator Local SEO & Built for Taxes Plan – Full Year"),
							},
							UnitAmount: stripe.Int64(annualCents),
						},
						Quantity: stripe.Int64(1),
					},
				}
				mode = string(stripe.CheckoutSessionModePayment)
			}
			amountInCents = annualCents
		}

		params := &stripe.CheckoutSessionParams{
			CustomerEmail: stripe.String(input.Email),
			PaymentMethodTypes: stripe.StringSlice([]string{
				"card",
			}),
			LineItems:  lineItems,
			Mode:       stripe.String(mode),
			SuccessURL: stripe.String(successURL),
			CancelURL:  stripe.String(cancelURL),
			Metadata: map[string]string{
				"plan_slug":       input.PlanSlug,
				"billing_type":    billingType,
				"billing_cycle":   input.BillingCycle, // kept for backward compat
				"email":           input.Email,
				"stripe_price_id": stripePriceID,
			},
		}

		// Attach free trial to the subscription when billing_type is "monthly":
		// customer pays the one-time fee today, then gets trialDays free before
		// the monthly recurring charge begins.
		if trialDays > 0 {
			params.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
				TrialPeriodDays: stripe.Int64(trialDays),
				Metadata: map[string]string{
					"plan_slug":    input.PlanSlug,
					"billing_type": billingType,
				},
			}
		}

		// Attempt Stripe API session creation
		sess, err := session.New(params)
		if err != nil {
			stripeErr := err.Error()
			log.Printf("[ERROR] Stripe session.New failed for plan=%s billing_type=%s: %v", input.PlanSlug, billingType, err)

			// --- Smart retry for annual: if a stored Price ID was rejected, retry with dynamic pricing ---
			if billingType == "annual" && stripeAnnualPriceID != "" && annualCents > 0 {
				log.Printf("[INFO] Retrying annual checkout with dynamic price_data (stored price ID may be invalid or test-only)")
				fallbackParams := &stripe.CheckoutSessionParams{
					CustomerEmail:      stripe.String(input.Email),
					PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
					LineItems: []*stripe.CheckoutSessionLineItemParams{
						{
							PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
								Currency: stripe.String("usd"),
								ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
									Name:        stripe.String(planName + " (Annual)"),
									Description: stripe.String("GoogleDominator Local SEO & Built for Taxes Plan – Full Year"),
								},
								UnitAmount: stripe.Int64(annualCents),
							},
							Quantity: stripe.Int64(1),
						},
					},
					Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
					SuccessURL: stripe.String(successURL),
					CancelURL:  stripe.String(cancelURL),
					Metadata: map[string]string{
						"plan_slug":    input.PlanSlug,
						"billing_type": billingType,
						"email":        input.Email,
					},
				}
				sess, err = session.New(fallbackParams)
			}

			// If still failing, determine response based on key type
			if err != nil {
				isLiveKey := strings.HasPrefix(cfg.StripeSecretKey, "sk_live")
				if isLiveKey {
					// Live mode: never return a fake URL — surface the real error
					log.Printf("[ERROR] Stripe LIVE mode error — returning 502: %v", err)
					c.JSON(http.StatusBadGateway, gin.H{
						"success":      false,
						"error":        "Stripe checkout session could not be created",
						"stripe_error": stripeErr,
						"hint":         "Check that the Stripe Price ID stored for this plan is valid and matches your live Stripe account",
					})
					return
				}

				// Test / dev mode: return mock URL so local development still works
				log.Printf("[INFO] Stripe test mode — returning mock checkout URL")
				mockSessionID := fmt.Sprintf("cs_test_mock_%s_%s", input.PlanSlug, billingType)
				mockCheckoutURL := fmt.Sprintf("%s/checkout/mock?session_id=%s", cfg.FrontendURL, mockSessionID)
				amountVal := float64(amountInCents) / 100.0

				if db.Instance != nil && db.Instance.IsConnected {
					_, _ = db.Instance.Prisma.Order.CreateOne(
						db.Order.Email.Set(input.Email),
						db.Order.StripeSessionID.Set(mockSessionID),
						db.Order.Amount.Set(amountVal),
						db.Order.PlanSlug.Set(input.PlanSlug),
						db.Order.Status.Set("pending"),
					).Exec(ctx)
				} else {
					mockOrdersLock.Lock()
					mockOrders = append(mockOrders, OrderResponse{
						ID:              fmt.Sprintf("order-%d", time.Now().UnixNano()),
						Email:           input.Email,
						StripeSessionID: mockSessionID,
						PlanSlug:        input.PlanSlug,
						Amount:          amountVal,
						Currency:        "usd",
						Status:          "pending",
						CreatedAt:       time.Now(),
					})
					mockOrdersLock.Unlock()
				}

				c.JSON(http.StatusOK, gin.H{
					"success":         true,
					"message":         "Stripe Checkout Session created (mock test mode)",
					"session_id":      mockSessionID,
					"checkout_url":    mockCheckoutURL,
					"amount":          amountVal,
					"currency":        "usd",
					"plan_name":       planName,
					"mode":            mode,
					"stripe_price_id": stripeAnnualPriceID,
				})
				return
			}
		}

		// DB recording
		amountVal := float64(amountInCents) / 100.0
		if db.Instance != nil && db.Instance.IsConnected {
			_, _ = db.Instance.Prisma.Order.CreateOne(
				db.Order.Email.Set(input.Email),
				db.Order.StripeSessionID.Set(sess.ID),
				db.Order.Amount.Set(amountVal),
				db.Order.PlanSlug.Set(input.PlanSlug),
				db.Order.Status.Set("pending"),
			).Exec(ctx)
		}

		c.JSON(http.StatusOK, gin.H{
			"success":         true,
			"session_id":      sess.ID,
			"checkout_url":    sess.URL,
			"amount":          amountVal,
			"currency":        "usd",
			"plan_name":       planName,
			"mode":            mode,
			"stripe_price_id": stripePriceID,
		})
	}
}

func GetCheckoutSession(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.Param("session_id")
		stripe.Key = cfg.StripeSecretKey

		// Attempt Stripe session fetch
		sess, err := session.Get(sessionID, nil)
		if err == nil && sess != nil {
			c.JSON(http.StatusOK, gin.H{
				"success":        true,
				"session_id":     sess.ID,
				"payment_status": string(sess.PaymentStatus),
				"status":         string(sess.Status),
				"customer_email": sess.CustomerEmail,
				"amount_total":   float64(sess.AmountTotal) / 100.0,
				"currency":       string(sess.Currency),
				"metadata":       sess.Metadata,
			})
			return
		}

		// Fallback database / mock search
		ctx := c.Request.Context()
		if db.Instance != nil && db.Instance.IsConnected {
			ord, err := db.Instance.Prisma.Order.FindUnique(
				db.Order.StripeSessionID.Equals(sessionID),
			).Exec(ctx)

			if err == nil && ord != nil {
				c.JSON(http.StatusOK, gin.H{
					"success":        true,
					"session_id":     ord.StripeSessionID,
					"payment_status": ord.Status,
					"status":         "complete",
					"customer_email": ord.Email,
					"amount_total":   ord.Amount,
					"currency":       ord.Currency,
				})
				return
			}
		}

		mockOrdersLock.RLock()
		defer mockOrdersLock.RUnlock()

		for _, o := range mockOrders {
			if o.StripeSessionID == sessionID {
				c.JSON(http.StatusOK, gin.H{
					"success":        true,
					"session_id":     o.StripeSessionID,
					"payment_status": o.Status,
					"status":         "complete",
					"customer_email": o.Email,
					"amount_total":   o.Amount,
					"currency":       o.Currency,
				})
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Checkout session not found",
		})
	}
}

func HandleStripeWebhook(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		const MaxBodyBytes = int64(65536)
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
		payload, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		sigHeader := c.GetHeader("Stripe-Signature")
		var event stripe.Event

		if cfg.StripeWebhookSecret != "" && sigHeader != "" {
			event, err = webhook.ConstructEvent(payload, sigHeader, cfg.StripeWebhookSecret)
			if err != nil {
				log.Printf("[WARNING] Webhook signature verification failed: %v", err)
			}
		} else {
			// Unmarshal raw json payload for development / test payload testing
			_ = json.Unmarshal(payload, &event)
		}

		log.Printf("[INFO] Received Stripe Webhook Event: %s (ID: %s)", event.Type, event.ID)
		ctx := c.Request.Context()

		switch event.Type {
		case "checkout.session.completed":
			var sess stripe.CheckoutSession
			if err := json.Unmarshal(event.Data.Raw, &sess); err == nil {
				log.Printf("[INFO] Stripe Checkout Session %s completed for %s", sess.ID, sess.CustomerEmail)

				if db.Instance != nil && db.Instance.IsConnected {
					// Update order status in DB
					_, _ = db.Instance.Prisma.Order.FindUnique(
						db.Order.StripeSessionID.Equals(sess.ID),
					).Update(
						db.Order.Status.Set("completed"),
					).Exec(ctx)

					// Register active subscription in DB if metadata contains plan details
					planSlug := sess.Metadata["plan_slug"]
					billingType := sess.Metadata["billing_type"]
					if billingType == "" {
						billingType = sess.Metadata["billing_cycle"] // fallback for old sessions
					}
					customerEmail := sess.CustomerEmail

					if planSlug != "" && customerEmail != "" {
						subID := sess.ID
						if sess.Subscription != nil {
							subID = sess.Subscription.ID
						}
						subOpts := []db.SubscriptionSetParam{
							db.Subscription.BillingCycle.Set(billingType),
							db.Subscription.Status.Set("active"),
						}
						if sess.Customer != nil {
							subOpts = append(subOpts, db.Subscription.StripeCustomerID.Set(sess.Customer.ID))
						}

						_, _ = db.Instance.Prisma.Subscription.UpsertOne(
							db.Subscription.StripeSubscriptionID.Equals(subID),
						).Create(
							db.Subscription.Email.Set(customerEmail),
							db.Subscription.PlanSlug.Set(planSlug),
							db.Subscription.BillingCycle.Set(billingType),
							db.Subscription.Status.Set("active"),
							db.Subscription.StripeSubscriptionID.Set(subID),
						).Update(subOpts...).Exec(ctx)
					}
				} else {
					mockOrdersLock.Lock()
					for i, o := range mockOrders {
						if o.StripeSessionID == sess.ID {
							mockOrders[i].Status = "completed"
						}
					}
					mockOrdersLock.Unlock()
				}
			}
		case "payment_intent.succeeded":
			log.Println("[INFO] Stripe Payment Intent Succeeded!")
		case "customer.subscription.created", "customer.subscription.updated":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err == nil {
				log.Printf("[INFO] Stripe Subscription %s status: %s", sub.ID, sub.Status)
				if db.Instance != nil && db.Instance.IsConnected {
					_, _ = db.Instance.Prisma.Subscription.FindUnique(
						db.Subscription.StripeSubscriptionID.Equals(sub.ID),
					).Update(
						db.Subscription.Status.Set(string(sub.Status)),
					).Exec(ctx)
				}
			}
		case "customer.subscription.deleted":
			var sub stripe.Subscription
			if err := json.Unmarshal(event.Data.Raw, &sub); err == nil {
				log.Printf("[INFO] Stripe Subscription %s canceled", sub.ID)
				if db.Instance != nil && db.Instance.IsConnected {
					_, _ = db.Instance.Prisma.Subscription.FindUnique(
						db.Subscription.StripeSubscriptionID.Equals(sub.ID),
					).Update(
						db.Subscription.Status.Set("canceled"),
					).Exec(ctx)
				}
			}
		default:
			log.Printf("[INFO] Handled event type: %s", event.Type)
		}

		c.JSON(http.StatusOK, gin.H{
			"received": true,
			"event":    event.Type,
		})
	}
}

func GetOrders(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		orders, err := db.Instance.Prisma.Order.FindMany().Exec(ctx)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    orders,
			})
			return
		}
	}

	mockOrdersLock.RLock()
	defer mockOrdersLock.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mockOrders,
	})
}
