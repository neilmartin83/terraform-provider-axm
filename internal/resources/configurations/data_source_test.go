// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configurations_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/configurations"
)

func TestConfigurationsDataSourceMetadata(t *testing.T) {
	ds := configurations.NewConfigurationsDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_configurations" {
		t.Errorf("expected TypeName %q, got %q", "axm_configurations", resp.TypeName)
	}
}

func TestConfigurationsDataSourceSchema(t *testing.T) {
	ds := configurations.NewConfigurationsDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	configsAttr, ok := resp.Schema.Attributes["configurations"]
	if !ok {
		t.Fatal("attribute 'configurations' not found")
	}
	listNested, ok := configsAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'configurations' to be a ListNestedAttribute")
	}

	nestedAttrs := listNested.NestedObject.Attributes
	expectedNested := []string{
		"id", "type", "name", "configuration_type",
		"configured_for_platforms", "created_date_time", "updated_date_time",
	}
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in configurations", name)
		}
	}

	platformsAttr, ok := nestedAttrs["configured_for_platforms"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'configured_for_platforms' to be a ListAttribute")
	}
	if platformsAttr.ElementType != types.StringType {
		t.Errorf("expected 'configured_for_platforms' ElementType to be StringType")
	}
}
