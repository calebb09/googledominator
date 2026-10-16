package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"googledominator-backend/db"
)

func TestPickTemplateURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	SetOnboardingURLs("https://picker.example/", "https://api.example/")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "http://api.example/api/v1/onboarding", nil)
	c.Request.Host = "api.example"
	c.Request.Header.Set("X-Forwarded-Proto", "https")

	got := pickTemplateURL(c, "tok_test")
	want := "https://picker.example/pick-template?token=tok_test&submitUrl=https://api.example/api/templates/pick"
	if got != want {
		t.Fatalf("pickTemplateURL() = %q, want %q", got, want)
	}
}

func TestNormalizeOnboardingDataJSONFields(t *testing.T) {
	got := normalizeOnboardingData(map[string]any{
		"colorScheme": `{"id":"custom","primary":"#e11d48"}`,
		"font":        `{"heading":"playfair","body":"source-sans"}`,
	}).(map[string]any)

	colorScheme, ok := got["colorScheme"].(map[string]any)
	if !ok || colorScheme["id"] != "custom" {
		t.Fatalf("colorScheme was not returned as an object: %#v", got["colorScheme"])
	}
	font, ok := got["font"].(map[string]any)
	if !ok || font["body"] != "source-sans" {
		t.Fatalf("font was not returned as an object: %#v", got["font"])
	}
}

func TestUpdateOnboardingDesignMock(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db.Instance = nil
	token := "tok_design_test"
	mockSubmissionsLock.Lock()
	previous := mockSubmissions
	mockSubmissions = []OnboardingSubmissionResponse{{ID: "sub-test", Token: token}}
	mockSubmissionsLock.Unlock()
	t.Cleanup(func() {
		mockSubmissionsLock.Lock()
		mockSubmissions = previous
		mockSubmissionsLock.Unlock()
	})

	body := `{
		"colorScheme":{"id":"custom","name":"Custom","primary":"#e11d48","secondary":"#111827","accent":"#fef3c7","swatches":["#111827","#e11d48","#fef3c7"]},
		"font":{"heading":"playfair","body":"source-sans"},
		"editorPath":"/editor/business?colorScheme=custom"
	}`
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/templates/pick?token="+token, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateOnboardingDesign(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
	var response struct {
		Data OnboardingSubmissionResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Data.ColorScheme == nil || response.Data.ColorScheme.Primary != "#e11d48" {
		t.Fatalf("color scheme was not saved: %#v", response.Data.ColorScheme)
	}
	if response.Data.Font == nil || response.Data.Font.Body != "source-sans" {
		t.Fatalf("font was not saved: %#v", response.Data.Font)
	}
}

func TestCreateOnboardingSubmission_MultipartForm(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db.Instance = nil // Ensure mock memory mode is tested

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("business_name", "Test Tax Office")
	_ = writer.WriteField("dba", "TTO Services")
	_ = writer.WriteField("contact_name", "Alice Walker")
	_ = writer.WriteField("phone", "+14045550199")
	_ = writer.WriteField("email", "alice@example.com")
	_ = writer.WriteField("office_addresses", "100 Peachtree St, Atlanta GA")
	_ = writer.WriteField("primary_category", "Tax Preparation")
	_ = writer.WriteField("primary_product_service", "Income Tax & Bookkeeping")
	_ = writer.WriteField("additional_services", "Payroll, Advisory")
	_ = writer.WriteField("target_locations", "Atlanta, GA\nMarietta, GA")
	_ = writer.WriteField("keywords", "atlanta tax return, tax preparer near me")
	_ = writer.WriteField("visit_model", "In-Person & Online")
	_ = writer.WriteField("twitter_url", "https://x.com/testtax")
	_ = writer.WriteField("agreement_business_name", "Test Tax Office LLC")
	_ = writer.WriteField("signer_first_name", "Alice")
	_ = writer.WriteField("signer_last_name", "Walker")
	_ = writer.WriteField("agreed_terms", "true")
	_ = writer.WriteField("template_id", "dominator-prime")

	// Add file upload for logo
	logoPart, err := writer.CreateFormFile("logo", "logo.png")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = logoPart.Write([]byte("fake-logo-content-bytes"))

	// Add file upload for backgrounds
	bgPart, err := writer.CreateFormFile("backgrounds", "bg1.jpg")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = bgPart.Write([]byte("fake-bg-content-bytes"))

	_ = writer.Close()

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/onboarding", &body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	CreateOnboardingSubmission(c)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var res struct {
		Success bool                         `json:"success"`
		Data    OnboardingSubmissionResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if res.Data.BusinessName != "Test Tax Office" {
		t.Errorf("expected BusinessName 'Test Tax Office', got %q", res.Data.BusinessName)
	}
	if res.Data.Dba != "TTO Services" {
		t.Errorf("expected Dba 'TTO Services', got %q", res.Data.Dba)
	}
	if !strings.HasPrefix(res.Data.LogoUrl, "/uploads/onboarding/logos/") {
		t.Errorf("expected LogoUrl to start with '/uploads/onboarding/logos/', got %q", res.Data.LogoUrl)
	}
	if res.Data.BackgroundImages == nil {
		t.Errorf("expected BackgroundImages to be set, got nil")
	}
	if res.Data.SignerFirstName != "Alice" {
		t.Errorf("expected SignerFirstName 'Alice', got %q", res.Data.SignerFirstName)
	}
	if !res.Data.AgreedTerms {
		t.Errorf("expected AgreedTerms true, got false")
	}
}

func TestCreateOnboardingSubmission_JSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db.Instance = nil

	jsonPayload := `{
		"businessName": "Vanguard Tax & CPA",
		"dba": "Vanguard Advisory",
		"contactName": "Robert Brown",
		"phone": "+12125550188",
		"email": "robert@vanguardtax.com",
		"primaryCategory": "CPA Practice",
		"primaryProductService": "Corporate Tax & Resolution",
		"targetLocations": "New York, NY\nJersey City, NJ",
		"keywords": "cpa nyc, corporate tax",
		"customerVisitType": ["Virtual", "In-Person"],
		"facebookUrl": "https://facebook.com/vanguardcpa",
		"selectedTemplate": "dominator-prime",
		"signerFirstName": "Robert",
		"signerLastName": "Brown",
		"agreedTerms": true
	}`

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/onboarding", strings.NewReader(jsonPayload))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateOnboardingSubmission(c)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var res struct {
		Success bool                         `json:"success"`
		Data    OnboardingSubmissionResponse `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if res.Data.BusinessName != "Vanguard Tax & CPA" {
		t.Errorf("expected BusinessName 'Vanguard Tax & CPA', got %q", res.Data.BusinessName)
	}
	if res.Data.Dba != "Vanguard Advisory" {
		t.Errorf("expected Dba 'Vanguard Advisory', got %q", res.Data.Dba)
	}
	if res.Data.FacebookUrl != "https://facebook.com/vanguardcpa" {
		t.Errorf("expected FacebookUrl, got %q", res.Data.FacebookUrl)
	}
	if res.Data.TemplateID != "dominator-prime" {
		t.Errorf("expected TemplateID 'dominator-prime', got %q", res.Data.TemplateID)
	}
	if !res.Data.AgreedTerms {
		t.Errorf("expected AgreedTerms true, got false")
	}
}

func TestGetOnboardingSubmissionByID_PopulatesTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db.Instance = nil

	targetID := "d879dc7f-334a-4ed4-8f88-0736c3926179"
	mockSubmissionsLock.Lock()
	previous := mockSubmissions
	mockSubmissions = []OnboardingSubmissionResponse{
		{
			ID:           targetID,
			BusinessName: "Target Submission Co",
			ContactName:  "Jane Smith",
			Email:        "jane@target.com",
			Phone:        "+14045551212",
			TemplateID:   "tpl-001-tax-pro",
		},
	}
	mockSubmissionsLock.Unlock()
	t.Cleanup(func() {
		mockSubmissionsLock.Lock()
		mockSubmissions = previous
		mockSubmissionsLock.Unlock()
	})

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: targetID}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/onboarding/"+targetID, nil)

	GetOnboardingSubmissionByID(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", recorder.Code, recorder.Body.String())
	}

	var rawMap map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &rawMap); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	data, ok := rawMap["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data object, got %#v", rawMap["data"])
	}

	// Verify template_id is populated
	if tid, ok := data["template_id"].(string); !ok || tid != "tpl-001-tax-pro" {
		t.Errorf("expected template_id 'tpl-001-tax-pro', got %v", data["template_id"])
	}

	// Verify template object is populated with details
	tmpl, ok := data["template"].(map[string]any)
	if !ok || tmpl == nil {
		t.Fatalf("expected template object to be populated, got nil")
	}
	if tmpl["id"] != "tpl-001-tax-pro" {
		t.Errorf("expected template.id 'tpl-001-tax-pro', got %v", tmpl["id"])
	}
	if tmpl["title"] != "Tax Practice Dominator Pro" {
		t.Errorf("expected template.title 'Tax Practice Dominator Pro', got %v", tmpl["title"])
	}
}


