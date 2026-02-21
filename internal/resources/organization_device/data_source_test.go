// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organization_device_test

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
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organization_device"
)

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"axm": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, envVar := range []string{"AXM_CLIENT_ID", "AXM_KEY_ID", "AXM_PRIVATE_KEY", "AXM_SCOPE"} {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

func TestOrganizationDeviceDataSourceMetadata(t *testing.T) {
	ds := organization_device.NewOrganizationDeviceDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_organization_device" {
		t.Errorf("expected TypeName %q, got %q", "axm_organization_device", resp.TypeName)
	}
}

func TestOrganizationDeviceDataSourceSchema(t *testing.T) {
	ds := organization_device.NewOrganizationDeviceDataSource()
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
		"type", "serial_number", "added_to_org_date_time", "released_from_org_date_time",
		"updated_date_time", "device_model", "product_family", "product_type",
		"device_capacity", "part_number", "order_number", "color", "status",
		"order_date_time", "eid", "purchase_source_id", "purchase_source_type",
		"wifi_mac_address", "bluetooth_mac_address",
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

	listAttrs := []string{"imei", "meid", "ethernet_mac_address"}
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

func TestAccOrganizationDeviceDataSource(t *testing.T) {
	deviceID := os.Getenv("AXM_TEST_DEVICE_ID")
	if deviceID == "" {
		t.Skip("AXM_TEST_DEVICE_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_organization_device" "test" {
						id = %q
					}
				`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_organization_device.test", "id", deviceID),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "serial_number"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "device_model"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "product_family"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "status"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "added_to_org_date_time"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "updated_date_time"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "color"),
				),
			},
		},
	})
}
