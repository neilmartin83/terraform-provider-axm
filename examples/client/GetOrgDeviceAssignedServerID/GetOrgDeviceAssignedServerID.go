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

	// Example device ID - replace with an actual device ID
	deviceID := "FAKE12345678"

	if teamID == "" || clientID == "" || keyID == "" || privateKey == "" || baseURL == "" {
		log.Fatal("Missing required environment variables: AXM_TEAM_ID, AXM_CLIENT_ID, AXM_KEY_ID, AXM_PRIVATE_KEY, AXM_BASE_URL")
	}

	client, err := client.NewClient(baseURL, teamID, clientID, keyID, scope, privateKey)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	serverData, err := client.GetOrgDeviceAssignedServerID(context.Background(), deviceID)
	if err != nil {
		log.Fatalf("Error getting assigned server ID: %v", err)
	}

	fmt.Printf("Device Assignment Details:\n"+
		"Device ID: %s\n"+
		"Server ID: %s\n"+
		"Resource Type: %s\n",
		deviceID,
		serverData.ID,
		serverData.Type,
	)
}
