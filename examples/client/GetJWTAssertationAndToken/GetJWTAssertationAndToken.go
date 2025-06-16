package main

import (
	"context"
	"fmt"
	"log"
	"time"

	axm "github.com/neilmartin83/terraform-provider-axm/internal/provider"
)

func main() {
	config := &axm.ClientConfig{
		TeamID:   "BUSINESSAPI.123e4567-e89b-12d3-a456-426614174000",
		ClientID: "BUSINESSAPI.123e4567-e89b-12d3-a456-426614174000",
		KeyID:    "123e4567-e89b-12d3-a456-426614174000",
		PrivateKey: []byte(`-----BEGIN EC PRIVATE KEY-----
FAKEAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wTESTBAQQgZxY8ytVhyXPLdHlj
TESTx9TSUTcFK29+lHvA1DybmFAKEyhRANCAAQXv+VXUiVv511AIa4nEXBrTESTD+
FAKEFigCMU45fN5v94OvEUUV2eUR3t4UZpZ4tHbCNdzEyXNIbFAKEY2xAc
-----END EC PRIVATE KEY-----`),
		Scope: "business.api",
	}

	// Create new client
	client, err := axm.NewAppleOAuthClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create JWT assertion
	assertion, err := client.CreateClientAssertion()
	if err != nil {
		log.Fatalf("Failed to create assertion: %v", err)
	}

	fmt.Println("=== JWT Assertion ===")
	fmt.Printf("%s\n\n", assertion)

	// Request new token
	ctx := context.Background()
	token, err := client.GetValidToken(ctx)
	if err != nil {
		log.Fatalf("Failed to get token: %v", err)
	}

	fmt.Println("=== Token Information ===")
	fmt.Printf("Access Token: %s\n", token.AccessToken)
	fmt.Printf("Token Type: %s\n", token.TokenType)
	fmt.Printf("Expires In: %d seconds\n", token.ExpiresIn)
	fmt.Printf("Scope: %s\n", token.Scope)
	fmt.Printf("Expires At: %s\n", token.ExpiresAt.Format(time.RFC3339))
	fmt.Printf("Time Until Expiry: %s\n", time.Until(token.ExpiresAt).Round(time.Second))

	// Verify token validity
	fmt.Printf("\nToken Valid: %v\n", client.IsTokenValid())
}
