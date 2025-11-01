package device_management_service

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// extractStrings converts a types.Set containing string values into a slice of strings,
// handling null and unknown values appropriately.
func extractStrings(set types.Set) []string {
	var result []string
	if set.IsNull() || set.IsUnknown() {
		return result
	}
	for _, v := range set.Elements() {
		if strVal, ok := v.(types.String); ok && !strVal.IsUnknown() && !strVal.IsNull() {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

// validateDevices checks all devices and returns a list of validation errors
func (r *DeviceManagementServiceResource) validateDevices(ctx context.Context, deviceIDs []string) []error {
	queryParams := url.Values{}
	queryParams.Add("fields[orgDevices]", "serialNumber")

	var errors []error
	for _, id := range deviceIDs {
		device, err := r.client.GetOrgDevice(ctx, id, queryParams)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to validate device %s: %s", id, err))
			continue
		}
		if device == nil {
			errors = append(errors, fmt.Errorf("device %s not found in Apple Business Manager", id))
		}
	}
	return errors
}

// waitForActivityCompletion polls the activity status until it completes, fails, or times out
func (r *DeviceManagementServiceResource) waitForActivityCompletion(ctx context.Context, activityID string) error {
	maxAttempts := 30
	retryInterval := 5 * time.Second

	queryParams := url.Values{}
	queryParams.Add("fields[orgDeviceActivities]", "status,subStatus")

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
		}

		activity, err := r.client.GetOrgDeviceActivity(ctx, activityID, queryParams)
		if err != nil {
			return fmt.Errorf("error checking activity status: %w", err)
		}

		switch activity.Attributes.Status {
		case "COMPLETED":
			return nil
		case "FAILED":
			return fmt.Errorf("activity failed with sub-status: %s", activity.Attributes.SubStatus)
		case "STOPPED":
			return fmt.Errorf("activity stopped with sub-status: %s", activity.Attributes.SubStatus)
		case "IN_PROGRESS", "PENDING":
			if attempt == maxAttempts {
				return fmt.Errorf("timed out waiting for activity to complete after %d attempts", maxAttempts)
			}
		default:
			return fmt.Errorf("unknown activity status: %s", activity.Attributes.Status)
		}
	}

	return fmt.Errorf("unexpected error monitoring activity status")
}
