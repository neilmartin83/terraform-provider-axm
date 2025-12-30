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

func newDeviceManagementServiceTimeoutsNullValue() timeouts.Value {
	return ensureDeviceManagementServiceTimeouts(timeouts.Value{})
}

func ensureDeviceManagementServiceTimeouts(value timeouts.Value) timeouts.Value {
	if value.Object.IsNull() && !value.Object.IsUnknown() {
		value.Object = types.ObjectNull(deviceManagementServiceTimeoutAttributeTypes)
	}
	return value
}
