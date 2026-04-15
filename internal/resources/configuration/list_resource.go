// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ list.ListResource = &ConfigurationListResource{}
var _ list.ListResourceWithConfigure = &ConfigurationListResource{}

// NewConfigurationListResource returns a new list resource for Configurations.
func NewConfigurationListResource() list.ListResource {
	return &ConfigurationListResource{}
}

// ConfigurationListResource implements terraform query list support for Configurations.
type ConfigurationListResource struct {
	client *client.Client
}

func (r *ConfigurationListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration"
}

func (r *ConfigurationListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "List")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_configuration list resource") {
		return
	}
	r.client = c
}

func (r *ConfigurationListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Searches for Apple Business Manager Configurations.",
		Attributes: map[string]listschema.Attribute{
			"name": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive exact Configuration name.",
			},
			"name_contains": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive substring match on the Configuration name.",
			},
		},
	}
}

func (r *ConfigurationListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unconfigured Provider",
				"The provider has not been configured yet. Re-run the command after `terraform init` has completed successfully.",
			),
		})
		return
	}

	var config ConfigurationListResourceModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	configurations, err := r.client.GetConfigurations(ctx, nil)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to list Configurations",
				err.Error(),
			),
		})
		return
	}

	filtered := filterConfigurationList(configurations, config)

	maxResults := req.Limit
	if maxResults <= 0 || maxResults > int64(len(filtered)) {
		maxResults = int64(len(filtered))
	}

	results := make([]list.ListResult, 0, int(maxResults))
	var emitted int64

	for _, configuration := range filtered {
		if maxResults > 0 && emitted >= maxResults {
			break
		}

		result := req.NewListResult(ctx)
		result.DisplayName = configuration.Attributes.Name
		identity := configurationIdentityModel{
			ID: types.StringValue(configuration.ID),
		}

		result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

		if req.IncludeResource {
			resourceState, err := r.buildConfigurationState(ctx, configuration.ID)
			if err != nil {
				stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to read Configuration",
						err.Error(),
					),
				})
				return
			}
			result.Diagnostics.Append(result.Resource.Set(ctx, resourceState)...)
		}

		results = append(results, result)
		emitted++
	}

	tflog.Debug(ctx, "Listed Configurations", map[string]any{
		"requested_limit": req.Limit,
		"returned":        len(results),
		"filters": map[string]string{
			"name":          config.Name.ValueString(),
			"name_contains": config.NameContains.ValueString(),
		},
	})

	if len(results) == 0 {
		stream.Results = list.NoListResults
		return
	}

	stream.Results = slices.Values(results)
}

func filterConfigurationList(configurations []client.Configuration, cfg ConfigurationListResourceModel) []client.Configuration {
	filtered := make([]client.Configuration, 0, len(configurations))

	exactName, hasExact := common.NormalizedFilterString(cfg.Name)
	containsName, hasContains := common.NormalizedFilterString(cfg.NameContains)

	exactName = strings.ToLower(exactName)
	containsName = strings.ToLower(containsName)

	for _, configuration := range configurations {
		currentName := strings.ToLower(strings.TrimSpace(configuration.Attributes.Name))

		if hasExact && currentName != exactName {
			continue
		}

		if hasContains && !strings.Contains(currentName, containsName) {
			continue
		}

		filtered = append(filtered, configuration)
	}

	return filtered
}

func (r *ConfigurationListResource) buildConfigurationState(ctx context.Context, configurationID string) (ConfigurationModel, error) {
	resource := &ConfigurationResource{client: r.client}
	var state ConfigurationModel
	if err := resource.refreshConfigurationState(ctx, configurationID, &state); err != nil {
		return ConfigurationModel{}, err
	}
	state.Timeouts = newConfigurationTimeoutsNullValue()
	return state, nil
}
