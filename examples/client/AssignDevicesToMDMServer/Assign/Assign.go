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

	if teamID == "" || clientID == "" || keyID == "" || privateKey == "" || baseURL == "" {
		log.Fatal("Missing required environment variables: AXM_TEAM_ID, AXM_CLIENT_ID, AXM_KEY_ID, AXM_PRIVATE_KEY, AXM_BASE_URL")
	}

	client, err := client.NewClient(baseURL, teamID, clientID, keyID, scope, privateKey)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	serverID := "FAKE98765432TEST98765432DUMMY0000"
	deviceIDs := []string{"FAKE12345678"}

	fmt.Println("Assigning devices to MDM server...")
	activity, err := client.AssignDevicesToMDMServer(context.Background(), serverID, deviceIDs, true)
	if err != nil {
		log.Fatalf("Error assigning devices: %v", err)
	}

	fmt.Printf("Assignment completed successfully:\n"+
		"  ID: %s\n"+
		"  Status: %s\n"+
		"  SubStatus: %s\n"+
		"  Created: %s\n",
		activity.ID,
		activity.Attributes.Status,
		activity.Attributes.SubStatus,
		activity.Attributes.CreatedDateTime,
	)

	if activity.Attributes.DownloadURL != "" {
		fmt.Printf("Results available at: %s\n", activity.Attributes.DownloadURL)
	}
}
