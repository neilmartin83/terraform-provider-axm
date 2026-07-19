// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organizational_units_test

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
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organizational_units"
)

func TestOrganizationalUnitsDataSourceMetadata(t *testing.T) {
	ds := organizational_units.NewOrganizationalUnitsDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_organizational_units" {
		t.Errorf("expected TypeName %q, got %q", "axm_organizational_units", resp.TypeName)
	}
}

func TestOrganizationalUnitsDataSourceSchema(t *testing.T) {
	ds := organizational_units.NewOrganizationalUnitsDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	ouAttr, ok := resp.Schema.Attributes["organizational_units"]
	if !ok {
		t.Fatal("attribute 'organizational_units' not found")
	}
	listNested, ok := ouAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'organizational_units' to be a ListNestedAttribute")
	}

	expectedNested := []string{
		"id", "type", "name", "description", "created_date_time", "updated_date_time",
	}
	nestedAttrs := listNested.NestedObject.Attributes
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in organizational_units", name)
		}
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

func TestAccOrganizationalUnitsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "axm_organizational_units" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.axm_organizational_units.all", "id"),
					resource.TestCheckResourceAttrSet("data.axm_organizational_units.all", "organizational_units.#"),
				),
			},
		},
	})
}
