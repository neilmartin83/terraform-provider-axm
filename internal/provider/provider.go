package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/device_management_service"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/device_management_service_serialnumbers"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/device_management_services"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organization_device"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organization_device_assigned_server_information"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/organization_devices"
)

// Constants for environment variable names.
const (
	envTeamID     = "AXM_TEAM_ID"
	envClientID   = "AXM_CLIENT_ID"
	envKeyID      = "AXM_KEY_ID"
	envPrivateKey = "AXM_PRIVATE_KEY"
	envScope      = "AXM_SCOPE"
)

// Ensure AxmProvider satisfies the provider.Provider interface.
var _ provider.Provider = &AxmProvider{}

// AxmProvider defines the provider implementation.
type AxmProvider struct {
	client  *client.Client
	version string
}

// AxmProviderModel describes the provider data model for configuration.
type AxmProviderModel struct {
	TeamID     types.String `tfsdk:"team_id"`
	ClientID   types.String `tfsdk:"client_id"`
	KeyID      types.String `tfsdk:"key_id"`
	PrivateKey types.String `tfsdk:"private_key"`
	Scope      types.String `tfsdk:"scope"`
}

func (p *AxmProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "axm"
	resp.Version = p.version
}

func (p *AxmProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Automate device management actions and access data about devices that enroll using Automated Device Enrollment with the Apple School and Business Manager API. https://developer.apple.com/documentation/apple-school-and-business-manager-api",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Optional:    true,
				Description: "Team ID for Apple Business and School Manager authentication. If not specified, client_id will be used. Can also be set via the AXM_TEAM_ID environment variable.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "Client ID for Apple Business and School Manager authentication. Can also be set via the AXM_CLIENT_ID environment variable.",
			},
			"key_id": schema.StringAttribute{
				Optional:    true,
				Description: "Key ID for the private key. Can also be set via the AXM_KEY_ID environment variable.",
			},
			"private_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Contents of the private key downloaded from Apple Business or School Manager. Can also be set via the AXM_PRIVATE_KEY environment variable.",
			},
			"scope": schema.StringAttribute{
				Optional:    true,
				Description: "API scope to use. Valid values are 'business.api' or 'school.api'. Can also be set via the AXM_SCOPE environment variable.",
				Validators: []validator.String{
					stringvalidator.OneOf("business.api", "school.api"),
				},
			},
		},
	}
}

func (p *AxmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AxmProviderModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	teamID := data.TeamID.ValueString()
	if teamID == "" {
		teamID = getenv(envTeamID)
	}
	clientID := data.ClientID.ValueString()
	if clientID == "" {
		clientID = getenv(envClientID)
	}
	keyID := data.KeyID.ValueString()
	if keyID == "" {
		keyID = getenv(envKeyID)
	}
	privateKey := data.PrivateKey.ValueString()
	if privateKey == "" {
		privateKey = getenv(envPrivateKey)
	}
	scope := data.Scope.ValueString()
	if scope == "" {
		scope = getenv(envScope)
	}
	if scope == "" {
		scope = "business.api"
	}

	if clientID == "" {
		resp.Diagnostics.AddError(
			"Missing Client ID",
			"client_id must be provided either in the provider configuration or via the AXM_CLIENT_ID environment variable.",
		)
	}
	if keyID == "" {
		resp.Diagnostics.AddError(
			"Missing Key ID",
			"key_id must be provided either in the provider configuration or via the AXM_KEY_ID environment variable.",
		)
	}
	if privateKey == "" {
		resp.Diagnostics.AddError(
			"Missing Private Key",
			"private_key must be provided either in the provider configuration or via the AXM_PRIVATE_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
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

	if teamID == "" {
		teamID = clientID
	}

	clientObj, err := client.NewClient(
		baseURL,
		teamID,
		clientID,
		keyID,
		scope,
		privateKey,
	)
	if err != nil {
		resp.Diagnostics.AddError("AXM Client Init Failed", err.Error())
		return
	}

	clientObj.SetLogger(NewTerraformLogger())

	p.client = clientObj
	resp.DataSourceData = clientObj
	resp.ResourceData = clientObj
}

func (p *AxmProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		device_management_service.NewDeviceManagementServiceResource,
	}
}

func (p *AxmProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		organization_device.NewOrganizationDeviceDataSource,
		organization_devices.NewOrganizationDevicesDataSource,
		device_management_services.NewDeviceManagementServicesDataSource,
		device_management_service_serialnumbers.NewDeviceManagementServiceSerialNumbersDataSource,
		organization_device_assigned_server_information.NewOrganizationDeviceAssignedServerInformationDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AxmProvider{
			version: version,
		}
	}
}

// getenv is a helper to get an environment variable, returns empty string if not set.
func getenv(key string) string {
	v, _ := os.LookupEnv(key)
	return v
}
