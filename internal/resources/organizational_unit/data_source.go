// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organizational_unit

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

var _ datasource.DataSource = &OrganizationalUnitDataSource{}

func NewOrganizationalUnitDataSource() datasource.DataSource {
	return &OrganizationalUnitDataSource{}
}

type OrganizationalUnitDataSource struct {
	client *client.Client
}

type OrganizationalUnitDataSourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
	Type            types.String   `tfsdk:"type"`
	Name            types.String   `tfsdk:"name"`
	Description     types.String   `tfsdk:"description"`
	CreatedDateTime types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime types.String   `tfsdk:"updated_date_time"`
	UserIDs         []types.String `tfsdk:"user_ids"`
}

func (d *OrganizationalUnitDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizational_unit"
}

func (d *OrganizationalUnitDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific organizational unit from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The organizational unit ID.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The resource type.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the organizational unit.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A description of the organizational unit.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the organizational unit was created.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the organizational unit was last modified.",
			},
			"user_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "User IDs associated with the organizational unit.",
			},
		},
	}
}

func (d *OrganizationalUnitDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_organizational_unit data source") {
		return
	}
	d.client = c
}

func (d *OrganizationalUnitDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationalUnitDataSourceModel

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

	ou, err := d.client.GetOrganizationalUnit(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organizational unit", err.Error())
		return
	}

	userIDs, err := d.client.GetOrganizationalUnitUserIDs(readCtx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organizational unit users", err.Error())
		return
	}

	data.Type = types.StringValue(ou.Type)
	data.Name = types.StringValue(ou.Attributes.Name)
	data.Description = types.StringPointerValue(common.StringPointerOrNil(ou.Attributes.Description))
	data.CreatedDateTime = types.StringPointerValue(common.StringPointerOrNil(ou.Attributes.CreatedDateTime))
	data.UpdatedDateTime = types.StringPointerValue(common.StringPointerOrNil(ou.Attributes.UpdatedDateTime))
	data.UserIDs = common.StringsToTypesStrings(userIDs)

	tflog.Debug(ctx, "Read organizational unit", map[string]any{
		"organizational_unit_id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
