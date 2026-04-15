// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package packageinfo_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/package"
)

func TestPackageDataSourceMetadata(t *testing.T) {
	ds := packageinfo.NewPackageDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_package" {
		t.Errorf("expected TypeName %q, got %q", "axm_package", resp.TypeName)
	}
}

func TestPackageDataSourceSchema(t *testing.T) {
	ds := packageinfo.NewPackageDataSource()
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

	bundleAttr, ok := resp.Schema.Attributes["bundle_ids"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'bundle_ids' to be a ListAttribute")
	}
	if bundleAttr.ElementType != types.StringType {
		t.Errorf("expected 'bundle_ids' ElementType to be StringType")
	}
}
