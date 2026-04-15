// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/blueprint"
)

func TestBlueprintDataSourceMetadata(t *testing.T) {
	ds := blueprint.NewBlueprintDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_blueprint" {
		t.Errorf("expected TypeName %q, got %q", "axm_blueprint", resp.TypeName)
	}
}

func TestBlueprintDataSourceSchema(t *testing.T) {
	ds := blueprint.NewBlueprintDataSource()
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

	appIDsAttr, ok := resp.Schema.Attributes["app_ids"].(dsschema.SetAttribute)
	if !ok {
		t.Fatal("expected 'app_ids' to be a SetAttribute")
	}
	if appIDsAttr.ElementType != types.StringType {
		t.Errorf("expected 'app_ids' ElementType to be StringType")
	}

	expectedAttrs := []string{
		"id", "timeouts", "name", "description", "status",
		"app_license_deficient", "created_date_time", "updated_date_time",
		"app_ids", "configuration_ids", "package_ids",
		"device_ids", "user_ids", "user_group_ids",
	}
	for _, name := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[name]; !ok {
			t.Errorf("attribute %q not found", name)
		}
	}
}
