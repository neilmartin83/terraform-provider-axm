package organization_device_applecare_coverage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ datasource.DataSource = &OrganizationDeviceAppleCareCoverageDataSource{}

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
	AppleCareCoverageResources []OrganizationDeviceAppleCareCoverageModel `tfsdk:"applecare_coverage_resources"`
}

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
			"applecare_coverage_resources": schema.ListNestedAttribute{
				Description: "List of AppleCare coverage resources associated with the device.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier for the AppleCare coverage.",
							Computed:    true,
						},
						"agreement_number": schema.StringAttribute{
							Description: "The agreement number associated with the AppleCare coverage.",
							Computed:    true,
						},
						"contract_cancel_date_time": schema.StringAttribute{
							Description: "The date and time when the AppleCare contract was canceled, if applicable.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "A description of the AppleCare coverage.",
							Computed:    true,
						},
						"end_date_time": schema.StringAttribute{
							Description: "The end date and time of the AppleCare coverage.",
							Computed:    true,
						},
						"is_canceled": schema.BoolAttribute{
							Description: "Indicates whether the AppleCare coverage has been canceled.",
							Computed:    true,
						},
						"is_renewable": schema.BoolAttribute{
							Description: "Indicates whether the AppleCare coverage is renewable.",
							Computed:    true,
						},
						"payment_type": schema.StringAttribute{
							Description: "The payment type for the AppleCare coverage.",
							Computed:    true,
						},
						"start_date_time": schema.StringAttribute{
							Description: "The start date and time of the AppleCare coverage.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The status of the AppleCare coverage.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationDeviceAppleCareCoverageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDeviceAppleCareCoverageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDeviceAppleCareCoverageDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	applecarecoverage, err := d.client.GetOrgDeviceAppleCareCoverage(ctx, data.ID.ValueString(), nil)
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

	tflog.Debug(ctx, "Read organization device applecare coverage information", map[string]interface{}{
		"resource_count": len(data.AppleCareCoverageResources),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
