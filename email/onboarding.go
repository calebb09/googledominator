package email

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	senderEmail = "info@googledominator.co"
	adminEmail  = "eglobalellc@gmail.com"
)

var graphClient = &http.Client{Timeout: 30 * time.Second}

// SendOnboardingNotifications sends through Microsoft Graph using application
// credentials. The Entra application must have the Mail.Send application
// permission with administrator consent.
func SendOnboardingNotifications(userEmail, contactName, businessName string) error {
	from := strings.TrimSpace(os.Getenv("EMAIL_ADDRESS"))
	if !strings.EqualFold(from, senderEmail) {
		return fmt.Errorf("EMAIL_ADDRESS must be configured as %s", senderEmail)
	}

	tenantID := strings.TrimSpace(os.Getenv("MICROSOFT_TENANT_ID"))
	clientID := strings.TrimSpace(os.Getenv("MICROSOFT_CLIENT_ID"))
	clientSecret := os.Getenv("MICROSOFT_CLIENT_SECRET")
	if tenantID == "" || clientID == "" || clientSecret == "" {
		return errors.New("MICROSOFT_TENANT_ID, MICROSOFT_CLIENT_ID, and MICROSOFT_CLIENT_SECRET must be configured")
	}

	accessToken, err := getGraphAccessToken(tenantID, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("authenticate with Microsoft Graph: %w", err)
	}

	userSubject := "GoogleDominator onboarding form received"
	userBody := fmt.Sprintf("Hello %s,\n\nYour onboarding form for %s was successfully submitted to GoogleDominator.\n\nThank you,\nThe GoogleDominator Team", contactName, businessName)
	adminSubject := "New GoogleDominator onboarding submission"
	adminBody := fmt.Sprintf("A user has submitted a new onboarding form.\n\nContact: %s\nBusiness: %s\nEmail: %s\n\nPlease log in to your admin account and check the submission.", contactName, businessName, userEmail)

	var sendErrors []error
	if err = sendGraphMail(accessToken, userEmail, userSubject, userBody); err != nil {
		log.Printf("[ERROR] Failed to send onboarding confirmation email to %s: %v", userEmail, err)
		sendErrors = append(sendErrors, fmt.Errorf("send submitter notification: %w", err))
	} else {
		log.Printf("[INFO] Onboarding confirmation email successfully sent from %s to %s", senderEmail, userEmail)
	}
	if err = sendGraphMail(accessToken, adminEmail, adminSubject, adminBody); err != nil {
		log.Printf("[ERROR] Failed to send onboarding admin notification to %s: %v", adminEmail, err)
		sendErrors = append(sendErrors, fmt.Errorf("send admin notification: %w", err))
	} else {
		log.Printf("[INFO] Onboarding admin notification successfully sent from %s to %s", senderEmail, adminEmail)
	}
	return errors.Join(sendErrors...)
}

func getGraphAccessToken(tenantID, clientID, clientSecret string) (string, error) {
	endpoint := "https://login.microsoftonline.com/" + url.PathEscape(tenantID) + "/oauth2/v2.0/token"
	form := url.Values{
		"client_id": {clientID}, "client_secret": {clientSecret},
		"scope": {"https://graph.microsoft.com/.default"}, "grant_type": {"client_credentials"},
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := graphClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err = json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}
	if tokenResponse.AccessToken == "" {
		return "", errors.New("Microsoft Graph returned an empty access token")
	}
	return tokenResponse.AccessToken, nil
}

func sendGraphMail(accessToken, recipient, subject, body string) error {
	address, err := mail.ParseAddress(recipient)
	if err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	payload := map[string]any{
		"message": map[string]any{
			"subject":      subject,
			"body":         map[string]string{"contentType": "Text", "content": body},
			"toRecipients": []map[string]any{{"emailAddress": map[string]string{"address": address.Address}}},
		},
		"saveToSentItems": true,
	}
	encodedPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	endpoint := "https://graph.microsoft.com/v1.0/users/" + url.PathEscape(senderEmail) + "/sendMail"
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(encodedPayload)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := graphClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Microsoft Graph returned %s: %s", resp.Status, strings.TrimSpace(string(responseBody)))
	}
	return nil
}
