// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/device_management_service"
)

func TestDeviceManagementServiceDataSourceMetadata(t *testing.T) {
	ds := device_management_service.NewDeviceManagementServiceDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_device_management_service" {
		t.Errorf("expected TypeName %q, got %q", "axm_device_management_service", resp.TypeName)
	}
}

func TestDeviceManagementServiceDataSourceSchema(t *testing.T) {
	ds := device_management_service.NewDeviceManagementServiceDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("attribute 'id' not found")
	}
	if !idAttr.IsRequired() {
		t.Error("expected 'id' to be Required")
	}

	computed := []string{"type", "server_name", "server_type", "status", "device_count", "default_product_families", "last_connected_date_time", "last_connected_ip", "allow_release", "created_date_time", "updated_date_time"}
	for _, name := range computed {
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

func TestAccDeviceManagementServiceDataSource(t *testing.T) {
	serverID := os.Getenv("AXM_TEST_SERVER_ID")
	if serverID == "" {
		t.Skip("AXM_TEST_SERVER_ID must be set for acceptance tests")
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccResourcePreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "axm_device_management_service" "test" {
					id = "` + serverID + `"
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_device_management_service.test", "id", serverID),
					resource.TestCheckResourceAttrSet("data.axm_device_management_service.test", "server_name"),
					resource.TestCheckResourceAttrSet("data.axm_device_management_service.test", "server_type"),
					resource.TestCheckResourceAttrSet("data.axm_device_management_service.test", "created_date_time"),
					resource.TestCheckResourceAttrSet("data.axm_device_management_service.test", "updated_date_time"),
				),
			},
		},
	})
}
