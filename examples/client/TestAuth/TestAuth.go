package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func main() {
	c, err := client.NewClient(
		envOrDefault("AXM_BASE_URL", "https://api-business.apple.com"),
		requireEnv("AXM_TEAM_ID"),
		requireEnv("AXM_CLIENT_ID"),
		requireEnv("AXM_KEY_ID"),
		envOrDefault("AXM_SCOPE", "business.api"),
		requireEnv("AXM_PRIVATE_KEY"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	assertion, assertionExpiry, token, err := c.TestAuth()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("=== JWT Client Assertion ===")
	fmt.Printf("%s\n\n", assertion)
	fmt.Printf("Assertion Expires At: %s\n", assertionExpiry.Format(time.RFC3339))
	fmt.Printf("Time Until Assertion Expiry: %s\n\n", time.Until(assertionExpiry).Round(time.Second))

	fmt.Println("=== OAuth Token ===")
	fmt.Printf("Access Token: %s\n", token.AccessToken)
	fmt.Printf("Token Type: %s\n", token.TokenType)
	fmt.Printf("Expires At: %s\n", token.Expiry.Format(time.RFC3339))
	fmt.Printf("Time Until Token Expiry: %s\n", time.Until(token.Expiry).Round(time.Second))
	fmt.Printf("Token Valid: %v\n", token.Valid())
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return v
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
