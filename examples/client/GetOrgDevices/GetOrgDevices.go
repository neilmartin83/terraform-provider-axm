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

	devices, err := c.GetOrgDevices(context.Background(), nil)
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
