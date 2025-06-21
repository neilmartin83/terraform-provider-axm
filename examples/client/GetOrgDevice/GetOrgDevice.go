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

	device, err := client.GetOrgDevice(context.Background(), deviceID, nil)
	if err != nil {
		log.Fatalf("Error getting org device: %v", err)
	}

	fmt.Printf("Device Details:\n"+
		"ID: %s\n"+
		"Serial Number: %s\n"+
		"Model: %s\n"+
		"Product Family: %s\n"+
		"Product Type: %s\n"+
		"Status: %s\n"+
		"Added to Org: %s\n"+
		"Last Updated: %s\n"+
		"Device Capacity: %s\n"+
		"Part Number: %s\n"+
		"Order Number: %s\n"+
		"Color: %s\n",
		device.ID,
		device.Attributes.SerialNumber,
		device.Attributes.DeviceModel,
		device.Attributes.ProductFamily,
		device.Attributes.ProductType,
		device.Attributes.Status,
		device.Attributes.AddedToOrgDateTime,
		device.Attributes.UpdatedDateTime,
		device.Attributes.DeviceCapacity,
		device.Attributes.PartNumber,
		device.Attributes.OrderNumber,
		device.Attributes.Color,
	)

	// Print IMEI numbers if they exist
	if len(device.Attributes.IMEI) > 0 {
		fmt.Println("IMEI Numbers:")
		for _, imei := range device.Attributes.IMEI {
			fmt.Printf("- %s\n", imei)
		}
	}

	// Print MEID numbers if they exist
	if len(device.Attributes.MEID) > 0 {
		fmt.Println("MEID Numbers:")
		for _, meid := range device.Attributes.MEID {
			fmt.Printf("- %s\n", meid)
		}
	}

	// Print EID if it exists
	if device.Attributes.EID != "" {
		fmt.Printf("EID: %s\n", device.Attributes.EID)
	}
}
