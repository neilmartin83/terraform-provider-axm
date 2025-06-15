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

	servers, err := client.GetDeviceManagementServices(context.Background())
	if err != nil {
		log.Fatalf("Error getting MDM servers: %v", err)
	}

	fmt.Printf("Found %d MDM server(s)\n\n", len(servers))

	for _, server := range servers {
		fmt.Printf("MDM Server Details:\n"+
			"ID: %s\n"+
			"Type: %s\n"+
			"Name: %s\n"+
			"Server Type: %s\n"+
			"Created: %s\n"+
			"Last Updated: %s\n\n",
			server.ID,
			server.Type,
			server.Attributes.ServerName,
			server.Attributes.ServerType,
			server.Attributes.CreatedDateTime,
			server.Attributes.UpdatedDateTime,
		)

		// If you want to show the number of devices assigned to this server
		if server.Relationships.Devices.Meta.Paging.Total > 0 {
			fmt.Printf("Total Assigned Devices: %d\n\n",
				server.Relationships.Devices.Meta.Paging.Total)
		}
	}
}
