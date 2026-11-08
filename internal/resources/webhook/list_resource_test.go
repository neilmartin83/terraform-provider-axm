// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/list"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func TestWebhookListResourceMetadata(t *testing.T) {
	lr := NewWebhookListResource()
	resp := tfresource.MetadataResponse{}
	lr.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)
	if resp.TypeName != "axm_webhook" {
		t.Fatalf("expected axm_webhook, got %q", resp.TypeName)
	}
}

func TestWebhookListResourceSchema(t *testing.T) {
	lr := NewWebhookListResource()
	resp := list.ListResourceSchemaResponse{}
	lr.ListResourceConfigSchema(context.Background(), list.ListResourceSchemaRequest{}, &resp)
	for _, name := range []string{"name", "enabled", "event_type"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok || !attr.IsOptional() {
			t.Errorf("expected optional %q filter", name)
		}
	}
}

func TestFilterWebhookList(t *testing.T) {
	webhooks := []client.Webhook{
		{
			ID: "one",
			Attributes: client.WebhookAttributes{
				Name:       "Device Changes",
				Enabled:    true,
				EventTypes: []string{"ORG_DEVICE_ATTRIBUTE_CHANGED"},
			},
		},
		{
			ID: "two",
			Attributes: client.WebhookAttributes{
				Name:       "Warranty",
				Enabled:    false,
				EventTypes: []string{"ORG_DEVICE_RELATIONSHIP_WARRANTY_CHANGED"},
			},
		},
	}

	filtered := filterWebhookList(webhooks, WebhookListResourceModel{
		Name:      types.StringValue("device changes"),
		Enabled:   types.BoolValue(true),
		EventType: types.StringValue("ORG_DEVICE_ATTRIBUTE_CHANGED"),
	})
	if len(filtered) != 1 || filtered[0].ID != "one" {
		t.Fatalf("unexpected filtered webhooks: %#v", filtered)
	}
}
