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

	// Example activity ID - replace with an actual activity ID
	activityID := "123e4567-e89b-12d3-a456-426614174000"

	if teamID == "" || clientID == "" || keyID == "" || privateKey == "" || baseURL == "" {
		log.Fatal("Missing required environment variables: AXM_TEAM_ID, AXM_CLIENT_ID, AXM_KEY_ID, AXM_PRIVATE_KEY, AXM_BASE_URL")
	}

	client, err := axm.NewClient(baseURL, teamID, clientID, keyID, scope, privateKey)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	activity, err := client.GetOrgDeviceActivity(context.Background(), activityID)
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

	// Print additional details based on status
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

	// Print all available timestamps
	fmt.Printf("\nTimeline:\n"+
		"Created: %s\n",
		activity.Attributes.CreatedDateTime)

	if activity.Attributes.CompletedDateTime != "" {
		fmt.Printf("Completed: %s\n", activity.Attributes.CompletedDateTime)
	}

	// Print any additional metadata
	if activity.Attributes.SubStatus != "" {
		fmt.Printf("\nAdditional Information:\n"+
			"Sub-Status: %s\n",
			activity.Attributes.SubStatus)
	}

	// Status-specific messages
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
