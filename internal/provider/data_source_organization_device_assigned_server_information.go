package axm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OrganizationDeviceAssignedServerInformationDataSource{}

type OrganizationDeviceAssignedServerInformationDataSource struct {
	client *Client
}

type OrganizationDeviceAssignedServerInformationDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	DeviceID        types.String `tfsdk:"device_id"`
	ServerID        types.String `tfsdk:"server_id"`
	ServerName      types.String `tfsdk:"server_name"`
	ServerType      types.String `tfsdk:"server_type"`
	CreatedDateTime types.String `tfsdk:"created_date_time"`
	UpdatedDateTime types.String `tfsdk:"updated_date_time"`
}

func NewOrganizationDeviceAssignedServerInformationDataSource() datasource.DataSource {
	return &OrganizationDeviceAssignedServerInformationDataSource{}
}

func (d *OrganizationDeviceAssignedServerInformationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_device_assigned_server_information"
}

func (d *OrganizationDeviceAssignedServerInformationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about the MDM server assigned to a specific device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"device_id": schema.StringAttribute{
				Description: "The identifier (serial number) of the device to look up.",
				Required:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "The identifier of the assigned MDM server.",
				Computed:    true,
			},
			"server_name": schema.StringAttribute{
				Description: "The name of the assigned MDM server.",
				Computed:    true,
			},
			"server_type": schema.StringAttribute{
				Description: "The type of the assigned server (MDM, APPLE_CONFIGURATOR, APPLE_MDM).",
				Computed:    true,
			},
			"created_date_time": schema.StringAttribute{
				Description: "The creation date and time of the server assignment.",
				Computed:    true,
			},
			"updated_date_time": schema.StringAttribute{
				Description: "The last update date and time of the server assignment.",
				Computed:    true,
			},
		},
	}
}

func (d *OrganizationDeviceAssignedServerInformationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *OrganizationDeviceAssignedServerInformationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OrganizationDeviceAssignedServerInformationDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server, err := d.client.GetOrgDeviceAssignedServer(ctx, state.DeviceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device Assigned Server",
			err.Error(),
		)
		return
	}

	state.ID = state.DeviceID
	state.ServerID = types.StringValue(server.ID)
	state.ServerName = types.StringValue(server.Attributes.ServerName)
	state.ServerType = types.StringValue(server.Attributes.ServerType)
	state.CreatedDateTime = types.StringValue(server.Attributes.CreatedDateTime)
	state.UpdatedDateTime = types.StringValue(server.Attributes.UpdatedDateTime)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
