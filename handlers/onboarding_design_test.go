package handlers

import (
	"encoding/json"
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
