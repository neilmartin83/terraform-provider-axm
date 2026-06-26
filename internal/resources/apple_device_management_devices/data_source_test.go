// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package apple_device_management_devices_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
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

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"axm": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}
	for _, envVar := range []string{"AXM_CLIENT_ID", "AXM_KEY_ID", "AXM_PRIVATE_KEY", "AXM_SCOPE"} {
		if os.Getenv(envVar) == "" {
			t.Skipf("%s must be set for acceptance tests", envVar)
		}
	}
}

func TestAccAppleDeviceManagementDevicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "axm_apple_device_management_devices" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.axm_apple_device_management_devices.all", "id"),
				),
			},
		},
	})
}
