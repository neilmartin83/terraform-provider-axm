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
		log.Fatal("Usage: GetDeviceManagementServiceSerialNumbers <server-id>")
	}
	serverID := os.Args[1]

	serialNumbers, err := c.GetDeviceManagementServiceSerialNumbers(context.Background(), serverID)
	if err != nil {
		log.Fatalf("Error getting device serial numbers: %v", err)
	}

	fmt.Printf("Found %d device(s) assigned to MDM server %s\n\n", len(serialNumbers), serverID)

	for i, serialNumber := range serialNumbers {
		fmt.Printf("%d. %s\n", i+1, serialNumber)
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
