// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

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

var _ list.ListResource = &BlueprintListResource{}
var _ list.ListResourceWithConfigure = &BlueprintListResource{}

// NewBlueprintListResource returns a new list resource for Blueprints.
func NewBlueprintListResource() list.ListResource {
	return &BlueprintListResource{}
}

// BlueprintListResource implements terraform query list support for Blueprints.
type BlueprintListResource struct {
	client *client.Client
}

func (r *BlueprintListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (r *BlueprintListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "List")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_blueprint list resource") {
		return
	}
	r.client = c
}

func (r *BlueprintListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Searches for Apple Business Manager Blueprints.",
		Attributes: map[string]listschema.Attribute{
			"name": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive exact Blueprint name.",
			},
			"name_contains": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive substring match on the Blueprint name.",
			},
			"status": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive Blueprint status.",
			},
		},
	}
}

func (r *BlueprintListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unconfigured Provider",
				"The provider has not been configured yet. Re-run the command after `terraform init` has completed successfully.",
			),
		})
		return
	}

	var config BlueprintListResourceModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	blueprints, err := r.client.GetBlueprints(ctx, nil)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to list Blueprints",
				err.Error(),
			),
		})
		return
	}

	filtered := filterBlueprintList(blueprints, config)

	maxResults := req.Limit
	if maxResults <= 0 || maxResults > int64(len(filtered)) {
		maxResults = int64(len(filtered))
	}

	results := make([]list.ListResult, 0, int(maxResults))
	var emitted int64

	for _, blueprint := range filtered {
		if maxResults > 0 && emitted >= maxResults {
			break
		}

		result := req.NewListResult(ctx)
		result.DisplayName = blueprint.Attributes.Name
		identity := blueprintIdentityModel{
			ID: types.StringValue(blueprint.ID),
		}

		result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

		if req.IncludeResource {
			resourceState, err := r.buildBlueprintState(ctx, blueprint.ID)
			if err != nil {
				stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to read Blueprint",
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

	tflog.Debug(ctx, "Listed Blueprints", map[string]any{
		"requested_limit": req.Limit,
		"returned":        len(results),
		"filters": map[string]string{
			"name":          config.Name.ValueString(),
			"name_contains": config.NameContains.ValueString(),
			"status":        config.Status.ValueString(),
		},
	})

	if len(results) == 0 {
		stream.Results = list.NoListResults
		return
	}

	stream.Results = slices.Values(results)
}

func filterBlueprintList(blueprints []client.Blueprint, cfg BlueprintListResourceModel) []client.Blueprint {
	filtered := make([]client.Blueprint, 0, len(blueprints))

	exactName, hasExact := common.NormalizedFilterString(cfg.Name)
	containsName, hasContains := common.NormalizedFilterString(cfg.NameContains)
	status, hasStatus := common.NormalizedFilterString(cfg.Status)

	exactName = strings.ToLower(exactName)
	containsName = strings.ToLower(containsName)
	status = strings.ToLower(status)

	for _, blueprint := range blueprints {
		currentName := strings.ToLower(strings.TrimSpace(blueprint.Attributes.Name))
		currentStatus := strings.ToLower(strings.TrimSpace(blueprint.Attributes.Status))

		if hasExact && currentName != exactName {
			continue
		}

		if hasContains && !strings.Contains(currentName, containsName) {
			continue
		}

		if hasStatus && currentStatus != status {
			continue
		}

		filtered = append(filtered, blueprint)
	}

	return filtered
}

func (r *BlueprintListResource) buildBlueprintState(ctx context.Context, blueprintID string) (BlueprintModel, error) {
	resource := &BlueprintResource{client: r.client}
	var state BlueprintModel
	if err := resource.refreshBlueprintAttributes(ctx, blueprintID, &state); err != nil {
		return BlueprintModel{}, err
	}
	if err := resource.populateRelationshipSets(ctx, blueprintID, &state); err != nil {
		return BlueprintModel{}, err
	}
	state.Timeouts = newBlueprintTimeoutsNullValue()
	return state, nil
}
