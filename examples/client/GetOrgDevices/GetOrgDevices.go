package main

import (
	"context"
	"fmt"
	"log"

	axm "github.com/neilmartin83/terraform-provider-axm/internal/provider"
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

	if teamID == "" || clientID == "" || keyID == "" || privateKey == "" || baseURL == "" {
		log.Fatal("Missing required environment variables: AXM_TEAM_ID, AXM_CLIENT_ID, AXM_KEY_ID, AXM_PRIVATE_KEY, AXM_BASE_URL")
	}

	client, err := axm.NewClient(baseURL, teamID, clientID, keyID, scope, privateKey)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	devices, err := client.GetOrgDevices(context.Background())
	if err != nil {
		log.Fatalf("Error getting org devices: %v", err)
	}

	for _, d := range devices {
		fmt.Printf("Device: ID=%s Serial=%s Model=%s ProductFamily=%s ProductType=%s Status=%s\n",
			d.ID,
			d.Attributes.SerialNumber,
			d.Attributes.DeviceModel,
			d.Attributes.ProductFamily,
			d.Attributes.ProductType,
			d.Attributes.Status,
		)
	}
}
