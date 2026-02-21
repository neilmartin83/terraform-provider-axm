package organization_device_test

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

func TestAccOrganizationDeviceDataSource(t *testing.T) {
	deviceID := os.Getenv("AXM_TEST_DEVICE_ID")
	if deviceID == "" {
		t.Skip("AXM_TEST_DEVICE_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_organization_device" "test" {
						id = %q
					}
				`, deviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.axm_organization_device.test", "id", deviceID),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "serial_number"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "device_model"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "product_family"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "status"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "added_to_org_date_time"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "updated_date_time"),
					resource.TestCheckResourceAttrSet("data.axm_organization_device.test", "color"),
				),
			},
		},
	})
}
