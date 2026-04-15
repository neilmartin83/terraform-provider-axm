// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package audit_events

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ datasource.DataSource = &AuditEventsDataSource{}

// NewAuditEventsDataSource returns a new data source for audit events.
func NewAuditEventsDataSource() datasource.DataSource {
	return &AuditEventsDataSource{}
}

// AuditEventsDataSource defines the data source implementation.
type AuditEventsDataSource struct {
	client *client.Client
}

// AuditEventsDataSourceModel describes the data source data model.
type AuditEventsDataSourceModel struct {
	ID             types.String      `tfsdk:"id"`
	Timeouts       timeouts.Value    `tfsdk:"timeouts"`
	StartTimestamp types.String      `tfsdk:"start_timestamp"`
	EndTimestamp   types.String      `tfsdk:"end_timestamp"`
	ActorID        types.String      `tfsdk:"actor_id"`
	SubjectID      types.String      `tfsdk:"subject_id"`
	EventType      types.String      `tfsdk:"event_type"`
	Limit          types.Int64       `tfsdk:"limit"`
	Fields         []types.String    `tfsdk:"fields"`
	Cursor         types.String      `tfsdk:"cursor"`
	Events         []AuditEventModel `tfsdk:"events"`
}

// AuditEventModel describes an audit event.
type AuditEventModel struct {
	ID                   types.String `tfsdk:"id"`
	Type                 types.String `tfsdk:"type"`
	EventDateTime        types.String `tfsdk:"event_date_time"`
	EventType            types.String `tfsdk:"event_type"`
	Category             types.String `tfsdk:"category"`
	ActorType            types.String `tfsdk:"actor_type"`
	ActorID              types.String `tfsdk:"actor_id"`
	ActorName            types.String `tfsdk:"actor_name"`
	SubjectType          types.String `tfsdk:"subject_type"`
	SubjectID            types.String `tfsdk:"subject_id"`
	SubjectName          types.String `tfsdk:"subject_name"`
	Outcome              types.String `tfsdk:"outcome"`
	GroupID              types.String `tfsdk:"group_id"`
	EventDataPropertyKey types.String `tfsdk:"event_data_property_key"`
	EventDataJSON        types.String `tfsdk:"event_data_json"`
}

func (d *AuditEventsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_audit_events"
}

func (d *AuditEventsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches audit events that match the provided criteria.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"start_timestamp": schema.StringAttribute{
				Required:    true,
				Description: "ISO8601 start timestamp for the query range.",
			},
			"end_timestamp": schema.StringAttribute{
				Required:    true,
				Description: "ISO8601 end timestamp for the query range.",
			},
			"actor_id": schema.StringAttribute{
				Optional:    true,
				Description: "Actor ID to filter by.",
			},
			"subject_id": schema.StringAttribute{
				Optional:    true,
				Description: "Subject ID to filter by.",
			},
			"event_type": schema.StringAttribute{
				Optional:    true,
				Description: "Event type to filter by.",
			},
			"limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of events to request per page.",
			},
			"fields": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Fields to include in the response.",
			},
			"cursor": schema.StringAttribute{
				Optional:    true,
				Description: "Pagination cursor for the first page.",
			},
			"events": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of audit events.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The audit event ID.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The resource type.",
						},
						"event_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "Timestamp when the event occurred.",
						},
						"event_type": schema.StringAttribute{
							Computed:    true,
							Description: "The event type.",
						},
						"category": schema.StringAttribute{
							Computed:    true,
							Description: "The event category.",
						},
						"actor_type": schema.StringAttribute{
							Computed:    true,
							Description: "The actor type.",
						},
						"actor_id": schema.StringAttribute{
							Computed:    true,
							Description: "The actor ID.",
						},
						"actor_name": schema.StringAttribute{
							Computed:    true,
							Description: "The actor name.",
						},
						"subject_type": schema.StringAttribute{
							Computed:    true,
							Description: "The subject type.",
						},
						"subject_id": schema.StringAttribute{
							Computed:    true,
							Description: "The subject ID.",
						},
						"subject_name": schema.StringAttribute{
							Computed:    true,
							Description: "The subject name.",
						},
						"outcome": schema.StringAttribute{
							Computed:    true,
							Description: "The outcome of the event.",
						},
						"group_id": schema.StringAttribute{
							Computed:    true,
							Description: "The event group ID.",
						},
						"event_data_property_key": schema.StringAttribute{
							Computed:    true,
							Description: "The event data property key.",
						},
						"event_data_json": schema.StringAttribute{
							Computed:    true,
							Description: "JSON payload for event-specific data.",
						},
					},
				},
			},
		},
	}
}

func (d *AuditEventsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_audit_events data source") {
		return
	}
	d.client = c
}

func (d *AuditEventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AuditEventsDataSourceModel

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

	queryParams := url.Values{}
	queryParams.Set("filter[startTimestamp]", data.StartTimestamp.ValueString())
	queryParams.Set("filter[endTimestamp]", data.EndTimestamp.ValueString())

	if !data.ActorID.IsNull() && !data.ActorID.IsUnknown() && data.ActorID.ValueString() != "" {
		queryParams.Set("filter[actorId]", data.ActorID.ValueString())
	}
	if !data.SubjectID.IsNull() && !data.SubjectID.IsUnknown() && data.SubjectID.ValueString() != "" {
		queryParams.Set("filter[subjectId]", data.SubjectID.ValueString())
	}
	if !data.EventType.IsNull() && !data.EventType.IsUnknown() && data.EventType.ValueString() != "" {
		queryParams.Set("filter[type]", data.EventType.ValueString())
	}
	if !data.Limit.IsNull() && !data.Limit.IsUnknown() && data.Limit.ValueInt64() > 0 {
		queryParams.Set("limit", strconv.FormatInt(data.Limit.ValueInt64(), 10))
	}
	if len(data.Fields) > 0 {
		fieldValues := make([]string, 0, len(data.Fields))
		for _, field := range data.Fields {
			if !field.IsNull() && !field.IsUnknown() && field.ValueString() != "" {
				fieldValues = append(fieldValues, field.ValueString())
			}
		}
		if len(fieldValues) > 0 {
			queryParams.Set("fields[auditEvents]", strings.Join(fieldValues, ","))
		}
	}
	if !data.Cursor.IsNull() && !data.Cursor.IsUnknown() && data.Cursor.ValueString() != "" {
		queryParams.Set("cursor", data.Cursor.ValueString())
	}

	events, err := d.client.GetAuditEvents(readCtx, queryParams)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read audit events", err.Error())
		return
	}

	data.Events = make([]AuditEventModel, 0, len(events))
	for _, event := range events {
		data.Events = append(data.Events, flattenAuditEvent(event))
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read audit events", map[string]any{
		"event_count": len(data.Events),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenAuditEvent(event client.AuditEvent) AuditEventModel {
	additional, _ := json.Marshal(event.Attributes.Additional)

	return AuditEventModel{
		ID:                   types.StringValue(event.ID),
		Type:                 types.StringValue(event.Type),
		EventDateTime:        types.StringPointerValue(common.StringPointerOrNil(event.Attributes.EventDateTime)),
		EventType:            types.StringPointerValue(common.StringPointerOrNil(event.Attributes.Type)),
		Category:             types.StringPointerValue(common.StringPointerOrNil(event.Attributes.Category)),
		ActorType:            types.StringPointerValue(common.StringPointerOrNil(event.Attributes.ActorType)),
		ActorID:              types.StringPointerValue(common.StringPointerOrNil(event.Attributes.ActorID)),
		ActorName:            types.StringPointerValue(common.StringPointerOrNil(event.Attributes.ActorName)),
		SubjectType:          types.StringPointerValue(common.StringPointerOrNil(event.Attributes.SubjectType)),
		SubjectID:            types.StringPointerValue(common.StringPointerOrNil(event.Attributes.SubjectID)),
		SubjectName:          types.StringPointerValue(common.StringPointerOrNil(event.Attributes.SubjectName)),
		Outcome:              types.StringPointerValue(common.StringPointerOrNil(event.Attributes.Outcome)),
		GroupID:              types.StringPointerValue(common.StringPointerOrNil(event.Attributes.GroupID)),
		EventDataPropertyKey: types.StringPointerValue(common.StringPointerOrNil(event.Attributes.EventDataPropertyKey)),
		EventDataJSON:        types.StringValue(string(additional)),
	}
}
