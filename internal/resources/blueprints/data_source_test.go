// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprints_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

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

	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Fatal("attribute 'id' not found")
	}
	if _, ok := resp.Schema.Attributes["timeouts"]; !ok {
		t.Fatal("attribute 'timeouts' not found")
	}
	if _, ok := resp.Schema.Attributes["blueprints"]; !ok {
		t.Fatal("attribute 'blueprints' not found")
	}
}
