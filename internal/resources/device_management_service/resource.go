package device_management_service

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

var _ resource.Resource = &deviceManagementServiceResource{}

type deviceManagementServiceResource struct {
	client *client.Client
}

type mdmDeviceAssignmentModel struct {
	ID        types.String `tfsdk:"id"`
	DeviceIDs types.List   `tfsdk:"device_ids"`
}

// NewDeviceManagementServiceResource creates a new instance of deviceManagementServiceResource
// with the provided client for managing device assignments.
func NewDeviceManagementServiceResource(client *client.Client) resource.Resource {
	return &deviceManagementServiceResource{
		client: client,
	}
}

// Metadata sets the provider type name for the resource.
func (r *deviceManagementServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service"
}

// Schema defines the schema for the resource.
func (r *deviceManagementServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages device assignments to a specific Apple Business Manager MDM server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "MDM server ID. This is a unique ID for the server and is visible in the browser address bar when navigating to Preferences and selecting the desired 'Device Management Service'. Required until creation is supported.",
			},
			"device_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "A list of device IDs to assign to the MDM server. These are device serial numbers.",
			},
		},
	}
}
