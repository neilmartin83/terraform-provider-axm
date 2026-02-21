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
		log.Fatal("Usage: GetOrgDeviceAssignedServerID <device-id>")
	}
	deviceID := os.Args[1]

	serverData, err := c.GetOrgDeviceAssignedServerID(context.Background(), deviceID)
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

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return val
}

func envOrDefault(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}
