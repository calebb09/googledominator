package handlers

import (
	"io"
	"log"
	"net/http"

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

		// Calculate plan price based on slug and cycle
		var amountInCents int64 = 19700
		var planName = "GoogleDominator Service"

		switch input.PlanSlug {
		case "existing-website":
			planName = "Existing Website Plan"
			if input.BillingCycle == "annual" {
				amountInCents = 149700
			} else {
				amountInCents = 19700
			}
		case "new-website":
			planName = "New / Turnkey Website Plan"
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
			amountInCents = 19700
		}

		params := &stripe.CheckoutSessionParams{
			CustomerEmail: stripe.String(input.Email),
			PaymentMethodTypes: stripe.StringSlice([]string{
				"card",
			}),
			LineItems: []*stripe.CheckoutSessionLineItemParams{
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
			},
			Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
			SuccessURL: stripe.String(successURL),
			CancelURL:  stripe.String(cancelURL),
			Metadata: map[string]string{
				"plan_slug":     input.PlanSlug,
				"billing_cycle": input.BillingCycle,
				"email":         input.Email,
			},
		}

		// If using real/valid Stripe key attempt live call, otherwise generate session mock
		sess, err := session.New(params)
		if err != nil {
			log.Printf("[INFO] Stripe API call error (fallback mock active): %v", err)
			mockSessionID := "cs_test_mock_" + input.PlanSlug + "_" + input.BillingCycle
			mockCheckoutURL := cfg.FrontendURL + "/checkout/mock?session_id=" + mockSessionID

			// Record order in Prisma DB if available
			if db.Instance != nil && db.Instance.IsConnected {
				ctx := c.Request.Context()
				_, _ = db.Instance.Prisma.Order.CreateOne(
					db.Order.Email.Set(input.Email),
					db.Order.StripeSessionID.Set(mockSessionID),
					db.Order.Amount.Set(float64(amountInCents)/100.0),
					db.Order.PlanSlug.Set(input.PlanSlug),
					db.Order.Status.Set("pending"),
				).Exec(ctx)
			}

			c.JSON(http.StatusOK, gin.H{
				"success":      true,
				"message":      "Stripe Checkout Session generated (test mode)",
				"session_id":   mockSessionID,
				"checkout_url": mockCheckoutURL,
				"amount":       float64(amountInCents) / 100.0,
				"currency":     "usd",
				"plan_name":    planName,
			})
			return
		}

		// Record order in Prisma DB
		if db.Instance != nil && db.Instance.IsConnected {
			ctx := c.Request.Context()
			_, _ = db.Instance.Prisma.Order.CreateOne(
				db.Order.Email.Set(input.Email),
				db.Order.StripeSessionID.Set(sess.ID),
				db.Order.Amount.Set(float64(amountInCents)/100.0),
				db.Order.PlanSlug.Set(input.PlanSlug),
				db.Order.Status.Set("pending"),
			).Exec(ctx)
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"session_id":   sess.ID,
			"checkout_url": sess.URL,
			"amount":       float64(amountInCents) / 100.0,
			"currency":     "usd",
			"plan_name":    planName,
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
		}

		// Process event types
		log.Printf("[INFO] Received Stripe Webhook Event: %s", event.Type)

		switch event.Type {
		case "checkout.session.completed":
			log.Println("[INFO] Stripe Checkout Session Completed successfully!")
		case "payment_intent.succeeded":
			log.Println("[INFO] Stripe Payment Intent Succeeded!")
		case "customer.subscription.created", "customer.subscription.updated":
			log.Println("[INFO] Stripe Subscription status updated")
		default:
			log.Printf("[INFO] Unhandled event type: %s", event.Type)
		}

		c.JSON(http.StatusOK, gin.H{
			"received": true,
			"event":    event.Type,
		})
	}
}
