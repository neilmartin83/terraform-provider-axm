// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/blueprint"
)

func TestBlueprintResourceMetadata(t *testing.T) {
	r := blueprint.NewBlueprintResource()
	resp := tfresource.MetadataResponse{}
	r.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_blueprint" {
		t.Errorf("expected TypeName %q, got %q", "axm_blueprint", resp.TypeName)
	}
}

func TestBlueprintResourceSchema(t *testing.T) {
	r := blueprint.NewBlueprintResource()
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
		{"description", false, true, false},
		{"status", false, false, true},
		{"app_license_deficient", false, false, true},
		{"created_date_time", false, false, true},
		{"updated_date_time", false, false, true},
		{"app_ids", false, true, true},
		{"configuration_ids", false, true, true},
		{"package_ids", false, true, true},
		{"device_ids", false, true, true},
		{"user_ids", false, true, true},
		{"user_group_ids", false, true, true},
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

	setAttrs := []string{
		"app_ids",
		"configuration_ids",
		"package_ids",
		"device_ids",
		"user_ids",
		"user_group_ids",
	}

	for _, name := range setAttrs {
		attr, ok := resp.Schema.Attributes[name].(resourceschema.SetAttribute)
		if !ok {
			t.Fatalf("attribute %q is not a SetAttribute", name)
		}
		if attr.ElementType != types.StringType {
			t.Errorf("expected %q ElementType to be StringType", name)
		}
	}
}

func TestBlueprintResourceIdentitySchema(t *testing.T) {
	r := blueprint.NewBlueprintResource()

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

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"axm": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}
	for _, envVar := range []string{"AXM_CLIENT_ID", "AXM_KEY_ID", "AXM_PRIVATE_KEY", "AXM_SCOPE"} {
		if os.Getenv(envVar) == "" {
			t.Skipf("%s must be set for acceptance tests", envVar)
		}
	}
}

func TestAccBlueprintResource_basic(t *testing.T) {
	testAccPreCheck(t)
	name := "tf-acc-test-blueprint-basic"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_blueprint" "test" {
						name        = %q
						description = "Test blueprint created by Terraform acceptance test"
					}
				`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("axm_blueprint.test", "id"),
					resource.TestCheckResourceAttr("axm_blueprint.test", "name", name),
					resource.TestCheckResourceAttr("axm_blueprint.test", "description", "Test blueprint created by Terraform acceptance test"),
					resource.TestCheckResourceAttrSet("axm_blueprint.test", "status"),
					resource.TestCheckResourceAttrSet("axm_blueprint.test", "created_date_time"),
					resource.TestCheckResourceAttrSet("axm_blueprint.test", "updated_date_time"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "axm_blueprint" "test" {
						name        = %q
						description = "Updated description"
					}
				`, name+"-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("axm_blueprint.test", "name", name+"-updated"),
					resource.TestCheckResourceAttr("axm_blueprint.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccBlueprintResource_import(t *testing.T) {
	testAccPreCheck(t)
	name := "tf-acc-test-blueprint-import"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_blueprint" "test" {
						name        = %q
						description = "Blueprint for import test"
					}
				`, name),
			},
			{
				ResourceName:            "axm_blueprint.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}
