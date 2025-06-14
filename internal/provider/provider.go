package axm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type providerModel struct {
	TeamID     types.String `tfsdk:"team_id"`
	ClientID   types.String `tfsdk:"client_id"`
	KeyID      types.String `tfsdk:"key_id"`
	PrivateKey types.String `tfsdk:"private_key"`
	Scope      types.String `tfsdk:"scope"`
}

type axmProvider struct{}

func (p *axmProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "axm"
}

func (p *axmProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Automate device management actions and access data about devices that enroll using Automated Device Enrollment with the Apple School and Business Manager API. https://developer.apple.com/documentation/apple-school-and-business-manager-api",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Optional:    true,
				Description: "Team ID for Apple Business and School Manager authentication. If not specified, client_id will be used.",
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Client ID for Apple Business and School Manager authentication",
			},
			"key_id": schema.StringAttribute{
				Required:    true,
				Description: "Key ID for the private key.",
			},
			"private_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Contents of the private key downloaded from Apple Business or School Manager.",
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "API scope to use. Valid values are 'business.api' or 'school.api'.",
				Validators: []validator.String{
					ScopeValidator{},
				},
			},
		},
	}
}

func (p *axmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	scope := config.Scope.ValueString()
	if scope == "" {
		scope = "business.api"
	}

	var baseURL string
	switch scope {
	case "business.api":
		baseURL = "https://api-business.apple.com"
	case "school.api":
		baseURL = "https://api-school.apple.com"
	default:
		resp.Diagnostics.AddError(
			"Invalid Scope",
			fmt.Sprintf("Scope must be either 'business.api' or 'school.api', got: %s", scope),
		)
		return
	}

	teamID := config.TeamID.ValueString()
	if teamID == "" {
		teamID = config.ClientID.ValueString()
	}

	client, err := NewClient(
		baseURL,
		teamID,
		config.ClientID.ValueString(),
		config.KeyID.ValueString(),
		config.Scope.ValueString(),
		config.PrivateKey.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("AXM Client Init Failed", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *axmProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *axmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDevicesDataSource,
		NewOrganizationDeviceDataSource,
		NewDeviceManagementServicesDataSource,
		NewDeviceManagementServiceSerialNumbersDataSource,
		NewOrganizationDeviceAssignedServerInformationDataSource,
	}
}

func New() provider.Provider {
	return &axmProvider{}
}

type ScopeValidator struct{}

func (v ScopeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if value != "business.api" && value != "school.api" {
		resp.Diagnostics.AddError(
			"Invalid Scope",
			"Scope must be either 'business.api' or 'school.api'",
		)
	}
}

func (v ScopeValidator) Description(ctx context.Context) string {
	return "Validates that the scope is either 'business.api' or 'school.api'"
}

func (v ScopeValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that the scope is either `business.api` or `school.api`"
}
