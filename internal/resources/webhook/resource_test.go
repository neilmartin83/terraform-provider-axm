// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/webhook"
)

func TestWebhookResourceMetadata(t *testing.T) {
	r := webhook.NewWebhookResource()
	resp := tfresource.MetadataResponse{}
	r.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)
	if resp.TypeName != "axm_webhook" {
		t.Fatalf("expected axm_webhook, got %q", resp.TypeName)
	}
}

func TestWebhookResourceSchema(t *testing.T) {
	r := webhook.NewWebhookResource()
	resp := tfresource.SchemaResponse{}
	r.Schema(context.Background(), tfresource.SchemaRequest{}, &resp)

	secret, ok := resp.Schema.Attributes["secret_wo"].(schema.StringAttribute)
	if !ok {
		t.Fatal("secret is not a StringAttribute")
	}
	if !secret.Required || !secret.WriteOnly || !secret.Sensitive {
		t.Errorf("secret must be required, write-only, and sensitive: %#v", secret)
	}
	if attr, ok := resp.Schema.Attributes["secret_wo_version"]; !ok || !attr.IsRequired() {
		t.Error("secret_wo_version must be required")
	}

	eventTypes, ok := resp.Schema.Attributes["event_types"].(schema.SetAttribute)
	if !ok {
		t.Fatal("event_types is not a SetAttribute")
	}
	if !eventTypes.Required {
		t.Error("event_types must be required")
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

func TestAccWebhookResource_basic(t *testing.T) {
	testAccPreCheck(t)
	webhookURL := os.Getenv("AXM_TEST_WEBHOOK_URL")
	secret := os.Getenv("AXM_TEST_WEBHOOK_SECRET")
	if webhookURL == "" || secret == "" {
		t.Skip("AXM_TEST_WEBHOOK_URL and AXM_TEST_WEBHOOK_SECRET must be set to create a webhook")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_webhook" "test" {
						name        = "tf-acc-webhook"
						event_types = ["ORG_DEVICE_ATTRIBUTE_CHANGED"]
						secret_wo   = %q
						secret_wo_version = 1
						url         = %q
					}
				`, secret, webhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("axm_webhook.test", "id"),
					resource.TestCheckResourceAttr("axm_webhook.test", "name", "tf-acc-webhook"),
					resource.TestCheckResourceAttr("axm_webhook.test", "enabled", "true"),
					resource.TestCheckResourceAttr("axm_webhook.test", "event_types.#", "1"),
					resource.TestCheckResourceAttr("axm_webhook.test", "url", webhookURL),
					resource.TestCheckNoResourceAttr("axm_webhook.test", "secret_wo"),
					resource.TestCheckResourceAttr("axm_webhook.test", "secret_wo_version", "1"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "axm_webhook" "test" {
						name        = "tf-acc-webhook-updated"
						enabled     = false
						event_types = ["ORG_DEVICE_RELATIONSHIP_ASSIGNED_SERVER_CHANGED"]
						secret_wo   = %q
						secret_wo_version = 2
						url         = %q
					}
				`, secret, webhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("axm_webhook.test", "name", "tf-acc-webhook-updated"),
					resource.TestCheckResourceAttr("axm_webhook.test", "enabled", "false"),
					resource.TestCheckResourceAttr("axm_webhook.test", "event_types.#", "1"),
					resource.TestCheckNoResourceAttr("axm_webhook.test", "secret_wo"),
					resource.TestCheckResourceAttr("axm_webhook.test", "secret_wo_version", "2"),
				),
			},
		},
	})
}

func TestAccWebhookResource_liveCRUD(t *testing.T) {
	testAccPreCheck(t)

	teamID := os.Getenv("AXM_TEAM_ID")
	if teamID == "" {
		teamID = os.Getenv("AXM_CLIENT_ID")
	}
	c, err := client.NewClient(
		"https://api-business.apple.com",
		teamID,
		os.Getenv("AXM_CLIENT_ID"),
		os.Getenv("AXM_KEY_ID"),
		os.Getenv("AXM_SCOPE"),
		os.Getenv("AXM_PRIVATE_KEY"),
	)
	if err != nil {
		t.Fatalf("failed to initialize Apple API client: %v", err)
	}

	webhookURL := "https://webhook-test.invalid/terraform-provider-axm"
	secret := os.Getenv("AXM_TEST_WEBHOOK_SECRET")
	if secret == "" {
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			t.Fatalf("failed to generate temporary webhook secret: %v", err)
		}
		secret = hex.EncodeToString(secretBytes)
	}
	name := fmt.Sprintf("tf-acc-webhook-%d", time.Now().UnixNano())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_webhook" "test" {
						name              = %q
						enabled           = false
						event_types       = ["ORG_DEVICE_ATTRIBUTE_CHANGED"]
						secret_wo         = %q
						secret_wo_version = 1
						url               = %q
					}
				`, name, secret, webhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("axm_webhook.test", "id"),
					resource.TestCheckResourceAttr("axm_webhook.test", "enabled", "false"),
					resource.TestCheckResourceAttr("axm_webhook.test", "secret_wo_version", "1"),
					resource.TestCheckNoResourceAttr("axm_webhook.test", "secret_wo"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "axm_webhook" "test" {
						name              = %q
						enabled           = false
						event_types       = ["ORG_DEVICE_ATTRIBUTE_CHANGED"]
						secret_wo         = %q
						secret_wo_version = 2
						url               = %q
					}
				`, name+"-updated", secret, webhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("axm_webhook.test", "name", name+"-updated"),
					resource.TestCheckResourceAttr("axm_webhook.test", "secret_wo_version", "2"),
					resource.TestCheckNoResourceAttr("axm_webhook.test", "secret_wo"),
				),
			},
		},
	})

	webhooks, err := c.GetWebhooks(context.Background(), nil)
	if err != nil {
		t.Fatalf("failed to verify webhook cleanup: %v", err)
	}
	for _, webhook := range webhooks {
		if webhook.Attributes.Name == name || webhook.Attributes.Name == name+"-updated" {
			t.Fatalf("test webhook %q still exists after acceptance test cleanup", webhook.Attributes.Name)
		}
	}
}
