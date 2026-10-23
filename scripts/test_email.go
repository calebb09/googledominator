//go:build ignore

// Run with: go run scripts/test_email.go
// Tests the Microsoft Graph email integration end-to-end.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	gdemail "googledominator-backend/email"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] No .env file loaded, using system env")
	}

	tenantID := os.Getenv("MICROSOFT_TENANT_ID")
	clientID := os.Getenv("MICROSOFT_CLIENT_ID")
	clientSecret := os.Getenv("MICROSOFT_CLIENT_SECRET")
	emailAddr := os.Getenv("EMAIL_ADDRESS")

	fmt.Println("=== Email Config Check ===")
	fmt.Printf("EMAIL_ADDRESS       : %s\n", emailAddr)
	fmt.Printf("MICROSOFT_TENANT_ID : %s\n", tenantID)
	fmt.Printf("MICROSOFT_CLIENT_ID : %s\n", clientID)
	fmt.Printf("MICROSOFT_CLIENT_SECRET length: %d chars\n", len(clientSecret))

	if tenantID == "" || clientID == "" || clientSecret == "" || emailAddr == "" {
		log.Fatal("[FATAL] One or more required env vars are missing")
	}

	fmt.Println("\n=== Sending Test Email ===")
	err := gdemail.SendOnboardingNotifications(
		emailAddr,           // send to yourself as test
		"Test Contact Name",
		"Test Business LLC",
	)
	if err != nil {
		log.Fatalf("[FAIL] Email send failed: %v", err)
	}
	fmt.Println("[OK] Email sent successfully! Check your inbox.")
}

