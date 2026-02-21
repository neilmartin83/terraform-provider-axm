// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organization_device_applecare_coverage_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organization_device_applecare_coverage"
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

func TestOrganizationDeviceAppleCareCoverageDataSourceMetadata(t *testing.T) {
	ds := organization_device_applecare_coverage.NewOrganizationDeviceAppleCareCoverageDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_organization_device_applecare_coverage" {
		t.Errorf("expected TypeName %q, got %q", "axm_organization_device_applecare_coverage", resp.TypeName)
	}
}

func TestOrganizationDeviceAppleCareCoverageDataSourceSchema(t *testing.T) {
	ds := organization_device_applecare_coverage.NewOrganizationDeviceAppleCareCoverageDataSource()
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

	coverageAttr, ok := resp.Schema.Attributes["applecare_coverage_resources"]
	if !ok {
		t.Fatal("attribute 'applecare_coverage_resources' not found")
	}
	listNested, ok := coverageAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'applecare_coverage_resources' to be a ListNestedAttribute")
	}
	if !coverageAttr.IsComputed() {
		t.Error("expected 'applecare_coverage_resources' to be Computed")
	}

	nestedAttrs := listNested.NestedObject.Attributes
	expectedNested := []string{
		"id", "agreement_number", "contract_cancel_date_time", "description",
		"end_date_time", "is_canceled", "is_renewable", "payment_type",
		"start_date_time", "status",
	}
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in applecare_coverage_resources", name)
		}
	}

	for _, boolAttr := range []string{"is_canceled", "is_renewable"} {
		attr, ok := nestedAttrs[boolAttr]
		if !ok {
			continue
		}
		if _, ok := attr.(dsschema.BoolAttribute); !ok {
			t.Errorf("expected nested attribute %q to be a BoolAttribute", boolAttr)
		}
	}
}

func TestAccOrganizationDeviceAppleCareCoverageDataSource(t *testing.T) {
	deviceID := os.Getenv("AXM_TEST_DEVICE_WITH_COVERAGE_ID")
	if deviceID == "" {
		t.Skip("AXM_TEST_DEVICE_WITH_COVERAGE_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_organization_device_applecare_coverage" "test" {
						id = %q
					}
				`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_organization_device_applecare_coverage.test", "id", deviceID),
					resource.TestCheckResourceAttrSet("data.axm_organization_device_applecare_coverage.test", "applecare_coverage_resources.#"),
				),
			},
		},
	})
}
