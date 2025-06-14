package axm

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &OrganizationDevicesDataSource{}

type OrganizationDevicesDataSource struct {
	client *Client
}

type OrganizationDevicesDataSourceModel struct {
	ID      types.String              `tfsdk:"id"`
	Devices []OrganizationDeviceModel `tfsdk:"devices"`
}

type OrganizationDeviceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Type               types.String   `tfsdk:"type"`
	SerialNumber       types.String   `tfsdk:"serial_number"`
	AddedDateTime      types.String   `tfsdk:"added_to_org_date_time"`
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

func NewOrganizationDevicesDataSource() datasource.DataSource {
	return &OrganizationDevicesDataSource{}
}

func (d *OrganizationDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_devices"
}

func (d *OrganizationDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of devices from Apple Business or School Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier of the data source.",
				Computed:    true,
			},
			"devices": schema.ListNestedAttribute{
				Description: "List of organization devices.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Device identifier.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Device type.",
							Computed:    true,
						},
						"serial_number": schema.StringAttribute{
							Description: "Device serial number.",
							Computed:    true,
						},
						"added_to_org_date_time": schema.StringAttribute{
							Description: "Date and time when device was added to organization.",
							Computed:    true,
						},
						"updated_date_time": schema.StringAttribute{
							Description: "Last update date and time.",
							Computed:    true,
						},
						"device_model": schema.StringAttribute{
							Description: "Device model.",
							Computed:    true,
						},
						"product_family": schema.StringAttribute{
							Description: "Product family.",
							Computed:    true,
						},
						"product_type": schema.StringAttribute{
							Description: "Product type.",
							Computed:    true,
						},
						"device_capacity": schema.StringAttribute{
							Description: "Device capacity.",
							Computed:    true,
						},
						"part_number": schema.StringAttribute{
							Description: "Part number.",
							Computed:    true,
						},
						"order_number": schema.StringAttribute{
							Description: "Order number.",
							Computed:    true,
						},
						"color": schema.StringAttribute{
							Description: "Device color.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Device status.",
							Computed:    true,
						},
						"order_date_time": schema.StringAttribute{
							Description: "Order date and time.",
							Computed:    true,
						},
						"imei": schema.ListAttribute{
							Description: "IMEI numbers.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"meid": schema.ListAttribute{
							Description: "MEID numbers.",
							Computed:    true,
							ElementType: types.StringType,
						},
						"eid": schema.StringAttribute{
							Description: "EID number.",
							Computed:    true,
						},
						"purchase_source_id": schema.StringAttribute{
							Description: "Purchase source identifier.",
							Computed:    true,
						},
						"purchase_source_type": schema.StringAttribute{
							Description: "Purchase source type.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationDevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *Client, got: unknown. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = client
}

func (d *OrganizationDevicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OrganizationDevicesDataSourceModel

	devices, err := d.client.GetOrgDevices(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Organization Devices",
			err.Error(),
		)
		return
	}

	state.Devices = make([]OrganizationDeviceModel, 0, len(devices))
	for _, device := range devices {
		deviceModel := OrganizationDeviceModel{
			ID:                 types.StringValue(device.ID),
			Type:               types.StringValue(device.Type),
			SerialNumber:       types.StringValue(device.Attributes.SerialNumber),
			AddedDateTime:      types.StringValue(device.Attributes.AddedToOrgDateTime),
			UpdatedDateTime:    types.StringValue(device.Attributes.UpdatedDateTime),
			DeviceModel:        types.StringValue(device.Attributes.DeviceModel),
			ProductFamily:      types.StringValue(device.Attributes.ProductFamily),
			ProductType:        types.StringValue(device.Attributes.ProductType),
			DeviceCapacity:     types.StringValue(device.Attributes.DeviceCapacity),
			PartNumber:         types.StringValue(device.Attributes.PartNumber),
			OrderNumber:        types.StringValue(device.Attributes.OrderNumber),
			Color:              types.StringValue(device.Attributes.Color),
			Status:             types.StringValue(device.Attributes.Status),
			OrderDateTime:      types.StringValue(device.Attributes.OrderDateTime),
			EID:                types.StringValue(device.Attributes.EID),
			PurchaseSourceID:   types.StringValue(device.Attributes.PurchaseSourceID),
			PurchaseSourceType: types.StringValue(device.Attributes.PurchaseSourceType),
		}

		deviceModel.IMEI = make([]types.String, len(device.Attributes.IMEI))
		for i, imei := range device.Attributes.IMEI {
			deviceModel.IMEI[i] = types.StringValue(imei)
		}

		deviceModel.MEID = make([]types.String, len(device.Attributes.MEID))
		for i, meid := range device.Attributes.MEID {
			deviceModel.MEID[i] = types.StringValue(meid)
		}

		state.Devices = append(state.Devices, deviceModel)
	}

	state.ID = types.StringValue(time.Now().UTC().String())

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
