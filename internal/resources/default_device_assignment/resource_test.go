// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package default_device_assignment_test

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sort"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
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

func TestAccDefaultDeviceAssignmentResource_basic(t *testing.T) {
	testAccPreCheck(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `resource "axm_default_device_assignment" "this" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("axm_default_device_assignment.this", "id", "default"),
				),
			},
		},
	})
}

func testAccBaseURL() string {
	if os.Getenv("AXM_SCOPE") == "school.api" {
		return "https://api-school.apple.com"
	}
	return "https://api-business.apple.com"
}

// testAccNewClient creates a real API client for verifying live server state.
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

func testAccAssignmentPreCheck(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	if os.Getenv("AXM_SCOPE") != "business.api" {
		t.Skip("default device assignment requires the business.api scope")
	}
	if os.Getenv("AXM_TEST_DEVICE_MANAGEMENT_SERVICE_CERTIFICATE") == "" {
		t.Skip("AXM_TEST_DEVICE_MANAGEMENT_SERVICE_CERTIFICATE must be set to create device management services")
	}
}

// testAccCheckLiveFamilies asserts the live API reports exactly the wanted default
// product families for the server tracked by the named resource.
func testAccCheckLiveFamilies(t *testing.T, resourceName string, want ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		srv, err := testAccNewClient(t).GetDeviceManagementService(context.Background(), rs.Primary.ID, nil)
		if err != nil {
			return fmt.Errorf("GET %s: %w", rs.Primary.ID, err)
		}
		got := make([]string, 0, len(srv.Attributes.DefaultProductFamilies))
		for _, f := range srv.Attributes.DefaultProductFamilies {
			got = append(got, string(f))
		}
		sort.Strings(got)
		wantSorted := slices.Clone(want)
		sort.Strings(wantSorted)
		if !slices.Equal(got, wantSorted) {
			return fmt.Errorf("%s live defaultProductFamilies = %v, want %v", resourceName, got, wantSorted)
		}
		return nil
	}
}

// TestAccDefaultDeviceAssignmentResource_moveAndClear covers the PATCH bodies that
// an empty defaultProductFamilies slice must produce: moving a family between
// servers clears the source, and unassigning clears the last remaining family.
func TestAccDefaultDeviceAssignmentResource_moveAndClear(t *testing.T) {
	testAccAssignmentPreCheck(t)

	cert := os.Getenv("AXM_TEST_DEVICE_MANAGEMENT_SERVICE_CERTIFICATE")
	suffix := time.Now().UnixNano()

	config := func(macTarget string) string {
		return fmt.Sprintf(`
			resource "axm_device_management_service" "a" {
				name = "tf-acc-dda-a-%[1]d"
				server_certificate = {
					name = "tf-acc-dda-a-%[1]d-cert"
					data = %[2]q
				}
			}

			resource "axm_device_management_service" "b" {
				name = "tf-acc-dda-b-%[1]d"
				server_certificate = {
					name = "tf-acc-dda-b-%[1]d-cert"
					data = %[2]q
				}
			}

			resource "axm_default_device_assignment" "this" {
				mac = %[3]s
			}
		`, suffix, cert, macTarget)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccAssignmentPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config("axm_device_management_service.a.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"axm_default_device_assignment.this", "mac",
						"axm_device_management_service.a", "id"),
					testAccCheckLiveFamilies(t, "axm_device_management_service.a", "MAC"),
					testAccCheckLiveFamilies(t, "axm_device_management_service.b"),
				),
			},
			{
				Config: config("axm_device_management_service.b.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"axm_default_device_assignment.this", "mac",
						"axm_device_management_service.b", "id"),
					testAccCheckLiveFamilies(t, "axm_device_management_service.a"),
					testAccCheckLiveFamilies(t, "axm_device_management_service.b", "MAC"),
				),
			},
			{
				Config: config(`""`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLiveFamilies(t, "axm_device_management_service.a"),
					testAccCheckLiveFamilies(t, "axm_device_management_service.b"),
				),
			},
		},
	})
}
