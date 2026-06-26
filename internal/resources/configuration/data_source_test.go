// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/configuration"
)

func TestConfigurationDataSourceMetadata(t *testing.T) {
	ds := configuration.NewConfigurationDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_configuration" {
		t.Errorf("expected TypeName %q, got %q", "axm_configuration", resp.TypeName)
	}
}

func TestConfigurationDataSourceSchema(t *testing.T) {
	ds := configuration.NewConfigurationDataSource()
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

	platformsAttr, ok := resp.Schema.Attributes["configured_for_platforms"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'configured_for_platforms' to be a ListAttribute")
	}
	if platformsAttr.ElementType != types.StringType {
		t.Errorf("expected 'configured_for_platforms' ElementType to be StringType")
	}

	expectedAttrs := []string{
		"id", "type", "name", "configuration_type",
		"configured_for_platforms", "configuration_profile", "filename",
		"created_date_time", "updated_date_time", "timeouts",
	}
	for _, name := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("attribute %q not found", name)
		}
	}
}

func TestAccConfigurationDataSource(t *testing.T) {
	configurationID := os.Getenv("AXM_TEST_CONFIGURATION_ID")
	if configurationID == "" {
		t.Skip("AXM_TEST_CONFIGURATION_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "axm_configuration" "test" { id = %q }`, configurationID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_configuration.test", "id", configurationID),
					resource.TestCheckResourceAttrSet("data.axm_configuration.test", "name"),
					resource.TestCheckResourceAttrSet("data.axm_configuration.test", "type"),
				),
			},
		},
	})
}
