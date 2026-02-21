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
		log.Fatal("Usage: GetOrgDeviceActivity <activity-id>")
	}
	activityID := os.Args[1]

	activity, err := c.GetOrgDeviceActivity(context.Background(), activityID, nil)
	if err != nil {
		log.Fatalf("Error getting activity status: %v", err)
	}

	fmt.Printf("Activity Details:\n"+
		"ID: %s\n"+
		"Type: %s\n"+
		"Status: %s\n",
		activity.ID,
		activity.Type,
		activity.Attributes.Status,
	)

	switch activity.Attributes.Status {
	case "COMPLETED":
		fmt.Printf("Completed At: %s\n", activity.Attributes.CompletedDateTime)
		if activity.Attributes.DownloadURL != "" {
			fmt.Printf("Download URL: %s\n", activity.Attributes.DownloadURL)
		}
	case "FAILED", "STOPPED":
		fmt.Printf("Sub-Status: %s\n", activity.Attributes.SubStatus)
	case "IN_PROGRESS":
		fmt.Printf("Started At: %s\n", activity.Attributes.CreatedDateTime)
	}

	fmt.Printf("\nTimeline:\n"+
		"Created: %s\n",
		activity.Attributes.CreatedDateTime)

	if activity.Attributes.CompletedDateTime != "" {
		fmt.Printf("Completed: %s\n", activity.Attributes.CompletedDateTime)
	}

	if activity.Attributes.SubStatus != "" {
		fmt.Printf("\nAdditional Information:\n"+
			"Sub-Status: %s\n",
			activity.Attributes.SubStatus)
	}

	fmt.Printf("\nStatus Summary: ")
	switch activity.Attributes.Status {
	case "COMPLETED":
		fmt.Println("The activity has completed successfully.")
	case "FAILED":
		fmt.Printf("The activity failed. Reason: %s\n", activity.Attributes.SubStatus)
	case "STOPPED":
		fmt.Printf("The activity was stopped. Reason: %s\n", activity.Attributes.SubStatus)
	case "IN_PROGRESS":
		fmt.Println("The activity is currently in progress.")
	default:
		fmt.Printf("The activity is in an unknown state: %s\n", activity.Attributes.Status)
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
