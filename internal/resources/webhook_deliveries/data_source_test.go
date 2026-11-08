// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook_deliveries_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/webhook_deliveries"
)

func TestWebhookDeliveriesDataSourceMetadata(t *testing.T) {
	ds := webhook_deliveries.NewWebhookDeliveriesDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)
	if resp.TypeName != "axm_webhook_deliveries" {
		t.Fatalf("expected axm_webhook_deliveries, got %q", resp.TypeName)
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
	if os.Getenv("AXM_SCOPE") != "business.api" {
		t.Skip("webhooks require the business.api scope")
	}
}

func TestAccWebhookDeliveriesDataSource(t *testing.T) {
	testAccPreCheck(t)
	webhookID := os.Getenv("AXM_TEST_WEBHOOK_ID")
	if webhookID == "" {
		t.Skip("AXM_TEST_WEBHOOK_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "axm_webhook_deliveries" "test" { webhook_id = %q }`, webhookID),
			Check:  resource.TestCheckResourceAttrSet("data.axm_webhook_deliveries.test", "id"),
		}},
	})
}
