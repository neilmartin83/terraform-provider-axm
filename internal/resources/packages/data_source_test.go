// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package packages_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/packages"
)

func TestPackagesDataSourceMetadata(t *testing.T) {
	ds := packages.NewPackagesDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_packages" {
		t.Errorf("expected TypeName %q, got %q", "axm_packages", resp.TypeName)
	}
}

func TestPackagesDataSourceSchema(t *testing.T) {
	ds := packages.NewPackagesDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	packagesAttr, ok := resp.Schema.Attributes["packages"]
	if !ok {
		t.Fatal("attribute 'packages' not found")
	}
	listNested, ok := packagesAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'packages' to be a ListNestedAttribute")
	}

	nestedAttrs := listNested.NestedObject.Attributes
	expectedNested := []string{
		"id", "type", "name", "url", "hash", "bundle_ids",
		"description", "version", "created_date_time", "updated_date_time",
	}
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in packages", name)
		}
	}

	bundleAttr, ok := nestedAttrs["bundle_ids"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'bundle_ids' to be a ListAttribute")
	}
	if bundleAttr.ElementType != types.StringType {
		t.Errorf("expected 'bundle_ids' ElementType to be StringType")
	}
}
