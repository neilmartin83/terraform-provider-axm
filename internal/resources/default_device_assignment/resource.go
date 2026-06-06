// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package default_device_assignment

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ resource.Resource = &DefaultDeviceAssignmentResource{}

// NewDefaultDeviceAssignmentResource returns a new resource for managing org-wide default device assignments.
func NewDefaultDeviceAssignmentResource() resource.Resource {
	return &DefaultDeviceAssignmentResource{}
}

// DefaultDeviceAssignmentResource manages which MDM server each Apple device family defaults to.
type DefaultDeviceAssignmentResource struct {
	client *client.Client
}

func (r *DefaultDeviceAssignmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_default_device_assignment"
}

// Schema defines the schema for the resource.
func (r *DefaultDeviceAssignmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	familyAttr := func(family string) schema.StringAttribute {
		return schema.StringAttribute{
			Optional:    true,
			Description: "MDM server ID to set as the default for " + family + ` devices. Set to "" to explicitly unassign. Omit to leave unmanaged.`,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		}
	}

	resp.Schema = schema.Schema{
		Description: "Manages the organisation-wide default MDM server assignment for each Apple device family. This is a singleton resource representing the default device assignment settings in Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always \"default\". Identifies this singleton resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"apple_tv":         familyAttr("Apple TV"),
			"apple_vision_pro": familyAttr("Apple Vision Pro"),
			"ipad":             familyAttr("iPad"),
			"iphone":           familyAttr("iPhone"),
			"ipod":             familyAttr("iPod"),
			"mac":              familyAttr("Mac"),
		},
	}
}

func (r *DefaultDeviceAssignmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Resource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = c
}
