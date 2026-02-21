// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organization_devices_test

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
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organization_devices"
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

func TestOrganizationDevicesDataSourceMetadata(t *testing.T) {
	ds := organization_devices.NewOrganizationDevicesDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_organization_devices" {
		t.Errorf("expected TypeName %q, got %q", "axm_organization_devices", resp.TypeName)
	}
}

func TestOrganizationDevicesDataSourceSchema(t *testing.T) {
	ds := organization_devices.NewOrganizationDevicesDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("attribute 'id' not found")
	}
	if !idAttr.IsComputed() {
		t.Error("expected 'id' to be Computed")
	}

	devicesAttr, ok := resp.Schema.Attributes["devices"]
	if !ok {
		t.Fatal("attribute 'devices' not found")
	}
	listNested, ok := devicesAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'devices' to be a ListNestedAttribute")
	}
	if !devicesAttr.IsComputed() {
		t.Error("expected 'devices' to be Computed")
	}

	nestedAttrs := listNested.NestedObject.Attributes
	expectedNested := []string{
		"id", "type", "serial_number", "added_to_org_date_time",
		"updated_date_time", "device_model", "product_family", "status",
		"imei", "meid", "ethernet_mac_address",
	}
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in devices", name)
		}
	}
}

func TestAccOrganizationDevicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "axm_organization_devices" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.axm_organization_devices.all", "id"),
					resource.TestCheckResourceAttrSet("data.axm_organization_devices.all", "devices.#"),
				),
			},
		},
	})
}
