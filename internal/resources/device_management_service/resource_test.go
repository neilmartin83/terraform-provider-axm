// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

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
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}
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

// testAccMigrationSerialPreCheck skips unless API creds and the
// AXM_TEST_DEVICE_SERIAL_1 test device are present. The migration tests create
// a fresh server, so AXM_TEST_SERVER_ID is not required.
func testAccMigrationSerialPreCheck(t *testing.T) {
	t.Helper()
	testAccPreCheck(t)
	if os.Getenv("AXM_TEST_DEVICE_SERIAL_1") == "" {
		t.Skip("AXM_TEST_DEVICE_SERIAL_1 must be set for migration acceptance tests")
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
		{"id", false, false, true},
		{"name", true, false, false},
		{"type", false, false, true},
		{"status", false, false, true},
		{"device_count", false, false, true},
		{"default_product_families", false, false, true},
		{"last_connected_date_time", false, false, true},
		{"last_connected_ip", false, false, true},
		{"created_date_time", false, false, true},
		{"updated_date_time", false, false, true},
		{"allow_release", false, true, true},
		{"device_ids", false, true, true},
		{"migrated_devices", false, true, false},
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

	serverCertAttr, ok := resp.Schema.Attributes["server_certificate"].(resourceschema.SingleNestedAttribute)
	if !ok {
		t.Fatal("server_certificate is not a SingleNestedAttribute")
	}
	if !serverCertAttr.IsOptional() {
		t.Error("expected server_certificate to be Optional")
	}
	for _, name := range []string{"name", "data"} {
		if _, ok := serverCertAttr.Attributes[name]; !ok {
			t.Errorf("nested attribute %q not found in server_certificate", name)
		}
	}
	dataAttr, ok := serverCertAttr.Attributes["data"].(resourceschema.StringAttribute)
	if !ok {
		t.Fatal("server_certificate.data is not a StringAttribute")
	}
	if !dataAttr.Sensitive {
		t.Error("expected server_certificate.data to be Sensitive")
	}

	migratedAttr, ok := resp.Schema.Attributes["migrated_devices"].(resourceschema.SetNestedAttribute)
	if !ok {
		t.Fatal("migrated_devices is not a SetNestedAttribute")
	}
	if !migratedAttr.IsOptional() {
		t.Error("expected migrated_devices to be Optional")
	}
	for _, name := range []string{"id", "migration_deadline"} {
		if _, ok := migratedAttr.NestedObject.Attributes[name]; !ok {
			t.Errorf("nested attribute %q not found in migrated_devices", name)
		}
	}
	if idAttr, ok := migratedAttr.NestedObject.Attributes["id"].(resourceschema.StringAttribute); !ok || !idAttr.IsRequired() {
		t.Error("expected migrated_devices.id to be a Required StringAttribute")
	}
	if deadlineAttr, ok := migratedAttr.NestedObject.Attributes["migration_deadline"].(resourceschema.StringAttribute); !ok || !deadlineAttr.IsRequired() {
		t.Error("expected migrated_devices.migration_deadline to be a Required StringAttribute")
	} else if len(deadlineAttr.Validators) == 0 {
		t.Error("expected migrated_devices.migration_deadline to have at least one validator")
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

// testAccServerCertificateData returns the base64-encoded PEM certificate used
// when creating a new device management service.
func testAccServerCertificateData(t *testing.T) string {
	t.Helper()
	pubKey := os.Getenv("AXM_TEST_SERVER_PUBLIC_KEY")
	if pubKey == "" {
		t.Skip("AXM_TEST_SERVER_PUBLIC_KEY must be set to create a device management service")
	}
	return base64.StdEncoding.EncodeToString([]byte(pubKey))
}

// testAccIsMdmMigrationCapable reports whether the live API marks the device as
// eligible for MDM migration.
func testAccIsMdmMigrationCapable(t *testing.T, c *client.Client, serial string) bool {
	t.Helper()
	device, err := c.GetOrgDevice(context.Background(), serial, nil)
	if err != nil {
		t.Fatalf("failed to query device %s: %v", serial, err)
	}
	return device.Attributes.IsMdmMigrationCapable
}

// testAccMigrationSerial reports whether the AXM_TEST_DEVICE_SERIAL_1 test
// device is eligible for MDM migration, logging the serial it checked.
func testAccMigrationSerial(t *testing.T) bool {
	t.Helper()
	testAccMigrationSerialPreCheck(t)
	serial := os.Getenv("AXM_TEST_DEVICE_SERIAL_1")
	c := testAccNewClient(t)
	capable := testAccIsMdmMigrationCapable(t, c, serial)
	t.Logf("device %s isMdmMigrationCapable=%t", serial, capable)
	return capable
}

// migratedDeviceHCL builds the HCL object literal for a single migrated_devices
// entry.
func migratedDeviceHCL(serial, deadline string) string {
	return fmt.Sprintf("{ id = %q, migration_deadline = %q }", serial, deadline)
}

func TestAccDeviceManagementServiceResource_migration(t *testing.T) {
	testAccMigrationSerialPreCheck(t)
	serial := os.Getenv("AXM_TEST_DEVICE_SERIAL_1")
	if !testAccMigrationSerial(t) {
		t.Skip("AXM_TEST_DEVICE_SERIAL_1 is not eligible for MDM migration; a device must be enrolled in another MDM server to become eligible")
	}

	deadline := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	name := fmt.Sprintf("tf-acc-migration-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMigrationSerialPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_device_management_service" "test" {
						name = %q

						server_certificate = {
							name = "PublicKey.pem"
							data = %q
						}

						device_ids = [
							%q,
						]

						migrated_devices = [
							%s,
						]

						timeouts = {
							create = "15m"
							update = "15m"
							delete = "15m"
						}
					}
				`,
					name,
					testAccServerCertificateData(t),
					serial,
					migratedDeviceHCL(serial, deadline),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("axm_device_management_service.test", "id"),
					resource.TestCheckResourceAttr("axm_device_management_service.test", "name", name),
					resource.TestCheckResourceAttr("axm_device_management_service.test", "device_ids.#", "1"),
					resource.TestCheckResourceAttr("axm_device_management_service.test", "migrated_devices.#", "1"),
				),
			},
		},
	})
}

func TestAccDeviceManagementServiceResource_migrationNotCapable(t *testing.T) {
	testAccMigrationSerialPreCheck(t)
	serial := os.Getenv("AXM_TEST_DEVICE_SERIAL_1")
	if testAccMigrationSerial(t) {
		t.Skip("AXM_TEST_DEVICE_SERIAL_1 supports MDM migration; negative test not applicable")
	}

	deadline := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	name := fmt.Sprintf("tf-acc-migration-lock-%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccMigrationSerialPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_device_management_service" "test" {
						name = %q

						server_certificate = {
							name = "PublicKey.pem"
							data = %q
						}

						device_ids = [
							%q,
						]

						migrated_devices = [
							%s,
						]
					}
				`,
					name,
					testAccServerCertificateData(t),
					serial,
					migratedDeviceHCL(serial, deadline),
				),
				ExpectError: regexp.MustCompile(`not eligible for MDM migration`),
			},
		},
	})
}
