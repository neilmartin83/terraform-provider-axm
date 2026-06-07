// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package apple_device_management_devices_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/apple_device_management_devices"
)

func TestAppleDeviceManagementDevicesDataSourceMetadata(t *testing.T) {
	ds := apple_device_management_devices.NewAppleDeviceManagementDevicesDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_apple_device_management_devices" {
		t.Errorf("expected TypeName %q, got %q", "axm_apple_device_management_devices", resp.TypeName)
	}
}

func TestAppleDeviceManagementDevicesDataSourceSchema(t *testing.T) {
	ds := apple_device_management_devices.NewAppleDeviceManagementDevicesDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	devicesAttr, ok := resp.Schema.Attributes["devices"]
	if !ok {
		t.Fatal("expected 'devices' attribute")
	}
	listAttr, ok := devicesAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'devices' to be ListNestedAttribute")
	}

	expectedNestedAttrs := []string{
		"id", "type", "device_name", "enrolled_user_id", "product_family", "serial_number",
	}
	for _, name := range expectedNestedAttrs {
		attr, ok := listAttr.NestedObject.Attributes[name]
		if !ok {
			t.Errorf("nested attribute %q not found in devices", name)
			continue
		}
		if !attr.IsComputed() && name != "id" {
			t.Errorf("expected nested attribute %q to be Computed", name)
		}
	}
}

func TestAppleDeviceManagementDevicesDataSource_ListAttributes(t *testing.T) {
	ds := apple_device_management_devices.NewAppleDeviceManagementDevicesDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	devicesAttr, ok := resp.Schema.Attributes["devices"]
	if !ok {
		t.Fatal("expected 'devices' attribute")
	}
	listAttr, ok := devicesAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'devices' to be ListNestedAttribute")
	}

	for name, attr := range listAttr.NestedObject.Attributes {
		if _, ok := attr.(dsschema.StringAttribute); !ok {
			t.Errorf("expected nested attribute %q to be StringAttribute", name)
		}
	}
}
