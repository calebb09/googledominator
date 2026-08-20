package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"googledominator-backend/db"
)

type PricingPlanResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Slug                 string   `json:"slug"`
	Description          string   `json:"description"`
	PriceMonthly         float64  `json:"price_monthly"`
	PriceAnnual          float64  `json:"price_annual"`
	StripePriceIdMonthly string   `json:"stripe_price_id_monthly"`
	StripePriceIdAnnual  string   `json:"stripe_price_id_annual"`
	Features             []string `json:"features"`
	RecurringPayment     int      `json:"recurring_payment"`
	IsPopular            bool     `json:"is_popular"`
	IsActive             bool     `json:"is_active"`
}

type CreatePricingPlanInput struct {
	Name                 string   `json:"name" binding:"required"`
	Slug                 string   `json:"slug" binding:"required"`
	Description          string   `json:"description"`
	PriceMonthly         float64  `json:"price_monthly" binding:"required"`
	PriceAnnual          float64  `json:"price_annual" binding:"required"`
	StripePriceIdMonthly string   `json:"stripe_price_id_monthly"`
	StripePriceIdAnnual  string   `json:"stripe_price_id_annual"`
	Features             []string `json:"features" binding:"required"`
	RecurringPayment     *int     `json:"recurring_payment"`
	IsPopular            bool     `json:"is_popular"`
}

var defaultPricingPlans = []PricingPlanResponse{
	{
		ID:                   "plan-existing-website",
		Name:                 "Existing Website Plan",
		Slug:                 "existing-website",
		Description:          "Ideal for established businesses and tax practices looking to dominate Google rankings with an existing domain.",
		PriceMonthly:         197.00,
		PriceAnnual:          1497.00,
		StripePriceIdMonthly: "price_1MockExistingWebsiteMonthly",
		StripePriceIdAnnual:  "price_1MockExistingWebsiteAnnual",
		Features: []string{
			"Google Business Profile Optimization",
			"Local Tax Service Keywords Targeting",
			"On-Page SEO & Schema Markup for Tax Firms",
			"Monthly Citation Cleanup & Building",
			"Review Generation Engine & SMS Alerts",
			"Dedicated Account Manager & Monthly Reports",
		},
		RecurringPayment: 197,
		IsPopular:        false,
		IsActive:         true,
	},
	{
		ID:                   "plan-new-website",
		Name:                 "New / Turnkey Website Plan",
		Slug:                 "new-website",
		Description:          "Complete turnkey solution for new tax firms and local businesses needing custom high-converting website build + local SEO domination.",
		PriceMonthly:         297.00,
		PriceAnnual:          1997.00,
		StripePriceIdMonthly: "price_1MockNewWebsiteMonthly",
		StripePriceIdAnnual:  "price_1MockNewWebsiteAnnual",
		Features: []string{
			"Custom Tax Firm High-Converting Website Build",
			"Full Google Business Profile Verification & Optimization",
			"Built-for-Taxes Form 1040/1099 Client Lead Funnels",
			"Speed-Optimized Next.js/Go Infrastructure",
			"Automated Lead Notification via Email/SMS",
			"Google Local Service Ads (LSA) Integration",
			"Priority 24/7 VIP Support",
		},
		RecurringPayment: 297,
		IsPopular:        true,
		IsActive:         true,
	},
	{
		ID:                   "plan-enterprise-tax",
		Name:                 "Enterprise Tax Practice Dominator",
		Slug:                 "enterprise-tax",
		Description:          "Multi-location tax firms, CPA franchises, and large accounting practices seeking complete market control.",
		PriceMonthly:         497.00,
		PriceAnnual:          3997.00,
		StripePriceIdMonthly: "price_1MockEnterpriseMonthly",
		StripePriceIdAnnual:  "price_1MockEnterpriseAnnual",
		Features: []string{
			"Multi-Location Google Maps Dominator Strategy",
			"Advanced Tax Calculator & Document Upload Portal",
			"Custom Stripe Checkout & Client Deposit System",
			"Automated Review & Reputation Management Suite",
			"Bi-Weekly Strategy Calls with Growth Executive",
			"100% Ranking Guarantee or Money Back",
		},
		RecurringPayment: 497,
		IsPopular:        false,
		IsActive:         true,
	},
}

func GetPricingPlans(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance == nil || !db.Instance.IsConnected {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Database is unavailable",
		})
		return
	}

	plans, err := db.Instance.Prisma.PricingPlan.FindMany(
		db.PricingPlan.IsActive.Equals(true),
	).Exec(ctx)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch pricing plans from PostgreSQL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch pricing plans from database",
		})
		return
	}

	resp := []PricingPlanResponse{}
	for _, p := range plans {
		featuresList := []string{}
		if p.Features != "" {
			featuresList = strings.Split(p.Features, "||")
		}
		monthlyStripe := ""
		if val, ok := p.StripePriceIDMonthly(); ok {
			monthlyStripe = val
		}
		annualStripe := ""
		if val, ok := p.StripePriceIDAnnual(); ok {
			annualStripe = val
		}
		resp = append(resp, PricingPlanResponse{
			ID:                   p.ID,
			Name:                 p.Name,
			Slug:                 p.Slug,
			Description:          p.Description,
			PriceMonthly:         p.PriceMonthly,
			PriceAnnual:          p.PriceAnnual,
			StripePriceIdMonthly: monthlyStripe,
			StripePriceIdAnnual:  annualStripe,
			Features:             featuresList,
			RecurringPayment:     p.RecurringPayment,
			IsPopular:            p.IsPopular,
			IsActive:             p.IsActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

func GetPricingPlanByID(c *gin.Context) {
	idOrSlug := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance == nil || !db.Instance.IsConnected {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Database is unavailable",
		})
		return
	}

	plan, err := db.Instance.Prisma.PricingPlan.FindFirst(
		db.PricingPlan.Or(
			db.PricingPlan.ID.Equals(idOrSlug),
			db.PricingPlan.Slug.Equals(idOrSlug),
		),
	).Exec(ctx)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch pricing plan from PostgreSQL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch pricing plan from database",
		})
		return
	}

	if plan != nil {
		featuresList := []string{}
		if plan.Features != "" {
			featuresList = strings.Split(plan.Features, "||")
		}
		monthlyStripe := ""
		if val, ok := plan.StripePriceIDMonthly(); ok {
			monthlyStripe = val
		}
		annualStripe := ""
		if val, ok := plan.StripePriceIDAnnual(); ok {
			annualStripe = val
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": PricingPlanResponse{
				ID:                   plan.ID,
				Name:                 plan.Name,
				Slug:                 plan.Slug,
				Description:          plan.Description,
				PriceMonthly:         plan.PriceMonthly,
				PriceAnnual:          plan.PriceAnnual,
				StripePriceIdMonthly: monthlyStripe,
				StripePriceIdAnnual:  annualStripe,
				Features:             featuresList,
				RecurringPayment:     plan.RecurringPayment,
				IsPopular:            plan.IsPopular,
				IsActive:             plan.IsActive,
			},
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Pricing plan not found",
	})
}

func CreatePricingPlan(c *gin.Context) {
	var input CreatePricingPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	featuresStr := strings.Join(input.Features, "||")
	var recurringPayment int
	if input.RecurringPayment != nil {
		recurringPayment = *input.RecurringPayment
	}
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		plan, err := db.Instance.Prisma.PricingPlan.CreateOne(
			db.PricingPlan.Name.Set(input.Name),
			db.PricingPlan.Slug.Set(input.Slug),
			db.PricingPlan.Description.Set(input.Description),
			db.PricingPlan.PriceMonthly.Set(input.PriceMonthly),
			db.PricingPlan.PriceAnnual.Set(input.PriceAnnual),
			db.PricingPlan.Features.Set(featuresStr),
			db.PricingPlan.RecurringPayment.Set(recurringPayment),
			db.PricingPlan.StripePriceIDMonthly.Set(input.StripePriceIdMonthly),
			db.PricingPlan.StripePriceIDAnnual.Set(input.StripePriceIdAnnual),
			db.PricingPlan.IsPopular.Set(input.IsPopular),
		).Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Pricing plan created successfully",
			"data":    plan,
		})
		return
	}

	newPlan := PricingPlanResponse{
		ID:                   "plan-" + input.Slug,
		Name:                 input.Name,
		Slug:                 input.Slug,
		Description:          input.Description,
		PriceMonthly:         input.PriceMonthly,
		PriceAnnual:          input.PriceAnnual,
		StripePriceIdMonthly: input.StripePriceIdMonthly,
		StripePriceIdAnnual:  input.StripePriceIdAnnual,
		Features:             input.Features,
		RecurringPayment:     recurringPayment,
		IsPopular:            input.IsPopular,
		IsActive:             true,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Pricing plan created (mock memory mode)",
		"data":    newPlan,
	})
}

type UpdatePricingPlanInput struct {
	Name                 string   `json:"name"`
	Slug                 string   `json:"slug"`
	Description          string   `json:"description"`
	PriceMonthly         float64  `json:"price_monthly"`
	PriceAnnual          float64  `json:"price_annual"`
	StripePriceIdMonthly string   `json:"stripe_price_id_monthly"`
	StripePriceIdAnnual  string   `json:"stripe_price_id_annual"`
	Features             []string `json:"features"`
	RecurringPayment     *int     `json:"recurring_payment"`
	IsPopular            *bool    `json:"is_popular"`
	IsActive             *bool    `json:"is_active"`
}

func UpdatePricingPlan(c *gin.Context) {
	id := c.Param("id")
	var input UpdatePricingPlanInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		var updates []db.PricingPlanSetParam
		if input.Name != "" {
			updates = append(updates, db.PricingPlan.Name.Set(input.Name))
		}
		if input.Slug != "" {
			updates = append(updates, db.PricingPlan.Slug.Set(input.Slug))
		}
		if input.Description != "" {
			updates = append(updates, db.PricingPlan.Description.Set(input.Description))
		}
		if input.PriceMonthly > 0 {
			updates = append(updates, db.PricingPlan.PriceMonthly.Set(input.PriceMonthly))
		}
		if input.PriceAnnual > 0 {
			updates = append(updates, db.PricingPlan.PriceAnnual.Set(input.PriceAnnual))
		}
		if len(input.Features) > 0 {
			featuresStr := strings.Join(input.Features, "||")
			updates = append(updates, db.PricingPlan.Features.Set(featuresStr))
		}
		if input.StripePriceIdMonthly != "" {
			updates = append(updates, db.PricingPlan.StripePriceIDMonthly.Set(input.StripePriceIdMonthly))
		}
		if input.StripePriceIdAnnual != "" {
			updates = append(updates, db.PricingPlan.StripePriceIDAnnual.Set(input.StripePriceIdAnnual))
		}
		if input.IsPopular != nil {
			updates = append(updates, db.PricingPlan.IsPopular.Set(*input.IsPopular))
		}
		if input.RecurringPayment != nil {
			updates = append(updates, db.PricingPlan.RecurringPayment.Set(*input.RecurringPayment))
		}
		if input.IsActive != nil {
			updates = append(updates, db.PricingPlan.IsActive.Set(*input.IsActive))
		}

		plan, err := db.Instance.Prisma.PricingPlan.FindUnique(
			db.PricingPlan.ID.Equals(id),
		).Update(updates...).Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pricing plan updated successfully",
			"data":    plan,
		})
		return
	}

	for i, p := range defaultPricingPlans {
		if p.ID == id || p.Slug == id {
			if input.Name != "" {
				defaultPricingPlans[i].Name = input.Name
			}
			if input.Slug != "" {
				defaultPricingPlans[i].Slug = input.Slug
			}
			if input.Description != "" {
				defaultPricingPlans[i].Description = input.Description
			}
			if input.PriceMonthly > 0 {
				defaultPricingPlans[i].PriceMonthly = input.PriceMonthly
			}
			if input.PriceAnnual > 0 {
				defaultPricingPlans[i].PriceAnnual = input.PriceAnnual
			}
			if len(input.Features) > 0 {
				defaultPricingPlans[i].Features = input.Features
			}
			if input.IsPopular != nil {
				defaultPricingPlans[i].IsPopular = *input.IsPopular
			}
			if input.RecurringPayment != nil {
				defaultPricingPlans[i].RecurringPayment = *input.RecurringPayment
			}
			if input.IsActive != nil {
				defaultPricingPlans[i].IsActive = *input.IsActive
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Pricing plan updated (mock memory mode)",
				"data":    defaultPricingPlans[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Pricing plan not found",
	})
}

func DeletePricingPlan(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		_, err := db.Instance.Prisma.PricingPlan.FindUnique(
			db.PricingPlan.ID.Equals(id),
		).Delete().Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pricing plan deleted successfully",
		})
		return
	}

	for i, p := range defaultPricingPlans {
		if p.ID == id || p.Slug == id {
			defaultPricingPlans = append(defaultPricingPlans[:i], defaultPricingPlans[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Pricing plan deleted (mock memory mode)",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Pricing plan not found",
	})
}
