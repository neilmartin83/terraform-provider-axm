// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ resource.Resource = &DeviceManagementServiceResource{}
var _ resource.ResourceWithIdentity = &DeviceManagementServiceResource{}
var _ resource.ResourceWithImportState = &DeviceManagementServiceResource{}

const (
	defaultCreateTimeout = 90 * time.Second
	defaultUpdateTimeout = 90 * time.Second
)

// NewDeviceManagementServiceResource returns a new resource for managing MDM servers.
func NewDeviceManagementServiceResource() resource.Resource {
	return &DeviceManagementServiceResource{}
}

// DeviceManagementServiceResource implements the Terraform resource for MDM servers.
type DeviceManagementServiceResource struct {
	client *client.Client
}

func (r *DeviceManagementServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service"
}

// Schema defines the schema for the resource.
func (r *DeviceManagementServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Apple Business Manager MDM server and its device assignments. Server creation, update, and deletion require business scope.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Device management service ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "MDM server name.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "MDM server type (MDM, APPLE_CONFIGURATOR, etc.).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_release": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Allow this service to release devices.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"server_certificate": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "X.509 MDM certificate. Required when creating a new server. Not returned by the API; stored in state as provided.",
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required:    true,
						Description: "Certificate filename.",
					},
					"data": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "Base64-encoded DER certificate data.",
					},
				},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
			}),
			"device_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Set of device serial numbers to assign to this MDM server.",
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
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
	c, diags := common.ConfigureClient(req.ProviderData, "Resource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	r.client = c
}

func (r *DeviceManagementServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
