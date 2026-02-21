// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var deviceManagementServiceTimeoutAttributeTypes = map[string]attr.Type{
	"create": types.StringType,
	"read":   types.StringType,
	"update": types.StringType,
}

// newDeviceManagementServiceTimeoutsNullValue returns a timeouts.Value with all attributes set to null.
func newDeviceManagementServiceTimeoutsNullValue() timeouts.Value {
	return ensureDeviceManagementServiceTimeouts(timeouts.Value{})
}

// ensureDeviceManagementServiceTimeouts ensures that the timeouts.Value is not null by
// initializing it with null attributes if necessary.
func ensureDeviceManagementServiceTimeouts(value timeouts.Value) timeouts.Value {
	if value.IsNull() && !value.IsUnknown() {
		value.Object = types.ObjectNull(deviceManagementServiceTimeoutAttributeTypes)
	}
	return value
}
