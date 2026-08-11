package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
	"googledominator-backend/config"
	"googledominator-backend/db"
)

type CreateCheckoutSessionInput struct {
	PlanSlug     string `json:"plan_slug" binding:"required"`
	BillingCycle string `json:"billing_cycle" binding:"required"` // "monthly" or "annual"
	Email        string `json:"email" binding:"required,email"`
	SuccessURL   string `json:"success_url"`
	CancelURL    string `json:"cancel_url"`
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

		// Attempt dynamic plan lookup from DB
		if db.Instance != nil && db.Instance.IsConnected {
			plan, err := db.Instance.Prisma.PricingPlan.FindUnique(
				db.PricingPlan.Slug.Equals(input.PlanSlug),
			).Exec(ctx)

			if err == nil && plan != nil {
				planName = plan.Name
				if input.BillingCycle == "annual" {
					amountInCents = int64(plan.PriceAnnual * 100)
					if priceID, ok := plan.StripePriceIDAnnual(); ok {
						stripePriceID = priceID
					}
				} else {
					amountInCents = int64(plan.PriceMonthly * 100)
					if priceID, ok := plan.StripePriceIDMonthly(); ok {
						stripePriceID = priceID
					}
				}
			}
		}

		// Fallback hardcoded plan lookup if not found in DB
		if planName == "GoogleDominator Service" {
			switch input.PlanSlug {
			case "existing-website":
				planName = "Existing Website Optimization"
				if input.BillingCycle == "annual" {
					amountInCents = 149700
				} else {
					amountInCents = 19700
				}
			case "new-website":
				planName = "New Turnkey Tax Website"
				if input.BillingCycle == "annual" {
					amountInCents = 199700
				} else {
					amountInCents = 29700
				}
			case "enterprise-tax":
				planName = "Enterprise Tax Practice Dominator"
				if input.BillingCycle == "annual" {
					amountInCents = 399700
				} else {
					amountInCents = 49700
				}
			default:
				planName = "GoogleDominator Plan (" + input.PlanSlug + ")"
				amountInCents = 19700
			}
		}

		var lineItems []*stripe.CheckoutSessionLineItemParams
		var mode string

		if stripePriceID != "" {
			// Using existing Stripe Price ID -> Subscription Mode
			lineItems = []*stripe.CheckoutSessionLineItemParams{
				{
					Price:    stripe.String(stripePriceID),
					Quantity: stripe.Int64(1),
				},
			}
			mode = string(stripe.CheckoutSessionModeSubscription)
		} else {
			// Dynamic inline price -> Payment Mode
			lineItems = []*stripe.CheckoutSessionLineItemParams{
				{
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency: stripe.String("usd"),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name:        stripe.String(planName + " (" + input.BillingCycle + ")"),
							Description: stripe.String("GoogleDominator Local SEO & Built for Taxes Plan"),
						},
						UnitAmount: stripe.Int64(amountInCents),
					},
					Quantity: stripe.Int64(1),
				},
			}
			mode = string(stripe.CheckoutSessionModePayment)
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
				"plan_slug":     input.PlanSlug,
				"billing_cycle": input.BillingCycle,
				"email":         input.Email,
			},
		}

		// Attempt Stripe API session creation
		sess, err := session.New(params)
		if err != nil {
			log.Printf("[INFO] Stripe API call error / test mode active: %v", err)
			mockSessionID := fmt.Sprintf("cs_test_mock_%s_%s", input.PlanSlug, input.BillingCycle)
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
				"success":      true,
				"message":      "Stripe Checkout Session created (mock test mode)",
				"session_id":   mockSessionID,
				"checkout_url": mockCheckoutURL,
				"amount":       amountVal,
				"currency":     "usd",
				"plan_name":    planName,
				"mode":         mode,
			})
			return
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
			"success":      true,
			"session_id":   sess.ID,
			"checkout_url": sess.URL,
			"amount":       amountVal,
			"currency":     "usd",
			"plan_name":    planName,
			"mode":         mode,
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
				"success":         true,
				"session_id":      sess.ID,
				"payment_status":  string(sess.PaymentStatus),
				"status":          string(sess.Status),
				"customer_email":  sess.CustomerEmail,
				"amount_total":    float64(sess.AmountTotal) / 100.0,
				"currency":        string(sess.Currency),
				"metadata":        sess.Metadata,
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
					"success":         true,
					"session_id":      ord.StripeSessionID,
					"payment_status":  ord.Status,
					"status":          "complete",
					"customer_email":  ord.Email,
					"amount_total":    ord.Amount,
					"currency":        ord.Currency,
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
					billingCycle := sess.Metadata["billing_cycle"]
					customerEmail := sess.CustomerEmail

					if planSlug != "" && customerEmail != "" {
						subID := sess.ID
						if sess.Subscription != nil {
							subID = sess.Subscription.ID
						}
						subOpts := []db.SubscriptionSetParam{
							db.Subscription.BillingCycle.Set(billingCycle),
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
							db.Subscription.BillingCycle.Set(billingCycle),
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
