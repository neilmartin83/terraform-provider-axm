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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
				Description: "The opaque resource ID that uniquely identifies the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The device management service's name.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of device management service: MDM, APPLE_CONFIGURATOR, APPLE_MDM. Read only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The operational status of the device management service. Read only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"device_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of devices currently assigned to this device management service. Read only.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"default_product_families": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The product families that are assigned by default to this device management service. Read/update only.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"last_connected_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the device management service last connected to Apple's servers. Read only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_connected_ip": schema.StringAttribute{
				Computed:    true,
				Description: "The IP address from which the device management service last connected to Apple's servers. Read only.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the creation of the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time of the most-recent update for the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_release": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A Boolean value that indicates whether the device management service is allowed to disown its enrolled devices.",
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
