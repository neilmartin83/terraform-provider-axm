// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprints_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/blueprints"
)

func TestBlueprintsDataSourceMetadata(t *testing.T) {
	ds := blueprints.NewBlueprintsDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_blueprints" {
		t.Errorf("expected TypeName %q, got %q", "axm_blueprints", resp.TypeName)
	}
}

func TestBlueprintsDataSourceSchema(t *testing.T) {
	ds := blueprints.NewBlueprintsDataSource()
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

	if _, ok := resp.Schema.Attributes["timeouts"]; !ok {
		t.Fatal("attribute 'timeouts' not found")
	}

	blueprintsAttr, ok := resp.Schema.Attributes["blueprints"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'blueprints' to be a ListNestedAttribute")
	}
	if !blueprintsAttr.IsComputed() {
		t.Error("expected 'blueprints' to be Computed")
	}

	nestedAttrs := blueprintsAttr.NestedObject.Attributes
	allExpectedNested := []string{
		"id", "name", "description", "status",
		"app_license_deficient", "created_date_time", "updated_date_time",
	}
	for _, name := range allExpectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in blueprints", name)
		}
	}

	if _, ok := nestedAttrs["app_license_deficient"].(dsschema.BoolAttribute); !ok {
		t.Error("expected 'app_license_deficient' to be a BoolAttribute")
	}
}
