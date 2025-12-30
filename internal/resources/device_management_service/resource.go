package device_management_service

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ resource.Resource = &DeviceManagementServiceResource{}
var _ resource.ResourceWithIdentity = &DeviceManagementServiceResource{}
var _ resource.ResourceWithImportState = &DeviceManagementServiceResource{}

const (
	defaultCreateTimeout = 90 * time.Second
	defaultReadTimeout   = 90 * time.Second
	defaultUpdateTimeout = 90 * time.Second
)

func NewDeviceManagementServiceResource() resource.Resource {
	return &DeviceManagementServiceResource{}
}

type DeviceManagementServiceResource struct {
	client *client.Client
}

func (r *DeviceManagementServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service"
}

// Schema defines the schema for the resource.
func (r *DeviceManagementServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages device assignments to a specific Apple Business Manager MDM server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Device management service ID. This is a unique ID for the service and is visible in the browser address bar when navigating to Preferences and selecting the desired 'Device Management Service'. Required until creation is supported.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Device management service name as reported by Apple Business Manager.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Device management service type (for example MDM, APPLE_CONFIGURATOR).",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
			}),
			"device_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "A set of device IDs to assign to the device management service. These are device serial numbers.",
			},
		},
	}
}

func (r *DeviceManagementServiceResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				Description:       "Device management service ID used to uniquely identify the Apple Business Manager server.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *DeviceManagementServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *DeviceManagementServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
