// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var configurationTimeoutAttributeTypes = map[string]attr.Type{
	"create": types.StringType,
	"read":   types.StringType,
	"update": types.StringType,
}

// newConfigurationTimeoutsNullValue returns a timeouts.Value with all attributes set to null.
func newConfigurationTimeoutsNullValue() timeouts.Value {
	return ensureConfigurationTimeouts(timeouts.Value{})
}

// ensureConfigurationTimeouts ensures that the timeouts.Value is initialized with null attributes if needed.
func ensureConfigurationTimeouts(value timeouts.Value) timeouts.Value {
	if value.IsNull() && !value.IsUnknown() {
		value.Object = types.ObjectNull(configurationTimeoutAttributeTypes)
	}
	return value
}
