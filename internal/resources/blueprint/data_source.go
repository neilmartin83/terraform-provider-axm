// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ datasource.DataSource = &BlueprintDataSource{}

// NewBlueprintDataSource returns a new data source for a single Blueprint.
func NewBlueprintDataSource() datasource.DataSource {
	return &BlueprintDataSource{}
}

// BlueprintDataSource defines the data source implementation.
type BlueprintDataSource struct {
	client *client.Client
}

// BlueprintDataSourceModel describes the data source data model.
type BlueprintDataSourceModel struct {
	ID                  types.String   `tfsdk:"id"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
	Name                types.String   `tfsdk:"name"`
	Description         types.String   `tfsdk:"description"`
	Status              types.String   `tfsdk:"status"`
	AppLicenseDeficient types.Bool     `tfsdk:"app_license_deficient"`
	CreatedDateTime     types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime     types.String   `tfsdk:"updated_date_time"`
	AppIDs              types.Set      `tfsdk:"app_ids"`
	ConfigurationIDs    types.Set      `tfsdk:"configuration_ids"`
	PackageIDs          types.Set      `tfsdk:"package_ids"`
	DeviceIDs           types.Set      `tfsdk:"device_ids"`
	UserIDs             types.Set      `tfsdk:"user_ids"`
	UserGroupIDs        types.Set      `tfsdk:"user_group_ids"`
}

func (d *BlueprintDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (d *BlueprintDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific Blueprint from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The Blueprint ID.",
			},
			"timeouts": timeouts.Attributes(ctx),
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
			"app_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "App IDs associated with the Blueprint.",
			},
			"configuration_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Configuration IDs associated with the Blueprint.",
			},
			"package_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Package IDs associated with the Blueprint.",
			},
			"device_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Device IDs associated with the Blueprint.",
			},
			"user_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "User IDs associated with the Blueprint.",
			},
			"user_group_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "User group IDs associated with the Blueprint.",
			},
		},
	}
}

func (d *BlueprintDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_blueprint data source") {
		return
	}
	d.client = c
}

func (d *BlueprintDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BlueprintDataSourceModel

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

	blueprintID := data.ID.ValueString()

	bp, err := d.client.GetBlueprint(readCtx, blueprintID, nil)
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			resp.Diagnostics.AddError("Blueprint not found", fmt.Sprintf("Blueprint with ID %q not found.", blueprintID))
			return
		}
		resp.Diagnostics.AddError("Unable to read Blueprint", err.Error())
		return
	}

	data.Name = types.StringValue(bp.Attributes.Name)
	data.Description = types.StringPointerValue(common.StringPointerOrNil(bp.Attributes.Description))
	data.Status = types.StringValue(bp.Attributes.Status)
	data.AppLicenseDeficient = types.BoolValue(bp.Attributes.AppLicenseDeficient)
	data.CreatedDateTime = types.StringValue(bp.Attributes.CreatedDateTime)
	data.UpdatedDateTime = types.StringValue(bp.Attributes.UpdatedDateTime)

	relationships := []struct {
		name string
		dest *types.Set
	}{
		{relationshipApps, &data.AppIDs},
		{relationshipConfigurations, &data.ConfigurationIDs},
		{relationshipPackages, &data.PackageIDs},
		{relationshipOrgDevices, &data.DeviceIDs},
		{relationshipUsers, &data.UserIDs},
		{relationshipUserGroups, &data.UserGroupIDs},
	}

	for _, rel := range relationships {
		ids, err := d.client.GetBlueprintRelationshipIDs(readCtx, blueprintID, rel.name)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Unable to read Blueprint %s", rel.name),
				err.Error(),
			)
			return
		}

		set, setDiags := common.StringsToSet(ids)
		if setDiags.HasError() {
			resp.Diagnostics.Append(setDiags...)
			return
		}
		*rel.dest = set
	}

	tflog.Debug(ctx, "Read blueprint", map[string]any{
		"blueprint_id": blueprintID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
