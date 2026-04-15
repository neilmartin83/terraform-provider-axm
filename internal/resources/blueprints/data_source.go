// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprints

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

var _ datasource.DataSource = &BlueprintsDataSource{}

// NewBlueprintsDataSource returns a new data source for all Blueprints.
func NewBlueprintsDataSource() datasource.DataSource {
	return &BlueprintsDataSource{}
}

// BlueprintsDataSource defines the data source implementation.
type BlueprintsDataSource struct {
	client *client.Client
}

// BlueprintsDataSourceModel describes the data source data model.
type BlueprintsDataSourceModel struct {
	ID         types.String     `tfsdk:"id"`
	Timeouts   timeouts.Value   `tfsdk:"timeouts"`
	Blueprints []BlueprintModel `tfsdk:"blueprints"`
}

// BlueprintModel describes a Blueprint in the list.
type BlueprintModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	Status              types.String `tfsdk:"status"`
	AppLicenseDeficient types.Bool   `tfsdk:"app_license_deficient"`
	CreatedDateTime     types.String `tfsdk:"created_date_time"`
	UpdatedDateTime     types.String `tfsdk:"updated_date_time"`
}

func (d *BlueprintsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprints"
}

func (d *BlueprintsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of Blueprints from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"blueprints": schema.ListNestedAttribute{
				Description: "List of Blueprints.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The Blueprint ID.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The Blueprint name.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "The Blueprint description.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "The Blueprint status.",
						},
						"app_license_deficient": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the Blueprint is missing app licenses.",
						},
						"created_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time the Blueprint was created.",
						},
						"updated_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The date and time the Blueprint was last updated.",
						},
					},
				},
			},
		},
	}
}

func (d *BlueprintsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_blueprints data source") {
		return
	}
	d.client = c
}

func (d *BlueprintsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BlueprintsDataSourceModel

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

	bps, err := d.client.GetBlueprints(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Blueprints", err.Error())
		return
	}

	data.Blueprints = make([]BlueprintModel, 0, len(bps))
	for _, bp := range bps {
		data.Blueprints = append(data.Blueprints, BlueprintModel{
			ID:                  types.StringValue(bp.ID),
			Name:                types.StringValue(bp.Attributes.Name),
			Description:         types.StringPointerValue(common.StringPointerOrNil(bp.Attributes.Description)),
			Status:              types.StringValue(bp.Attributes.Status),
			AppLicenseDeficient: types.BoolValue(bp.Attributes.AppLicenseDeficient),
			CreatedDateTime:     types.StringValue(bp.Attributes.CreatedDateTime),
			UpdatedDateTime:     types.StringValue(bp.Attributes.UpdatedDateTime),
		})
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read blueprints", map[string]any{
		"blueprint_count": len(data.Blueprints),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
