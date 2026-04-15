// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration_test

import (
	"context"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/configuration"
)

func TestConfigurationResourceMetadata(t *testing.T) {
	r := configuration.NewConfigurationResource()
	resp := tfresource.MetadataResponse{}
	r.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_configuration" {
		t.Errorf("expected TypeName %q, got %q", "axm_configuration", resp.TypeName)
	}
}

func TestConfigurationResourceSchema(t *testing.T) {
	r := configuration.NewConfigurationResource()
	resp := tfresource.SchemaResponse{}
	r.Schema(context.Background(), tfresource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	tests := []struct {
		name     string
		required bool
		optional bool
		computed bool
	}{
		{"id", false, false, true},
		{"name", true, false, false},
		{"type", false, false, true},
		{"configured_for_platforms", false, true, true},
		{"configuration_profile", false, true, true},
		{"filename", false, true, true},
		{"created_date_time", false, false, true},
		{"updated_date_time", false, false, true},
		{"timeouts", false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, ok := resp.Schema.Attributes[tt.name]
			if !ok {
				t.Fatalf("attribute %q not found in schema", tt.name)
			}
			if attr.IsRequired() != tt.required {
				t.Errorf("expected Required=%v, got %v", tt.required, attr.IsRequired())
			}
			if attr.IsOptional() != tt.optional {
				t.Errorf("expected Optional=%v, got %v", tt.optional, attr.IsOptional())
			}
			if attr.IsComputed() != tt.computed {
				t.Errorf("expected Computed=%v, got %v", tt.computed, attr.IsComputed())
			}
		})
	}

	configuredAttr, ok := resp.Schema.Attributes["configured_for_platforms"].(resourceschema.SetAttribute)
	if !ok {
		t.Fatal("configured_for_platforms is not a SetAttribute")
	}
	if configuredAttr.ElementType != types.StringType {
		t.Errorf("expected configured_for_platforms ElementType to be StringType")
	}

	configProfileAttr, ok := resp.Schema.Attributes["configuration_profile"].(resourceschema.StringAttribute)
	if !ok {
		t.Fatal("configuration_profile is not a StringAttribute")
	}
	if configProfileAttr.Sensitive {
		t.Error("expected configuration_profile to not be Sensitive")
	}
}

func TestConfigurationResourceIdentitySchema(t *testing.T) {
	r := configuration.NewConfigurationResource()

	ri, ok := r.(tfresource.ResourceWithIdentity)
	if !ok {
		t.Fatal("resource does not implement ResourceWithIdentity")
	}

	resp := tfresource.IdentitySchemaResponse{}
	ri.IdentitySchema(context.Background(), tfresource.IdentitySchemaRequest{}, &resp)

	idAttr, ok := resp.IdentitySchema.Attributes["id"]
	if !ok {
		t.Fatal("identity schema missing 'id' attribute")
	}

	idIdentityAttr, ok := idAttr.(identityschema.StringAttribute)
	if !ok {
		t.Fatal("identity 'id' attribute is not a StringAttribute")
	}
	if !idIdentityAttr.RequiredForImport {
		t.Error("expected identity 'id' to have RequiredForImport=true")
	}
}
