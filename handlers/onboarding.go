package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"googledominator-backend/db"
	gdemail "googledominator-backend/email"

	"github.com/gin-gonic/gin"
)

const (
	OnboardingUploadDir  = "./uploads/onboarding"
	LogoUploadDir        = "./uploads/onboarding/logos"
	BackgroundsUploadDir = "./uploads/onboarding/backgrounds"
	OfficeUploadDir      = "./uploads/onboarding/office"
	RecentWinsUploadDir  = "./uploads/onboarding/recent_wins"
)

func init() {
	_ = os.MkdirAll(LogoUploadDir, os.ModePerm)
	_ = os.MkdirAll(BackgroundsUploadDir, os.ModePerm)
	_ = os.MkdirAll(OfficeUploadDir, os.ModePerm)
	_ = os.MkdirAll(RecentWinsUploadDir, os.ModePerm)
}

type CreateOnboardingSubmissionInput struct {
	BusinessName             string `json:"business_name" form:"business_name"`
	BusinessNameAlt          string `json:"businessName" form:"businessName"`
	Dba                      string `json:"dba" form:"dba"`
	ContactName              string `json:"contact_name" form:"contact_name"`
	ContactNameAlt           string `json:"contactName" form:"contactName"`
	FirstName                string `json:"first_name" form:"first_name"`
	FirstNameAlt             string `json:"firstName" form:"firstName"`
	LastName                 string `json:"last_name" form:"last_name"`
	LastNameAlt              string `json:"lastName" form:"lastName"`
	CountryCode              string `json:"country_code" form:"country_code"`
	CountryCodeAlt           string `json:"countryCode" form:"countryCode"`
	Phone                    string `json:"phone" form:"phone"`
	Email                    string `json:"email" form:"email"`
	HasExistingWebsite       *bool  `json:"has_existing_website" form:"has_existing_website"`
	HasExistingWebsiteAlt    *bool  `json:"hasExistingWebsite" form:"hasExistingWebsite"`
	WebsiteUrl               string `json:"website_url" form:"website_url"`
	WebsiteUrlAlt            string `json:"websiteUrl" form:"websiteUrl"`
	HasGoogleBusinessProfile *bool  `json:"has_google_business_profile" form:"has_google_business_profile"`
	HasGoogleBusinessProfileAlt *bool `json:"hasGoogleBusinessProfile" form:"hasGoogleBusinessProfile"`
	GbpLink                  string `json:"gbp_link" form:"gbp_link"`
	GbpLinkAlt               string `json:"gbpLink" form:"gbpLink"`
	GmbUrl                   string `json:"gmb_url" form:"gmb_url"`
	GmbUrlAlt                string `json:"gmbUrl" form:"gmbUrl"`
	OfficeAddresses          string `json:"office_addresses" form:"office_addresses"`
	OfficeAddressesAlt       string `json:"officeAddresses" form:"officeAddresses"`
	StreetAddress            string `json:"street_address" form:"street_address"`
	StreetAddressAlt         string `json:"streetAddress" form:"streetAddress"`
	City                     string `json:"city" form:"city"`
	State                    string `json:"state" form:"state"`
	ZipCode                  string `json:"zip_code" form:"zip_code"`
	ZipCodeAlt               string `json:"zipCode" form:"zipCode"`
	PrimaryCategory          string `json:"primary_category" form:"primary_category"`
	PrimaryCategoryAlt       string `json:"primaryCategory" form:"primaryCategory"`
	PrimaryProductService    string `json:"primary_product_service" form:"primary_product_service"`
	PrimaryProductServiceAlt string `json:"primaryProductService" form:"primaryProductService"`
	ServicesOffered          string `json:"services_offered" form:"services_offered"`
	ServicesOfferedAlt       string `json:"servicesOffered" form:"servicesOffered"`
	AdditionalServices       string `json:"additional_services" form:"additional_services"`
	AdditionalServicesAlt    string `json:"additionalServices" form:"additionalServices"`
	TargetLocations          string `json:"target_locations" form:"target_locations"`
	TargetLocationsAlt       string `json:"targetLocations" form:"targetLocations"`
	Keywords                 string `json:"keywords" form:"keywords"`
	KeywordsAlt              string `json:"keywordsAlt" form:"keywordsAlt"`
	VisitModel               string `json:"visit_model" form:"visit_model"`
	VisitModelAlt            string `json:"visitModel" form:"visitModel"`
	CustomerVisitType        any    `json:"customer_visit_type" form:"customer_visit_type"`
	CustomerVisitTypeAlt     any    `json:"customerVisitType" form:"customerVisitType"`
	ConsentTransactional     *bool  `json:"consent_transactional" form:"consent_transactional"`
	ConsentTransactionalAlt  *bool  `json:"consentTransactional" form:"consentTransactional"`
	ConsentMarketing         *bool  `json:"consent_marketing" form:"consent_marketing"`
	ConsentMarketingAlt      *bool  `json:"consentMarketing" form:"consentMarketing"`
	TwitterUrl               string `json:"twitter_url" form:"twitter_url"`
	TwitterUrlAlt            string `json:"twitterUrl" form:"twitterUrl"`
	FacebookUrl              string `json:"facebook_url" form:"facebook_url"`
	FacebookUrlAlt           string `json:"facebookUrl" form:"facebookUrl"`
	InstagramUrl             string `json:"instagram_url" form:"instagram_url"`
	InstagramUrlAlt          string `json:"instagramUrl" form:"instagramUrl"`
	YoutubeUrl               string `json:"youtube_url" form:"youtube_url"`
	YoutubeUrlAlt            string `json:"youtubeUrl" form:"youtubeUrl"`
	OtherSocialUrls          any    `json:"other_social_urls" form:"other_social_urls"`
	OtherSocialUrlsAlt       any    `json:"otherSocialUrls" form:"otherSocialUrls"`
	UseExistingImages        string `json:"use_existing_images" form:"use_existing_images"`
	UseExistingImagesAlt     string `json:"useExistingImages" form:"useExistingImages"`
	LogoUrl                  string `json:"logo_url" form:"logo_url"`
	LogoUrlAlt               string `json:"logoUrl" form:"logoUrl"`
	BackgroundImages         any    `json:"background_images" form:"background_images"`
	BackgroundImagesAlt      any    `json:"backgroundImages" form:"backgroundImages"`
	OfficeImages             any    `json:"office_images" form:"office_images"`
	OfficeImagesAlt          any    `json:"officeImages" form:"officeImages"`
	RecentWinsImages         any    `json:"recent_wins_images" form:"recent_wins_images"`
	RecentWinsImagesAlt      any    `json:"recentWinsImages" form:"recentWinsImages"`
	AgreementBusinessName    string `json:"agreement_business_name" form:"agreement_business_name"`
	AgreementBusinessNameAlt string `json:"agreementBusinessName" form:"agreementBusinessName"`
	SignerFirstName          string `json:"signer_first_name" form:"signer_first_name"`
	SignerFirstNameAlt       string `json:"signerFirstName" form:"signerFirstName"`
	SignerLastName           string `json:"signer_last_name" form:"signer_last_name"`
	SignerLastNameAlt        string `json:"signerLastName" form:"signerLastName"`
	SignerPhone              string `json:"signer_phone" form:"signer_phone"`
	SignerPhoneAlt           string `json:"signerPhone" form:"signerPhone"`
	AgreedTerms              *bool  `json:"agreed_terms" form:"agreed_terms"`
	AgreedTermsAlt           *bool  `json:"agreedTerms" form:"agreedTerms"`
	TemplateID               string `json:"template_id" form:"template_id"`
	TemplateIDAlt            string `json:"templateId" form:"templateId"`
	SelectedTemplate         string `json:"selected_template" form:"selected_template"`
	SelectedTemplateAlt      string `json:"selectedTemplate" form:"selectedTemplate"`
}

func (input *CreateOnboardingSubmissionInput) Clean() {
	if input.BusinessName == "" && input.BusinessNameAlt != "" {
		input.BusinessName = input.BusinessNameAlt
	}
	if input.ContactName == "" {
		if input.ContactNameAlt != "" {
			input.ContactName = input.ContactNameAlt
		} else if input.FirstName != "" || input.LastName != "" {
			input.ContactName = strings.TrimSpace(input.FirstName + " " + input.LastName)
		} else if input.SignerFirstName != "" || input.SignerLastName != "" {
			input.ContactName = strings.TrimSpace(input.SignerFirstName + " " + input.SignerLastName)
		}
	}
	if input.ContactName == "" {
		input.ContactName = "Primary Contact"
	}
	if input.Phone == "" {
		if input.SignerPhone != "" {
			input.Phone = input.SignerPhone
		} else if input.SignerPhoneAlt != "" {
			input.Phone = input.SignerPhoneAlt
		}
	}
	if input.CountryCode == "" {
		if input.CountryCodeAlt != "" {
			input.CountryCode = input.CountryCodeAlt
		} else {
			input.CountryCode = "+1"
		}
	}
	if input.WebsiteUrl == "" && input.WebsiteUrlAlt != "" {
		input.WebsiteUrl = input.WebsiteUrlAlt
	}
	if input.GbpLink == "" {
		if input.GbpLinkAlt != "" {
			input.GbpLink = input.GbpLinkAlt
		} else if input.GmbUrl != "" {
			input.GbpLink = input.GmbUrl
		} else if input.GmbUrlAlt != "" {
			input.GbpLink = input.GmbUrlAlt
		}
	}
	if input.HasGoogleBusinessProfile == nil {
		if input.HasGoogleBusinessProfileAlt != nil {
			input.HasGoogleBusinessProfile = input.HasGoogleBusinessProfileAlt
		} else {
			hasGbp := input.GbpLink != ""
			input.HasGoogleBusinessProfile = &hasGbp
		}
	}
	if input.HasExistingWebsite == nil {
		if input.HasExistingWebsiteAlt != nil {
			input.HasExistingWebsite = input.HasExistingWebsiteAlt
		} else {
			hasWeb := input.WebsiteUrl != "" || strings.EqualFold(input.UseExistingImages, "yes") || strings.EqualFold(input.UseExistingImagesAlt, "yes")
			input.HasExistingWebsite = &hasWeb
		}
	}
	if input.OfficeAddresses == "" && input.OfficeAddressesAlt != "" {
		input.OfficeAddresses = input.OfficeAddressesAlt
	}
	if input.StreetAddress == "" && input.StreetAddressAlt != "" {
		input.StreetAddress = input.StreetAddressAlt
	}
	if input.StreetAddress == "" && input.OfficeAddresses != "" {
		input.StreetAddress = input.OfficeAddresses
	}
	if input.OfficeAddresses == "" && input.StreetAddress != "" {
		input.OfficeAddresses = input.StreetAddress
	}
	if input.PrimaryCategory == "" && input.PrimaryCategoryAlt != "" {
		input.PrimaryCategory = input.PrimaryCategoryAlt
	}
	if input.PrimaryProductService == "" && input.PrimaryProductServiceAlt != "" {
		input.PrimaryProductService = input.PrimaryProductServiceAlt
	}
	if input.ServicesOffered == "" && input.ServicesOfferedAlt != "" {
		input.ServicesOffered = input.ServicesOfferedAlt
	}
	if input.AdditionalServices == "" && input.AdditionalServicesAlt != "" {
		input.AdditionalServices = input.AdditionalServicesAlt
	}
	if input.ServicesOffered == "" {
		parts := []string{}
		if input.PrimaryProductService != "" {
			parts = append(parts, input.PrimaryProductService)
		}
		if input.AdditionalServices != "" {
			parts = append(parts, input.AdditionalServices)
		}
		input.ServicesOffered = strings.Join(parts, ", ")
	}
	if input.TargetLocations == "" && input.TargetLocationsAlt != "" {
		input.TargetLocations = input.TargetLocationsAlt
	}
	if input.Keywords == "" {
		if input.KeywordsAlt != "" {
			input.Keywords = input.KeywordsAlt
		}
	}
	if input.VisitModel == "" {
		if input.VisitModelAlt != "" {
			input.VisitModel = input.VisitModelAlt
		} else {
			vtVal := input.CustomerVisitType
			if vtVal == nil {
				vtVal = input.CustomerVisitTypeAlt
			}
			if vtVal != nil {
				switch vt := vtVal.(type) {
				case []string:
					input.VisitModel = strings.Join(vt, ", ")
				case []any:
					var strParts []string
					for _, p := range vt {
						if s, ok := p.(string); ok && s != "" {
							strParts = append(strParts, s)
						}
					}
					input.VisitModel = strings.Join(strParts, ", ")
				case string:
					input.VisitModel = vt
				}
			}
		}
	}
	if input.VisitModel == "" {
		input.VisitModel = "Both In-Person & Virtual"
	}
	if input.TwitterUrl == "" && input.TwitterUrlAlt != "" {
		input.TwitterUrl = input.TwitterUrlAlt
	}
	if input.FacebookUrl == "" && input.FacebookUrlAlt != "" {
		input.FacebookUrl = input.FacebookUrlAlt
	}
	if input.InstagramUrl == "" && input.InstagramUrlAlt != "" {
		input.InstagramUrl = input.InstagramUrlAlt
	}
	if input.YoutubeUrl == "" && input.YoutubeUrlAlt != "" {
		input.YoutubeUrl = input.YoutubeUrlAlt
	}
	if input.OtherSocialUrls == nil && input.OtherSocialUrlsAlt != nil {
		input.OtherSocialUrls = input.OtherSocialUrlsAlt
	}
	if input.UseExistingImages == "" {
		if input.UseExistingImagesAlt != "" {
			input.UseExistingImages = input.UseExistingImagesAlt
		} else {
			input.UseExistingImages = "yes"
		}
	}
	if input.TemplateID == "" {
		if input.TemplateIDAlt != "" {
			input.TemplateID = input.TemplateIDAlt
		} else if input.SelectedTemplate != "" {
			input.TemplateID = input.SelectedTemplate
		} else if input.SelectedTemplateAlt != "" {
			input.TemplateID = input.SelectedTemplateAlt
		}
	}
	if input.AgreementBusinessName == "" {
		if input.AgreementBusinessNameAlt != "" {
			input.AgreementBusinessName = input.AgreementBusinessNameAlt
		} else {
			input.AgreementBusinessName = input.BusinessName
		}
	}
	if input.SignerFirstName == "" && input.SignerFirstNameAlt != "" {
		input.SignerFirstName = input.SignerFirstNameAlt
	}
	if input.SignerLastName == "" && input.SignerLastNameAlt != "" {
		input.SignerLastName = input.SignerLastNameAlt
	}
	if input.SignerPhone == "" {
		if input.SignerPhoneAlt != "" {
			input.SignerPhone = input.SignerPhoneAlt
		} else {
			input.SignerPhone = input.Phone
		}
	}
	if input.LogoUrl == "" && input.LogoUrlAlt != "" {
		input.LogoUrl = input.LogoUrlAlt
	}
	if input.BackgroundImages == nil && input.BackgroundImagesAlt != nil {
		input.BackgroundImages = input.BackgroundImagesAlt
	}
	if input.OfficeImages == nil && input.OfficeImagesAlt != nil {
		input.OfficeImages = input.OfficeImagesAlt
	}
	if input.RecentWinsImages == nil && input.RecentWinsImagesAlt != nil {
		input.RecentWinsImages = input.RecentWinsImagesAlt
	}
	if input.ConsentTransactional == nil {
		if input.ConsentTransactionalAlt != nil {
			input.ConsentTransactional = input.ConsentTransactionalAlt
		} else {
			f := false
			input.ConsentTransactional = &f
		}
	}
	if input.ConsentMarketing == nil {
		if input.ConsentMarketingAlt != nil {
			input.ConsentMarketing = input.ConsentMarketingAlt
		} else {
			f := false
			input.ConsentMarketing = &f
		}
	}
	if input.AgreedTerms == nil {
		if input.AgreedTermsAlt != nil {
			input.AgreedTerms = input.AgreedTermsAlt
		} else {
			f := false
			input.AgreedTerms = &f
		}
	}
}

type OnboardingSubmissionResponse struct {
	ID                       string            `json:"id"`
	Token                    string            `json:"token,omitempty"`
	BusinessName             string            `json:"business_name"`
	Dba                      string            `json:"dba,omitempty"`
	ContactName              string            `json:"contact_name"`
	CountryCode              string            `json:"country_code,omitempty"`
	Phone                    string            `json:"phone"`
	Email                    string            `json:"email"`
	HasExistingWebsite       bool              `json:"has_existing_website"`
	WebsiteUrl               string            `json:"website_url,omitempty"`
	HasGoogleBusinessProfile bool              `json:"has_google_business_profile"`
	GbpLink                  string            `json:"gbp_link,omitempty"`
	OfficeAddresses          string            `json:"office_addresses,omitempty"`
	StreetAddress            string            `json:"street_address,omitempty"`
	City                     string            `json:"city,omitempty"`
	State                    string            `json:"state,omitempty"`
	ZipCode                  string            `json:"zip_code,omitempty"`
	PrimaryCategory          string            `json:"primary_category,omitempty"`
	PrimaryProductService    string            `json:"primary_product_service,omitempty"`
	ServicesOffered          string            `json:"services_offered,omitempty"`
	AdditionalServices       string            `json:"additional_services,omitempty"`
	TargetLocations          string            `json:"target_locations,omitempty"`
	Keywords                 string            `json:"keywords,omitempty"`
	VisitModel               string            `json:"visit_model,omitempty"`
	ConsentTransactional     bool              `json:"consent_transactional"`
	ConsentMarketing         bool              `json:"consent_marketing"`
	TwitterUrl               string            `json:"twitter_url,omitempty"`
	FacebookUrl              string            `json:"facebook_url,omitempty"`
	InstagramUrl             string            `json:"instagram_url,omitempty"`
	YoutubeUrl               string            `json:"youtube_url,omitempty"`
	OtherSocialUrls          any               `json:"other_social_urls,omitempty"`
	UseExistingImages        string            `json:"use_existing_images,omitempty"`
	LogoUrl                  string            `json:"logo_url,omitempty"`
	BackgroundImages         any               `json:"background_images,omitempty"`
	OfficeImages             any               `json:"office_images,omitempty"`
	RecentWinsImages         any               `json:"recent_wins_images,omitempty"`
	AgreementBusinessName    string            `json:"agreement_business_name,omitempty"`
	SignerFirstName          string            `json:"signer_first_name,omitempty"`
	SignerLastName           string            `json:"signer_last_name,omitempty"`
	SignerPhone              string            `json:"signer_phone,omitempty"`
	AgreedTerms              bool              `json:"agreed_terms"`
	TemplateID               string            `json:"template_id,omitempty"`
	Template                 *TemplateResponse `json:"template,omitempty"`
	ColorScheme              *ColorScheme      `json:"colorScheme,omitempty"`
	Font                     *FontChoice       `json:"font,omitempty"`
	EditorPath               string            `json:"editorPath,omitempty"`
	Status                   string            `json:"status"`
	CreatedAt                time.Time         `json:"created_at"`
	UpdatedAt                time.Time         `json:"updated_at,omitempty"`
}

func toJSONBytes(val any) []byte {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return nil
		}
		if (strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) ||
			(strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) {
			return []byte(trimmed)
		}
		b, _ := json.Marshal([]string{trimmed})
		return b
	case []string:
		if len(v) == 0 {
			return nil
		}
		b, _ := json.Marshal(v)
		return b
	case []any:
		if len(v) == 0 {
			return nil
		}
		b, _ := json.Marshal(v)
		return b
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		return b
	}
}

func getFormString(c *gin.Context, keys ...string) string {
	for _, k := range keys {
		if val := strings.TrimSpace(c.PostForm(k)); val != "" {
			return val
		}
	}
	return ""
}

func getFormBool(c *gin.Context, defaultVal bool, keys ...string) bool {
	for _, k := range keys {
		val := strings.TrimSpace(strings.ToLower(c.PostForm(k)))
		if val != "" {
			return val == "true" || val == "1" || val == "yes" || val == "on"
		}
	}
	return defaultVal
}

func saveSingleUploadedFile(c *gin.Context, fieldNames []string, targetDir, urlPrefix string) string {
	urls := saveUploadedFiles(c, fieldNames, targetDir, urlPrefix)
	if len(urls) > 0 {
		return urls[0]
	}
	return ""
}

func saveUploadedFiles(c *gin.Context, fieldNames []string, targetDir, urlPrefix string) []string {
	var savedURLs []string
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return savedURLs
	}

	for _, field := range fieldNames {
		files := form.File[field]
		for _, file := range files {
			if file == nil || file.Filename == "" {
				continue
			}
			_ = os.MkdirAll(targetDir, os.ModePerm)
			ext := filepath.Ext(file.Filename)
			if ext == "" {
				ext = ".png"
			}
			cleanBase := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
					return r
				}
				return '_'
			}, strings.TrimSuffix(file.Filename, ext))
			if len(cleanBase) > 25 {
				cleanBase = cleanBase[:25]
			}
			filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), cleanBase, ext)
			dst := filepath.Join(targetDir, filename)

			if err := c.SaveUploadedFile(file, dst); err == nil {
				savedURLs = append(savedURLs, urlPrefix+"/"+filename)
			} else {
				log.Printf("[WARN] Failed to save uploaded file %s: %v", file.Filename, err)
			}
		}
	}
	return savedURLs
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
	submitURL := submitBaseURL + "/api/templates/pick"
	return onboardingBaseURL + "/pick-template?token=" + url.QueryEscape(token) + "&submitUrl=" + submitURL
}

func resolveTemplateObject(templateID string) *TemplateResponse {
	trimmedID := strings.TrimSpace(templateID)
	if trimmedID == "" {
		trimmedID = "tpl-001-tax-pro"
	}

	// 1. Try DB lookup if connected
	if db.Instance != nil && db.Instance.IsConnected {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if t, err := db.Instance.Prisma.Template.FindUnique(db.Template.ID.Equals(trimmedID)).Exec(ctx); err == nil && t != nil {
			sub := ""
			if s, ok := t.Subtitle(); ok {
				sub = s
			}
			return &TemplateResponse{
				ID:          t.ID,
				Title:       t.Title,
				Subtitle:    sub,
				Description: t.Description,
				Image:       t.Image,
				IsActive:    t.IsActive,
				CreatedAt:   t.CreatedAt,
				UpdatedAt:   t.UpdatedAt,
			}
		}
	}

	// 2. Try mock/in-memory templates
	mockTemplatesLock.RLock()
	defer mockTemplatesLock.RUnlock()
	for _, t := range mockTemplates {
		if strings.EqualFold(t.ID, trimmedID) || strings.EqualFold(t.Title, trimmedID) {
			return &t
		}
	}

	// 3. Fallback matching for dominator-prime or default templates
	if len(mockTemplates) > 0 {
		tpl := mockTemplates[0]
		tpl.ID = trimmedID
		if strings.Contains(strings.ToLower(trimmedID), "small-biz") && len(mockTemplates) > 1 {
			tpl = mockTemplates[1]
			tpl.ID = trimmedID
		}
		return &tpl
	}

	return &TemplateResponse{
		ID:          trimmedID,
		Title:       "Tax Practice Dominator Pro",
		Subtitle:    "High-converting turnkey tax practice website template",
		Description: "Optimized for CPA firms, tax preparers, and bookkeeping practices to capture local search leads.",
		Image:       "/uploads/templates/tax_practice_dominator.jpg",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func normalizeSubmissionItem(m map[string]any) {
	// Extract template ID from either template_id or templateId
	var templateID string
	if tid, ok := m["template_id"].(string); ok && strings.TrimSpace(tid) != "" {
		templateID = strings.TrimSpace(tid)
	} else if tid, ok := m["templateId"].(string); ok && strings.TrimSpace(tid) != "" {
		templateID = strings.TrimSpace(tid)
	}

	if templateID == "" {
		templateID = "tpl-001-tax-pro"
	}

	// Always populate BOTH snake_case and camelCase template ID keys so no client fails
	m["template_id"] = templateID
	m["templateId"] = templateID

	// Populate template object if missing or null
	var hasValidTemplate bool
	if tmplRaw, ok := m["template"]; ok && tmplRaw != nil {
		if tmplMap, ok := tmplRaw.(map[string]any); ok && len(tmplMap) > 0 {
			if idVal, ok := tmplMap["id"].(string); ok && idVal != "" {
				hasValidTemplate = true
			}
		}
	}

	if !hasValidTemplate {
		m["template"] = resolveTemplateObject(templateID)
	}
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
			if (key == "colorScheme" || key == "font" ||
				key == "otherSocialUrls" || key == "other_social_urls" ||
				key == "backgroundImages" || key == "background_images" ||
				key == "officeImages" || key == "office_images" ||
				key == "recentWinsImages" || key == "recent_wins_images") && item != nil {
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
		normalizeSubmissionItem(current)
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
	contentType := c.GetHeader("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") || strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		_ = c.Request.ParseMultipartForm(50 << 20) // 50MB max file memory

		input.BusinessName = getFormString(c, "business_name", "businessName")
		input.Dba = getFormString(c, "dba")
		input.ContactName = getFormString(c, "contact_name", "contactName")
		input.FirstName = getFormString(c, "first_name", "firstName")
		input.LastName = getFormString(c, "last_name", "lastName")
		input.CountryCode = getFormString(c, "country_code", "countryCode")
		input.Phone = getFormString(c, "phone")
		input.Email = getFormString(c, "email")
		hasExistingWeb := getFormBool(c, false, "has_existing_website", "hasExistingWebsite")
		input.HasExistingWebsite = &hasExistingWeb
		input.WebsiteUrl = getFormString(c, "website_url", "websiteUrl")
		hasGbp := getFormBool(c, false, "has_google_business_profile", "hasGoogleBusinessProfile")
		input.HasGoogleBusinessProfile = &hasGbp
		input.GbpLink = getFormString(c, "gbp_link", "gbpLink", "gmb_url", "gmbUrl")
		input.OfficeAddresses = getFormString(c, "office_addresses", "officeAddresses")
		input.StreetAddress = getFormString(c, "street_address", "streetAddress")
		input.City = getFormString(c, "city")
		input.State = getFormString(c, "state")
		input.ZipCode = getFormString(c, "zip_code", "zipCode")
		input.PrimaryCategory = getFormString(c, "primary_category", "primaryCategory")
		input.PrimaryProductService = getFormString(c, "primary_product_service", "primaryProductService")
		input.ServicesOffered = getFormString(c, "services_offered", "servicesOffered")
		input.AdditionalServices = getFormString(c, "additional_services", "additionalServices")
		input.TargetLocations = getFormString(c, "target_locations", "targetLocations")
		input.Keywords = getFormString(c, "keywords")
		input.VisitModel = getFormString(c, "visit_model", "visitModel", "customer_visit_type", "customerVisitType")
		consentTrans := getFormBool(c, false, "consent_transactional", "consentTransactional")
		input.ConsentTransactional = &consentTrans
		consentMkt := getFormBool(c, false, "consent_marketing", "consentMarketing")
		input.ConsentMarketing = &consentMkt
		input.TwitterUrl = getFormString(c, "twitter_url", "twitterUrl")
		input.FacebookUrl = getFormString(c, "facebook_url", "facebookUrl")
		input.InstagramUrl = getFormString(c, "instagram_url", "instagramUrl")
		input.YoutubeUrl = getFormString(c, "youtube_url", "youtubeUrl")
		input.OtherSocialUrls = getFormString(c, "other_social_urls", "otherSocialUrls")
		input.UseExistingImages = getFormString(c, "use_existing_images", "useExistingImages")
		input.AgreementBusinessName = getFormString(c, "agreement_business_name", "agreementBusinessName")
		input.SignerFirstName = getFormString(c, "signer_first_name", "signerFirstName")
		input.SignerLastName = getFormString(c, "signer_last_name", "signerLastName")
		input.SignerPhone = getFormString(c, "signer_phone", "signerPhone")
		agreed := getFormBool(c, false, "agreed_terms", "agreedTerms")
		input.AgreedTerms = &agreed
		input.TemplateID = getFormString(c, "template_id", "templateId", "selected_template", "selectedTemplate")

		// Handle file uploads
		uploadedLogo := saveSingleUploadedFile(c, []string{"logo", "logo_file", "logoFile", "logoImage"}, LogoUploadDir, "/uploads/onboarding/logos")
		if uploadedLogo != "" {
			input.LogoUrl = uploadedLogo
		} else {
			input.LogoUrl = getFormString(c, "logo_url", "logoUrl")
		}

		uploadedBg := saveUploadedFiles(c, []string{"backgrounds", "background_images", "backgroundImages"}, BackgroundsUploadDir, "/uploads/onboarding/backgrounds")
		if len(uploadedBg) > 0 {
			input.BackgroundImages = uploadedBg
		} else if bgForm := getFormString(c, "background_images", "backgroundImages"); bgForm != "" {
			input.BackgroundImages = bgForm
		}

		uploadedOffice := saveUploadedFiles(c, []string{"office", "office_images", "officeImages"}, OfficeUploadDir, "/uploads/onboarding/office")
		if len(uploadedOffice) > 0 {
			input.OfficeImages = uploadedOffice
		} else if officeForm := getFormString(c, "office_images", "officeImages"); officeForm != "" {
			input.OfficeImages = officeForm
		}

		uploadedRecentWins := saveUploadedFiles(c, []string{"recentWins", "recent_wins", "recentWinsImages", "recent_wins_images"}, RecentWinsUploadDir, "/uploads/onboarding/recent_wins")
		if len(uploadedRecentWins) > 0 {
			input.RecentWinsImages = uploadedRecentWins
		} else if rwForm := getFormString(c, "recent_wins_images", "recentWinsImages"); rwForm != "" {
			input.RecentWinsImages = rwForm
		}
	} else {
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
	}

	input.Clean()

	if strings.TrimSpace(input.BusinessName) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "business_name is required",
		})
		return
	}
	if strings.TrimSpace(input.Email) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "email is required",
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

	otherSocialJSON := toJSONBytes(input.OtherSocialUrls)
	bgImagesJSON := toJSONBytes(input.BackgroundImages)
	officeImagesJSON := toJSONBytes(input.OfficeImages)
	recentWinsJSON := toJSONBytes(input.RecentWinsImages)

	hasExistingWebsiteVal := false
	if input.HasExistingWebsite != nil {
		hasExistingWebsiteVal = *input.HasExistingWebsite
	}
	hasGoogleBusinessProfileVal := false
	if input.HasGoogleBusinessProfile != nil {
		hasGoogleBusinessProfileVal = *input.HasGoogleBusinessProfile
	}
	consentTransactionalVal := false
	if input.ConsentTransactional != nil {
		consentTransactionalVal = *input.ConsentTransactional
	}
	consentMarketingVal := false
	if input.ConsentMarketing != nil {
		consentMarketingVal = *input.ConsentMarketing
	}
	agreedTermsVal := false
	if input.AgreedTerms != nil {
		agreedTermsVal = *input.AgreedTerms
	}

	if db.Instance != nil && db.Instance.IsConnected {
		createParams := []db.OnboardingSubmissionSetParam{
			db.OnboardingSubmission.Token.Set(token),
			db.OnboardingSubmission.HasExistingWebsite.Set(hasExistingWebsiteVal),
			db.OnboardingSubmission.HasGoogleBusinessProfile.Set(hasGoogleBusinessProfileVal),
			db.OnboardingSubmission.ConsentTransactional.Set(consentTransactionalVal),
			db.OnboardingSubmission.ConsentMarketing.Set(consentMarketingVal),
			db.OnboardingSubmission.AgreedTerms.Set(agreedTermsVal),
			db.OnboardingSubmission.Status.Set("PENDING"),
		}

		if input.Dba != "" {
			createParams = append(createParams, db.OnboardingSubmission.Dba.Set(input.Dba))
		}
		if input.CountryCode != "" {
			createParams = append(createParams, db.OnboardingSubmission.CountryCode.Set(input.CountryCode))
		}
		if input.WebsiteUrl != "" {
			createParams = append(createParams, db.OnboardingSubmission.WebsiteURL.Set(input.WebsiteUrl))
		}
		if input.GbpLink != "" {
			createParams = append(createParams, db.OnboardingSubmission.GbpLink.Set(input.GbpLink))
		}
		if input.OfficeAddresses != "" {
			createParams = append(createParams, db.OnboardingSubmission.OfficeAddresses.Set(input.OfficeAddresses))
		}
		if input.StreetAddress != "" {
			createParams = append(createParams, db.OnboardingSubmission.StreetAddress.Set(input.StreetAddress))
		}
		if input.City != "" {
			createParams = append(createParams, db.OnboardingSubmission.City.Set(input.City))
		}
		if input.State != "" {
			createParams = append(createParams, db.OnboardingSubmission.State.Set(input.State))
		}
		if input.ZipCode != "" {
			createParams = append(createParams, db.OnboardingSubmission.ZipCode.Set(input.ZipCode))
		}
		if input.PrimaryCategory != "" {
			createParams = append(createParams, db.OnboardingSubmission.PrimaryCategory.Set(input.PrimaryCategory))
		}
		if input.PrimaryProductService != "" {
			createParams = append(createParams, db.OnboardingSubmission.PrimaryProductService.Set(input.PrimaryProductService))
		}
		if input.ServicesOffered != "" {
			createParams = append(createParams, db.OnboardingSubmission.ServicesOffered.Set(input.ServicesOffered))
		}
		if input.AdditionalServices != "" {
			createParams = append(createParams, db.OnboardingSubmission.AdditionalServices.Set(input.AdditionalServices))
		}
		if input.TargetLocations != "" {
			createParams = append(createParams, db.OnboardingSubmission.TargetLocations.Set(input.TargetLocations))
		}
		if input.Keywords != "" {
			createParams = append(createParams, db.OnboardingSubmission.Keywords.Set(input.Keywords))
		}
		if input.VisitModel != "" {
			createParams = append(createParams, db.OnboardingSubmission.VisitModel.Set(input.VisitModel))
		}
		if input.TwitterUrl != "" {
			createParams = append(createParams, db.OnboardingSubmission.TwitterURL.Set(input.TwitterUrl))
		}
		if input.FacebookUrl != "" {
			createParams = append(createParams, db.OnboardingSubmission.FacebookURL.Set(input.FacebookUrl))
		}
		if input.InstagramUrl != "" {
			createParams = append(createParams, db.OnboardingSubmission.InstagramURL.Set(input.InstagramUrl))
		}
		if input.YoutubeUrl != "" {
			createParams = append(createParams, db.OnboardingSubmission.YoutubeURL.Set(input.YoutubeUrl))
		}
		if len(otherSocialJSON) > 0 {
			createParams = append(createParams, db.OnboardingSubmission.OtherSocialUrls.Set(db.JSON(otherSocialJSON)))
		}
		if input.UseExistingImages != "" {
			createParams = append(createParams, db.OnboardingSubmission.UseExistingImages.Set(input.UseExistingImages))
		}
		if input.LogoUrl != "" {
			createParams = append(createParams, db.OnboardingSubmission.LogoURL.Set(input.LogoUrl))
		}
		if len(bgImagesJSON) > 0 {
			createParams = append(createParams, db.OnboardingSubmission.BackgroundImages.Set(db.JSON(bgImagesJSON)))
		}
		if len(officeImagesJSON) > 0 {
			createParams = append(createParams, db.OnboardingSubmission.OfficeImages.Set(db.JSON(officeImagesJSON)))
		}
		if len(recentWinsJSON) > 0 {
			createParams = append(createParams, db.OnboardingSubmission.RecentWinsImages.Set(db.JSON(recentWinsJSON)))
		}
		if input.AgreementBusinessName != "" {
			createParams = append(createParams, db.OnboardingSubmission.AgreementBusinessName.Set(input.AgreementBusinessName))
		}
		if input.SignerFirstName != "" {
			createParams = append(createParams, db.OnboardingSubmission.SignerFirstName.Set(input.SignerFirstName))
		}
		if input.SignerLastName != "" {
			createParams = append(createParams, db.OnboardingSubmission.SignerLastName.Set(input.SignerLastName))
		}
		if input.SignerPhone != "" {
			createParams = append(createParams, db.OnboardingSubmission.SignerPhone.Set(input.SignerPhone))
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
		Dba:                      input.Dba,
		ContactName:              input.ContactName,
		CountryCode:              input.CountryCode,
		Phone:                    input.Phone,
		Email:                    input.Email,
		HasExistingWebsite:       hasExistingWebsiteVal,
		WebsiteUrl:               input.WebsiteUrl,
		HasGoogleBusinessProfile: hasGoogleBusinessProfileVal,
		GbpLink:                  input.GbpLink,
		OfficeAddresses:          input.OfficeAddresses,
		StreetAddress:            input.StreetAddress,
		City:                     input.City,
		State:                    input.State,
		ZipCode:                  input.ZipCode,
		PrimaryCategory:          input.PrimaryCategory,
		PrimaryProductService:    input.PrimaryProductService,
		ServicesOffered:          input.ServicesOffered,
		AdditionalServices:       input.AdditionalServices,
		TargetLocations:          input.TargetLocations,
		Keywords:                 input.Keywords,
		VisitModel:               input.VisitModel,
		ConsentTransactional:     consentTransactionalVal,
		ConsentMarketing:         consentMarketingVal,
		TwitterUrl:               input.TwitterUrl,
		FacebookUrl:              input.FacebookUrl,
		InstagramUrl:             input.InstagramUrl,
		YoutubeUrl:               input.YoutubeUrl,
		OtherSocialUrls:          input.OtherSocialUrls,
		UseExistingImages:        input.UseExistingImages,
		LogoUrl:                  input.LogoUrl,
		BackgroundImages:         input.BackgroundImages,
		OfficeImages:             input.OfficeImages,
		RecentWinsImages:         input.RecentWinsImages,
		AgreementBusinessName:    input.AgreementBusinessName,
		SignerFirstName:          input.SignerFirstName,
		SignerLastName:           input.SignerLastName,
		SignerPhone:              input.SignerPhone,
		AgreedTerms:              agreedTermsVal,
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
		subs, err := db.Instance.Prisma.OnboardingSubmission.FindMany().With(
			db.OnboardingSubmission.Template.Fetch(),
		).Exec(ctx)
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
		"data":    normalizeOnboardingData(mockSubmissions),
	})
}

func GetOnboardingSubmissionByID(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	if db.Instance != nil && db.Instance.IsConnected {
		sub, err := db.Instance.Prisma.OnboardingSubmission.FindUnique(
			db.OnboardingSubmission.ID.Equals(id),
		).With(
			db.OnboardingSubmission.Template.Fetch(),
		).Exec(ctx)

		if err == nil && sub != nil {
			// If templateId is empty in DB, auto-populate default template
			if tid, ok := sub.TemplateID(); !ok || strings.TrimSpace(tid) == "" {
				go func(subID string) {
					bgCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					_, _ = db.Instance.Prisma.OnboardingSubmission.FindUnique(
						db.OnboardingSubmission.ID.Equals(subID),
					).Update(
						db.OnboardingSubmission.TemplateID.Set("tpl-001-tax-pro"),
					).Exec(bgCtx)
				}(id)
			}

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
				"data":    normalizeOnboardingData(s),
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
	Dba                      string `json:"dba"`
	ContactName              string `json:"contact_name"`
	CountryCode              string `json:"country_code"`
	Phone                    string `json:"phone"`
	Email                    string `json:"email"`
	HasExistingWebsite       *bool  `json:"has_existing_website"`
	WebsiteUrl               string `json:"website_url"`
	HasGoogleBusinessProfile *bool  `json:"has_google_business_profile"`
	GbpLink                  string `json:"gbp_link"`
	OfficeAddresses          string `json:"office_addresses"`
	StreetAddress            string `json:"street_address"`
	City                     string `json:"city"`
	State                    string `json:"state"`
	ZipCode                  string `json:"zip_code"`
	PrimaryCategory          string `json:"primary_category"`
	PrimaryProductService    string `json:"primary_product_service"`
	ServicesOffered          string `json:"services_offered"`
	AdditionalServices       string `json:"additional_services"`
	TargetLocations          string `json:"target_locations"`
	Keywords                 string `json:"keywords"`
	VisitModel               string `json:"visit_model"`
	ConsentTransactional     *bool  `json:"consent_transactional"`
	ConsentMarketing         *bool  `json:"consent_marketing"`
	TwitterUrl               string `json:"twitter_url"`
	FacebookUrl              string `json:"facebook_url"`
	InstagramUrl             string `json:"instagram_url"`
	YoutubeUrl               string `json:"youtube_url"`
	OtherSocialUrls          any    `json:"other_social_urls"`
	UseExistingImages        string `json:"use_existing_images"`
	LogoUrl                  string `json:"logo_url"`
	BackgroundImages         any    `json:"background_images"`
	OfficeImages             any    `json:"office_images"`
	RecentWinsImages         any    `json:"recent_wins_images"`
	AgreementBusinessName    string `json:"agreement_business_name"`
	SignerFirstName          string `json:"signer_first_name"`
	SignerLastName           string `json:"signer_last_name"`
	SignerPhone              string `json:"signer_phone"`
	AgreedTerms              *bool  `json:"agreed_terms"`
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
		if input.Dba != "" {
			updates = append(updates, db.OnboardingSubmission.Dba.Set(input.Dba))
		}
		if input.ContactName != "" {
			updates = append(updates, db.OnboardingSubmission.ContactName.Set(input.ContactName))
		}
		if input.CountryCode != "" {
			updates = append(updates, db.OnboardingSubmission.CountryCode.Set(input.CountryCode))
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
		if input.OfficeAddresses != "" {
			updates = append(updates, db.OnboardingSubmission.OfficeAddresses.Set(input.OfficeAddresses))
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
		if input.PrimaryProductService != "" {
			updates = append(updates, db.OnboardingSubmission.PrimaryProductService.Set(input.PrimaryProductService))
		}
		if input.ServicesOffered != "" {
			updates = append(updates, db.OnboardingSubmission.ServicesOffered.Set(input.ServicesOffered))
		}
		if input.AdditionalServices != "" {
			updates = append(updates, db.OnboardingSubmission.AdditionalServices.Set(input.AdditionalServices))
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
		if input.TwitterUrl != "" {
			updates = append(updates, db.OnboardingSubmission.TwitterURL.Set(input.TwitterUrl))
		}
		if input.FacebookUrl != "" {
			updates = append(updates, db.OnboardingSubmission.FacebookURL.Set(input.FacebookUrl))
		}
		if input.InstagramUrl != "" {
			updates = append(updates, db.OnboardingSubmission.InstagramURL.Set(input.InstagramUrl))
		}
		if input.YoutubeUrl != "" {
			updates = append(updates, db.OnboardingSubmission.YoutubeURL.Set(input.YoutubeUrl))
		}
		if input.OtherSocialUrls != nil {
			if b := toJSONBytes(input.OtherSocialUrls); len(b) > 0 {
				updates = append(updates, db.OnboardingSubmission.OtherSocialUrls.Set(db.JSON(b)))
			}
		}
		if input.UseExistingImages != "" {
			updates = append(updates, db.OnboardingSubmission.UseExistingImages.Set(input.UseExistingImages))
		}
		if input.LogoUrl != "" {
			updates = append(updates, db.OnboardingSubmission.LogoURL.Set(input.LogoUrl))
		}
		if input.BackgroundImages != nil {
			if b := toJSONBytes(input.BackgroundImages); len(b) > 0 {
				updates = append(updates, db.OnboardingSubmission.BackgroundImages.Set(db.JSON(b)))
			}
		}
		if input.OfficeImages != nil {
			if b := toJSONBytes(input.OfficeImages); len(b) > 0 {
				updates = append(updates, db.OnboardingSubmission.OfficeImages.Set(db.JSON(b)))
			}
		}
		if input.RecentWinsImages != nil {
			if b := toJSONBytes(input.RecentWinsImages); len(b) > 0 {
				updates = append(updates, db.OnboardingSubmission.RecentWinsImages.Set(db.JSON(b)))
			}
		}
		if input.AgreementBusinessName != "" {
			updates = append(updates, db.OnboardingSubmission.AgreementBusinessName.Set(input.AgreementBusinessName))
		}
		if input.SignerFirstName != "" {
			updates = append(updates, db.OnboardingSubmission.SignerFirstName.Set(input.SignerFirstName))
		}
		if input.SignerLastName != "" {
			updates = append(updates, db.OnboardingSubmission.SignerLastName.Set(input.SignerLastName))
		}
		if input.SignerPhone != "" {
			updates = append(updates, db.OnboardingSubmission.SignerPhone.Set(input.SignerPhone))
		}
		if input.AgreedTerms != nil {
			updates = append(updates, db.OnboardingSubmission.AgreedTerms.Set(*input.AgreedTerms))
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
			if input.Dba != "" {
				mockSubmissions[i].Dba = input.Dba
			}
			if input.ContactName != "" {
				mockSubmissions[i].ContactName = input.ContactName
			}
			if input.CountryCode != "" {
				mockSubmissions[i].CountryCode = input.CountryCode
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
			if input.OfficeAddresses != "" {
				mockSubmissions[i].OfficeAddresses = input.OfficeAddresses
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
			if input.PrimaryProductService != "" {
				mockSubmissions[i].PrimaryProductService = input.PrimaryProductService
			}
			if input.ServicesOffered != "" {
				mockSubmissions[i].ServicesOffered = input.ServicesOffered
			}
			if input.AdditionalServices != "" {
				mockSubmissions[i].AdditionalServices = input.AdditionalServices
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
			if input.TwitterUrl != "" {
				mockSubmissions[i].TwitterUrl = input.TwitterUrl
			}
			if input.FacebookUrl != "" {
				mockSubmissions[i].FacebookUrl = input.FacebookUrl
			}
			if input.InstagramUrl != "" {
				mockSubmissions[i].InstagramUrl = input.InstagramUrl
			}
			if input.YoutubeUrl != "" {
				mockSubmissions[i].YoutubeUrl = input.YoutubeUrl
			}
			if input.OtherSocialUrls != nil {
				mockSubmissions[i].OtherSocialUrls = input.OtherSocialUrls
			}
			if input.UseExistingImages != "" {
				mockSubmissions[i].UseExistingImages = input.UseExistingImages
			}
			if input.LogoUrl != "" {
				mockSubmissions[i].LogoUrl = input.LogoUrl
			}
			if input.BackgroundImages != nil {
				mockSubmissions[i].BackgroundImages = input.BackgroundImages
			}
			if input.OfficeImages != nil {
				mockSubmissions[i].OfficeImages = input.OfficeImages
			}
			if input.RecentWinsImages != nil {
				mockSubmissions[i].RecentWinsImages = input.RecentWinsImages
			}
			if input.AgreementBusinessName != "" {
				mockSubmissions[i].AgreementBusinessName = input.AgreementBusinessName
			}
			if input.SignerFirstName != "" {
				mockSubmissions[i].SignerFirstName = input.SignerFirstName
			}
			if input.SignerLastName != "" {
				mockSubmissions[i].SignerLastName = input.SignerLastName
			}
			if input.SignerPhone != "" {
				mockSubmissions[i].SignerPhone = input.SignerPhone
			}
			if input.AgreedTerms != nil {
				mockSubmissions[i].AgreedTerms = *input.AgreedTerms
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
