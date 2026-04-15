// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
