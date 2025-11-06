package main

import (
	"context"
	"fmt"
	"log"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func main() {
	teamID := "BUSINESSAPI.123e4567-e89b-12d3-a456-426614174000"
	clientID := "BUSINESSAPI.123e4567-e89b-12d3-a456-426614174000"
	keyID := "123e4567-e89b-12d3-a456-426614174000"
	privateKey := `-----BEGIN EC PRIVATE KEY-----
FAKEAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wTESTBAQQgZxY8ytVhyXPLdHlj
TESTx9TSUTcFK29+lHvA1DybmFAKEyhRANCAAQXv+VXUiVv511AIa4nEXBrTESTD+
FAKEFigCMU45fN5v94OvEUUV2eUR3t4UZpZ4tHbCNdzEyXNIbFAKEY2xAc
-----END EC PRIVATE KEY-----`
	baseURL := "https://api-business.apple.com"
	scope := "business.api"

	deviceID := "FAKE12345678" // Replace with actual device ID

	if teamID == "" || clientID == "" || keyID == "" || privateKey == "" || baseURL == "" || deviceID == "" {
		log.Fatal("Missing required environment variables or deviceID")
	}

	client, err := client.NewClient(baseURL, teamID, clientID, keyID, scope, privateKey)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	coverages, err := client.GetOrgDeviceAppleCareCoverage(context.Background(), deviceID, nil)
	if err != nil {
		log.Fatalf("Error getting AppleCare coverage: %v", err)
	}

	if len(coverages) == 0 {
		fmt.Println("No AppleCare coverage found for this device")
		return
	}

	for _, coverage := range coverages {
		fmt.Printf("Coverage: ID=%s Type=%s\n", coverage.ID, coverage.Type)
		fmt.Printf("  Status: %s\n", coverage.Attributes.Status)
		fmt.Printf("  Payment Type: %s\n", coverage.Attributes.PaymentType)
		fmt.Printf("  Description: %s\n", coverage.Attributes.Description)
		fmt.Printf("  Start Date: %s\n", coverage.Attributes.StartDateTime)
		fmt.Printf("  End Date: %s\n", coverage.Attributes.EndDateTime)
		fmt.Printf("  Is Renewable: %t\n", coverage.Attributes.IsRenewable)
		fmt.Printf("  Is Canceled: %t\n", coverage.Attributes.IsCanceled)
		if coverage.Attributes.ContractCancelDateTime != "" {
			fmt.Printf("  Cancel Date: %s\n", coverage.Attributes.ContractCancelDateTime)
		}
		if coverage.Attributes.AgreementNumber != "" {
			fmt.Printf("  Agreement Number: %s\n", coverage.Attributes.AgreementNumber)
		}
		fmt.Println()
	}
}
