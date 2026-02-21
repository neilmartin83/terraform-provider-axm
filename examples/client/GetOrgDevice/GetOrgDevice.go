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
		log.Fatal("Usage: GetOrgDevice <device-id>")
	}
	deviceID := os.Args[1]

	device, err := c.GetOrgDevice(context.Background(), deviceID, nil)
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

	if len(device.Attributes.IMEI) > 0 {
		fmt.Println("IMEI Numbers:")
		for _, imei := range device.Attributes.IMEI {
			fmt.Printf("- %s\n", imei)
		}
	}

	if len(device.Attributes.MEID) > 0 {
		fmt.Println("MEID Numbers:")
		for _, meid := range device.Attributes.MEID {
			fmt.Printf("- %s\n", meid)
		}
	}

	if device.Attributes.EID != "" {
		fmt.Printf("EID: %s\n", device.Attributes.EID)
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
