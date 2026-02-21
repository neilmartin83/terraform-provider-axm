// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organization_device_assigned_server_information_test

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

func TestAccOrganizationDeviceAssignedServerInformationDataSource(t *testing.T) {
	deviceID := os.Getenv("AXM_TEST_ASSIGNED_DEVICE_ID")
	if deviceID == "" {
		t.Skip("AXM_TEST_ASSIGNED_DEVICE_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_organization_device_assigned_server_information" "test" {
						device_id = %q
					}
				`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_organization_device_assigned_server_information.test", "id", deviceID),
					resource.TestCheckResourceAttrSet("data.axm_organization_device_assigned_server_information.test", "server_id"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device_assigned_server_information.test", "server_name"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device_assigned_server_information.test", "server_type"),
				),
			},
		},
	})
}
