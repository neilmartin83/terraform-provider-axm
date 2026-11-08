// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package webhook

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var secret types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("secret_wo"), &secret)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if secret.IsNull() || secret.IsUnknown() || secret.ValueString() == "" {
		resp.Diagnostics.AddError("Missing Webhook Secret", "A non-empty secret is required to create a webhook.")
		return
	}

	createTimeout := defaultCreateTimeout
	if !plan.Timeouts.IsNull() && !plan.Timeouts.IsUnknown() {
		configuredTimeout, timeoutDiags := plan.Timeouts.Create(ctx, defaultCreateTimeout)
		resp.Diagnostics.Append(timeoutDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createTimeout = configuredTimeout
	}
	createCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	eventTypes := common.SetToStrings(plan.EventTypes)
	created, err := r.client.CreateWebhook(createCtx, client.WebhookCreateRequest{
		Data: client.WebhookCreateRequestData{
			Type: webhookResourceType,
			Attributes: client.WebhookCreateRequestAttributes{
				Name:       plan.Name.ValueString(),
				Enabled:    plan.Enabled.ValueBool(),
				EventTypes: eventTypes,
				Secret:     secret.ValueString(),
				URL:        plan.URL.ValueString(),
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Webhook", err.Error())
		return
	}

	state := plan
	r.refreshWebhookModel(ctx, created, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Timeouts = ensureWebhookTimeouts(state.Timeouts)

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, webhookIdentityModel{ID: state.ID})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readCtx, cancel, timeoutDiags := common.ResolveReadTimeout(ctx, state.Timeouts, common.DefaultReadTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	defer cancel()

	webhook, err := r.client.GetWebhook(readCtx, state.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Webhook", err.Error())
		return
	}

	r.refreshWebhookModel(ctx, webhook, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Timeouts = ensureWebhookTimeouts(state.Timeouts)

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, webhookIdentityModel{ID: state.ID})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WebhookModel
	var state WebhookModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout := defaultUpdateTimeout
	if !plan.Timeouts.IsNull() && !plan.Timeouts.IsUnknown() {
		configuredTimeout, timeoutDiags := plan.Timeouts.Update(ctx, defaultUpdateTimeout)
		resp.Diagnostics.Append(timeoutDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateTimeout = configuredTimeout
	}
	updateCtx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	attributes := client.WebhookUpdateRequestAttributes{}
	if plan.Name.ValueString() != state.Name.ValueString() {
		name := plan.Name.ValueString()
		attributes.Name = &name
	}
	if plan.Enabled.ValueBool() != state.Enabled.ValueBool() {
		enabled := plan.Enabled.ValueBool()
		attributes.Enabled = &enabled
	}
	planEventTypes := common.SetToStrings(plan.EventTypes)
	stateEventTypes := common.SetToStrings(state.EventTypes)
	if !slices.Equal(planEventTypes, stateEventTypes) {
		attributes.EventTypes = &planEventTypes
	}
	if plan.URL.ValueString() != state.URL.ValueString() {
		url := plan.URL.ValueString()
		attributes.URL = &url
	}
	if plan.SecretWoVersion.ValueInt64() != state.SecretWoVersion.ValueInt64() {
		var secret types.String
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("secret_wo"), &secret)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if secret.IsNull() || secret.IsUnknown() || secret.ValueString() == "" {
			resp.Diagnostics.AddError("Missing Webhook Secret", "A non-empty secret is required when rotating a webhook secret.")
			return
		}
		secretValue := secret.ValueString()
		attributes.Secret = &secretValue
	}

	updated, err := r.client.UpdateWebhook(updateCtx, client.WebhookUpdateRequest{
		Data: client.WebhookUpdateRequestData{
			Type:       webhookResourceType,
			ID:         state.ID.ValueString(),
			Attributes: attributes,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to update Webhook", err.Error())
		return
	}

	newState := plan
	r.refreshWebhookModel(ctx, updated, &newState, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	newState.Timeouts = ensureWebhookTimeouts(newState.Timeouts)

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, webhookIdentityModel{ID: newState.ID})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WebhookModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout := defaultDeleteTimeout
	if !state.Timeouts.IsNull() && !state.Timeouts.IsUnknown() {
		configuredTimeout, timeoutDiags := state.Timeouts.Delete(ctx, defaultDeleteTimeout)
		resp.Diagnostics.Append(timeoutDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		deleteTimeout = configuredTimeout
	}
	deleteCtx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	if err := r.client.DeleteWebhook(deleteCtx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete Webhook", err.Error())
	}
}

func (r *WebhookResource) refreshWebhookModel(ctx context.Context, webhook *client.Webhook, state *WebhookModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(webhook.ID)
	state.Name = types.StringValue(webhook.Attributes.Name)
	state.Enabled = types.BoolValue(webhook.Attributes.Enabled)
	state.URL = types.StringValue(webhook.Attributes.URL)
	state.CreatedDateTime = types.StringValue(webhook.Attributes.CreatedDateTime)
	state.UpdatedDateTime = types.StringValue(webhook.Attributes.UpdatedDateTime)
	eventTypes, eventDiags := common.StringsToSet(webhook.Attributes.EventTypes)
	diags.Append(eventDiags...)
	state.EventTypes = eventTypes

	tflog.Debug(ctx, "Refreshed webhook attributes", map[string]any{
		"webhook_id": webhook.ID,
	})
}

var webhookTimeoutAttributeTypes = map[string]attr.Type{
	"create": types.StringType,
	"read":   types.StringType,
	"update": types.StringType,
	"delete": types.StringType,
}

func ensureWebhookTimeouts(value timeouts.Value) timeouts.Value {
	if value.IsNull() && !value.IsUnknown() {
		value.Object = types.ObjectNull(webhookTimeoutAttributeTypes)
	}
	return value
}
