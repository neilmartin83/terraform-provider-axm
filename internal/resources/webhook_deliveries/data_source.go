// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook_deliveries

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

var _ datasource.DataSource = &WebhookDeliveriesDataSource{}

// NewWebhookDeliveriesDataSource returns a data source for a webhook's delivery history.
func NewWebhookDeliveriesDataSource() datasource.DataSource {
	return &WebhookDeliveriesDataSource{}
}

// WebhookDeliveriesDataSource fetches delivery history for one webhook.
type WebhookDeliveriesDataSource struct {
	client *client.Client
}

// WebhookDeliveriesDataSourceModel describes the delivery history data source.
type WebhookDeliveriesDataSourceModel struct {
	ID         types.String               `tfsdk:"id"`
	WebhookID  types.String               `tfsdk:"webhook_id"`
	Timeouts   timeouts.Value             `tfsdk:"timeouts"`
	Deliveries []WebhookDeliveryDataModel `tfsdk:"deliveries"`
}

// WebhookDeliveryDataModel describes one delivery.
type WebhookDeliveryDataModel struct {
	ID                 types.String `tfsdk:"id"`
	CreatedDateTime    types.String `tfsdk:"created_date_time"`
	DeliveryState      types.String `tfsdk:"delivery_state"`
	ErrorMessage       types.String `tfsdk:"error_message"`
	Redelivery         types.Bool   `tfsdk:"redelivery"`
	SentDateTime       types.String `tfsdk:"sent_date_time"`
	RequestURL         types.String `tfsdk:"request_url"`
	ResponseHTTPStatus types.Int64  `tfsdk:"response_http_status_code"`
	ResponseBody       types.String `tfsdk:"response_body"`
	EventID            types.String `tfsdk:"event_id"`
}

func (d *WebhookDeliveriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_deliveries"
}

func (d *WebhookDeliveriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches delivery history for an Apple Business Manager webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this data source.",
			},
			"webhook_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the webhook whose deliveries are fetched.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"deliveries": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The webhook's deliveries.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The delivery ID.",
						},
						"created_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time the delivery was created.",
						},
						"delivery_state": schema.StringAttribute{
							Computed:    true,
							Description: "The delivery state: SUCCEEDED, FAILED, or PENDING.",
						},
						"error_message": schema.StringAttribute{
							Computed:    true,
							Description: "The error message, if delivery failed.",
						},
						"redelivery": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this delivery is a redelivery.",
						},
						"sent_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time the delivery was sent.",
						},
						"request_url": schema.StringAttribute{
							Computed:    true,
							Description: "The webhook URL to which the delivery was sent.",
						},
						"response_http_status_code": schema.Int64Attribute{
							Computed:    true,
							Description: "The HTTP status code returned by the webhook endpoint.",
						},
						"response_body": schema.StringAttribute{
							Computed:    true,
							Description: "The response body returned by the webhook endpoint.",
						},
						"event_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the event associated with the delivery.",
						},
					},
				},
			},
		},
	}
}

func (d *WebhookDeliveriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_webhook_deliveries data source") {
		return
	}
	d.client = c
}

func (d *WebhookDeliveriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data WebhookDeliveriesDataSourceModel
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

	deliveries, err := d.client.GetWebhookDeliveries(readCtx, data.WebhookID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Webhook Deliveries", err.Error())
		return
	}

	data.Deliveries = make([]WebhookDeliveryDataModel, 0, len(deliveries))
	for _, delivery := range deliveries {
		data.Deliveries = append(data.Deliveries, WebhookDeliveryDataModel{
			ID:                 types.StringValue(delivery.ID),
			CreatedDateTime:    types.StringValue(delivery.Attributes.CreatedDateTime),
			DeliveryState:      types.StringValue(delivery.Attributes.DeliveryState),
			ErrorMessage:       types.StringValue(delivery.Attributes.ErrorMessage),
			Redelivery:         types.BoolValue(delivery.Attributes.Redelivery),
			SentDateTime:       types.StringValue(delivery.Attributes.SentDateTime),
			RequestURL:         types.StringValue(delivery.Attributes.Request.URL),
			ResponseHTTPStatus: types.Int64Value(int64(delivery.Attributes.Response.HTTPStatusCode)),
			ResponseBody:       types.StringValue(delivery.Attributes.Response.Body),
			EventID:            types.StringValue(delivery.Relationships.Event.Data.ID),
		})
	}
	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read webhook deliveries", map[string]any{
		"webhook_id":     data.WebhookID.ValueString(),
		"delivery_count": len(data.Deliveries),
	})
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
