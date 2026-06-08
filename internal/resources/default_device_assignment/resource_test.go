// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package default_device_assignment_test

import (
	"context"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/default_device_assignment"
)

func TestDefaultDeviceAssignmentResourceMetadata(t *testing.T) {
	r := default_device_assignment.NewDefaultDeviceAssignmentResource()
	resp := tfresource.MetadataResponse{}
	r.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_default_device_assignment" {
		t.Errorf("expected TypeName %q, got %q", "axm_default_device_assignment", resp.TypeName)
	}
}

func TestDefaultDeviceAssignmentResourceSchema(t *testing.T) {
	r := default_device_assignment.NewDefaultDeviceAssignmentResource()
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
		{"apple_tv", false, true, false},
		{"apple_vision_pro", false, true, false},
		{"ipad", false, true, false},
		{"iphone", false, true, false},
		{"ipod", false, true, false},
		{"mac", false, true, false},
		{"watch", false, true, false},
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

	if len(resp.Schema.Attributes) != 8 {
		t.Errorf("expected 8 attributes, got %d", len(resp.Schema.Attributes))
	}
}

func TestDefaultDeviceAssignmentResourceDoesNotImplementIdentity(t *testing.T) {
	r := default_device_assignment.NewDefaultDeviceAssignmentResource()
	_, ok := r.(tfresource.ResourceWithIdentity)
	if ok {
		t.Error("expected resource to NOT implement ResourceWithIdentity")
	}
}
