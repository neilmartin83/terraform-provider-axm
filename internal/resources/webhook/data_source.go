// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ datasource.DataSource = &WebhookDataSource{}

// NewWebhookDataSource returns a new data source for a single webhook.
func NewWebhookDataSource() datasource.DataSource {
	return &WebhookDataSource{}
}

// WebhookDataSource fetches information about a webhook.
type WebhookDataSource struct {
	client *client.Client
}

// WebhookDataSourceModel describes the webhook data source model.
type WebhookDataSourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
	Name            types.String   `tfsdk:"name"`
	Enabled         types.Bool     `tfsdk:"enabled"`
	EventTypes      []types.String `tfsdk:"event_types"`
	URL             types.String   `tfsdk:"url"`
	CreatedDateTime types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime types.String   `tfsdk:"updated_date_time"`
}

func (d *WebhookDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (d *WebhookDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific Apple Business Manager webhook. The webhook secret is never returned by Apple.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The webhook ID.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the webhook.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the webhook is enabled.",
			},
			"event_types": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The event types to which the webhook subscribes.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL that receives webhook notifications.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the webhook was created.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the webhook was last updated.",
			},
		},
	}
}

func (d *WebhookDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_webhook data source") {
		return
	}
	d.client = c
}

func (d *WebhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WebhookDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readCtx, cancel, timeoutDiags := common.ResolveReadTimeout(ctx, data.Timeouts, common.DefaultReadTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	defer cancel()

	webhook, err := d.client.GetWebhook(readCtx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Webhook", err.Error())
		return
	}

	data.ID = types.StringValue(webhook.ID)
	data.Name = types.StringValue(webhook.Attributes.Name)
	data.Enabled = types.BoolValue(webhook.Attributes.Enabled)
	data.URL = types.StringValue(webhook.Attributes.URL)
	data.CreatedDateTime = types.StringValue(webhook.Attributes.CreatedDateTime)
	data.UpdatedDateTime = types.StringValue(webhook.Attributes.UpdatedDateTime)
	data.EventTypes = make([]types.String, 0, len(webhook.Attributes.EventTypes))
	for _, eventType := range webhook.Attributes.EventTypes {
		data.EventTypes = append(data.EventTypes, types.StringValue(eventType))
	}

	tflog.Debug(ctx, "Read webhook", map[string]any{"webhook_id": data.ID.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
