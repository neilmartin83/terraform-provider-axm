// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organization_device_assigned_server_information_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organization_device_assigned_server_information"
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

func TestOrganizationDeviceAssignedServerInformationDataSourceMetadata(t *testing.T) {
	ds := organization_device_assigned_server_information.NewOrganizationDeviceAssignedServerInformationDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_organization_device_assigned_server_information" {
		t.Errorf("expected TypeName %q, got %q", "axm_organization_device_assigned_server_information", resp.TypeName)
	}
}

func TestOrganizationDeviceAssignedServerInformationDataSourceSchema(t *testing.T) {
	ds := organization_device_assigned_server_information.NewOrganizationDeviceAssignedServerInformationDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	deviceIDAttr, ok := resp.Schema.Attributes["device_id"]
	if !ok {
		t.Fatal("attribute 'device_id' not found")
	}
	if !deviceIDAttr.IsRequired() {
		t.Error("expected 'device_id' to be Required")
	}

	computedAttrs := []string{"id", "server_id", "server_name", "server_type", "created_date_time", "updated_date_time"}
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
}

func TestAccOrganizationDeviceAssignedServerInformationDataSource(t *testing.T) {
	deviceID := os.Getenv("AXM_TEST_ASSIGNED_DEVICE_ID")
	if deviceID == "" {
		t.Skip("AXM_TEST_ASSIGNED_DEVICE_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_organization_device_assigned_server_information" "test" {
						device_id = %q
					}
				`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_organization_device_assigned_server_information.test", "id", deviceID),
					resource.TestCheckResourceAttrSet("data.axm_organization_device_assigned_server_information.test", "server_id"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device_assigned_server_information.test", "server_name"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device_assigned_server_information.test", "server_type"),
				),
			},
		},
	})
}
