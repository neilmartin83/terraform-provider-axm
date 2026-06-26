// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprints_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/blueprints"
)

func TestBlueprintsDataSourceMetadata(t *testing.T) {
	ds := blueprints.NewBlueprintsDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_blueprints" {
		t.Errorf("expected TypeName %q, got %q", "axm_blueprints", resp.TypeName)
	}
}

func TestBlueprintsDataSourceSchema(t *testing.T) {
	ds := blueprints.NewBlueprintsDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("attribute 'id' not found")
	}
	if !idAttr.IsComputed() {
		t.Error("expected 'id' to be Computed")
	}

	if _, ok := resp.Schema.Attributes["timeouts"]; !ok {
		t.Fatal("attribute 'timeouts' not found")
	}

	blueprintsAttr, ok := resp.Schema.Attributes["blueprints"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'blueprints' to be a ListNestedAttribute")
	}
	if !blueprintsAttr.IsComputed() {
		t.Error("expected 'blueprints' to be Computed")
	}

	nestedAttrs := blueprintsAttr.NestedObject.Attributes
	allExpectedNested := []string{
		"id", "name", "description", "status",
		"app_license_deficient", "created_date_time", "updated_date_time",
	}
	for _, name := range allExpectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in blueprints", name)
		}
	}

	if _, ok := nestedAttrs["app_license_deficient"].(dsschema.BoolAttribute); !ok {
		t.Error("expected 'app_license_deficient' to be a BoolAttribute")
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

func TestAccBlueprintsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "axm_blueprints" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.axm_blueprints.all", "id"),
				),
			},
		},
	})
}
