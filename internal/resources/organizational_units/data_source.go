// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package organizational_units

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

var _ datasource.DataSource = &OrganizationalUnitsDataSource{}

func NewOrganizationalUnitsDataSource() datasource.DataSource {
	return &OrganizationalUnitsDataSource{}
}

type OrganizationalUnitsDataSource struct {
	client *client.Client
}

type OrganizationalUnitsDataSourceModel struct {
	ID                  types.String              `tfsdk:"id"`
	Timeouts            timeouts.Value            `tfsdk:"timeouts"`
	OrganizationalUnits []OrganizationalUnitModel `tfsdk:"organizational_units"`
}

type OrganizationalUnitModel struct {
	ID              types.String `tfsdk:"id"`
	Type            types.String `tfsdk:"type"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
	CreatedDateTime types.String `tfsdk:"created_date_time"`
	UpdatedDateTime types.String `tfsdk:"updated_date_time"`
}

func (d *OrganizationalUnitsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizational_units"
}

func (d *OrganizationalUnitsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of organizational units from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"organizational_units": schema.ListNestedAttribute{
				Description: "List of organizational units.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The organizational unit ID.",
						},
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
					},
				},
			},
		},
	}
}

func (d *OrganizationalUnitsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_organizational_units data source") {
		return
	}
	d.client = c
}

func (d *OrganizationalUnitsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationalUnitsDataSourceModel

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

	ous, err := d.client.GetOrganizationalUnits(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read organizational units", err.Error())
		return
	}

	data.OrganizationalUnits = make([]OrganizationalUnitModel, 0, len(ous))
	for _, ou := range ous {
		data.OrganizationalUnits = append(data.OrganizationalUnits, OrganizationalUnitModel{
			ID:              types.StringValue(ou.ID),
			Type:            types.StringValue(ou.Type),
			Name:            types.StringValue(ou.Attributes.Name),
			Description:     types.StringPointerValue(common.StringPointerOrNil(ou.Attributes.Description)),
			CreatedDateTime: types.StringPointerValue(common.StringPointerOrNil(ou.Attributes.CreatedDateTime)),
			UpdatedDateTime: types.StringPointerValue(common.StringPointerOrNil(ou.Attributes.UpdatedDateTime)),
		})
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read organizational units", map[string]any{
		"organizational_unit_count": len(data.OrganizationalUnits),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
