// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var blueprintTimeoutAttributeTypes = map[string]attr.Type{
	"create": types.StringType,
	"read":   types.StringType,
	"update": types.StringType,
}

// newBlueprintTimeoutsNullValue returns a timeouts.Value with all attributes set to null.
func newBlueprintTimeoutsNullValue() timeouts.Value {
	return ensureBlueprintTimeouts(timeouts.Value{})
}

// ensureBlueprintTimeouts ensures that the timeouts.Value is initialized with null attributes if needed.
func ensureBlueprintTimeouts(value timeouts.Value) timeouts.Value {
	if value.IsNull() && !value.IsUnknown() {
		value.Object = types.ObjectNull(blueprintTimeoutAttributeTypes)
	}
	return value
}
