// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service_serialnumbers_test

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
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/device_management_service_serialnumbers"
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

func TestDeviceManagementServiceSerialNumbersDataSourceMetadata(t *testing.T) {
	ds := device_management_service_serialnumbers.NewDeviceManagementServiceSerialNumbersDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_device_management_service_serial_numbers" {
		t.Errorf("expected TypeName %q, got %q", "axm_device_management_service_serial_numbers", resp.TypeName)
	}
}

func TestDeviceManagementServiceSerialNumbersDataSourceSchema(t *testing.T) {
	ds := device_management_service_serialnumbers.NewDeviceManagementServiceSerialNumbersDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	serverIDAttr, ok := resp.Schema.Attributes["server_id"]
	if !ok {
		t.Fatal("attribute 'server_id' not found")
	}
	if !serverIDAttr.IsRequired() {
		t.Error("expected 'server_id' to be Required")
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("attribute 'id' not found")
	}
	if !idAttr.IsComputed() {
		t.Error("expected 'id' to be Computed")
	}

	snAttr, ok := resp.Schema.Attributes["serial_numbers"]
	if !ok {
		t.Fatal("attribute 'serial_numbers' not found")
	}
	listAttr, ok := snAttr.(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'serial_numbers' to be a ListAttribute")
	}
	if listAttr.ElementType != types.StringType {
		t.Error("expected 'serial_numbers' ElementType to be StringType")
	}
	if !snAttr.IsComputed() {
		t.Error("expected 'serial_numbers' to be Computed")
	}
}

func TestAccDeviceManagementServiceSerialNumbersDataSource(t *testing.T) {
	serverID := os.Getenv("AXM_TEST_SERVER_ID")
	if serverID == "" {
		t.Skip("AXM_TEST_SERVER_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_device_management_service_serial_numbers" "test" {
						server_id = %q
					}
				`, serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_device_management_service_serial_numbers.test", "id", serverID),
					resource.TestCheckResourceAttrSet("data.axm_device_management_service_serial_numbers.test", "serial_numbers.#"),
				),
			},
		},
	})
}
