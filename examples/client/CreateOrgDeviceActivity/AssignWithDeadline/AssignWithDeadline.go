package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
		log.Fatal("Usage: AssignWithDeadline <server-id> <device-id> [device-id...]")
	}
	serverID := os.Args[1]
	deviceIDs := os.Args[2:]

	deadline := time.Now().Add(48 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z")
	fmt.Printf("Assigning devices to MDM server %s with migration deadline %s...\n", serverID, deadline)
	activity, err := c.CreateOrgDeviceActivity(
		context.Background(),
		client.OrgDeviceActivityAssignWithMDMMigrationDeadline,
		deviceIDs,
		client.WithMdmServer(serverID),
		client.WithMigrationDeadline(deadline),
	)
	if err != nil {
		log.Fatalf("Error assigning devices with deadline: %v", err)
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
