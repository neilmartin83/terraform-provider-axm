package device_management_service

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// extractStrings converts a types.List containing string values into a slice of strings,
// handling null and unknown values appropriately.
func extractStrings(list types.List) []string {
	var result []string
	if list.IsNull() || list.IsUnknown() {
		return result
	}
	for _, v := range list.Elements() {
		if strVal, ok := v.(types.String); ok && !strVal.IsUnknown() && !strVal.IsNull() {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

// setsEqual compares two sets represented as maps and returns true if they contain
// exactly the same elements, false otherwise.
func setsEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, exists := b[k]; !exists {
			return false
		}
	}
	return true
}

// validateDevices checks all devices and returns a list of validation errors
func (r *deviceManagementServiceResource) validateDevices(ctx context.Context, deviceIDs []string) []error {
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
