package device_management_service

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ resource.Resource = &DeviceManagementServiceResource{}
var _ resource.ResourceWithImportState = &DeviceManagementServiceResource{}

func NewDeviceManagementServiceResource() resource.Resource {
	return &DeviceManagementServiceResource{}
}

type DeviceManagementServiceResource struct {
	client *client.Client
}

// MdmDeviceAssignmentModel describes the resource data model.
type MdmDeviceAssignmentModel struct {
	ID        types.String `tfsdk:"id"`
	DeviceIDs types.Set    `tfsdk:"device_ids"`
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
				Description: "MDM server ID. This is a unique ID for the server and is visible in the browser address bar when navigating to Preferences and selecting the desired 'Device Management Service'. Required until creation is supported.",
			},
			"device_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "A set of device IDs to assign to the MDM server. These are device serial numbers.",
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
