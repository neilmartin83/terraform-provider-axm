// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package users

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

var _ datasource.DataSource = &UsersDataSource{}

// NewUsersDataSource returns a new data source for all users.
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// UsersDataSource defines the data source implementation.
type UsersDataSource struct {
	client *client.Client
}

// UsersDataSourceModel describes the data source data model.
type UsersDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
	Users    []UserModel    `tfsdk:"users"`
}

// UserModel describes a user.
type UserModel struct {
	ID                  types.String           `tfsdk:"id"`
	Type                types.String           `tfsdk:"type"`
	FirstName           types.String           `tfsdk:"first_name"`
	LastName            types.String           `tfsdk:"last_name"`
	MiddleName          types.String           `tfsdk:"middle_name"`
	Status              types.String           `tfsdk:"status"`
	ManagedAppleAccount types.String           `tfsdk:"managed_apple_account"`
	IsExternalUser      types.Bool             `tfsdk:"is_external_user"`
	RoleOuList          []UserRoleOuModel      `tfsdk:"role_ou_list"`
	Email               types.String           `tfsdk:"email"`
	EmployeeNumber      types.String           `tfsdk:"employee_number"`
	CostCenter          types.String           `tfsdk:"cost_center"`
	Division            types.String           `tfsdk:"division"`
	Department          types.String           `tfsdk:"department"`
	JobTitle            types.String           `tfsdk:"job_title"`
	StartDateTime       types.String           `tfsdk:"start_date_time"`
	CreatedDateTime     types.String           `tfsdk:"created_date_time"`
	UpdatedDateTime     types.String           `tfsdk:"updated_date_time"`
	PhoneNumbers        []UserPhoneNumberModel `tfsdk:"phone_numbers"`
}

// UserRoleOuModel describes a user role and organizational unit mapping.
type UserRoleOuModel struct {
	RoleName types.String `tfsdk:"role_name"`
	OuID     types.String `tfsdk:"ou_id"`
}

// UserPhoneNumberModel describes a user phone number.
type UserPhoneNumberModel struct {
	PhoneNumber types.String `tfsdk:"phone_number"`
	Type        types.String `tfsdk:"type"`
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of users from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"users": schema.ListNestedAttribute{
				Description: "List of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The user ID.",
						},
						"type": schema.StringAttribute{
							Computed:    true,
							Description: "The resource type.",
						},
						"first_name": schema.StringAttribute{
							Computed:    true,
							Description: "The user's first name.",
						},
						"last_name": schema.StringAttribute{
							Computed:    true,
							Description: "The user's last name.",
						},
						"middle_name": schema.StringAttribute{
							Computed:    true,
							Description: "The user's middle name.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "The user's status.",
						},
						"managed_apple_account": schema.StringAttribute{
							Computed:    true,
							Description: "The user's managed Apple account.",
						},
						"is_external_user": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the user is external.",
						},
						"role_ou_list": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Role and organizational unit mappings.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"role_name": schema.StringAttribute{
										Computed:    true,
										Description: "The role name.",
									},
									"ou_id": schema.StringAttribute{
										Computed:    true,
										Description: "The organizational unit ID.",
									},
								},
							},
						},
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "The user's email address.",
						},
						"employee_number": schema.StringAttribute{
							Computed:    true,
							Description: "The employee number.",
						},
						"cost_center": schema.StringAttribute{
							Computed:    true,
							Description: "The cost center.",
						},
						"division": schema.StringAttribute{
							Computed:    true,
							Description: "The division.",
						},
						"department": schema.StringAttribute{
							Computed:    true,
							Description: "The department.",
						},
						"job_title": schema.StringAttribute{
							Computed:    true,
							Description: "The job title.",
						},
						"start_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The start date and time.",
						},
						"created_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The created date and time.",
						},
						"updated_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The updated date and time.",
						},
						"phone_numbers": schema.ListNestedAttribute{
							Computed:    true,
							Description: "Phone numbers for the user.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"phone_number": schema.StringAttribute{
										Computed:    true,
										Description: "The phone number.",
									},
									"type": schema.StringAttribute{
										Computed:    true,
										Description: "The phone number type.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_users data source") {
		return
	}
	d.client = c
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel

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

	users, err := d.client.GetUsers(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read users", err.Error())
		return
	}

	data.Users = make([]UserModel, 0, len(users))
	for _, user := range users {
		data.Users = append(data.Users, flattenUser(user))
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read users", map[string]any{
		"user_count": len(data.Users),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func flattenUser(user client.User) UserModel {
	roleMappings := make([]UserRoleOuModel, 0, len(user.Attributes.RoleOuList))
	for _, role := range user.Attributes.RoleOuList {
		roleMappings = append(roleMappings, UserRoleOuModel{
			RoleName: types.StringValue(role.RoleName),
			OuID:     types.StringValue(role.OuID),
		})
	}

	phoneNumbers := make([]UserPhoneNumberModel, 0, len(user.Attributes.PhoneNumbers))
	for _, phone := range user.Attributes.PhoneNumbers {
		phoneNumbers = append(phoneNumbers, UserPhoneNumberModel{
			PhoneNumber: types.StringValue(phone.PhoneNumber),
			Type:        types.StringValue(phone.Type),
		})
	}

	return UserModel{
		ID:                  types.StringValue(user.ID),
		Type:                types.StringValue(user.Type),
		FirstName:           types.StringValue(user.Attributes.FirstName),
		LastName:            types.StringValue(user.Attributes.LastName),
		MiddleName:          types.StringPointerValue(common.StringPointerOrNil(user.Attributes.MiddleName)),
		Status:              types.StringValue(user.Attributes.Status),
		ManagedAppleAccount: types.StringValue(user.Attributes.ManagedAppleAccount),
		IsExternalUser:      types.BoolValue(user.Attributes.IsExternalUser),
		RoleOuList:          roleMappings,
		Email:               types.StringPointerValue(common.StringPointerOrNil(user.Attributes.Email)),
		EmployeeNumber:      types.StringPointerValue(common.StringPointerOrNil(user.Attributes.EmployeeNumber)),
		CostCenter:          types.StringPointerValue(common.StringPointerOrNil(user.Attributes.CostCenter)),
		Division:            types.StringPointerValue(common.StringPointerOrNil(user.Attributes.Division)),
		Department:          types.StringPointerValue(common.StringPointerOrNil(user.Attributes.Department)),
		JobTitle:            types.StringPointerValue(common.StringPointerOrNil(user.Attributes.JobTitle)),
		StartDateTime:       types.StringPointerValue(common.StringPointerOrNil(user.Attributes.StartDateTime)),
		CreatedDateTime:     types.StringPointerValue(common.StringPointerOrNil(user.Attributes.CreatedDateTime)),
		UpdatedDateTime:     types.StringPointerValue(common.StringPointerOrNil(user.Attributes.UpdatedDateTime)),
		PhoneNumbers:        phoneNumbers,
	}
}
