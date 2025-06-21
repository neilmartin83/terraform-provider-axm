package organization_device

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ datasource.DataSource = &OrganizationDeviceDataSource{}

type OrganizationDeviceDataSource struct {
	client *client.Client
}

type OrganizationDeviceDataSourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Type               types.String   `tfsdk:"type"`
	SerialNumber       types.String   `tfsdk:"serial_number"`
	AddedToOrgDateTime types.String   `tfsdk:"added_to_org_date_time"`
	UpdatedDateTime    types.String   `tfsdk:"updated_date_time"`
	DeviceModel        types.String   `tfsdk:"device_model"`
	ProductFamily      types.String   `tfsdk:"product_family"`
	ProductType        types.String   `tfsdk:"product_type"`
	DeviceCapacity     types.String   `tfsdk:"device_capacity"`
	PartNumber         types.String   `tfsdk:"part_number"`
	OrderNumber        types.String   `tfsdk:"order_number"`
	Color              types.String   `tfsdk:"color"`
	Status             types.String   `tfsdk:"status"`
	OrderDateTime      types.String   `tfsdk:"order_date_time"`
	IMEI               []types.String `tfsdk:"imei"`
	MEID               []types.String `tfsdk:"meid"`
	EID                types.String   `tfsdk:"eid"`
	PurchaseSourceID   types.String   `tfsdk:"purchase_source_id"`
	PurchaseSourceType types.String   `tfsdk:"purchase_source_type"`
}

func NewOrganizationDeviceDataSource() datasource.DataSource {
	return &OrganizationDeviceDataSource{}
}

func (d *OrganizationDeviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_device"
}

func (d *OrganizationDeviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific device from Apple Business or School Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The identifier of the device to lookup.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the device.",
			},
			"serial_number": schema.StringAttribute{
				Computed:    true,
				Description: "The serial number of the device.",
			},
			"added_to_org_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the device was added to the organization.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the device was last updated.",
			},
			"device_model": schema.StringAttribute{
				Computed:    true,
				Description: "The model of the device.",
			},
			"product_family": schema.StringAttribute{
				Computed:    true,
				Description: "The product family of the device.",
			},
			"product_type": schema.StringAttribute{
				Computed:    true,
				Description: "The product type of the device.",
			},
			"device_capacity": schema.StringAttribute{
				Computed:    true,
				Description: "The storage capacity of the device.",
			},
			"part_number": schema.StringAttribute{
				Computed:    true,
				Description: "The part number of the device.",
			},
			"order_number": schema.StringAttribute{
				Computed:    true,
				Description: "The order number associated with the device.",
			},
			"color": schema.StringAttribute{
				Computed:    true,
				Description: "The color of the device.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the device.",
			},
			"order_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time when the device was ordered.",
			},
			"imei": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The IMEI numbers associated with the device.",
			},
			"meid": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The MEID numbers associated with the device.",
			},
			"eid": schema.StringAttribute{
				Computed:    true,
				Description: "The EID of the device.",
			},
			"purchase_source_id": schema.StringAttribute{
				Computed:    true,
				Description: "The identifier of the purchase source.",
			},
			"purchase_source_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the purchase source.",
			},
		},
	}
}

func (d *OrganizationDeviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	var state OrganizationDeviceDataSourceModel

	var config OrganizationDeviceDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	device, err := d.client.GetOrgDevice(ctx, config.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Organization Device",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue(device.ID)
	state.Type = types.StringValue(device.Type)
	state.SerialNumber = types.StringValue(device.Attributes.SerialNumber)
	state.AddedToOrgDateTime = types.StringValue(device.Attributes.AddedToOrgDateTime)
	state.UpdatedDateTime = types.StringValue(device.Attributes.UpdatedDateTime)
	state.DeviceModel = types.StringValue(device.Attributes.DeviceModel)
	state.ProductFamily = types.StringValue(device.Attributes.ProductFamily)
	state.ProductType = types.StringValue(device.Attributes.ProductType)
	state.DeviceCapacity = types.StringValue(device.Attributes.DeviceCapacity)
	state.PartNumber = types.StringValue(device.Attributes.PartNumber)
	state.OrderNumber = types.StringValue(device.Attributes.OrderNumber)
	state.Color = types.StringValue(device.Attributes.Color)
	state.Status = types.StringValue(device.Attributes.Status)
	state.OrderDateTime = types.StringValue(device.Attributes.OrderDateTime)
	state.EID = types.StringValue(device.Attributes.EID)
	state.PurchaseSourceID = types.StringValue(device.Attributes.PurchaseSourceID)
	state.PurchaseSourceType = types.StringValue(device.Attributes.PurchaseSourceType)

	state.IMEI = make([]types.String, len(device.Attributes.IMEI))
	for i, imei := range device.Attributes.IMEI {
		state.IMEI[i] = types.StringValue(imei)
	}

	state.MEID = make([]types.String, len(device.Attributes.MEID))
	for i, meid := range device.Attributes.MEID {
		state.MEID[i] = types.StringValue(meid)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
