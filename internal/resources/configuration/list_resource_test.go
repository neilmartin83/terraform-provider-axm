// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/configuration"
)

func TestConfigurationListResourceMetadata(t *testing.T) {
	lr := configuration.NewConfigurationListResource()
	resp := tfresource.MetadataResponse{}
	lr.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_configuration" {
		t.Errorf("expected TypeName %q, got %q", "axm_configuration", resp.TypeName)
	}
}

func TestConfigurationListResourceSchema(t *testing.T) {
	lr := configuration.NewConfigurationListResource()
	resp := list.ListResourceSchemaResponse{}
	lr.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	tests := []string{"name", "name_contains"}
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

	if len(resp.Schema.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(resp.Schema.Attributes))
	}
}
