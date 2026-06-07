// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package apple_device_management_devices

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

var _ datasource.DataSource = &AppleDeviceManagementDevicesDataSource{}

// NewAppleDeviceManagementDevicesDataSource returns a new data source for all
// MDM-enrolled devices.
func NewAppleDeviceManagementDevicesDataSource() datasource.DataSource {
	return &AppleDeviceManagementDevicesDataSource{}
}

// AppleDeviceManagementDevicesDataSource defines the data source implementation.
type AppleDeviceManagementDevicesDataSource struct {
	client *client.Client
}

// AppleDeviceManagementDevicesDataSourceModel describes the data source data model.
type AppleDeviceManagementDevicesDataSourceModel struct {
	ID      types.String                   `tfsdk:"id"`
	Timeouts timeouts.Value                `tfsdk:"timeouts"`
	Devices []AppleDeviceManagementDeviceModel `tfsdk:"devices"`
}

// AppleDeviceManagementDeviceModel describes an MDM-enrolled device.
type AppleDeviceManagementDeviceModel struct {
	ID             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	DeviceName     types.String `tfsdk:"device_name"`
	EnrolledUserID types.String `tfsdk:"enrolled_user_id"`
	ProductFamily  types.String `tfsdk:"product_family"`
	SerialNumber   types.String `tfsdk:"serial_number"`
}

func (d *AppleDeviceManagementDevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apple_device_management_devices"
}

func (d *AppleDeviceManagementDevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Apple devices enrolled in a device management service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"devices": schema.ListNestedAttribute{
				Description: "List of MDM-enrolled devices.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The opaque resource ID that uniquely identifies the resource.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the resource.",
						},
						"device_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the device.",
						},
						"enrolled_user_id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the user enrolled with the device.",
						},
						"product_family": schema.StringAttribute{
							Computed:    true,
							Description: "The product family of the device.",
						},
						"serial_number": schema.StringAttribute{
							Computed:    true,
							Description: "The serial number of the device.",
						},
					},
				},
			},
		},
	}
}

func (d *AppleDeviceManagementDevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = c
}

func (d *AppleDeviceManagementDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppleDeviceManagementDevicesDataSourceModel

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

	devices, err := d.client.GetMdmDevices(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Apple Device Management Devices",
			err.Error(),
		)
		return
	}

	data.Devices = make([]AppleDeviceManagementDeviceModel, 0, len(devices))
	for _, device := range devices {
		deviceModel := AppleDeviceManagementDeviceModel{
			ID:             types.StringValue(device.ID),
			Type:           types.StringValue(device.Type),
			DeviceName:     types.StringValue(device.Attributes.DeviceName),
			EnrolledUserID: types.StringValue(device.Attributes.EnrolledUserID),
			ProductFamily:  types.StringValue(device.Attributes.ProductFamily),
			SerialNumber:   types.StringValue(device.Attributes.SerialNumber),
		}

		data.Devices = append(data.Devices, deviceModel)
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read apple device management devices", map[string]any{
		"device_count": len(data.Devices),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
