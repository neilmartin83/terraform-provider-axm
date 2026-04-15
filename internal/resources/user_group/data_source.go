// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package user_group

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

var _ datasource.DataSource = &UserGroupDataSource{}

// NewUserGroupDataSource returns a new data source for a single user group.
func NewUserGroupDataSource() datasource.DataSource {
	return &UserGroupDataSource{}
}

// UserGroupDataSource defines the data source implementation.
type UserGroupDataSource struct {
	client *client.Client
}

// UserGroupDataSourceModel describes the data source data model.
type UserGroupDataSourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
	Type             types.String   `tfsdk:"type"`
	OuID             types.String   `tfsdk:"ou_id"`
	Name             types.String   `tfsdk:"name"`
	GroupType        types.String   `tfsdk:"group_type"`
	TotalMemberCount types.Int64    `tfsdk:"total_member_count"`
	Status           types.String   `tfsdk:"status"`
	CreatedDateTime  types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime  types.String   `tfsdk:"updated_date_time"`
	UserIDs          []types.String `tfsdk:"user_ids"`
}

func (d *UserGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_group"
}

func (d *UserGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific user group from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The user group ID.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The resource type.",
			},
			"ou_id": schema.StringAttribute{
				Computed:    true,
				Description: "The organizational unit ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The user group name.",
			},
			"group_type": schema.StringAttribute{
				Computed:    true,
				Description: "The user group type.",
			},
			"total_member_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The total member count.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The user group status.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The created date and time.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The updated date and time.",
			},
			"user_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "User IDs associated with the group.",
			},
		},
	}
}

func (d *UserGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_user_group data source") {
		return
	}
	d.client = c
}

func (d *UserGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserGroupDataSourceModel

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

	group, err := d.client.GetUserGroup(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read user group", err.Error())
		return
	}

	userIDs, err := d.client.GetUserGroupUserIDs(readCtx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read user group users", err.Error())
		return
	}

	data.Type = types.StringValue(group.Type)
	data.OuID = types.StringPointerValue(common.StringPointerOrNil(group.Attributes.OuID))
	data.Name = types.StringValue(group.Attributes.Name)
	data.GroupType = types.StringValue(group.Attributes.Type)
	data.TotalMemberCount = types.Int64Value(int64(group.Attributes.TotalMemberCount))
	data.Status = types.StringValue(group.Attributes.Status)
	data.CreatedDateTime = types.StringPointerValue(common.StringPointerOrNil(group.Attributes.CreatedDateTime))
	data.UpdatedDateTime = types.StringPointerValue(common.StringPointerOrNil(group.Attributes.UpdatedDateTime))
	data.UserIDs = common.StringsToTypesStrings(userIDs)

	tflog.Debug(ctx, "Read user group", map[string]any{
		"user_group_id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
