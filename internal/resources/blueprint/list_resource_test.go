// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/blueprint"
)

func TestBlueprintListResourceMetadata(t *testing.T) {
	lr := blueprint.NewBlueprintListResource()
	resp := tfresource.MetadataResponse{}
	lr.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_blueprint" {
		t.Errorf("expected TypeName %q, got %q", "axm_blueprint", resp.TypeName)
	}
}

func TestBlueprintListResourceSchema(t *testing.T) {
	lr := blueprint.NewBlueprintListResource()
	resp := list.ListResourceSchemaResponse{}
	lr.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	tests := []string{"name", "name_contains", "status"}
	for _, name := range tests {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q not found in schema", name)
		}
		if !attr.IsOptional() {
			t.Errorf("expected attribute %q to be Optional", name)
		}
		if attr.IsRequired() {
			t.Errorf("expected attribute %q to not be Required", name)
		}
	}

	if len(resp.Schema.Attributes) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(resp.Schema.Attributes))
	}
}
