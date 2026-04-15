// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package app_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/app"
)

func TestAppDataSourceMetadata(t *testing.T) {
	ds := app.NewAppDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_app" {
		t.Errorf("expected TypeName %q, got %q", "axm_app", resp.TypeName)
	}
}

func TestAppDataSourceSchema(t *testing.T) {
	ds := app.NewAppDataSource()
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

	supportedOSAttr, ok := resp.Schema.Attributes["supported_os"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'supported_os' to be a ListAttribute")
	}
	if supportedOSAttr.ElementType != types.StringType {
		t.Errorf("expected 'supported_os' ElementType to be StringType")
	}
}
