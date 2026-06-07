// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

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

var _ datasource.DataSource = &DeviceManagementServiceDataSource{}

// NewDeviceManagementServiceDataSource returns a new data source for a single device management service.
func NewDeviceManagementServiceDataSource() datasource.DataSource {
	return &DeviceManagementServiceDataSource{}
}

// DeviceManagementServiceDataSource defines the data source implementation.
type DeviceManagementServiceDataSource struct {
	client *client.Client
}

// DeviceManagementServiceDataSourceModel describes the data source data model.
type DeviceManagementServiceDataSourceModel struct {
	ID                     types.String   `tfsdk:"id"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
	Type                   types.String   `tfsdk:"type"`
	ServerName             types.String   `tfsdk:"server_name"`
	ServerType             types.String   `tfsdk:"server_type"`
	Status                 types.String   `tfsdk:"status"`
	DeviceCount            types.Int64    `tfsdk:"device_count"`
	DefaultProductFamilies types.List     `tfsdk:"default_product_families"`
	LastConnectedDateTime  types.String   `tfsdk:"last_connected_date_time"`
	LastConnectedIp        types.String   `tfsdk:"last_connected_ip"`
	AllowRelease           types.Bool     `tfsdk:"allow_release"`
	CreatedDateTime        types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime        types.String   `tfsdk:"updated_date_time"`
}

func (d *DeviceManagementServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service"
}

func (d *DeviceManagementServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific device management service (MDM server) from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The opaque resource ID that uniquely identifies the resource.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the resource (mdmServers).",
			},
			"server_name": schema.StringAttribute{
				Computed:    true,
				Description: "The device management service's name.",
			},
			"server_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of device management service: MDM, APPLE_CONFIGURATOR, APPLE_MDM. Read only.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The operational status of the device management service. Read only.",
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of devices currently assigned to this device management service. Read only.",
			},
			"default_product_families": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The product families that are assigned by default to this device management service. Read/update only.",
			},
			"last_connected_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the device management service last connected to Apple's servers. Read only.",
			},
			"last_connected_ip": schema.StringAttribute{
				Computed:    true,
				Description: "The IP address from which the device management service last connected to Apple's servers. Read only.",
			},
			"allow_release": schema.BoolAttribute{
				Computed:    true,
				Description: "A Boolean value that indicates whether the device management service is allowed to disown its enrolled devices.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the creation of the resource.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the most-recent update for the resource.",
			},
		},
	}
}

func (d *DeviceManagementServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = c
}

func (d *DeviceManagementServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceManagementServiceDataSourceModel

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

	srv, err := d.client.GetDeviceManagementService(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read device management service", err.Error())
		return
	}

	data.Type = types.StringValue(srv.Type)
	data.ServerName = types.StringValue(srv.Attributes.ServerName)
	data.ServerType = types.StringValue(srv.Attributes.ServerType)
	data.Status = common.StringPointerToTypesString(srv.Attributes.Status)
	data.DeviceCount = types.Int64PointerValue(srv.Attributes.DeviceCount)
	data.DefaultProductFamilies = common.StringsToList(ctx, srv.Attributes.DefaultProductFamilies)
	data.LastConnectedDateTime = types.StringPointerValue(srv.Attributes.LastConnectedDateTime)
	data.LastConnectedIp = types.StringPointerValue(srv.Attributes.LastConnectedIp)
	data.AllowRelease = types.BoolPointerValue(srv.Attributes.EnableMdmDisownFlag)
	data.CreatedDateTime = types.StringValue(srv.Attributes.CreatedDateTime)
	data.UpdatedDateTime = types.StringValue(srv.Attributes.UpdatedDateTime)

	tflog.Debug(ctx, "Read device management service", map[string]any{
		"server_id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
