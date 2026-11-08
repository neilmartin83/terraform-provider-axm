// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ list.ListResource = &WebhookListResource{}
var _ list.ListResourceWithConfigure = &WebhookListResource{}

// NewWebhookListResource returns a new list resource for webhooks.
func NewWebhookListResource() list.ListResource {
	return &WebhookListResource{}
}

// WebhookListResource supports querying webhooks.
type WebhookListResource struct {
	client *client.Client
}

func (r *WebhookListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "List")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_webhook list resource") {
		return
	}
	r.client = c
}

func (r *WebhookListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Searches Apple Business Manager webhooks.",
		Attributes: map[string]listschema.Attribute{
			"name": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive exact webhook name.",
			},
			"enabled": listschema.BoolAttribute{
				Optional:    true,
				Description: "Filters results by whether the webhook is enabled.",
			},
			"event_type": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results to webhooks subscribed to this event type.",
			},
		},
	}
}

func (r *WebhookListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unconfigured Provider",
				"The provider has not been configured yet. Re-run the command after `terraform init` has completed successfully.",
			),
		})
		return
	}

	var config WebhookListResourceModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	webhooks, err := r.client.GetWebhooks(ctx, nil)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic("Unable to list Webhooks", err.Error()),
		})
		return
	}

	filtered := filterWebhookList(webhooks, config)
	maxResults := req.Limit
	if maxResults <= 0 || maxResults > int64(len(filtered)) {
		maxResults = int64(len(filtered))
	}

	results := make([]list.ListResult, 0, int(maxResults))
	for _, webhook := range filtered {
		if int64(len(results)) >= maxResults {
			break
		}

		result := req.NewListResult(ctx)
		result.DisplayName = webhook.Attributes.Name
		identity := webhookIdentityModel{ID: types.StringValue(webhook.ID)}
		result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

		if req.IncludeResource {
			state, err := r.buildWebhookState(ctx, &webhook)
			if err != nil {
				stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
					diag.NewErrorDiagnostic("Unable to read Webhook", err.Error()),
				})
				return
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, state)...)
		}

		results = append(results, result)
	}

	tflog.Debug(ctx, "Listed webhooks", map[string]any{
		"requested_limit": req.Limit,
		"returned":        len(results),
	})
	if len(results) == 0 {
		stream.Results = list.NoListResults
		return
	}
	stream.Results = slices.Values(results)
}

func filterWebhookList(webhooks []client.Webhook, config WebhookListResourceModel) []client.Webhook {
	filtered := make([]client.Webhook, 0, len(webhooks))
	name, hasName := common.NormalizedFilterString(config.Name)
	eventType, hasEventType := common.NormalizedFilterString(config.EventType)

	for _, webhook := range webhooks {
		if hasName && !strings.EqualFold(strings.TrimSpace(webhook.Attributes.Name), name) {
			continue
		}
		if !config.Enabled.IsNull() && !config.Enabled.IsUnknown() &&
			webhook.Attributes.Enabled != config.Enabled.ValueBool() {
			continue
		}
		if hasEventType && !slices.Contains(webhook.Attributes.EventTypes, eventType) {
			continue
		}
		filtered = append(filtered, webhook)
	}
	return filtered
}

func (r *WebhookListResource) buildWebhookState(ctx context.Context, webhook *client.Webhook) (WebhookModel, error) {
	state := WebhookModel{Timeouts: ensureWebhookTimeouts(timeouts.Value{})}
	resource := &WebhookResource{client: r.client}
	var diags diag.Diagnostics
	resource.refreshWebhookModel(ctx, webhook, &state, &diags)
	if diags.HasError() {
		return WebhookModel{}, fmt.Errorf("unable to convert webhook attributes: %s", diags)
	}
	return state, nil
}
