// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ datasource.DataSource = &WebhooksDataSource{}

// NewWebhooksDataSource returns a new data source for all webhooks.
func NewWebhooksDataSource() datasource.DataSource {
	return &WebhooksDataSource{}
}

// WebhooksDataSource fetches all webhooks in the organization.
type WebhooksDataSource struct {
	client *client.Client
}

// WebhooksDataSourceModel describes the webhooks data source model.
type WebhooksDataSourceModel struct {
	ID       types.String             `tfsdk:"id"`
	Timeouts timeouts.Value           `tfsdk:"timeouts"`
	Webhooks []WebhookDataSourceModel `tfsdk:"webhooks"`
}

// WebhookDataSourceModel describes one webhook in the webhooks data source.
type WebhookDataSourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Enabled         types.Bool     `tfsdk:"enabled"`
	EventTypes      []types.String `tfsdk:"event_types"`
	URL             types.String   `tfsdk:"url"`
	CreatedDateTime types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime types.String   `tfsdk:"updated_date_time"`
}

func (d *WebhooksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhooks"
}

func (d *WebhooksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches Apple Business Manager webhooks. Apple does not return webhook secrets.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this data source.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"webhooks": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The webhooks in the organization.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The webhook ID.",
						},
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
				},
			},
		},
	}
}

func (d *WebhooksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_webhooks data source") {
		return
	}
	d.client = c
}

func (d *WebhooksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WebhooksDataSourceModel
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

	webhooks, err := d.client.GetWebhooks(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Webhooks", err.Error())
		return
	}

	data.Webhooks = make([]WebhookDataSourceModel, 0, len(webhooks))
	for _, webhook := range webhooks {
		eventTypes := make([]types.String, 0, len(webhook.Attributes.EventTypes))
		for _, eventType := range webhook.Attributes.EventTypes {
			eventTypes = append(eventTypes, types.StringValue(eventType))
		}
		data.Webhooks = append(data.Webhooks, WebhookDataSourceModel{
			ID:              types.StringValue(webhook.ID),
			Name:            types.StringValue(webhook.Attributes.Name),
			Enabled:         types.BoolValue(webhook.Attributes.Enabled),
			EventTypes:      eventTypes,
			URL:             types.StringValue(webhook.Attributes.URL),
			CreatedDateTime: types.StringValue(webhook.Attributes.CreatedDateTime),
			UpdatedDateTime: types.StringValue(webhook.Attributes.UpdatedDateTime),
		})
	}
	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read webhooks", map[string]any{"webhook_count": len(data.Webhooks)})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
