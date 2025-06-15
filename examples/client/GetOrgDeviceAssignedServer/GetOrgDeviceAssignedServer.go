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

	// Example device ID - replace with an actual device ID
	deviceID := "FAKE12345678"

	if teamID == "" || clientID == "" || keyID == "" || privateKey == "" || baseURL == "" {
		log.Fatal("Missing required environment variables: AXM_TEAM_ID, AXM_CLIENT_ID, AXM_KEY_ID, AXM_PRIVATE_KEY, AXM_BASE_URL")
	}

	client, err := axm.NewClient(baseURL, teamID, clientID, keyID, scope, privateKey)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	server, err := client.GetOrgDeviceAssignedServer(context.Background(), deviceID)
	if err != nil {
		log.Fatalf("Error getting assigned server: %v", err)
	}

	fmt.Printf("Device %s Assignment Details:\n\n", deviceID)

	fmt.Printf("Server Information:\n"+
		"ID: %s\n"+
		"Type: %s\n"+
		"Name: %s\n"+
		"Server Type: %s\n"+
		"Created: %s\n"+
		"Last Updated: %s\n",
		server.ID,
		server.Type,
		server.Attributes.ServerName,
		server.Attributes.ServerType,
		server.Attributes.CreatedDateTime,
		server.Attributes.UpdatedDateTime,
	)

	// If the server has device relationships
	if server.Relationships.Devices.Meta.Paging.Total > 0 {
		fmt.Printf("\nTotal Devices Assigned to Server: %d\n",
			server.Relationships.Devices.Meta.Paging.Total)
	}

	// Optional: Get the device details to show alongside server info
	device, err := client.GetOrgDevice(context.Background(), deviceID)
	if err != nil {
		log.Printf("Error getting device details: %v", err)
	} else {
		fmt.Printf("\nDevice Details:\n"+
			"Serial Number: %s\n"+
			"Model: %s\n"+
			"Product Family: %s\n"+
			"Status: %s\n"+
			"Added to Org: %s\n",
			device.Attributes.SerialNumber,
			device.Attributes.DeviceModel,
			device.Attributes.ProductFamily,
			device.Attributes.Status,
			device.Attributes.AddedToOrgDateTime,
		)
	}
}
