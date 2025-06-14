package axm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DeviceManagementServiceSerialNumbersDataSource{}

type DeviceManagementServiceSerialNumbersDataSource struct {
	client *Client
}

type DeviceManagementServiceSerialNumbersDataSourceModel struct {
	ID            types.String   `tfsdk:"id"`
	ServerID      types.String   `tfsdk:"server_id"`
	SerialNumbers []types.String `tfsdk:"serial_numbers"`
}

func NewDeviceManagementServiceSerialNumbersDataSource() datasource.DataSource {
	return &DeviceManagementServiceSerialNumbersDataSource{}
}

func (d *DeviceManagementServiceSerialNumbersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service_serial_numbers"
}

func (d *DeviceManagementServiceSerialNumbersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the list of device serial numbers assigned to a specific device management service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "The identifier of the MDM server to get serial numbers for.",
				Required:    true,
			},
			"serial_numbers": schema.ListAttribute{
				Description: "List of device serial numbers assigned to this MDM server.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *DeviceManagementServiceSerialNumbersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeviceManagementServiceSerialNumbersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DeviceManagementServiceSerialNumbersDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	serialNumbers, err := d.client.GetDeviceManagementServiceSerialNumbers(ctx, state.ServerID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device Management Service Serial Numbers",
			err.Error(),
		)
		return
	}

	state.SerialNumbers = make([]types.String, len(serialNumbers))
	for i, sn := range serialNumbers {
		state.SerialNumbers[i] = types.StringValue(sn)
	}

	state.ID = state.ServerID

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
