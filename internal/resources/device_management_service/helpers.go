package device_management_service

import (
	"context"
	"fmt"
	"net/url"

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
