package organization_device_applecare_coverage

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

var _ datasource.DataSource = &OrganizationDeviceAppleCareCoverageDataSource{}

// NewOrganizationDeviceAppleCareCoverageDataSource returns a new data source for device AppleCare coverage.
func NewOrganizationDeviceAppleCareCoverageDataSource() datasource.DataSource {
	return &OrganizationDeviceAppleCareCoverageDataSource{}
}

// OrganizationDeviceAppleCareCoverageDataSource defines the data source implementation.
type OrganizationDeviceAppleCareCoverageDataSource struct {
	client *client.Client
}

// OrganizationDeviceAppleCareCoverageDataSourceModel describes the data source data model.
type OrganizationDeviceAppleCareCoverageDataSourceModel struct {
	ID                         types.String                               `tfsdk:"id"`
	Timeouts                   timeouts.Value                             `tfsdk:"timeouts"`
	AppleCareCoverageResources []OrganizationDeviceAppleCareCoverageModel `tfsdk:"applecare_coverage_resources"`
}

// OrganizationDeviceAppleCareCoverageModel describes an AppleCare coverage resource.
type OrganizationDeviceAppleCareCoverageModel struct {
	ID                     types.String `tfsdk:"id"`
	AgreementNumber        types.String `tfsdk:"agreement_number"`
	ContractCancelDateTime types.String `tfsdk:"contract_cancel_date_time"`
	Description            types.String `tfsdk:"description"`
	EndDateTime            types.String `tfsdk:"end_date_time"`
	IsCanceled             types.Bool   `tfsdk:"is_canceled"`
	IsRenewable            types.Bool   `tfsdk:"is_renewable"`
	PaymentType            types.String `tfsdk:"payment_type"`
	StartDateTime          types.String `tfsdk:"start_date_time"`
	Status                 types.String `tfsdk:"status"`
}

func (d *OrganizationDeviceAppleCareCoverageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_device_applecare_coverage"
}

func (d *OrganizationDeviceAppleCareCoverageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about the AppleCare coverage for a specific device.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Device Identifier.",
				Required:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"applecare_coverage_resources": schema.ListNestedAttribute{
				Description: "List of AppleCare coverage resources associated with the device.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The opaque resource ID that uniquely identifies the resource.",
							Computed:    true,
						},
						"agreement_number": schema.StringAttribute{
							Description: "Agreement number associated with device coverage. This field isn't applicable for Limited Warranty and AppleCare+ for Business Essentials.",
							Computed:    true,
						},
						"contract_cancel_date_time": schema.StringAttribute{
							Description: "UTC date when coverage was canceled for the device. This field isn't applicable for Limited Warranty and AppleCare+ for Business Essentials.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of device coverage.",
							Computed:    true,
						},
						"end_date_time": schema.StringAttribute{
							Description: "UTC date when coverage period ends for the device. This field isn't applicable for AppleCare+ for Business Essentials.",
							Computed:    true,
						},
						"is_canceled": schema.BoolAttribute{
							Description: "Indicates whether coverage is canceled for the device. This field isn't applicable for Limited Warranty and AppleCare+ for Business Essentials.",
							Computed:    true,
						},
						"is_renewable": schema.BoolAttribute{
							Description: "Indicates whether coverage renews after endDateTime for the device. This field isn't applicable for Limited Warranty.",
							Computed:    true,
						},
						"payment_type": schema.StringAttribute{
							Description: "Payment type of device coverage. Possible values: 'ABE_SUBSCRIPTION', 'PAID_UP_FRONT', 'SUBSCRIPTION', 'NONE'.",
							Computed:    true,
						},
						"start_date_time": schema.StringAttribute{
							Description: "UTC date when coverage period commenced. For AppleCare+ for Business Essentials, it's UTC date when a device enrolls into the plan.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of device coverage. Possible values: 'ACTIVE', 'INACTIVE'",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationDeviceAppleCareCoverageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	d.client = c
}

func (d *OrganizationDeviceAppleCareCoverageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDeviceAppleCareCoverageDataSourceModel

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

	applecarecoverage, err := d.client.GetOrgDeviceAppleCareCoverage(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Organization Device AppleCare Coverage",
			err.Error(),
		)
		return
	}

	data.AppleCareCoverageResources = make([]OrganizationDeviceAppleCareCoverageModel, 0, len(applecarecoverage))
	for _, coverage := range applecarecoverage {
		coverageModel := OrganizationDeviceAppleCareCoverageModel{
			ID:                     types.StringValue(coverage.ID),
			AgreementNumber:        types.StringValue(coverage.Attributes.AgreementNumber),
			ContractCancelDateTime: types.StringValue(coverage.Attributes.ContractCancelDateTime),
			Description:            types.StringValue(coverage.Attributes.Description),
			EndDateTime:            types.StringValue(coverage.Attributes.EndDateTime),
			IsCanceled:             types.BoolValue(coverage.Attributes.IsCanceled),
			IsRenewable:            types.BoolValue(coverage.Attributes.IsRenewable),
			PaymentType:            types.StringValue(coverage.Attributes.PaymentType),
			StartDateTime:          types.StringValue(coverage.Attributes.StartDateTime),
			Status:                 types.StringValue(coverage.Attributes.Status),
		}

		data.AppleCareCoverageResources = append(data.AppleCareCoverageResources, coverageModel)
	}

	tflog.Debug(ctx, "Read organization device applecare coverage information", map[string]any{
		"resource_count": len(data.AppleCareCoverageResources),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
