// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package apple_device_management_device_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/apple_device_management_device"
)

func TestAppleDeviceManagementDeviceDataSourceMetadata(t *testing.T) {
	ds := apple_device_management_device.NewAppleDeviceManagementDeviceDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_apple_device_management_device" {
		t.Errorf("expected TypeName %q, got %q", "axm_apple_device_management_device", resp.TypeName)
	}
}

func TestAppleDeviceManagementDeviceDataSourceSchema(t *testing.T) {
	ds := apple_device_management_device.NewAppleDeviceManagementDeviceDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	requiredAttrs := []string{"id"}
	for _, name := range requiredAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found", name)
			continue
		}
		if !attr.IsRequired() {
			t.Errorf("expected attribute %q to be Required", name)
		}
	}

	computedAttrs := []string{
		"type", "bluetooth_mac_address", "device_erase_status", "device_lock_status",
		"device_model", "device_name", "ethernet_mac_address", "is_filevault_enabled",
		"is_firewall_enabled", "last_check_in_date_time", "lost_mode_status",
		"os_version", "platform", "serial_number", "storage_free_capacity",
		"storage_total_capacity", "wifi_mac_address",
	}
	for _, name := range computedAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("expected attribute %q to be Computed", name)
		}
	}

	listAttrs := []string{"imei", "meid"}
	for _, name := range listAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found", name)
			continue
		}
		listAttr, ok := attr.(dsschema.ListAttribute)
		if !ok {
			t.Errorf("expected attribute %q to be a ListAttribute", name)
			continue
		}
		if listAttr.ElementType != types.StringType {
			t.Errorf("expected attribute %q ElementType to be StringType", name)
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

func TestAccAppleDeviceManagementDeviceDataSource(t *testing.T) {
	deviceID := os.Getenv("AXM_TEST_ADM_DEVICE_ID")
	if deviceID == "" {
		t.Skip("AXM_TEST_ADM_DEVICE_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "axm_apple_device_management_device" "test" { id = %q }`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_apple_device_management_device.test", "id", deviceID),
					resource.TestCheckResourceAttrSet("data.axm_apple_device_management_device.test", "serial_number"),
					resource.TestCheckResourceAttrSet("data.axm_apple_device_management_device.test", "device_model"),
				),
			},
		},
	})
}
