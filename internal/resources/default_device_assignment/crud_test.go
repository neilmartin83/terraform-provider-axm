// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package default_device_assignment

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFamilyServerMap(t *testing.T) {
	data := DefaultDeviceAssignmentModel{
		AppleTV: types.StringValue("srv-1"),
		IPad:    types.StringValue("srv-2"),
		Mac:     types.StringValue("srv-1"),
	}

	m := familyServerMap(data)

	if m["APPLE_TV"] != "srv-1" {
		t.Errorf("expected APPLE_TV -> srv-1, got %q", m["APPLE_TV"])
	}
	if m["IPAD"] != "srv-2" {
		t.Errorf("expected IPAD -> srv-2, got %q", m["IPAD"])
	}
	if m["MAC"] != "srv-1" {
		t.Errorf("expected MAC -> srv-1, got %q", m["MAC"])
	}

	nullKeys := []string{"IPHONE", "IPOD", "VISION", "WATCH"}
	for _, k := range nullKeys {
		if _, ok := m[k]; ok {
			t.Errorf("expected %q to be omitted for null values", k)
		}
	}
}

func TestFamilyServerMap_EmptyString(t *testing.T) {
	data := DefaultDeviceAssignmentModel{
		AppleTV: types.StringValue(""),
		IPad:    types.StringValue("srv-2"),
	}

	m := familyServerMap(data)

	if m["APPLE_TV"] != "" {
		t.Errorf("expected APPLE_TV -> \"\", got %q", m["APPLE_TV"])
	}
}

func TestUniqueServerIDs(t *testing.T) {
	data := DefaultDeviceAssignmentModel{
		AppleTV: types.StringValue("srv-1"),
		IPad:    types.StringValue("srv-2"),
		Mac:     types.StringValue("srv-1"),
	}

	ids := uniqueServerIDs(data)

	if len(ids) != 2 {
		t.Fatalf("expected 2 unique IDs, got %d", len(ids))
	}
	if _, ok := ids["srv-1"]; !ok {
		t.Error("expected srv-1 in unique IDs")
	}
	if _, ok := ids["srv-2"]; !ok {
		t.Error("expected srv-2 in unique IDs")
	}
}

func TestUniqueServerIDs_SkipsNullUnknownEmpty(t *testing.T) {
	data := DefaultDeviceAssignmentModel{
		AppleTV: types.StringNull(),
		IPad:    types.StringUnknown(),
		Mac:     types.StringValue(""),
	}

	ids := uniqueServerIDs(data)

	if len(ids) != 0 {
		t.Errorf("expected 0 unique IDs, got %d", len(ids))
	}
}

func TestReconcileFamily(t *testing.T) {
	current := map[string]string{
		"IPAD": "srv-1",
		"MAC":  "srv-2",
	}

	t.Run("null_field", func(t *testing.T) {
		result := reconcileFamily(types.StringNull(), "IPAD", current)
		if !result.IsNull() {
			t.Error("expected null result for null input")
		}
	})

	t.Run("unknown_field", func(t *testing.T) {
		result := reconcileFamily(types.StringUnknown(), "IPAD", current)
		if !result.IsUnknown() {
			t.Error("expected unknown result for unknown input")
		}
	})

	t.Run("empty_sentinel", func(t *testing.T) {
		result := reconcileFamily(types.StringValue(""), "IPAD", current)
		if result.ValueString() != "" {
			t.Errorf("expected empty string, got %q", result.ValueString())
		}
	})

	t.Run("server_still_holds_family", func(t *testing.T) {
		result := reconcileFamily(types.StringValue("srv-1"), "IPAD", current)
		if result.ValueString() != "srv-1" {
			t.Errorf("expected srv-1, got %q", result.ValueString())
		}
	})

	t.Run("server_no_longer_holds_family", func(t *testing.T) {
		result := reconcileFamily(types.StringValue("srv-1"), "MAC", current)
		if !result.IsNull() {
			t.Errorf("expected null result when server no longer holds family, got %q", result.ValueString())
		}
	})

	t.Run("family_not_in_current", func(t *testing.T) {
		result := reconcileFamily(types.StringValue("srv-3"), "IPHONE", current)
		if !result.IsNull() {
			t.Errorf("expected null result when family not in current map")
		}
	})
}
