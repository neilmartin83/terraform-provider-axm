// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

const (
	webhookResourceType  = "webhooks"
	defaultCreateTimeout = 10 * time.Minute
	defaultUpdateTimeout = 10 * time.Minute
	defaultDeleteTimeout = 10 * time.Minute
)

var _ resource.Resource = &WebhookResource{}
var _ resource.ResourceWithIdentity = &WebhookResource{}
var _ resource.ResourceWithImportState = &WebhookResource{}

// NewWebhookResource returns a new resource for managing webhooks.
func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

// WebhookResource implements the Terraform resource for Apple Business webhooks.
type WebhookResource struct {
	client *client.Client
}

func (r *WebhookResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Apple Business Manager webhook.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The webhook ID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the webhook.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the webhook is enabled. Defaults to true.",
			},
			"event_types": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "The event types to subscribe to: ORG_DEVICE_ATTRIBUTE_CHANGED, ORG_DEVICE_ATTRIBUTE_MIGRATION_DETAILS_CHANGED, ORG_DEVICE_RELATIONSHIP_ASSIGNED_SERVER_CHANGED, or ORG_DEVICE_RELATIONSHIP_WARRANTY_CHANGED.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(stringvalidator.OneOf(
						"ORG_DEVICE_ATTRIBUTE_CHANGED",
						"ORG_DEVICE_ATTRIBUTE_MIGRATION_DETAILS_CHANGED",
						"ORG_DEVICE_RELATIONSHIP_ASSIGNED_SERVER_CHANGED",
						"ORG_DEVICE_RELATIONSHIP_WARRANTY_CHANGED",
					)),
				},
			},
			"secret_wo": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "The shared secret used by Apple to sign webhook notifications. This value is write-only and is never stored in Terraform state.",
			},
			"secret_wo_version": schema.Int64Attribute{
				Required:    true,
				Description: "Increment this value to trigger an update when rotating secret_wo.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The HTTPS URL that receives webhook notifications.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the webhook was created.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The date and time the webhook was last updated.",
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *WebhookResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"id": identityschema.StringAttribute{
				Description:       "Webhook ID used to uniquely identify the webhook.",
				RequiredForImport: true,
			},
		},
	}
}

func (r *WebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Resource")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_webhook resource") {
		return
	}
	r.client = c
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
