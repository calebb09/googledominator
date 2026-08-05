package db

import (
	"context"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func SeedDB(client *PrismaClient) {
	ctx := context.Background()
	log.Println("[INFO] Checking PostgreSQL database tables for initial seeding...")

	// 1. Seed Pricing Plans
	pricingCount, err := client.PricingPlan.FindMany().Exec(ctx)
	if err == nil && len(pricingCount) == 0 {
		log.Println("[INFO] Seeding initial Pricing Plans into PostgreSQL...")
		plans := []struct {
			name         string
			slug         string
			description  string
			monthly      float64
			annual       float64
			monthlyStripe string
			annualStripe  string
			features     string
			isPopular    bool
		}{
			{
				name:         "Existing Website Plan",
				slug:         "existing-website",
				description:  "Ideal for established businesses and tax practices looking to dominate Google rankings with an existing domain.",
				monthly:      197.00,
				annual:       1497.00,
				monthlyStripe: "price_1MockExistingWebsiteMonthly",
				annualStripe:  "price_1MockExistingWebsiteAnnual",
				features:     "Google Business Profile Optimization||Local Tax Service Keywords Targeting||On-Page SEO & Schema Markup for Tax Firms||Monthly Citation Cleanup & Building||Review Generation Engine & SMS Alerts||Dedicated Account Manager & Monthly Reports",
				isPopular:    false,
			},
			{
				name:         "New / Turnkey Website Plan",
				slug:         "new-website",
				description:  "Complete turnkey solution for new tax firms and local businesses needing custom high-converting website build + local SEO domination.",
				monthly:      297.00,
				annual:       1997.00,
				monthlyStripe: "price_1MockNewWebsiteMonthly",
				annualStripe:  "price_1MockNewWebsiteAnnual",
				features:     "Custom Tax Firm High-Converting Website Build||Full Google Business Profile Verification & Optimization||Built-for-Taxes Form 1040/1099 Client Lead Funnels||Speed-Optimized Next.js/Go Infrastructure||Automated Lead Notification via Email/SMS||Google Local Service Ads (LSA) Integration||Priority 24/7 VIP Support",
				isPopular:    true,
			},
			{
				name:         "Enterprise Tax Practice Dominator",
				slug:         "enterprise-tax",
				description:  "Multi-location tax firms, CPA franchises, and large accounting practices seeking complete market control.",
				monthly:      497.00,
				annual:       3997.00,
				monthlyStripe: "price_1MockEnterpriseMonthly",
				annualStripe:  "price_1MockEnterpriseAnnual",
				features:     "Multi-Location Google Maps Dominator Strategy||Advanced Tax Calculator & Document Upload Portal||Custom Stripe Checkout & Client Deposit System||Automated Review & Reputation Management Suite||Bi-Weekly Strategy Calls with Growth Executive||100% Ranking Guarantee or Money Back",
				isPopular:    false,
			},
		}

		for _, p := range plans {
			_, err := client.PricingPlan.CreateOne(
				PricingPlan.Name.Set(p.name),
				PricingPlan.Slug.Set(p.slug),
				PricingPlan.Description.Set(p.description),
				PricingPlan.PriceMonthly.Set(p.monthly),
				PricingPlan.PriceAnnual.Set(p.annual),
				PricingPlan.Features.Set(p.features),
				PricingPlan.StripePriceIDMonthly.Set(p.monthlyStripe),
				PricingPlan.StripePriceIDAnnual.Set(p.annualStripe),
				PricingPlan.IsPopular.Set(p.isPopular),
				PricingPlan.IsActive.Set(true),
			).Exec(ctx)
			if err != nil {
				log.Printf("[WARNING] Failed to seed pricing plan '%s': %v", p.slug, err)
			}
		}
		log.Println("[INFO] Pricing Plans seeded successfully into PostgreSQL!")
	}

	// 2. Seed Tax Services
	taxServicesCount, err := client.TaxService.FindMany().Exec(ctx)
	if err == nil && len(taxServicesCount) == 0 {
		log.Println("[INFO] Seeding initial Tax Services into PostgreSQL...")
		services := []struct {
			title       string
			slug        string
			description string
			category    string
			forms       string
			price       float64
		}{
			{
				title:       "Google Maps Domination for Tax Firms",
				slug:        "google-maps-tax-domination",
				description: "Target high-intent local taxpayers searching for CPA, Tax Prep, and Accountant near them during peak tax season.",
				category:    "Local SEO & GBP",
				forms:       "Form 1040||Form 1099||Schedule C||Form 1120S",
				price:       297.00,
			},
			{
				title:       "Built-for-Taxes Digital Client Intake Funnel",
				slug:        "tax-digital-intake-funnel",
				description: "Automated digital intake system for clients to submit Form 1040, W-2s, and 1099s securely online.",
				category:    "Lead Generation & Intake",
				forms:       "Form W-2||Form 1099-NEC||Form 1099-MISC||Form 1040-SR",
				price:       197.00,
			},
			{
				title:       "Seasonal Tax Sprint Campaign (Jan - April Peak)",
				slug:        "seasonal-tax-sprint-campaign",
				description: "Turnkey marketing and local search takeover engineered specifically for high-volume tax season profitability.",
				category:    "Seasonal Marketing",
				forms:       "Form 1040||Form 1120||Form 1065||Form 2848",
				price:       497.00,
			},
		}

		for _, s := range services {
			_, err := client.TaxService.CreateOne(
				TaxService.Title.Set(s.title),
				TaxService.Slug.Set(s.slug),
				TaxService.Description.Set(s.description),
				TaxService.Category.Set(s.category),
				TaxService.FormsIncluded.Set(s.forms),
				TaxService.PriceEstimate.Set(s.price),
				TaxService.IsActive.Set(true),
			).Exec(ctx)
			if err != nil {
				log.Printf("[WARNING] Failed to seed tax service '%s': %v", s.slug, err)
			}
		}
		log.Println("[INFO] Tax Services seeded successfully into PostgreSQL!")
	}

	// 3. Seed Tax Form Categories
	formCatsCount, err := client.TaxFormCategory.FindMany().Exec(ctx)
	if err == nil && len(formCatsCount) == 0 {
		log.Println("[INFO] Seeding initial Tax Form Categories into PostgreSQL...")
		cats := []struct {
			name        string
			code        string
			description string
			formType    string
		}{
			{
				name:        "Individual & Family Tax Returns",
				code:        "INDIVIDUAL-TAX",
				description: "Federal and State tax forms for individual taxpayers, freelancers, and sole proprietors.",
				formType:    "Individual",
			},
			{
				name:        "Corporate & Business Tax Filing",
				code:        "BUSINESS-TAX",
				description: "Entities, Partnerships, S-Corporations, and LLC business tax returns.",
				formType:    "Corporate",
			},
			{
				name:        "Tax Resolution & IRS Representation",
				code:        "TAX-RESOLUTION",
				description: "IRS power of attorney, offer in compromise, and audit representation forms.",
				formType:    "Resolution",
			},
		}

		for _, c := range cats {
			_, err := client.TaxFormCategory.CreateOne(
				TaxFormCategory.Name.Set(c.name),
				TaxFormCategory.Code.Set(c.code),
				TaxFormCategory.Description.Set(c.description),
				TaxFormCategory.FormType.Set(c.formType),
				TaxFormCategory.IsActive.Set(true),
			).Exec(ctx)
			if err != nil {
				log.Printf("[WARNING] Failed to seed tax form category '%s': %v", c.code, err)
			}
		}
		log.Println("[INFO] Tax Form Categories seeded successfully into PostgreSQL!")
	}

	// 4. Seed Onboarding Submission
	onboardingCount, err := client.OnboardingSubmission.FindMany().Exec(ctx)
	if err == nil && len(onboardingCount) == 0 {
		log.Println("[INFO] Seeding initial Onboarding Submission into PostgreSQL...")
		_, err := client.OnboardingSubmission.CreateOne(
			OnboardingSubmission.BusinessName.Set("Apex Tax & Financial Services"),
			OnboardingSubmission.ContactName.Set("Jane Doe, CPA"),
			OnboardingSubmission.Phone.Set("+1-555-019-2831"),
			OnboardingSubmission.Email.Set("jane@apextax.com"),
			OnboardingSubmission.HasExistingWebsite.Set(true),
			OnboardingSubmission.WebsiteURL.Set("https://apextax.com"),
			OnboardingSubmission.HasGoogleBusinessProfile.Set(true),
			OnboardingSubmission.GbpLink.Set("https://maps.google.com/?cid=123456"),
			OnboardingSubmission.StreetAddress.Set("100 Main St, Suite 400"),
			OnboardingSubmission.City.Set("Dallas"),
			OnboardingSubmission.State.Set("TX"),
			OnboardingSubmission.ZipCode.Set("75201"),
			OnboardingSubmission.PrimaryCategory.Set("Tax Preparation & CPA Practice"),
			OnboardingSubmission.ServicesOffered.Set("Form 1040, Form 1120, Payroll, Tax Resolution"),
			OnboardingSubmission.TargetLocations.Set("Dallas TX, Fort Worth TX, Plano TX"),
			OnboardingSubmission.Keywords.Set("tax prep dallas, CPA near me, tax resolution tax forms"),
			OnboardingSubmission.VisitModel.Set("Both In-Person & Online"),
			OnboardingSubmission.ConsentTransactional.Set(true),
			OnboardingSubmission.ConsentMarketing.Set(true),
			OnboardingSubmission.Status.Set("IN_REVIEW"),
		).Exec(ctx)
		if err != nil {
			log.Printf("[WARNING] Failed to seed onboarding submission: %v", err)
		} else {
			log.Println("[INFO] Onboarding Submission seeded successfully into PostgreSQL!")
		}
	}

	// 5. Seed Admin User
	adminCount, err := client.Admin.FindMany().Exec(ctx)
	if err == nil && len(adminCount) == 0 {
		log.Println("[INFO] Seeding default Admin user into PostgreSQL...")
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
		_, err := client.Admin.CreateOne(
			Admin.Email.Set("admin@googledominator.co"),
			Admin.Password.Set(string(hashedPassword)),
			Admin.Name.Set("GoogleDominator Admin"),
			Admin.Role.Set("ADMIN"),
		).Exec(ctx)
		if err != nil {
			log.Printf("[WARNING] Failed to seed default admin user: %v", err)
		} else {
			log.Println("[INFO] Default Admin user (admin@googledominator.co) seeded successfully into PostgreSQL!")
		}
	}
}

// Helper to convert delimited string to string array for API JSON outputs
func SplitDelimited(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	return strings.Split(s, "||")
}
