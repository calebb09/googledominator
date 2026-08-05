package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"googledominator-backend/db"
)

type TaxFormItem struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TaxFormCategoryResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Code        string        `json:"code"`
	Description string        `json:"description"`
	FormType    string        `json:"form_type"`
	Forms       []TaxFormItem `json:"forms"`
	IsActive    bool          `json:"is_active"`
}

type TaxServiceResponse struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	FormsIncluded []string `json:"forms_included"`
	PriceEstimate float64  `json:"price_estimate"`
	IsActive      bool     `json:"is_active"`
}

type CreateTaxServiceInput struct {
	Title         string   `json:"title" binding:"required"`
	Slug          string   `json:"slug" binding:"required"`
	Description   string   `json:"description" binding:"required"`
	Category      string   `json:"category" binding:"required"`
	FormsIncluded []string `json:"forms_included" binding:"required"`
	PriceEstimate float64  `json:"price_estimate"`
}

var defaultTaxFormCategories = []TaxFormCategoryResponse{
	{
		ID:          "cat-individual",
		Name:        "Individual & Family Tax Returns",
		Code:        "INDIVIDUAL-TAX",
		Description: "Federal and State tax forms for individual taxpayers, freelancers, and sole proprietors.",
		FormType:    "Individual",
		IsActive:    true,
		Forms: []TaxFormItem{
			{Code: "FORM-1040", Name: "Form 1040", Description: "U.S. Individual Income Tax Return"},
			{Code: "SCH-C", Name: "Schedule C", Description: "Profit or Loss From Business (Sole Proprietorship)"},
			{Code: "FORM-1099-NEC", Name: "Form 1099-NEC", Description: "Nonemployee Compensation for Contractors"},
			{Code: "FORM-W2", Name: "Form W-2", Description: "Wage and Tax Statement"},
		},
	},
	{
		ID:          "cat-business",
		Name:        "Corporate & Business Tax Filing",
		Code:        "BUSINESS-TAX",
		Description: "Entities, Partnerships, S-Corporations, and LLC business tax returns.",
		FormType:    "Corporate",
		IsActive:    true,
		Forms: []TaxFormItem{
			{Code: "FORM-1120", Name: "Form 1120", Description: "U.S. Corporation Income Tax Return"},
			{Code: "FORM-1120S", Name: "Form 1120-S", Description: "U.S. Income Tax Return for an S Corporation"},
			{Code: "FORM-1065", Name: "Form 1065", Description: "U.S. Return of Partnership Income"},
			{Code: "SCH-K1", Name: "Schedule K-1", Description: "Partner's Share of Income, Deductions, Credits"},
		},
	},
	{
		ID:          "cat-resolution",
		Name:        "Tax Resolution & IRS Representation",
		Code:        "TAX-RESOLUTION",
		Description: "IRS power of attorney, offer in compromise, and audit representation forms.",
		FormType:    "Resolution",
		IsActive:    true,
		Forms: []TaxFormItem{
			{Code: "FORM-2848", Name: "Form 2848", Description: "Power of Attorney and Declaration of Representative"},
			{Code: "FORM-8821", Name: "Form 8821", Description: "Tax Information Authorization"},
			{Code: "FORM-433-A", Name: "Form 433-A", Description: "Collection Information Statement for Wage Earners"},
		},
	},
}

var defaultTaxServices = []TaxServiceResponse{
	{
		ID:          "service-tax-local-seo",
		Title:       "Google Maps Domination for Tax Firms",
		Slug:        "google-maps-tax-domination",
		Description: "Target high-intent local taxpayers searching for CPA, Tax Prep, and Accountant near them during peak tax season.",
		Category:    "Local SEO & GBP",
		FormsIncluded: []string{
			"Form 1040", "Form 1099", "Schedule C", "Form 1120S",
		},
		PriceEstimate: 297.00,
		IsActive:      true,
	},
	{
		ID:          "service-tax-form-intake",
		Title:       "Built-for-Taxes Digital Client Intake Funnel",
		Slug:        "tax-digital-intake-funnel",
		Description: "Automated digital intake system for clients to submit Form 1040, W-2s, and 1099s securely online.",
		Category:    "Lead Generation & Intake",
		FormsIncluded: []string{
			"Form W-2", "Form 1099-NEC", "Form 1099-MISC", "Form 1040-SR",
		},
		PriceEstimate: 197.00,
		IsActive:      true,
	},
	{
		ID:          "service-tax-season-campaign",
		Title:       "Seasonal Tax Sprint Campaign (Jan - April Peak)",
		Slug:        "seasonal-tax-sprint-campaign",
		Description: "Turnkey marketing and local search takeover engineered specifically for high-volume tax season profitability.",
		Category:    "Seasonal Marketing",
		FormsIncluded: []string{
			"Form 1040", "Form 1120", "Form 1065", "Form 2848",
		},
		PriceEstimate: 497.00,
		IsActive:      true,
	},
}

func GetTaxServices(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		services, err := db.Instance.Prisma.TaxService.FindMany(
			db.TaxService.IsActive.Equals(true),
		).Exec(ctx)

		if err == nil {
			resp := []TaxServiceResponse{}
			for _, s := range services {
				forms := []string{}
				if s.FormsIncluded != "" {
					forms = strings.Split(s.FormsIncluded, "||")
				}
				priceEst := 0.0
				if val, ok := s.PriceEstimate(); ok {
					priceEst = val
				}
				resp = append(resp, TaxServiceResponse{
					ID:            s.ID,
					Title:         s.Title,
					Slug:          s.Slug,
					Description:   s.Description,
					Category:      s.Category,
					FormsIncluded: forms,
					PriceEstimate: priceEst,
					IsActive:      s.IsActive,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    resp,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    defaultTaxServices,
	})
}

func GetTaxServiceByID(c *gin.Context) {
	idOrSlug := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		service, err := db.Instance.Prisma.TaxService.FindFirst(
			db.TaxService.Or(
				db.TaxService.ID.Equals(idOrSlug),
				db.TaxService.Slug.Equals(idOrSlug),
			),
		).Exec(ctx)

		if err == nil && service != nil {
			forms := []string{}
			if service.FormsIncluded != "" {
				forms = strings.Split(service.FormsIncluded, "||")
			}
			priceEst := 0.0
			if val, ok := service.PriceEstimate(); ok {
				priceEst = val
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data": TaxServiceResponse{
					ID:            service.ID,
					Title:         service.Title,
					Slug:          service.Slug,
					Description:   service.Description,
					Category:      service.Category,
					FormsIncluded: forms,
					PriceEstimate: priceEst,
					IsActive:      service.IsActive,
				},
			})
			return
		}
	}

	for _, s := range defaultTaxServices {
		if s.ID == idOrSlug || s.Slug == idOrSlug {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    s,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Tax service not found",
	})
}

func GetTaxFormCategories(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		cats, err := db.Instance.Prisma.TaxFormCategory.FindMany(
			db.TaxFormCategory.IsActive.Equals(true),
		).Exec(ctx)

		if err == nil {
			resp := []TaxFormCategoryResponse{}
			for _, cat := range cats {
				resp = append(resp, TaxFormCategoryResponse{
					ID:          cat.ID,
					Name:        cat.Name,
					Code:        cat.Code,
					Description: cat.Description,
					FormType:    cat.FormType,
					Forms:       []TaxFormItem{},
					IsActive:    cat.IsActive,
				})
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    resp,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    defaultTaxFormCategories,
	})
}

func CreateTaxService(c *gin.Context) {
	var input CreateTaxServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	formsStr := strings.Join(input.FormsIncluded, "||")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		service, err := db.Instance.Prisma.TaxService.CreateOne(
			db.TaxService.Title.Set(input.Title),
			db.TaxService.Slug.Set(input.Slug),
			db.TaxService.Description.Set(input.Description),
			db.TaxService.Category.Set(input.Category),
			db.TaxService.FormsIncluded.Set(formsStr),
			db.TaxService.PriceEstimate.Set(input.PriceEstimate),
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
			"message": "Tax service created successfully",
			"data":    service,
		})
		return
	}

	newService := TaxServiceResponse{
		ID:            "service-" + input.Slug,
		Title:         input.Title,
		Slug:          input.Slug,
		Description:   input.Description,
		Category:      input.Category,
		FormsIncluded: input.FormsIncluded,
		PriceEstimate: input.PriceEstimate,
		IsActive:      true,
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tax service created (mock memory mode)",
		"data":    newService,
	})
}

type UpdateTaxServiceInput struct {
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	FormsIncluded []string `json:"forms_included"`
	PriceEstimate float64  `json:"price_estimate"`
	IsActive      *bool    `json:"is_active"`
}

func UpdateTaxService(c *gin.Context) {
	id := c.Param("id")
	var input UpdateTaxServiceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		var updates []db.TaxServiceSetParam
		if input.Title != "" {
			updates = append(updates, db.TaxService.Title.Set(input.Title))
		}
		if input.Slug != "" {
			updates = append(updates, db.TaxService.Slug.Set(input.Slug))
		}
		if input.Description != "" {
			updates = append(updates, db.TaxService.Description.Set(input.Description))
		}
		if input.Category != "" {
			updates = append(updates, db.TaxService.Category.Set(input.Category))
		}
		if len(input.FormsIncluded) > 0 {
			formsStr := strings.Join(input.FormsIncluded, "||")
			updates = append(updates, db.TaxService.FormsIncluded.Set(formsStr))
		}
		if input.PriceEstimate > 0 {
			updates = append(updates, db.TaxService.PriceEstimate.Set(input.PriceEstimate))
		}
		if input.IsActive != nil {
			updates = append(updates, db.TaxService.IsActive.Set(*input.IsActive))
		}

		service, err := db.Instance.Prisma.TaxService.FindUnique(
			db.TaxService.ID.Equals(id),
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
			"message": "Tax service updated successfully",
			"data":    service,
		})
		return
	}

	for i, s := range defaultTaxServices {
		if s.ID == id || s.Slug == id {
			if input.Title != "" {
				defaultTaxServices[i].Title = input.Title
			}
			if input.Slug != "" {
				defaultTaxServices[i].Slug = input.Slug
			}
			if input.Description != "" {
				defaultTaxServices[i].Description = input.Description
			}
			if input.Category != "" {
				defaultTaxServices[i].Category = input.Category
			}
			if len(input.FormsIncluded) > 0 {
				defaultTaxServices[i].FormsIncluded = input.FormsIncluded
			}
			if input.PriceEstimate > 0 {
				defaultTaxServices[i].PriceEstimate = input.PriceEstimate
			}
			if input.IsActive != nil {
				defaultTaxServices[i].IsActive = *input.IsActive
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Tax service updated (mock memory mode)",
				"data":    defaultTaxServices[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Tax service not found",
	})
}

func DeleteTaxService(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		_, err := db.Instance.Prisma.TaxService.FindUnique(
			db.TaxService.ID.Equals(id),
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
			"message": "Tax service deleted successfully",
		})
		return
	}

	for i, s := range defaultTaxServices {
		if s.ID == id || s.Slug == id {
			defaultTaxServices = append(defaultTaxServices[:i], defaultTaxServices[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Tax service deleted (mock memory mode)",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Tax service not found",
	})
}

type CreateTaxFormCategoryInput struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description" binding:"required"`
	FormType    string `json:"form_type" binding:"required"`
}

func CreateTaxFormCategory(c *gin.Context) {
	var input CreateTaxFormCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		cat, err := db.Instance.Prisma.TaxFormCategory.CreateOne(
			db.TaxFormCategory.Name.Set(input.Name),
			db.TaxFormCategory.Code.Set(input.Code),
			db.TaxFormCategory.Description.Set(input.Description),
			db.TaxFormCategory.FormType.Set(input.FormType),
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
			"message": "Tax form category created successfully",
			"data":    cat,
		})
		return
	}

	newCat := TaxFormCategoryResponse{
		ID:          "cat-" + strings.ToLower(input.Code),
		Name:        input.Name,
		Code:        input.Code,
		Description: input.Description,
		FormType:    input.FormType,
		Forms:       []TaxFormItem{},
		IsActive:    true,
	}
	defaultTaxFormCategories = append(defaultTaxFormCategories, newCat)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tax form category created (mock memory mode)",
		"data":    newCat,
	})
}

type UpdateTaxFormCategoryInput struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	FormType    string `json:"form_type"`
	IsActive    *bool  `json:"is_active"`
}

func UpdateTaxFormCategory(c *gin.Context) {
	id := c.Param("id")
	var input UpdateTaxFormCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		var updates []db.TaxFormCategorySetParam
		if input.Name != "" {
			updates = append(updates, db.TaxFormCategory.Name.Set(input.Name))
		}
		if input.Code != "" {
			updates = append(updates, db.TaxFormCategory.Code.Set(input.Code))
		}
		if input.Description != "" {
			updates = append(updates, db.TaxFormCategory.Description.Set(input.Description))
		}
		if input.FormType != "" {
			updates = append(updates, db.TaxFormCategory.FormType.Set(input.FormType))
		}
		if input.IsActive != nil {
			updates = append(updates, db.TaxFormCategory.IsActive.Set(*input.IsActive))
		}

		cat, err := db.Instance.Prisma.TaxFormCategory.FindUnique(
			db.TaxFormCategory.ID.Equals(id),
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
			"message": "Tax form category updated successfully",
			"data":    cat,
		})
		return
	}

	for i, cat := range defaultTaxFormCategories {
		if cat.ID == id || cat.Code == id {
			if input.Name != "" {
				defaultTaxFormCategories[i].Name = input.Name
			}
			if input.Code != "" {
				defaultTaxFormCategories[i].Code = input.Code
			}
			if input.Description != "" {
				defaultTaxFormCategories[i].Description = input.Description
			}
			if input.FormType != "" {
				defaultTaxFormCategories[i].FormType = input.FormType
			}
			if input.IsActive != nil {
				defaultTaxFormCategories[i].IsActive = *input.IsActive
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Tax form category updated (mock memory mode)",
				"data":    defaultTaxFormCategories[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Tax form category not found",
	})
}

func DeleteTaxFormCategory(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		_, err := db.Instance.Prisma.TaxFormCategory.FindUnique(
			db.TaxFormCategory.ID.Equals(id),
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
			"message": "Tax form category deleted successfully",
		})
		return
	}

	for i, cat := range defaultTaxFormCategories {
		if cat.ID == id || cat.Code == id {
			defaultTaxFormCategories = append(defaultTaxFormCategories[:i], defaultTaxFormCategories[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Tax form category deleted (mock memory mode)",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Tax form category not found",
	})
}

