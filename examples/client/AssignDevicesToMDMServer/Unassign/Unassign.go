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

	if len(os.Args) < 3 {
		log.Fatal("Usage: Unassign <server-id> <device-id> [device-id...]")
	}
	serverID := os.Args[1]
	deviceIDs := os.Args[2:]

	fmt.Println("Unassigning devices from MDM server...")
	activity, err := c.AssignDevicesToMDMServer(context.Background(), serverID, deviceIDs, false)
	if err != nil {
		log.Fatalf("Error unassigning devices: %v", err)
	}

	fmt.Printf("Unassignment completed successfully:\n"+
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
