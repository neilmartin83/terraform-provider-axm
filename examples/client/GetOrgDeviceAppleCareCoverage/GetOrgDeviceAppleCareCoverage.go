package main

import (
	"context"
	"fmt"
	"log"
	"os"

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
		log.Fatalf("Failed to initialize client: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatal("Usage: GetOrgDeviceAppleCareCoverage <device-id>")
	}
	deviceID := os.Args[1]

	coverages, err := c.GetOrgDeviceAppleCareCoverage(context.Background(), deviceID, nil)
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
