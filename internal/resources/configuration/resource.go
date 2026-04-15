// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration

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
	configurationResourceType = "configurations"
	customSettingType         = "CUSTOM_SETTING"
	defaultCreateTimeout      = 90 * time.Second
	defaultUpdateTimeout      = 90 * time.Second
)

var _ resource.Resource = &ConfigurationResource{}
var _ resource.ResourceWithIdentity = &ConfigurationResource{}
var _ resource.ResourceWithImportState = &ConfigurationResource{}

// NewConfigurationResource returns a new resource for managing Configurations.
func NewConfigurationResource() resource.Resource {
	return &ConfigurationResource{}
}

// ConfigurationResource implements the Terraform resource for Configurations.
type ConfigurationResource struct {
	client *client.Client
}

func (r *ConfigurationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration"
}

func (r *ConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages custom Configuration profiles in Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The Configuration ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The Configuration name.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The Configuration type.",
			},
			"configured_for_platforms": schema.SetAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Platforms that the Configuration targets.",
			},
			"configuration_profile": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The configuration profile payload (mobileconfig XML). Required for CUSTOM_SETTING configurations.",
			},
			"filename": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filename for the configuration profile. Required for CUSTOM_SETTING configurations.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the Configuration was created.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the Configuration was last updated.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
			}),
		},
	}
}

func (r *ConfigurationResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				Description:       "Configuration ID used to uniquely identify the Configuration.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *ConfigurationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Resource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_configuration resource") {
		return
	}
	r.client = c
}

func (r *ConfigurationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
