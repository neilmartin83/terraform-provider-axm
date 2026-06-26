// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package apps_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/apps"
)

func TestAppsDataSourceMetadata(t *testing.T) {
	ds := apps.NewAppsDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_apps" {
		t.Errorf("expected TypeName %q, got %q", "axm_apps", resp.TypeName)
	}
}

func TestAppsDataSourceSchema(t *testing.T) {
	ds := apps.NewAppsDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	appsAttr, ok := resp.Schema.Attributes["apps"]
	if !ok {
		t.Fatal("attribute 'apps' not found")
	}
	listNested, ok := appsAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'apps' to be a ListNestedAttribute")
	}

	nestedAttrs := listNested.NestedObject.Attributes
	expectedNested := []string{
		"id", "type", "name", "bundle_id", "website_url", "version",
		"supported_os", "is_custom_app", "app_store_url",
	}
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in apps", name)
		}
	}

	supportedOSAttr, ok := nestedAttrs["supported_os"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'supported_os' to be a ListAttribute")
	}
	if supportedOSAttr.ElementType != types.StringType {
		t.Errorf("expected 'supported_os' ElementType to be StringType")
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

func TestAccAppsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "axm_apps" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.axm_apps.all", "id"),
				),
			},
		},
	})
}
