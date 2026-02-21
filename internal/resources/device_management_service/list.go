package device_management_service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ list.ListResource = &DeviceManagementServiceListResource{}
var _ list.ListResourceWithConfigure = &DeviceManagementServiceListResource{}

// NewDeviceManagementServiceListResource returns a new list resource for device management services.
func NewDeviceManagementServiceListResource() list.ListResource {
	return &DeviceManagementServiceListResource{}
}

// DeviceManagementServiceListResource implements terraform query list support for the
// device_management_service resource type.
type DeviceManagementServiceListResource struct {
	client *client.Client
}

// DeviceManagementServiceListResourceModel captures filters supported by the list query.
type DeviceManagementServiceListResourceModel struct {
	Name         types.String `tfsdk:"name"`
	NameContains types.String `tfsdk:"name_contains"`
	ServerType   types.String `tfsdk:"server_type"`
}

func (r *DeviceManagementServiceListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service"
}

func (r *DeviceManagementServiceListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected List Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DeviceManagementServiceListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = listschema.Schema{
		Description: "Searches for Apple Business Manager device management services.",
		Attributes: map[string]listschema.Attribute{
			"server_type": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by the Apple Business Manager server type (MDM, APPLE_CONFIGURATOR, APPLE_MDM).",
				Validators: []validator.String{
					stringvalidator.OneOf("MDM", "APPLE_CONFIGURATOR", "APPLE_MDM"),
				},
			},
			"name": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive exact server name.",
			},
			"name_contains": listschema.StringAttribute{
				Optional:    true,
				Description: "Filters results by a case-insensitive substring match on the server name.",
			},
		},
	}
}

func (r *DeviceManagementServiceListResource) List(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	if r.client == nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unconfigured Provider",
				"The provider has not been configured yet. Re-run the command after `terraform init` has completed successfully.",
			),
		})
		return
	}

	var config DeviceManagementServiceListResourceModel
	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	servers, err := r.client.GetDeviceManagementServices(ctx, nil)
	if err != nil {
		stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Unable to list device management services",
				err.Error(),
			),
		})
		return
	}

	filtered := filterDeviceManagementServiceList(servers, config)

	maxResults := req.Limit
	if maxResults <= 0 || maxResults > int64(len(filtered)) {
		maxResults = int64(len(filtered))
	}

	results := make([]list.ListResult, 0, int(maxResults))
	var emitted int64

	for _, server := range filtered {
		if maxResults > 0 && emitted >= maxResults {
			break
		}

		result := req.NewListResult(ctx)
		result.DisplayName = server.Attributes.ServerName
		identity := deviceManagementServiceIdentityModel{
			ID: types.StringValue(server.ID),
		}

		result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

		if req.IncludeResource {
			serials, err := r.client.GetDeviceManagementServiceSerialNumbers(ctx, server.ID)
			if err != nil {
				stream.Results = list.ListResultsStreamDiagnostics(diag.Diagnostics{
					diag.NewErrorDiagnostic(
						"Unable to read device assignments",
						err.Error(),
					),
				})
				return
			}

			deviceSet, setDiags := stringsToSet(serials)
			if setDiags.HasError() {
				stream.Results = list.ListResultsStreamDiagnostics(setDiags)
				return
			}

			state := MdmDeviceAssignmentModel{
				ID:        types.StringValue(server.ID),
				Name:      types.StringValue(server.Attributes.ServerName),
				Type:      types.StringValue(server.Attributes.ServerType),
				DeviceIDs: deviceSet,
			}

			// Provide a null object with the expected shape so Terraform can coerce the timeouts attribute.
			state.Timeouts = newDeviceManagementServiceTimeoutsNullValue()

			result.Diagnostics.Append(result.Resource.Set(ctx, state)...)
		}

		results = append(results, result)
		emitted++
	}

	tflog.Debug(ctx, "Listed device management services", map[string]any{
		"requested_limit": req.Limit,
		"returned":        len(results),
		"filters": map[string]string{
			"name":          config.Name.ValueString(),
			"name_contains": config.NameContains.ValueString(),
			"server_type":   config.ServerType.ValueString(),
		},
	})

	if len(results) == 0 {
		stream.Results = list.NoListResults
		return
	}

	stream.Results = slices.Values(results)
}

func filterDeviceManagementServiceList(servers []client.MdmServer, cfg DeviceManagementServiceListResourceModel) []client.MdmServer {
	filtered := make([]client.MdmServer, 0, len(servers))

	exactName, hasExact := normalizedFilterString(cfg.Name)
	containsName, hasContains := normalizedFilterString(cfg.NameContains)
	serverType, hasType := normalizedFilterString(cfg.ServerType)

	exactName = strings.ToLower(exactName)
	containsName = strings.ToLower(containsName)
	serverType = strings.ToLower(serverType)

	for _, server := range servers {
		currentName := strings.ToLower(strings.TrimSpace(server.Attributes.ServerName))
		currentType := strings.ToLower(strings.TrimSpace(server.Attributes.ServerType))

		if hasType && currentType != serverType {
			continue
		}

		if hasExact && currentName != exactName {
			continue
		}

		if hasContains && !strings.Contains(currentName, containsName) {
			continue
		}

		filtered = append(filtered, server)
	}

	return filtered
}

func normalizedFilterString(value types.String) (string, bool) {
	if value.IsNull() || value.IsUnknown() {
		return "", false
	}

	trimmed := strings.TrimSpace(value.ValueString())
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
}
