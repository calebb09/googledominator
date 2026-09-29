package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"googledominator-backend/config"
)

func setupTestRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/v1/stripe/checkout-session", CreateCheckoutSession(cfg))
	return router
}

func TestCreateCheckoutSession_BillingCycle(t *testing.T) {
	cfg := &config.Config{
		StripeSecretKey: "sk_test_mock_secret",
		FrontendURL:     "https://googledominator.co",
	}
	router := setupTestRouter(cfg)

	tests := []struct {
		name                 string
		planSlug             string
		billingCycle         string
		expectedAmount       float64
		expectedStripePriceID string
	}{
		{
			name:                 "Monthly billing cycle gets monthly price and monthly stripe price id",
			planSlug:             "existing-website",
			billingCycle:         "monthly",
			expectedAmount:       197.00,
			expectedStripePriceID: "price_1MockExistingWebsiteMonthly",
		},
		{
			name:                 "Yearly billing cycle gets annual price and annual stripe price id",
			planSlug:             "existing-website",
			billingCycle:         "yearly",
			expectedAmount:       1497.00,
			expectedStripePriceID: "price_1MockExistingWebsiteAnnual",
		},
		{
			name:                 "Annual billing cycle gets annual price and annual stripe price id",
			planSlug:             "existing-website",
			billingCycle:         "annual",
			expectedAmount:       1497.00,
			expectedStripePriceID: "price_1MockExistingWebsiteAnnual",
		},
		{
			name:                 "New website plan with yearly billing cycle",
			planSlug:             "new-website",
			billingCycle:         "yearly",
			expectedAmount:       1997.00,
			expectedStripePriceID: "price_1MockNewWebsiteAnnual",
		},
		{
			name:                 "New website plan with monthly billing cycle",
			planSlug:             "new-website",
			billingCycle:         "monthly",
			expectedAmount:       297.00,
			expectedStripePriceID: "price_1MockNewWebsiteMonthly",
		},
		{
			name:                 "Enterprise tax plan with yearly billing cycle (case-insensitive)",
			planSlug:             "enterprise-tax",
			billingCycle:         "YEARLY",
			expectedAmount:       3997.00,
			expectedStripePriceID: "price_1MockEnterpriseAnnual",
		},
		{
			name:                 "Enterprise tax plan with monthly billing cycle",
			planSlug:             "enterprise-tax",
			billingCycle:         "monthly",
			expectedAmount:       497.00,
			expectedStripePriceID: "price_1MockEnterpriseMonthly",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := map[string]string{
				"plan_slug":     tc.planSlug,
				"billing_cycle": tc.billingCycle,
				"email":         "test@example.com",
			}
			body, _ := json.Marshal(payload)

			req, _ := http.NewRequest(http.MethodPost, "/api/v1/stripe/checkout-session", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to decode response JSON: %v", err)
			}

			if resp["success"] != true {
				t.Errorf("expected success to be true, got %v", resp["success"])
			}

			amount, ok := resp["amount"].(float64)
			if !ok || amount != tc.expectedAmount {
				t.Errorf("expected amount %v, got %v", tc.expectedAmount, resp["amount"])
			}

			priceID, _ := resp["stripe_price_id"].(string)
			if priceID != tc.expectedStripePriceID {
				t.Errorf("expected stripe_price_id %q, got %q", tc.expectedStripePriceID, priceID)
			}
		})
	}
}
