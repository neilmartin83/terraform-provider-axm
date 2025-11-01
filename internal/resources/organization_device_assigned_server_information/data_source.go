package organization_device_assigned_server_information

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ datasource.DataSource = &OrganizationDeviceAssignedServerInformationDataSource{}

func NewOrganizationDeviceAssignedServerInformationDataSource() datasource.DataSource {
	return &OrganizationDeviceAssignedServerInformationDataSource{}
}

// OrganizationDeviceAssignedServerInformationDataSource defines the data source implementation.
type OrganizationDeviceAssignedServerInformationDataSource struct {
	client *client.Client
}

// OrganizationDeviceAssignedServerInformationDataSourceModel describes the data source data model.
type OrganizationDeviceAssignedServerInformationDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	DeviceID        types.String `tfsdk:"device_id"`
	ServerID        types.String `tfsdk:"server_id"`
	ServerName      types.String `tfsdk:"server_name"`
	ServerType      types.String `tfsdk:"server_type"`
	CreatedDateTime types.String `tfsdk:"created_date_time"`
	UpdatedDateTime types.String `tfsdk:"updated_date_time"`
}

func (d *OrganizationDeviceAssignedServerInformationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_device_assigned_server_information"
}

func (d *OrganizationDeviceAssignedServerInformationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

func (d *OrganizationDeviceAssignedServerInformationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *OrganizationDeviceAssignedServerInformationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDeviceAssignedServerInformationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	server, err := d.client.GetOrgDeviceAssignedServer(ctx, data.DeviceID.ValueString(), nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device Assigned Server",
			err.Error(),
		)
		return
	}

	data.ID = data.DeviceID
	data.ServerID = types.StringValue(server.ID)
	data.ServerName = types.StringValue(server.Attributes.ServerName)
	data.ServerType = types.StringValue(server.Attributes.ServerType)
	data.CreatedDateTime = types.StringValue(server.Attributes.CreatedDateTime)
	data.UpdatedDateTime = types.StringValue(server.Attributes.UpdatedDateTime)

	tflog.Debug(ctx, "Read organization device assigned server information", map[string]interface{}{
		"data": data,
	})

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
}
