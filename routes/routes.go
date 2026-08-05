package routes

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"googledominator-backend/config"
	"googledominator-backend/handlers"
	"googledominator-backend/middleware"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	r.GET("/health", handlers.HealthCheck)

	authMiddleware := middleware.AuthMiddleware(cfg)

	// API v1 routes group
	v1 := r.Group("/api/v1")
	{
		// Admin Authentication endpoints
		adminAuth := v1.Group("/admin")
		{
			adminAuth.POST("/register", handlers.AdminRegister(cfg))
			adminAuth.POST("/login", handlers.AdminLogin(cfg))
		}

		// Auth alias group
		userAuth := v1.Group("/auth")
		{
			userAuth.POST("/register", handlers.AdminRegister(cfg))
			userAuth.POST("/login", handlers.AdminLogin(cfg))
		}

		// Pricing endpoints
		pricing := v1.Group("/pricing")
		{
			pricing.GET("", handlers.GetPricingPlans)
			pricing.GET("/:id", handlers.GetPricingPlanByID)
			// Protected POST, PUT, DELETE endpoints
			pricing.POST("", authMiddleware, handlers.CreatePricingPlan)
			pricing.PUT("/:id", authMiddleware, handlers.UpdatePricingPlan)
			pricing.DELETE("/:id", authMiddleware, handlers.DeletePricingPlan)
		}

		// Built for Taxes endpoints
		taxes := v1.Group("/taxes")
		{
			taxes.GET("/services", handlers.GetTaxServices)
			taxes.GET("/services/:id", handlers.GetTaxServiceByID)
			taxes.GET("/forms", handlers.GetTaxFormCategories)
			// Protected POST, PUT, DELETE endpoints
			taxes.POST("/services", authMiddleware, handlers.CreateTaxService)
			taxes.PUT("/services/:id", authMiddleware, handlers.UpdateTaxService)
			taxes.DELETE("/services/:id", authMiddleware, handlers.DeleteTaxService)
			taxes.POST("/forms", authMiddleware, handlers.CreateTaxFormCategory)
			taxes.PUT("/forms/:id", authMiddleware, handlers.UpdateTaxFormCategory)
			taxes.DELETE("/forms/:id", authMiddleware, handlers.DeleteTaxFormCategory)
		}

		// Onboarding Form Submissions endpoints
		onboarding := v1.Group("/onboarding")
		{
			onboarding.POST("", handlers.CreateOnboardingSubmission) // Public form intake
			// Protected Admin routes for viewing, updating and deleting submissions
			onboarding.GET("", authMiddleware, handlers.GetOnboardingSubmissions)
			onboarding.GET("/:id", authMiddleware, handlers.GetOnboardingSubmissionByID)
			onboarding.PUT("/:id", authMiddleware, handlers.UpdateOnboardingSubmission)
			onboarding.DELETE("/:id", authMiddleware, handlers.DeleteOnboardingSubmission)
		}

		// Stripe Payment Integration endpoints
		stripeGroup := v1.Group("/stripe")
		{
			stripeGroup.POST("/checkout-session", handlers.CreateCheckoutSession(cfg))
			stripeGroup.POST("/webhook", handlers.HandleStripeWebhook(cfg))
		}
	}

	return r
}
