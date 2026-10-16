package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"googledominator-backend/db"
	gdemail "googledominator-backend/email"

	"github.com/gin-gonic/gin"
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
	TemplateID               string `json:"template_id"`
}

type OnboardingSubmissionResponse struct {
	ID                       string            `json:"id"`
	Token                    string            `json:"token"`
	BusinessName             string            `json:"business_name"`
	ContactName              string            `json:"contact_name"`
	Phone                    string            `json:"phone"`
	Email                    string            `json:"email"`
	HasExistingWebsite       bool              `json:"has_existing_website"`
	WebsiteUrl               string            `json:"website_url,omitempty"`
	HasGoogleBusinessProfile bool              `json:"has_google_business_profile"`
	GbpLink                  string            `json:"gbp_link,omitempty"`
	StreetAddress            string            `json:"street_address,omitempty"`
	City                     string            `json:"city,omitempty"`
	State                    string            `json:"state,omitempty"`
	ZipCode                  string            `json:"zip_code,omitempty"`
	PrimaryCategory          string            `json:"primary_category,omitempty"`
	ServicesOffered          string            `json:"services_offered,omitempty"`
	TargetLocations          string            `json:"target_locations,omitempty"`
	Keywords                 string            `json:"keywords,omitempty"`
	VisitModel               string            `json:"visit_model,omitempty"`
	ConsentTransactional     bool              `json:"consent_transactional"`
	ConsentMarketing         bool              `json:"consent_marketing"`
	TemplateID               string            `json:"template_id,omitempty"`
	Template                 *TemplateResponse `json:"template,omitempty"`
	ColorScheme              *ColorScheme      `json:"colorScheme,omitempty"`
	Font                     *FontChoice       `json:"font,omitempty"`
	EditorPath               string            `json:"editorPath,omitempty"`
	Status                   string            `json:"status"`
	CreatedAt                time.Time         `json:"created_at"`
}

type ColorScheme struct {
	ID        string   `json:"id" binding:"required"`
	Name      string   `json:"name" binding:"required"`
	Primary   string   `json:"primary" binding:"required"`
	Secondary string   `json:"secondary" binding:"required"`
	Accent    string   `json:"accent" binding:"required"`
	Swatches  []string `json:"swatches" binding:"required"`
}

type FontChoice struct {
	Heading string `json:"heading" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

var (
	onboardingBaseURL string
	publicAPIBaseURL  string
)

func SetOnboardingURLs(baseURL, apiBaseURL string) {
	onboardingBaseURL = strings.TrimRight(baseURL, "/")
	publicAPIBaseURL = strings.TrimRight(apiBaseURL, "/")
}

func generateOnboardingToken() (string, error) {
	random := make([]byte, 24)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	return "tok_" + base64.RawURLEncoding.EncodeToString(random), nil
}

func pickTemplateURL(c *gin.Context, token string) string {
	submitBaseURL := publicAPIBaseURL
	if submitBaseURL == "" {
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			host = c.Request.Host
		}
		scheme := c.GetHeader("X-Forwarded-Proto")
		if scheme == "" {
			scheme = "https"
		}
		submitBaseURL = scheme + "://" + host
	}
	submitURL := submitBaseURL + "/api/v1/onboarding"
	return onboardingBaseURL + "/pick-template?token=" + url.QueryEscape(token) + "&submitUrl=" + submitURL
}

// normalizeOnboardingData works around prisma-client-go encoding JSON columns
// as JSON strings in response models. PostgreSQL still stores these as jsonb.
func normalizeOnboardingData(data any) any {
	encoded, err := json.Marshal(data)
	if err != nil {
		return data
	}
	var normalized any
	if err = json.Unmarshal(encoded, &normalized); err != nil {
		return data
	}
	normalizeJSONFields(normalized)
	return normalized
}

func normalizeJSONFields(value any) {
	switch current := value.(type) {
	case []any:
		for _, item := range current {
			normalizeJSONFields(item)
		}
	case map[string]any:
		for key, item := range current {
			if (key == "colorScheme" || key == "font") && item != nil {
				if raw, ok := item.(string); ok {
					var object any
					if json.Unmarshal([]byte(raw), &object) == nil {
						current[key] = object
						continue
					}
				}
			}
			normalizeJSONFields(item)
		}
	}
}

func respondToOnboardingCreation(c *gin.Context, message string, data any, pickerURL string) {
	// Native browser form posts follow the requested GET flow. JSON clients get
	// the same destination in both Location and the response body.
	if strings.Contains(c.GetHeader("Accept"), "text/html") {
		c.Redirect(http.StatusSeeOther, pickerURL)
		return
	}
	c.Header("Location", pickerURL)
	c.JSON(http.StatusCreated, gin.H{
		"success":           true,
		"message":           message,
		"data":              normalizeOnboardingData(data),
		"pick_template_url": pickerURL,
	})
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
	token, err := generateOnboardingToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to generate onboarding token"})
		return
	}
	pickerURL := pickTemplateURL(c, token)

	if db.Instance != nil && db.Instance.IsConnected {
		createParams := []db.OnboardingSubmissionSetParam{
			db.OnboardingSubmission.Token.Set(token),
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
		}
		if input.TemplateID != "" {
			createParams = append(createParams, db.OnboardingSubmission.TemplateID.Set(input.TemplateID))
		}

		submission, err := db.Instance.Prisma.OnboardingSubmission.CreateOne(
			db.OnboardingSubmission.BusinessName.Set(input.BusinessName),
			db.OnboardingSubmission.ContactName.Set(input.ContactName),
			db.OnboardingSubmission.Phone.Set(input.Phone),
			db.OnboardingSubmission.Email.Set(input.Email),
			createParams...,
		).Exec(ctx)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		sendOnboardingNotifications(input)

		respondToOnboardingCreation(c, "Onboarding submission received successfully", submission, pickerURL)
		return
	}

	mockSubmissionsLock.Lock()
	newSub := OnboardingSubmissionResponse{
		ID:                       "sub-" + time.Now().Format("20060102150405"),
		Token:                    token,
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
		TemplateID:               input.TemplateID,
		Status:                   "PENDING",
		CreatedAt:                time.Now(),
	}
	mockSubmissions = append(mockSubmissions, newSub)
	mockSubmissionsLock.Unlock()

	sendOnboardingNotifications(input)

	respondToOnboardingCreation(c, "Onboarding submission received successfully (mock memory mode)", newSub, pickerURL)
}

type UpdateOnboardingDesignInput struct {
	ColorScheme ColorScheme `json:"colorScheme" binding:"required"`
	Font        FontChoice  `json:"font" binding:"required"`
	EditorPath  string      `json:"editorPath" binding:"required"`
}

// UpdateOnboardingDesign receives the result of the external template picker.
// The opaque token limits the update to the submission that initiated it.
func UpdateOnboardingDesign(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "token query parameter is required"})
		return
	}

	var input UpdateOnboardingDesignInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if db.Instance != nil && db.Instance.IsConnected {
		colorJSON, err := json.Marshal(input.ColorScheme)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid colorScheme"})
			return
		}
		fontJSON, err := json.Marshal(input.Font)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid font"})
			return
		}

		submission, err := db.Instance.Prisma.OnboardingSubmission.FindUnique(
			db.OnboardingSubmission.Token.Equals(token),
		).Update(
			db.OnboardingSubmission.ColorScheme.Set(db.JSON(colorJSON)),
			db.OnboardingSubmission.Font.Set(db.JSON(fontJSON)),
			db.OnboardingSubmission.EditorPath.Set(input.EditorPath),
		).Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "onboarding submission not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Onboarding design updated successfully", "data": normalizeOnboardingData(submission)})
		return
	}

	mockSubmissionsLock.Lock()
	defer mockSubmissionsLock.Unlock()
	for i := range mockSubmissions {
		if mockSubmissions[i].Token == token {
			mockSubmissions[i].ColorScheme = &input.ColorScheme
			mockSubmissions[i].Font = &input.Font
			mockSubmissions[i].EditorPath = input.EditorPath
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Onboarding design updated (mock memory mode)", "data": mockSubmissions[i]})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "onboarding submission not found"})
}

func sendOnboardingNotifications(input CreateOnboardingSubmissionInput) {
	if err := gdemail.SendOnboardingNotifications(
		input.Email,
		input.ContactName,
		input.BusinessName,
	); err != nil {
		// The submission is already saved. Logging the notification failure avoids
		// returning an error that could cause the user to submit a duplicate.
		log.Printf("[ERROR] Onboarding submission saved, but email notification failed: %v", err)
	}
}

func GetOnboardingSubmissions(c *gin.Context) {
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		subs, err := db.Instance.Prisma.OnboardingSubmission.FindMany().Exec(ctx)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    normalizeOnboardingData(subs),
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
				"data":    normalizeOnboardingData(sub),
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
			"data":    normalizeOnboardingData(sub),
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
