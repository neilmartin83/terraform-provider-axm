package organization_device

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ datasource.DataSource = &OrganizationDeviceDataSource{}

const defaultReadTimeout = 90 * time.Second

func NewOrganizationDeviceDataSource() datasource.DataSource {
	return &OrganizationDeviceDataSource{}
}

// OrganizationDeviceDataSource defines the data source implementation.
type OrganizationDeviceDataSource struct {
	client *client.Client
}

// OrganizationDeviceDataSourceModel describes the data source data model.
type OrganizationDeviceDataSourceModel struct {
	ID                      types.String   `tfsdk:"id"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
	Type                    types.String   `tfsdk:"type"`
	SerialNumber            types.String   `tfsdk:"serial_number"`
	AddedToOrgDateTime      types.String   `tfsdk:"added_to_org_date_time"`
	ReleasedFromOrgDateTime types.String   `tfsdk:"released_from_org_date_time"`
	UpdatedDateTime         types.String   `tfsdk:"updated_date_time"`
	DeviceModel             types.String   `tfsdk:"device_model"`
	ProductFamily           types.String   `tfsdk:"product_family"`
	ProductType             types.String   `tfsdk:"product_type"`
	DeviceCapacity          types.String   `tfsdk:"device_capacity"`
	PartNumber              types.String   `tfsdk:"part_number"`
	OrderNumber             types.String   `tfsdk:"order_number"`
	Color                   types.String   `tfsdk:"color"`
	Status                  types.String   `tfsdk:"status"`
	OrderDateTime           types.String   `tfsdk:"order_date_time"`
	IMEI                    []types.String `tfsdk:"imei"`
	MEID                    []types.String `tfsdk:"meid"`
	EID                     types.String   `tfsdk:"eid"`
	PurchaseSourceID        types.String   `tfsdk:"purchase_source_id"`
	PurchaseSourceType      types.String   `tfsdk:"purchase_source_type"`
	WifiMacAddress          types.String   `tfsdk:"wifi_mac_address"`
	BluetoothMacAddress     types.String   `tfsdk:"bluetooth_mac_address"`
	EthernetMacAddress      []types.String `tfsdk:"ethernet_mac_address"`
}

func (d *OrganizationDeviceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_device"
}

func (d *OrganizationDeviceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific device from Apple Business or School Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The opaque resource ID that uniquely identifies the resource.",
			},
			"timeouts": timeouts.Attributes(ctx),
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
				Description: "The date and time the device was released from an organization. This will be null if the device hasn't been released. Currently only querying by a single device is supported. Batch device queries aren’t currently supported for this property.",
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
	}
}

func (d *OrganizationDeviceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDeviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDeviceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout := defaultReadTimeout
	if !data.Timeouts.IsNull() && !data.Timeouts.IsUnknown() {
		configuredTimeout, timeoutDiags := data.Timeouts.Read(ctx, defaultReadTimeout)
		resp.Diagnostics.Append(timeoutDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		readTimeout = configuredTimeout
	}

	readCtx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	device, err := d.client.GetOrgDevice(readCtx, data.ID.ValueString(), nil)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Organization Device",
			err.Error(),
		)
		return
	}

	data.ID = types.StringValue(device.ID)
	data.Type = types.StringValue(device.Type)
	data.SerialNumber = types.StringValue(device.Attributes.SerialNumber)
	data.AddedToOrgDateTime = types.StringValue(device.Attributes.AddedToOrgDateTime)
	data.ReleasedFromOrgDateTime = types.StringPointerValue(stringPointerOrNil(device.Attributes.ReleasedFromOrgDateTime))
	data.UpdatedDateTime = types.StringValue(device.Attributes.UpdatedDateTime)
	data.DeviceModel = types.StringValue(device.Attributes.DeviceModel)
	data.ProductFamily = types.StringValue(device.Attributes.ProductFamily)
	data.ProductType = types.StringValue(device.Attributes.ProductType)
	data.DeviceCapacity = types.StringValue(device.Attributes.DeviceCapacity)
	data.PartNumber = types.StringValue(device.Attributes.PartNumber)
	data.OrderNumber = types.StringValue(device.Attributes.OrderNumber)
	data.Color = types.StringValue(device.Attributes.Color)
	data.Status = types.StringValue(device.Attributes.Status)
	data.OrderDateTime = types.StringValue(device.Attributes.OrderDateTime)
	data.EID = types.StringValue(device.Attributes.EID)
	data.PurchaseSourceID = types.StringValue(device.Attributes.PurchaseSourceID)
	data.PurchaseSourceType = types.StringValue(device.Attributes.PurchaseSourceType)
	data.WifiMacAddress = types.StringValue(device.Attributes.WifiMacAddress)
	data.BluetoothMacAddress = types.StringValue(device.Attributes.BluetoothMacAddress)

	data.EthernetMacAddress = make([]types.String, len(device.Attributes.EthernetMacAddress))
	for i, ethernetMacAddress := range device.Attributes.EthernetMacAddress {
		data.EthernetMacAddress[i] = types.StringValue(ethernetMacAddress)
	}

	data.IMEI = make([]types.String, len(device.Attributes.IMEI))
	for i, imei := range device.Attributes.IMEI {
		data.IMEI[i] = types.StringValue(imei)
	}

	data.MEID = make([]types.String, len(device.Attributes.MEID))
	for i, meid := range device.Attributes.MEID {
		data.MEID[i] = types.StringValue(meid)
	}

	tflog.Debug(ctx, "Read organization device", map[string]interface{}{
		"device_id":     data.ID.ValueString(),
		"serial_number": data.SerialNumber.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// stringPointerOrNil returns a pointer to the string if it's not empty, otherwise returns nil
func stringPointerOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
