// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/device_management_service"
)

func TestListResourceMetadata(t *testing.T) {
	lr := device_management_service.NewDeviceManagementServiceListResource()
	resp := tfresource.MetadataResponse{}
	lr.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_device_management_service" {
		t.Errorf("expected TypeName %q, got %q", "axm_device_management_service", resp.TypeName)
	}
}

func TestListResourceSchema(t *testing.T) {
	lr := device_management_service.NewDeviceManagementServiceListResource()
	resp := list.ListResourceSchemaResponse{}
	lr.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	tests := []struct {
		name         string
		optional     bool
		hasValidator bool
	}{
		{"server_type", true, true},
		{"name", true, false},
		{"name_contains", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, ok := resp.Schema.Attributes[tt.name]
			if !ok {
				t.Fatalf("attribute %q not found in schema", tt.name)
			}
			if !attr.IsOptional() {
				t.Errorf("expected attribute %q to be Optional", tt.name)
			}
			if attr.IsRequired() {
				t.Errorf("expected attribute %q to not be Required", tt.name)
			}

			if tt.hasValidator {
				stringAttr, ok := attr.(listschema.StringAttribute)
				if !ok {
					t.Fatalf("attribute %q is not a StringAttribute", tt.name)
				}
				if len(stringAttr.Validators) == 0 {
					t.Errorf("expected attribute %q to have at least one validator", tt.name)
				}
			}
		})
	}

	if len(resp.Schema.Attributes) != 3 {
		t.Errorf("expected 3 attributes, got %d", len(resp.Schema.Attributes))
	}
}
