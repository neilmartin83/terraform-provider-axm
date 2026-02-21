package organization_devices

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

var _ datasource.DataSource = &OrganizationDevicesDataSource{}

func NewOrganizationDevicesDataSource() datasource.DataSource {
	return &OrganizationDevicesDataSource{}
}

// OrganizationDevicesDataSource defines the data source implementation.
type OrganizationDevicesDataSource struct {
	client *client.Client
}

// OrganizationDevicesDataSourceModel describes the data source data model.
type OrganizationDevicesDataSourceModel struct {
	ID       types.String              `tfsdk:"id"`
	Timeouts timeouts.Value            `tfsdk:"timeouts"`
	Devices  []OrganizationDeviceModel `tfsdk:"devices"`
}

// OrganizationDeviceModel describes an organization device.
type OrganizationDeviceModel struct {
	ID                  types.String   `tfsdk:"id"`
	Type                types.String   `tfsdk:"type"`
	SerialNumber        types.String   `tfsdk:"serial_number"`
	AddedDateTime       types.String   `tfsdk:"added_to_org_date_time"`
	ReleasedDateTime    types.String   `tfsdk:"released_from_org_date_time"`
	UpdatedDateTime     types.String   `tfsdk:"updated_date_time"`
	DeviceModel         types.String   `tfsdk:"device_model"`
	ProductFamily       types.String   `tfsdk:"product_family"`
	ProductType         types.String   `tfsdk:"product_type"`
	DeviceCapacity      types.String   `tfsdk:"device_capacity"`
	PartNumber          types.String   `tfsdk:"part_number"`
	OrderNumber         types.String   `tfsdk:"order_number"`
	Color               types.String   `tfsdk:"color"`
	Status              types.String   `tfsdk:"status"`
	OrderDateTime       types.String   `tfsdk:"order_date_time"`
	IMEI                []types.String `tfsdk:"imei"`
	MEID                []types.String `tfsdk:"meid"`
	EID                 types.String   `tfsdk:"eid"`
	PurchaseSourceID    types.String   `tfsdk:"purchase_source_id"`
	PurchaseSourceType  types.String   `tfsdk:"purchase_source_type"`
	WifiMacAddress      types.String   `tfsdk:"wifi_mac_address"`
	BluetoothMacAddress types.String   `tfsdk:"bluetooth_mac_address"`
	EthernetMacAddress  []types.String `tfsdk:"ethernet_mac_address"`
}

func (d *OrganizationDevicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_devices"
}

func (d *OrganizationDevicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of devices from Apple Business or School Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"devices": schema.ListNestedAttribute{
				Description: "List of organization devices.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The opaque resource ID that uniquely identifies the resource.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the device.",
						},
						"serial_number": schema.StringAttribute{
							Computed:    true,
							Description: "The device's serial number.",
						},
						"added_to_org_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time of adding the device to an organization.",
						},
						"released_from_org_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time the device was released from an organization. This will be null if the device hasn't been released. Currently only querying by a single device is supported. Batch device queries aren't currently supported for this property.",
						},
						"updated_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time of the most-recent update for the device.",
						},
						"device_model": schema.StringAttribute{
							Computed:    true,
							Description: "The model name.",
						},
						"product_family": schema.StringAttribute{
							Computed:    true,
							Description: "The device's Apple product family: iPhone, iPad,Mac, AppleTV, Watch, or Vision.",
						},
						"product_type": schema.StringAttribute{
							Computed:    true,
							Description: "The device's product type: (examples: iPhone14,3, iPad13,4, MacBookPro14,2).",
						},
						"device_capacity": schema.StringAttribute{
							Computed:    true,
							Description: "The capacity of the device.",
						},
						"part_number": schema.StringAttribute{
							Computed:    true,
							Description: "The part number of the device.",
						},
						"order_number": schema.StringAttribute{
							Computed:    true,
							Description: "The order number of the device.",
						},
						"color": schema.StringAttribute{
							Computed:    true,
							Description: "The color of the device.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "The device's status: ASSIGNED or UNASSIGNED. If ASSIGNED, use a separate API to get the information of the assigned server.",
						},
						"order_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time of placing the device's order.",
						},
						"imei": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "The device's IMEI (if available).",
						},
						"meid": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "The device's MEID (if available).",
						},
						"eid": schema.StringAttribute{
							Computed:    true,
							Description: "The device's EID (if available).",
						},
						"purchase_source_id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique ID of the purchase source type: Apple Customer Number or Reseller Number.",
						},
						"purchase_source_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the purchase source.",
						},
						"wifi_mac_address": schema.StringAttribute{
							Description: "The device's Wi-Fi MAC address.",
							Computed:    true,
						},
						"bluetooth_mac_address": schema.StringAttribute{
							Description: "The device's Bluetooth MAC address.",
							Computed:    true,
						},
						"ethernet_mac_address": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "The device's built-in Ethernet MAC addresses.",
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationDevicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = c
}

func (d *OrganizationDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDevicesDataSourceModel

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

	devices, err := d.client.GetOrgDevices(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Organization Devices",
			err.Error(),
		)
		return
	}

	data.Devices = make([]OrganizationDeviceModel, 0, len(devices))
	for _, device := range devices {
		deviceModel := OrganizationDeviceModel{
			ID:                  types.StringValue(device.ID),
			Type:                types.StringValue(device.Type),
			SerialNumber:        types.StringValue(device.Attributes.SerialNumber),
			AddedDateTime:       types.StringValue(device.Attributes.AddedToOrgDateTime),
			ReleasedDateTime:    types.StringValue(device.Attributes.ReleasedFromOrgDateTime),
			UpdatedDateTime:     types.StringValue(device.Attributes.UpdatedDateTime),
			DeviceModel:         types.StringValue(device.Attributes.DeviceModel),
			ProductFamily:       types.StringValue(device.Attributes.ProductFamily),
			ProductType:         types.StringValue(device.Attributes.ProductType),
			DeviceCapacity:      types.StringValue(device.Attributes.DeviceCapacity),
			PartNumber:          types.StringValue(device.Attributes.PartNumber),
			OrderNumber:         types.StringValue(device.Attributes.OrderNumber),
			Color:               types.StringValue(device.Attributes.Color),
			Status:              types.StringValue(device.Attributes.Status),
			OrderDateTime:       types.StringValue(device.Attributes.OrderDateTime),
			EID:                 types.StringValue(device.Attributes.EID),
			PurchaseSourceID:    types.StringValue(device.Attributes.PurchaseSourceID),
			PurchaseSourceType:  types.StringValue(device.Attributes.PurchaseSourceType),
			WifiMacAddress:      types.StringValue(device.Attributes.WifiMacAddress),
			BluetoothMacAddress: types.StringValue(device.Attributes.BluetoothMacAddress),
			EthernetMacAddress:  common.StringsToTypesStrings(device.Attributes.EthernetMacAddress),
			IMEI:                common.StringsToTypesStrings(device.Attributes.IMEI),
			MEID:                common.StringsToTypesStrings(device.Attributes.MEID),
		}

		data.Devices = append(data.Devices, deviceModel)
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read organization devices", map[string]any{
		"device_count": len(data.Devices),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
