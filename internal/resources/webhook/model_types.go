// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WebhookModel describes the Terraform state for a webhook.
type WebhookModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Enabled         types.Bool     `tfsdk:"enabled"`
	EventTypes      types.Set      `tfsdk:"event_types"`
	SecretWo        types.String   `tfsdk:"secret_wo"`
	SecretWoVersion types.Int64    `tfsdk:"secret_wo_version"`
	URL             types.String   `tfsdk:"url"`
	CreatedDateTime types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime types.String   `tfsdk:"updated_date_time"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

// webhookIdentityModel captures the webhook ID used by managed resource and list results.
type webhookIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

// WebhookListResourceModel captures filters supported by webhook list queries.
type WebhookListResourceModel struct {
	Name      types.String `tfsdk:"name"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	EventType types.String `tfsdk:"event_type"`
}
