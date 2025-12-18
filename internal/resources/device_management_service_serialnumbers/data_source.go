package device_management_service_serialnumbers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ datasource.DataSource = &DeviceManagementServiceSerialNumbersDataSource{}

func NewDeviceManagementServiceSerialNumbersDataSource() datasource.DataSource {
	return &DeviceManagementServiceSerialNumbersDataSource{}
}

// DeviceManagementServiceSerialNumbersDataSource defines the data source implementation.
type DeviceManagementServiceSerialNumbersDataSource struct {
	client *client.Client
}

// DeviceManagementServiceSerialNumbersDataSourceModel describes the data source data model.
type DeviceManagementServiceSerialNumbersDataSourceModel struct {
	ID            types.String   `tfsdk:"id"`
	ServerID      types.String   `tfsdk:"server_id"`
	SerialNumbers []types.String `tfsdk:"serial_numbers"`
}

func (d *DeviceManagementServiceSerialNumbersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service_serial_numbers"
}

func (d *DeviceManagementServiceSerialNumbersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the list of device serial numbers assigned to a specific device management service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The opaque resource ID that uniquely identifies the resource.",
				Computed:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "The opaque resource ID that uniquely identifies the device management service to get serial numbers for.",
				Required:    true,
			},
			"serial_numbers": schema.ListAttribute{
				Description: "List of device serial numbers assigned to this device management service.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *DeviceManagementServiceSerialNumbersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeviceManagementServiceSerialNumbersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DeviceManagementServiceSerialNumbersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	serialNumbers, err := d.client.GetDeviceManagementServiceSerialNumbers(ctx, data.ServerID.ValueString())

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Device Management Service Serial Numbers",
			err.Error(),
		)
		return
	}

	data.SerialNumbers = make([]types.String, len(serialNumbers))

	for i, sn := range serialNumbers {
		data.SerialNumbers[i] = types.StringValue(sn)
	}

	data.ID = data.ServerID

	tflog.Debug(ctx, "Read device management service serial numbers", map[string]interface{}{
		"server_id":      data.ServerID.ValueString(),
		"serial_numbers": serialNumbers,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
