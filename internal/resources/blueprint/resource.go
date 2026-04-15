// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

const (
	blueprintResourceType = "blueprints"
	defaultCreateTimeout  = 90 * time.Second
	defaultUpdateTimeout  = 90 * time.Second
)

var _ resource.Resource = &BlueprintResource{}
var _ resource.ResourceWithIdentity = &BlueprintResource{}
var _ resource.ResourceWithImportState = &BlueprintResource{}

// NewBlueprintResource returns a new resource for managing Blueprints.
func NewBlueprintResource() resource.Resource {
	return &BlueprintResource{}
}

// BlueprintResource implements the Terraform resource for Blueprints.
type BlueprintResource struct {
	client *client.Client
}

func (r *BlueprintResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint"
}

func (r *BlueprintResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Apple Business Manager Blueprints.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The Blueprint ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The Blueprint name.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The Blueprint description.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The Blueprint status.",
			},
			"app_license_deficient": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the Blueprint is missing app licenses.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the Blueprint was created.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the Blueprint was last updated.",
			},
			"app_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "App IDs associated with the Blueprint.",
			},
			"configuration_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Configuration IDs associated with the Blueprint.",
			},
			"package_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Package IDs associated with the Blueprint.",
			},
			"device_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Device IDs associated with the Blueprint.",
			},
			"user_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "User IDs associated with the Blueprint.",
			},
			"user_group_ids": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "User group IDs associated with the Blueprint.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
			}),
		},
	}
}

func (r *BlueprintResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				Description:       "Blueprint ID used to uniquely identify the Blueprint.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *BlueprintResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Resource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_blueprint resource") {
		return
	}
	r.client = c
}

func (r *BlueprintResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
