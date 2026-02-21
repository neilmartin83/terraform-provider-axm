// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service_serialnumbers_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
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

func TestAccDeviceManagementServiceSerialNumbersDataSource(t *testing.T) {
	serverID := os.Getenv("AXM_TEST_SERVER_ID")
	if serverID == "" {
		t.Skip("AXM_TEST_SERVER_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_device_management_service_serial_numbers" "test" {
						server_id = %q
					}
				`, serverID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_device_management_service_serial_numbers.test", "id", serverID),
					resource.TestCheckResourceAttrSet("data.axm_device_management_service_serial_numbers.test", "serial_numbers.#"),
				),
			},
		},
	})
}
