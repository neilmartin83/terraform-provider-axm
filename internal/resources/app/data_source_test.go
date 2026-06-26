// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package app_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/app"
)

func TestAppDataSourceMetadata(t *testing.T) {
	ds := app.NewAppDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_app" {
		t.Errorf("expected TypeName %q, got %q", "axm_app", resp.TypeName)
	}
}

func TestAppDataSourceSchema(t *testing.T) {
	ds := app.NewAppDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("attribute 'id' not found")
	}
	if !idAttr.IsRequired() {
		t.Error("expected 'id' to be Required")
	}

	supportedOSAttr, ok := resp.Schema.Attributes["supported_os"].(dsschema.ListAttribute)
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

func TestAccAppDataSource(t *testing.T) {
	appID := os.Getenv("AXM_TEST_APP_ID")
	if appID == "" {
		t.Skip("AXM_TEST_APP_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`data "axm_app" "test" { id = %q }`, appID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_app.test", "id", appID),
					resource.TestCheckResourceAttrSet("data.axm_app.test", "name"),
					resource.TestCheckResourceAttrSet("data.axm_app.test", "bundle_id"),
				),
			},
		},
	})
}
