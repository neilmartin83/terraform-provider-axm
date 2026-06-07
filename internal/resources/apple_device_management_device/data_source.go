// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package apple_device_management_device

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

var _ datasource.DataSource = &AppleDeviceManagementDeviceDataSource{}

// NewAppleDeviceManagementDeviceDataSource returns a new data source for a single
// MDM-enrolled device.
func NewAppleDeviceManagementDeviceDataSource() datasource.DataSource {
	return &AppleDeviceManagementDeviceDataSource{}
}

// AppleDeviceManagementDeviceDataSource defines the data source implementation.
type AppleDeviceManagementDeviceDataSource struct {
	client *client.Client
}

// AppleDeviceManagementDeviceDataSourceModel describes the data source data model.
type AppleDeviceManagementDeviceDataSourceModel struct {
	ID                   types.String   `tfsdk:"id"`
	Timeouts             timeouts.Value `tfsdk:"timeouts"`
	Type                 types.String   `tfsdk:"type"`
	BluetoothMacAddress  types.String   `tfsdk:"bluetooth_mac_address"`
	DeviceEraseStatus    types.String   `tfsdk:"device_erase_status"`
	DeviceLockStatus     types.String   `tfsdk:"device_lock_status"`
	DeviceModel          types.String   `tfsdk:"device_model"`
	DeviceName           types.String   `tfsdk:"device_name"`
	EthernetMacAddress   types.String   `tfsdk:"ethernet_mac_address"`
	IMEI                 []types.String `tfsdk:"imei"`
	IsFileVaultEnabled   types.Bool     `tfsdk:"is_filevault_enabled"`
	IsFirewallEnabled    types.Bool     `tfsdk:"is_firewall_enabled"`
	LastCheckInDateTime  types.String   `tfsdk:"last_check_in_date_time"`
	LostModeStatus       types.String   `tfsdk:"lost_mode_status"`
	MEID                 []types.String `tfsdk:"meid"`
	OsVersion            types.String   `tfsdk:"os_version"`
	Platform             types.String   `tfsdk:"platform"`
	SerialNumber         types.String   `tfsdk:"serial_number"`
	StorageFreeCapacity  types.Int64    `tfsdk:"storage_free_capacity"`
	StorageTotalCapacity types.Int64    `tfsdk:"storage_total_capacity"`
	WifiMacAddress       types.String   `tfsdk:"wifi_mac_address"`
}

func (d *AppleDeviceManagementDeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apple_device_management_device"
}

func (d *AppleDeviceManagementDeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches detailed information about a specific Apple device enrolled in a device management service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The opaque resource ID that uniquely identifies the resource.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the resource.",
			},
			"bluetooth_mac_address": schema.StringAttribute{
				Computed:    true,
				Description: "The Bluetooth MAC address of the device.",
			},
			"device_erase_status": schema.StringAttribute{
				Computed:    true,
				Description: "The erase status of the device.",
			},
			"device_lock_status": schema.StringAttribute{
				Computed:    true,
				Description: "The lock status of the device.",
			},
			"device_model": schema.StringAttribute{
				Computed:    true,
				Description: "The model of the device.",
			},
			"device_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the device.",
			},
			"ethernet_mac_address": schema.StringAttribute{
				Computed:    true,
				Description: "The Ethernet MAC address of the device.",
			},
			"imei": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The IMEI numbers of the device.",
			},
			"is_filevault_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "A Boolean value indicating whether FileVault is enabled on the device.",
			},
			"is_firewall_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "A Boolean value indicating whether the firewall is enabled on the device.",
			},
			"last_check_in_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the device last checked in.",
			},
			"lost_mode_status": schema.StringAttribute{
				Computed:    true,
				Description: "The lost mode status of the device.",
			},
			"meid": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The Mobile Equipment Identifier numbers of the device.",
			},
			"os_version": schema.StringAttribute{
				Computed:    true,
				Description: "The operating system version of the device.",
			},
			"platform": schema.StringAttribute{
				Computed:    true,
				Description: "The platform of the device.",
			},
			"serial_number": schema.StringAttribute{
				Computed:    true,
				Description: "The serial number of the device.",
			},
			"storage_free_capacity": schema.Int64Attribute{
				Computed:    true,
				Description: "The free storage capacity of the device, in bytes.",
			},
			"storage_total_capacity": schema.Int64Attribute{
				Computed:    true,
				Description: "The total storage capacity of the device, in bytes.",
			},
			"wifi_mac_address": schema.StringAttribute{
				Computed:    true,
				Description: "The Wi-Fi MAC address of the device.",
			},
		},
	}
}

func (d *AppleDeviceManagementDeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = c
}

func (d *AppleDeviceManagementDeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppleDeviceManagementDeviceDataSourceModel

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

	detail, err := d.client.GetMdmDeviceDetail(readCtx, data.ID.ValueString(), nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Apple Device Management Device",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(detail.ID)
	data.Type = types.StringValue(detail.Type)
	data.BluetoothMacAddress = types.StringValue(detail.Attributes.BluetoothMacAddress)
	data.DeviceEraseStatus = types.StringValue(detail.Attributes.DeviceEraseStatus)
	data.DeviceLockStatus = types.StringValue(detail.Attributes.DeviceLockStatus)
	data.DeviceModel = types.StringValue(detail.Attributes.DeviceModel)
	data.DeviceName = types.StringValue(detail.Attributes.DeviceName)
	data.EthernetMacAddress = types.StringValue(detail.Attributes.EthernetMacAddress)
	data.IMEI = common.StringsToTypesStrings(detail.Attributes.IMEI)
	data.IsFileVaultEnabled = types.BoolPointerValue(detail.Attributes.IsFileVaultEnabled)
	data.IsFirewallEnabled = types.BoolPointerValue(detail.Attributes.IsFirewallEnabled)
	data.LastCheckInDateTime = types.StringValue(detail.Attributes.LastCheckInDateTime)
	data.LostModeStatus = types.StringValue(detail.Attributes.LostModeStatus)
	data.MEID = common.StringsToTypesStrings(detail.Attributes.MEID)
	data.OsVersion = types.StringValue(detail.Attributes.OsVersion)
	data.Platform = types.StringValue(detail.Attributes.Platform)
	data.SerialNumber = types.StringValue(detail.Attributes.SerialNumber)
	data.StorageFreeCapacity = types.Int64PointerValue(detail.Attributes.StorageFreeCapacity)
	data.StorageTotalCapacity = types.Int64PointerValue(detail.Attributes.StorageTotalCapacity)
	data.WifiMacAddress = types.StringValue(detail.Attributes.WifiMacAddress)

	tflog.Debug(ctx, "Read apple device management device", map[string]any{
		"device_id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
