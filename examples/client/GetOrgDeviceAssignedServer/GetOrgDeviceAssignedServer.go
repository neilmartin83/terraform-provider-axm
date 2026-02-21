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
		log.Fatal("Usage: GetOrgDeviceAssignedServer <device-id>")
	}
	deviceID := os.Args[1]

	server, err := c.GetOrgDeviceAssignedServer(context.Background(), deviceID, nil)
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

	if server.Relationships.Devices.Meta.Paging.Total > 0 {
		fmt.Printf("\nTotal Devices Assigned to Server: %d\n",
			server.Relationships.Devices.Meta.Paging.Total)
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
