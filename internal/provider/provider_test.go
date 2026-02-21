// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"context"
	"os"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
)

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"axm": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, envVar := range []string{"AXM_CLIENT_ID", "AXM_KEY_ID", "AXM_PRIVATE_KEY", "AXM_SCOPE"} {
		if os.Getenv(envVar) == "" {
			t.Fatalf("%s must be set for acceptance tests", envVar)
		}
	}
}

func TestProviderMetadata(t *testing.T) {
	p := provider.New("1.2.3")()
	resp := tfprovider.MetadataResponse{}
	p.Metadata(context.Background(), tfprovider.MetadataRequest{}, &resp)

	if resp.TypeName != "axm" {
		t.Errorf("expected TypeName %q, got %q", "axm", resp.TypeName)
	}
	if resp.Version != "1.2.3" {
		t.Errorf("expected Version %q, got %q", "1.2.3", resp.Version)
	}
}

func TestProviderSchema(t *testing.T) {
	p := provider.New("test")()
	resp := tfprovider.SchemaResponse{}
	p.Schema(context.Background(), tfprovider.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	tests := []struct {
		name      string
		sensitive bool
	}{
		{"team_id", false},
		{"client_id", false},
		{"key_id", false},
		{"private_key", true},
		{"scope", false},
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
			if attr.IsComputed() {
				t.Errorf("expected attribute %q to not be Computed", tt.name)
			}
			if attr.IsSensitive() != tt.sensitive {
				t.Errorf("expected attribute %q Sensitive=%v, got %v", tt.name, tt.sensitive, attr.IsSensitive())
			}
		})
	}
}

func TestProviderSchema_ScopeHasValidator(t *testing.T) {
	p := provider.New("test")()
	resp := tfprovider.SchemaResponse{}
	p.Schema(context.Background(), tfprovider.SchemaRequest{}, &resp)

	attr, ok := resp.Schema.Attributes["scope"]
	if !ok {
		t.Fatal("scope attribute not found")
	}

	stringAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("scope attribute is not a StringAttribute")
	}
	if len(stringAttr.Validators) == 0 {
		t.Error("expected scope attribute to have at least one validator")
	}
}

func TestProviderResources(t *testing.T) {
	p := provider.New("test")()
	ctx := context.Background()
	resources := p.Resources(ctx)

	if len(resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(resources))
	}

	r := resources[0]()
	resp := tfresource.MetadataResponse{}
	r.Metadata(ctx, tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_device_management_service" {
		t.Errorf("expected TypeName %q, got %q", "axm_device_management_service", resp.TypeName)
	}
}

func TestProviderDataSources(t *testing.T) {
	p := provider.New("test")()
	ctx := context.Background()
	dataSources := p.DataSources(ctx)

	if len(dataSources) != 6 {
		t.Fatalf("expected 6 data sources, got %d", len(dataSources))
	}

	expected := []string{
		"axm_device_management_service_serial_numbers",
		"axm_device_management_services",
		"axm_organization_device",
		"axm_organization_device_applecare_coverage",
		"axm_organization_device_assigned_server_information",
		"axm_organization_devices",
	}

	var got []string
	for _, factory := range dataSources {
		ds := factory()
		resp := datasource.MetadataResponse{}
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)
		got = append(got, resp.TypeName)
	}

	sort.Strings(got)
	if len(got) != len(expected) {
		t.Fatalf("expected %d data sources, got %d", len(expected), len(got))
	}
	for i := range expected {
		if got[i] != expected[i] {
			t.Errorf("data source[%d]: expected %q, got %q", i, expected[i], got[i])
		}
	}
}

func TestProviderListResources(t *testing.T) {
	p := provider.New("test")()
	ctx := context.Background()

	plr, ok := p.(tfprovider.ProviderWithListResources)
	if !ok {
		t.Fatal("provider does not implement ProviderWithListResources")
	}

	listResources := plr.ListResources(ctx)
	if len(listResources) != 1 {
		t.Fatalf("expected 1 list resource, got %d", len(listResources))
	}

	lr := listResources[0]()
	resp := tfresource.MetadataResponse{}
	lr.Metadata(ctx, tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_device_management_service" {
		t.Errorf("expected TypeName %q, got %q", "axm_device_management_service", resp.TypeName)
	}
}

func TestAccProviderConfiguresSuccessfully(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
					data "axm_device_management_services" "test" {}
				`,
				Check: resource.TestCheckResourceAttrSet("data.axm_device_management_services.test", "id"),
			},
		},
	})
}
