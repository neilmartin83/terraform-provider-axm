// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package user_groups

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

var _ datasource.DataSource = &UserGroupsDataSource{}

// NewUserGroupsDataSource returns a new data source for all user groups.
func NewUserGroupsDataSource() datasource.DataSource {
	return &UserGroupsDataSource{}
}

// UserGroupsDataSource defines the data source implementation.
type UserGroupsDataSource struct {
	client *client.Client
}

// UserGroupsDataSourceModel describes the data source data model.
type UserGroupsDataSourceModel struct {
	ID       types.String     `tfsdk:"id"`
	Timeouts timeouts.Value   `tfsdk:"timeouts"`
	Groups   []UserGroupModel `tfsdk:"groups"`
}

// UserGroupModel describes a user group.
type UserGroupModel struct {
	ID               types.String `tfsdk:"id"`
	Type             types.String `tfsdk:"type"`
	OuID             types.String `tfsdk:"ou_id"`
	Name             types.String `tfsdk:"name"`
	GroupType        types.String `tfsdk:"group_type"`
	TotalMemberCount types.Int64  `tfsdk:"total_member_count"`
	Status           types.String `tfsdk:"status"`
	CreatedDateTime  types.String `tfsdk:"created_date_time"`
	UpdatedDateTime  types.String `tfsdk:"updated_date_time"`
}

func (d *UserGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_groups"
}

func (d *UserGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of user groups from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"groups": schema.ListNestedAttribute{
				Description: "List of user groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The user group ID.",
						},
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
					},
				},
			},
		},
	}
}

func (d *UserGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_user_groups data source") {
		return
	}
	d.client = c
}

func (d *UserGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserGroupsDataSourceModel

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

	groups, err := d.client.GetUserGroups(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read user groups", err.Error())
		return
	}

	data.Groups = make([]UserGroupModel, 0, len(groups))
	for _, group := range groups {
		data.Groups = append(data.Groups, UserGroupModel{
			ID:               types.StringValue(group.ID),
			Type:             types.StringValue(group.Type),
			OuID:             types.StringPointerValue(common.StringPointerOrNil(group.Attributes.OuID)),
			Name:             types.StringValue(group.Attributes.Name),
			GroupType:        types.StringValue(group.Attributes.Type),
			TotalMemberCount: types.Int64Value(int64(group.Attributes.TotalMemberCount)),
			Status:           types.StringValue(group.Attributes.Status),
			CreatedDateTime:  types.StringPointerValue(common.StringPointerOrNil(group.Attributes.CreatedDateTime)),
			UpdatedDateTime:  types.StringPointerValue(common.StringPointerOrNil(group.Attributes.UpdatedDateTime)),
		})
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read user groups", map[string]any{
		"group_count": len(data.Groups),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
