package device_management_services

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

type DeviceManagementServicesDataSource struct {
	client *client.Client
}

type DeviceManagementServicesDataSourceModel struct {
	ID      types.String                   `tfsdk:"id"`
	Servers []DeviceManagementServiceModel `tfsdk:"servers"`
}

type DeviceManagementServiceModel struct {
	ID              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	ServerName      types.String `tfsdk:"server_name"`
	ServerType      types.String `tfsdk:"server_type"`
	CreatedDateTime types.String `tfsdk:"created_date_time"`
	UpdatedDateTime types.String `tfsdk:"updated_date_time"`
}

func NewDeviceManagementServicesDataSource() datasource.DataSource {
	return &DeviceManagementServicesDataSource{}
}

func (d *DeviceManagementServicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_services"
}

func (d *DeviceManagementServicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of device management services.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"servers": schema.ListNestedAttribute{
				Description: "List of device management services.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The identifier of the MDM server.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of the resource (mdmServers).",
							Computed:    true,
						},
						"server_name": schema.StringAttribute{
							Description: "The name of the MDM server.",
							Computed:    true,
						},
						"server_type": schema.StringAttribute{
							Description: "The type of the server (MDM, APPLE_CONFIGURATOR, APPLE_MDM).",
							Computed:    true,
						},
						"created_date_time": schema.StringAttribute{
							Description: "The creation date and time of the server.",
							Computed:    true,
						},
						"updated_date_time": schema.StringAttribute{
							Description: "The last update date and time of the server.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DeviceManagementServicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *DeviceManagementServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DeviceManagementServicesDataSourceModel

	servers, err := d.client.GetDeviceManagementServices(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device Management Services",
			err.Error(),
		)
		return
	}

	state.Servers = make([]DeviceManagementServiceModel, 0, len(servers))
	for _, server := range servers {
		serverModel := DeviceManagementServiceModel{
			ID:              types.StringValue(server.ID),
			Type:            types.StringValue(server.Type),
			ServerName:      types.StringValue(server.Attributes.ServerName),
			ServerType:      types.StringValue(server.Attributes.ServerType),
			CreatedDateTime: types.StringValue(server.Attributes.CreatedDateTime),
			UpdatedDateTime: types.StringValue(server.Attributes.UpdatedDateTime),
		}
		state.Servers = append(state.Servers, serverModel)
	}

	state.ID = types.StringValue("device_management_services")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
