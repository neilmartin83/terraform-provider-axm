// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/webhook"
)

func TestWebhookDataSourceMetadata(t *testing.T) {
	ds := webhook.NewWebhookDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)
	if resp.TypeName != "axm_webhook" {
		t.Fatalf("expected axm_webhook, got %q", resp.TypeName)
	}
}

func TestAccWebhookDataSource(t *testing.T) {
	testAccPreCheck(t)
	if os.Getenv("AXM_SCOPE") != "business.api" {
		t.Skip("webhooks require the business.api scope")
	}
	webhookID := os.Getenv("AXM_TEST_WEBHOOK_ID")
	if webhookID == "" {
		t.Skip("AXM_TEST_WEBHOOK_ID must be set for this test")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`data "axm_webhook" "test" { id = %q }`, webhookID),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.axm_webhook.test", "id", webhookID),
				resource.TestCheckResourceAttrSet("data.axm_webhook.test", "name"),
			),
		}},
	})
}
