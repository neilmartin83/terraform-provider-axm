// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/device_management_service"
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
			t.Skipf("%s must be set for acceptance tests", envVar)
		}
	}
}

func testAccResourcePreCheck(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	for _, envVar := range []string{"AXM_TEST_SERVER_ID", "AXM_TEST_DEVICE_SERIAL_1", "AXM_TEST_DEVICE_SERIAL_2", "AXM_TEST_DEVICE_SERIAL_3"} {
		if os.Getenv(envVar) == "" {
			t.Skipf("%s must be set for resource acceptance tests", envVar)
		}
	}
}

// testAccBaseURL returns the API base URL derived from the AXM_SCOPE env var.
func testAccBaseURL() string {
	scope := os.Getenv("AXM_SCOPE")
	if scope == "school.api" {
		return "https://api-school.apple.com"
	}
	return "https://api-business.apple.com"
}

// testAccNewClient creates a real API client for pre-test queries.
func testAccNewClient(t *testing.T) *client.Client {
	t.Helper()
	teamID := os.Getenv("AXM_TEAM_ID")
	if teamID == "" {
		teamID = os.Getenv("AXM_CLIENT_ID")
	}
	c, err := client.NewClient(
		testAccBaseURL(),
		teamID,
		os.Getenv("AXM_CLIENT_ID"),
		os.Getenv("AXM_KEY_ID"),
		os.Getenv("AXM_SCOPE"),
		os.Getenv("AXM_PRIVATE_KEY"),
	)
	if err != nil {
		t.Fatalf("failed to create API client: %v", err)
	}
	return c
}

// testAccGetExistingSerials queries the server for all currently assigned device
// serial numbers so the test config can represent the full source of truth.
func testAccGetExistingSerials(t *testing.T, serverID string) []string {
	t.Helper()
	c := testAccNewClient(t)
	serials, err := c.GetDeviceManagementServiceSerialNumbers(context.Background(), serverID)
	if err != nil {
		t.Fatalf("failed to query existing serials for server %s: %v", serverID, err)
	}
	return serials
}

// deviceIDsHCL builds the HCL set literal for device_ids, merging test serials
// with any pre-existing serials already on the server (deduplicating).
func deviceIDsHCL(existing []string, testSerials ...string) string {
	seen := make(map[string]bool)
	var all []string
	for _, s := range existing {
		if !seen[s] {
			seen[s] = true
			all = append(all, s)
		}
	}
	for _, s := range testSerials {
		if !seen[s] {
			seen[s] = true
			all = append(all, s)
		}
	}
	quoted := make([]string, len(all))
	for i, s := range all {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func TestResourceMetadata(t *testing.T) {
	r := device_management_service.NewDeviceManagementServiceResource()
	resp := tfresource.MetadataResponse{}
	r.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_device_management_service" {
		t.Errorf("expected TypeName %q, got %q", "axm_device_management_service", resp.TypeName)
	}
}

func TestResourceSchema(t *testing.T) {
	r := device_management_service.NewDeviceManagementServiceResource()
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
		{"id", false, true, true},
		{"name", false, false, true},
		{"type", false, false, true},
		{"device_ids", true, false, false},
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

	deviceIDsAttr, ok := resp.Schema.Attributes["device_ids"].(resourceschema.SetAttribute)
	if !ok {
		t.Fatal("device_ids is not a SetAttribute")
	}
	if deviceIDsAttr.ElementType != types.StringType {
		t.Errorf("expected device_ids ElementType to be StringType, got %T", deviceIDsAttr.ElementType)
	}
}

func TestResourceIdentitySchema(t *testing.T) {
	r := device_management_service.NewDeviceManagementServiceResource()

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

func TestAccDeviceManagementServiceResource_basic(t *testing.T) {
	testAccResourcePreCheck(t)
	serverID := os.Getenv("AXM_TEST_SERVER_ID")
	serial1 := os.Getenv("AXM_TEST_DEVICE_SERIAL_1")
	serial2 := os.Getenv("AXM_TEST_DEVICE_SERIAL_2")
	serial3 := os.Getenv("AXM_TEST_DEVICE_SERIAL_3")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				PreConfig: func() {},
				Config: func() string {
					existing := testAccGetExistingSerials(t, serverID)
					return fmt.Sprintf(`
						resource "axm_device_management_service" "test" {
							id         = %q
							device_ids = %s

							timeouts = {
								create = "5m"
								read   = "2m"
								update = "5m"
							}
						}
					`, serverID, deviceIDsHCL(existing, serial1, serial2))
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("axm_device_management_service.test", "id", serverID),
					resource.TestCheckResourceAttrSet("axm_device_management_service.test", "name"),
					resource.TestCheckResourceAttrSet("axm_device_management_service.test", "type"),
				),
			},
			{
				Config: func() string {
					existing := testAccGetExistingSerials(t, serverID)
					return fmt.Sprintf(`
						resource "axm_device_management_service" "test" {
							id         = %q
							device_ids = %s

							timeouts = {
								create = "5m"
								read   = "2m"
								update = "5m"
							}
						}
					`, serverID, deviceIDsHCL(existing, serial1, serial2, serial3))
				}(),
			},
			{
				Config: func() string {
					existing := testAccGetExistingSerials(t, serverID)
					return fmt.Sprintf(`
						resource "axm_device_management_service" "test" {
							id         = %q
							device_ids = %s

							timeouts = {
								create = "5m"
								read   = "2m"
								update = "5m"
							}
						}
					`, serverID, deviceIDsHCL(existing, serial1, serial3))
				}(),
			},
		},
	})
}

func TestAccDeviceManagementServiceResource_import(t *testing.T) {
	testAccResourcePreCheck(t)
	serverID := os.Getenv("AXM_TEST_SERVER_ID")
	serial1 := os.Getenv("AXM_TEST_DEVICE_SERIAL_1")
	serial2 := os.Getenv("AXM_TEST_DEVICE_SERIAL_2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccResourcePreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: func() string {
					existing := testAccGetExistingSerials(t, serverID)
					return fmt.Sprintf(`
						resource "axm_device_management_service" "test" {
							id         = %q
							device_ids = %s

							timeouts = {
								create = "5m"
								read   = "2m"
								update = "5m"
							}
						}
					`, serverID, deviceIDsHCL(existing, serial1, serial2))
				}(),
			},
			{
				ResourceName:            "axm_device_management_service.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}
