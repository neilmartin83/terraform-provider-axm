package device_management_services

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

var _ datasource.DataSource = &DeviceManagementServicesDataSource{}

func NewDeviceManagementServicesDataSource() datasource.DataSource {
	return &DeviceManagementServicesDataSource{}
}

// DeviceManagementServicesDataSource defines the data source implementation.
type DeviceManagementServicesDataSource struct {
	client *client.Client
}

// DeviceManagementServicesDataSourceModel describes the data source data model.
type DeviceManagementServicesDataSourceModel struct {
	ID       types.String                   `tfsdk:"id"`
	Timeouts timeouts.Value                 `tfsdk:"timeouts"`
	Servers  []DeviceManagementServiceModel `tfsdk:"servers"`
}

// DeviceManagementServiceModel describes a device management service.
type DeviceManagementServiceModel struct {
	ID              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	ServerName      types.String `tfsdk:"server_name"`
	ServerType      types.String `tfsdk:"server_type"`
	CreatedDateTime types.String `tfsdk:"created_date_time"`
	UpdatedDateTime types.String `tfsdk:"updated_date_time"`
}

func (d *DeviceManagementServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_services"
}

func (d *DeviceManagementServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of device management services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"servers": schema.ListNestedAttribute{
				Description: "List of device management services.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The opaque resource ID that uniquely identifies the resource.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the resource (mdmServers).",
							Computed:    true,
						},
						"server_name": schema.StringAttribute{
							Description: "The device management service's name.",
							Computed:    true,
						},
						"server_type": schema.StringAttribute{
							Description: "The type of device management service: MDM, APPLE_CONFIGURATOR, APPLE_MDM.",
							Computed:    true,
						},
						"created_date_time": schema.StringAttribute{
							Description: "The date and time of the creation of the resource.",
							Computed:    true,
						},
						"updated_date_time": schema.StringAttribute{
							Description: "The date and time of the most-recent update for the resource.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DeviceManagementServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = c
}

func (d *DeviceManagementServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceManagementServicesDataSourceModel

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

	servers, err := d.client.GetDeviceManagementServices(readCtx, nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device Management Services",
			err.Error(),
		)
		return
	}

	data.Servers = make([]DeviceManagementServiceModel, 0, len(servers))
	for _, server := range servers {
		serverModel := DeviceManagementServiceModel{
			ID:              types.StringValue(server.ID),
			Type:            types.StringValue(server.Type),
			ServerName:      types.StringValue(server.Attributes.ServerName),
			ServerType:      types.StringValue(server.Attributes.ServerType),
			CreatedDateTime: types.StringValue(server.Attributes.CreatedDateTime),
			UpdatedDateTime: types.StringValue(server.Attributes.UpdatedDateTime),
		}
		data.Servers = append(data.Servers, serverModel)
	}

	data.ID = types.StringValue("device_management_services")

	tflog.Debug(ctx, "Read device management services", map[string]any{
		"data": data,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
