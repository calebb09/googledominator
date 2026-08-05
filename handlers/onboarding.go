package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"googledominator-backend/db"
)

type CreateOnboardingSubmissionInput struct {
	BusinessName             string `json:"business_name" binding:"required"`
	ContactName              string `json:"contact_name" binding:"required"`
	Phone                    string `json:"phone" binding:"required"`
	Email                    string `json:"email" binding:"required,email"`
	HasExistingWebsite       bool   `json:"has_existing_website"`
	WebsiteUrl               string `json:"website_url"`
	HasGoogleBusinessProfile bool   `json:"has_google_business_profile"`
	GbpLink                  string `json:"gbp_link"`
	StreetAddress            string `json:"street_address"`
	City                     string `json:"city"`
	State                    string `json:"state"`
	ZipCode                  string `json:"zip_code"`
	PrimaryCategory          string `json:"primary_category"`
	ServicesOffered          string `json:"services_offered"`
	TargetLocations          string `json:"target_locations"`
	Keywords                 string `json:"keywords"`
	VisitModel               string `json:"visit_model"`
	ConsentTransactional     bool   `json:"consent_transactional"`
	ConsentMarketing         bool   `json:"consent_marketing"`
}

type OnboardingSubmissionResponse struct {
	ID                       string    `json:"id"`
	BusinessName             string    `json:"business_name"`
	ContactName              string    `json:"contact_name"`
	Phone                    string    `json:"phone"`
	Email                    string    `json:"email"`
	HasExistingWebsite       bool      `json:"has_existing_website"`
	WebsiteUrl               string    `json:"website_url,omitempty"`
	HasGoogleBusinessProfile bool      `json:"has_google_business_profile"`
	GbpLink                  string    `json:"gbp_link,omitempty"`
	StreetAddress            string    `json:"street_address,omitempty"`
	City                     string    `json:"city,omitempty"`
	State                    string    `json:"state,omitempty"`
	ZipCode                  string    `json:"zip_code,omitempty"`
	PrimaryCategory          string    `json:"primary_category,omitempty"`
	ServicesOffered          string    `json:"services_offered,omitempty"`
	TargetLocations          string    `json:"target_locations,omitempty"`
	Keywords                 string    `json:"keywords,omitempty"`
	VisitModel               string    `json:"visit_model,omitempty"`
	ConsentTransactional     bool      `json:"consent_transactional"`
	ConsentMarketing         bool      `json:"consent_marketing"`
	Status                   string    `json:"status"`
	CreatedAt                time.Time `json:"created_at"`
}

var (
	mockSubmissionsLock sync.RWMutex
	mockSubmissions     = []OnboardingSubmissionResponse{
		{
			ID:                       "sub-sample-001",
			BusinessName:             "Apex Tax & Financial Services",
			ContactName:              "Jane Doe, CPA",
			Phone:                    "+1-555-019-2831",
			Email:                    "jane@apextax.com",
			HasExistingWebsite:       true,
			WebsiteUrl:               "https://apextax.com",
			HasGoogleBusinessProfile: true,
			GbpLink:                  "https://maps.google.com/?cid=123456",
			StreetAddress:            "100 Main St, Suite 400",
			City:                     "Dallas",
			State:                    "TX",
			ZipCode:                  "75201",
			PrimaryCategory:          "Tax Preparation & CPA Practice",
			ServicesOffered:          "Form 1040, Form 1120, Payroll, Tax Resolution",
			TargetLocations:          "Dallas TX, Fort Worth TX, Plano TX",
			Keywords:                 "tax prep dallas, CPA near me, tax resolution tax forms",
			VisitModel:               "Both In-Person & Online",
			ConsentTransactional:     true,
			ConsentMarketing:         true,
			Status:                   "IN_REVIEW",
			CreatedAt:                time.Now().Add(-24 * time.Hour),
		},
	}
)

func CreateOnboardingSubmission(c *gin.Context) {
	var input CreateOnboardingSubmissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		submission, err := db.Instance.Prisma.OnboardingSubmission.CreateOne(
			db.OnboardingSubmission.BusinessName.Set(input.BusinessName),
			db.OnboardingSubmission.ContactName.Set(input.ContactName),
			db.OnboardingSubmission.Phone.Set(input.Phone),
			db.OnboardingSubmission.Email.Set(input.Email),
			db.OnboardingSubmission.HasExistingWebsite.Set(input.HasExistingWebsite),
			db.OnboardingSubmission.WebsiteURL.Set(input.WebsiteUrl),
			db.OnboardingSubmission.HasGoogleBusinessProfile.Set(input.HasGoogleBusinessProfile),
			db.OnboardingSubmission.GbpLink.Set(input.GbpLink),
			db.OnboardingSubmission.StreetAddress.Set(input.StreetAddress),
			db.OnboardingSubmission.City.Set(input.City),
			db.OnboardingSubmission.State.Set(input.State),
			db.OnboardingSubmission.ZipCode.Set(input.ZipCode),
			db.OnboardingSubmission.PrimaryCategory.Set(input.PrimaryCategory),
			db.OnboardingSubmission.ServicesOffered.Set(input.ServicesOffered),
			db.OnboardingSubmission.TargetLocations.Set(input.TargetLocations),
			db.OnboardingSubmission.Keywords.Set(input.Keywords),
			db.OnboardingSubmission.VisitModel.Set(input.VisitModel),
			db.OnboardingSubmission.ConsentTransactional.Set(input.ConsentTransactional),
			db.OnboardingSubmission.ConsentMarketing.Set(input.ConsentMarketing),
			db.OnboardingSubmission.Status.Set("PENDING"),
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
			"message": "Onboarding submission received successfully",
			"data":    submission,
		})
		return
	}

	mockSubmissionsLock.Lock()
	newSub := OnboardingSubmissionResponse{
		ID:                       "sub-" + time.Now().Format("20060102150405"),
		BusinessName:             input.BusinessName,
		ContactName:              input.ContactName,
		Phone:                    input.Phone,
		Email:                    input.Email,
		HasExistingWebsite:       input.HasExistingWebsite,
		WebsiteUrl:               input.WebsiteUrl,
		HasGoogleBusinessProfile: input.HasGoogleBusinessProfile,
		GbpLink:                  input.GbpLink,
		StreetAddress:            input.StreetAddress,
		City:                     input.City,
		State:                    input.State,
		ZipCode:                  input.ZipCode,
		PrimaryCategory:          input.PrimaryCategory,
		ServicesOffered:          input.ServicesOffered,
		TargetLocations:          input.TargetLocations,
		Keywords:                 input.Keywords,
		VisitModel:               input.VisitModel,
		ConsentTransactional:     input.ConsentTransactional,
		ConsentMarketing:         input.ConsentMarketing,
		Status:                   "PENDING",
		CreatedAt:                time.Now(),
	}
	mockSubmissions = append(mockSubmissions, newSub)
	mockSubmissionsLock.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Onboarding submission received successfully (mock memory mode)",
		"data":    newSub,
	})
}

func GetOnboardingSubmissions(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		subs, err := db.Instance.Prisma.OnboardingSubmission.FindMany().Exec(ctx)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    subs,
			})
			return
		}
	}

	mockSubmissionsLock.RLock()
	defer mockSubmissionsLock.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    mockSubmissions,
	})
}

func GetOnboardingSubmissionByID(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		sub, err := db.Instance.Prisma.OnboardingSubmission.FindUnique(
			db.OnboardingSubmission.ID.Equals(id),
		).Exec(ctx)

		if err == nil && sub != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    sub,
			})
			return
		}
	}

	mockSubmissionsLock.RLock()
	defer mockSubmissionsLock.RUnlock()

	for _, s := range mockSubmissions {
		if s.ID == id {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    s,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Onboarding submission not found",
	})
}

type UpdateOnboardingSubmissionInput struct {
	Status                   string `json:"status"`
	BusinessName             string `json:"business_name"`
	ContactName              string `json:"contact_name"`
	Phone                    string `json:"phone"`
	Email                    string `json:"email"`
	HasExistingWebsite       *bool  `json:"has_existing_website"`
	WebsiteUrl               string `json:"website_url"`
	HasGoogleBusinessProfile *bool  `json:"has_google_business_profile"`
	GbpLink                  string `json:"gbp_link"`
	StreetAddress            string `json:"street_address"`
	City                     string `json:"city"`
	State                    string `json:"state"`
	ZipCode                  string `json:"zip_code"`
	PrimaryCategory          string `json:"primary_category"`
	ServicesOffered          string `json:"services_offered"`
	TargetLocations          string `json:"target_locations"`
	Keywords                 string `json:"keywords"`
	VisitModel               string `json:"visit_model"`
	ConsentTransactional     *bool  `json:"consent_transactional"`
	ConsentMarketing         *bool  `json:"consent_marketing"`
}

func UpdateOnboardingSubmission(c *gin.Context) {
	id := c.Param("id")
	var input UpdateOnboardingSubmissionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		var updates []db.OnboardingSubmissionSetParam
		if input.Status != "" {
			updates = append(updates, db.OnboardingSubmission.Status.Set(input.Status))
		}
		if input.BusinessName != "" {
			updates = append(updates, db.OnboardingSubmission.BusinessName.Set(input.BusinessName))
		}
		if input.ContactName != "" {
			updates = append(updates, db.OnboardingSubmission.ContactName.Set(input.ContactName))
		}
		if input.Phone != "" {
			updates = append(updates, db.OnboardingSubmission.Phone.Set(input.Phone))
		}
		if input.Email != "" {
			updates = append(updates, db.OnboardingSubmission.Email.Set(input.Email))
		}
		if input.HasExistingWebsite != nil {
			updates = append(updates, db.OnboardingSubmission.HasExistingWebsite.Set(*input.HasExistingWebsite))
		}
		if input.WebsiteUrl != "" {
			updates = append(updates, db.OnboardingSubmission.WebsiteURL.Set(input.WebsiteUrl))
		}
		if input.HasGoogleBusinessProfile != nil {
			updates = append(updates, db.OnboardingSubmission.HasGoogleBusinessProfile.Set(*input.HasGoogleBusinessProfile))
		}
		if input.GbpLink != "" {
			updates = append(updates, db.OnboardingSubmission.GbpLink.Set(input.GbpLink))
		}
		if input.StreetAddress != "" {
			updates = append(updates, db.OnboardingSubmission.StreetAddress.Set(input.StreetAddress))
		}
		if input.City != "" {
			updates = append(updates, db.OnboardingSubmission.City.Set(input.City))
		}
		if input.State != "" {
			updates = append(updates, db.OnboardingSubmission.State.Set(input.State))
		}
		if input.ZipCode != "" {
			updates = append(updates, db.OnboardingSubmission.ZipCode.Set(input.ZipCode))
		}
		if input.PrimaryCategory != "" {
			updates = append(updates, db.OnboardingSubmission.PrimaryCategory.Set(input.PrimaryCategory))
		}
		if input.ServicesOffered != "" {
			updates = append(updates, db.OnboardingSubmission.ServicesOffered.Set(input.ServicesOffered))
		}
		if input.TargetLocations != "" {
			updates = append(updates, db.OnboardingSubmission.TargetLocations.Set(input.TargetLocations))
		}
		if input.Keywords != "" {
			updates = append(updates, db.OnboardingSubmission.Keywords.Set(input.Keywords))
		}
		if input.VisitModel != "" {
			updates = append(updates, db.OnboardingSubmission.VisitModel.Set(input.VisitModel))
		}
		if input.ConsentTransactional != nil {
			updates = append(updates, db.OnboardingSubmission.ConsentTransactional.Set(*input.ConsentTransactional))
		}
		if input.ConsentMarketing != nil {
			updates = append(updates, db.OnboardingSubmission.ConsentMarketing.Set(*input.ConsentMarketing))
		}

		sub, err := db.Instance.Prisma.OnboardingSubmission.FindUnique(
			db.OnboardingSubmission.ID.Equals(id),
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
			"message": "Onboarding submission updated successfully",
			"data":    sub,
		})
		return
	}

	mockSubmissionsLock.Lock()
	defer mockSubmissionsLock.Unlock()

	for i, s := range mockSubmissions {
		if s.ID == id {
			if input.Status != "" {
				mockSubmissions[i].Status = input.Status
			}
			if input.BusinessName != "" {
				mockSubmissions[i].BusinessName = input.BusinessName
			}
			if input.ContactName != "" {
				mockSubmissions[i].ContactName = input.ContactName
			}
			if input.Phone != "" {
				mockSubmissions[i].Phone = input.Phone
			}
			if input.Email != "" {
				mockSubmissions[i].Email = input.Email
			}
			if input.HasExistingWebsite != nil {
				mockSubmissions[i].HasExistingWebsite = *input.HasExistingWebsite
			}
			if input.WebsiteUrl != "" {
				mockSubmissions[i].WebsiteUrl = input.WebsiteUrl
			}
			if input.HasGoogleBusinessProfile != nil {
				mockSubmissions[i].HasGoogleBusinessProfile = *input.HasGoogleBusinessProfile
			}
			if input.GbpLink != "" {
				mockSubmissions[i].GbpLink = input.GbpLink
			}
			if input.StreetAddress != "" {
				mockSubmissions[i].StreetAddress = input.StreetAddress
			}
			if input.City != "" {
				mockSubmissions[i].City = input.City
			}
			if input.State != "" {
				mockSubmissions[i].State = input.State
			}
			if input.ZipCode != "" {
				mockSubmissions[i].ZipCode = input.ZipCode
			}
			if input.PrimaryCategory != "" {
				mockSubmissions[i].PrimaryCategory = input.PrimaryCategory
			}
			if input.ServicesOffered != "" {
				mockSubmissions[i].ServicesOffered = input.ServicesOffered
			}
			if input.TargetLocations != "" {
				mockSubmissions[i].TargetLocations = input.TargetLocations
			}
			if input.Keywords != "" {
				mockSubmissions[i].Keywords = input.Keywords
			}
			if input.VisitModel != "" {
				mockSubmissions[i].VisitModel = input.VisitModel
			}
			if input.ConsentTransactional != nil {
				mockSubmissions[i].ConsentTransactional = *input.ConsentTransactional
			}
			if input.ConsentMarketing != nil {
				mockSubmissions[i].ConsentMarketing = *input.ConsentMarketing
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Onboarding submission updated (mock memory mode)",
				"data":    mockSubmissions[i],
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Onboarding submission not found",
	})
}

func DeleteOnboardingSubmission(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		_, err := db.Instance.Prisma.OnboardingSubmission.FindUnique(
			db.OnboardingSubmission.ID.Equals(id),
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
			"message": "Onboarding submission deleted successfully",
		})
		return
	}

	mockSubmissionsLock.Lock()
	defer mockSubmissionsLock.Unlock()

	for i, s := range mockSubmissions {
		if s.ID == id {
			mockSubmissions = append(mockSubmissions[:i], mockSubmissions[i+1:]...)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Onboarding submission deleted (mock memory mode)",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"success": false,
		"error":   "Onboarding submission not found",
	})
}

