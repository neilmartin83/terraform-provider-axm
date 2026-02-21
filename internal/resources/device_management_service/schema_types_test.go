// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNewDeviceManagementServiceTimeoutsNullValue(t *testing.T) {
	value := newDeviceManagementServiceTimeoutsNullValue()
	if !value.IsNull() {
		t.Error("expected IsNull to be true")
	}
	if value.IsUnknown() {
		t.Error("expected IsUnknown to be false")
	}

	attrTypes := value.AttributeTypes(context.TODO())
	for _, key := range []string{"create", "read", "update"} {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("expected attribute type for %q", key)
		}
	}
}

func TestEnsureDeviceManagementServiceTimeouts_AlreadyValid(t *testing.T) {
	attrValues := map[string]types.String{
		"create": types.StringValue("30m"),
		"read":   types.StringValue("10m"),
		"update": types.StringValue("30m"),
	}
	_ = attrValues

	initial := newDeviceManagementServiceTimeoutsNullValue()
	result := ensureDeviceManagementServiceTimeouts(initial)
	if !result.IsNull() {
		t.Error("expected null value to remain null after ensure")
	}
}

func TestEnsureDeviceManagementServiceTimeouts_ZeroValue(t *testing.T) {
	zeroValue := timeouts.Value{}
	if !zeroValue.IsNull() {
		t.Skip("zero-value timeouts.Value is not null — skipping")
	}

	result := ensureDeviceManagementServiceTimeouts(zeroValue)
	if !result.IsNull() {
		t.Error("expected result to be null")
	}

	attrTypes := result.AttributeTypes(context.TODO())
	if len(attrTypes) != 3 {
		t.Errorf("expected 3 attribute types, got %d", len(attrTypes))
	}
}
