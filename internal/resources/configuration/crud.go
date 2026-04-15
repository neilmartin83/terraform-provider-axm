// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

// Create creates a new Configuration.
func (r *ConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ConfigurationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
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

	request := client.ConfigurationCreateRequest{
		Data: client.ConfigurationCreateRequestData{
			Type: configurationResourceType,
			Attributes: client.ConfigurationCreateRequestAttributes{
				Type:                   customSettingType,
				Name:                   plan.Name.ValueString(),
				ConfiguredForPlatforms: common.SetToStrings(plan.ConfiguredForPlatforms),
				CustomSettingsValues: client.CustomSettingsValues{
					ConfigurationProfile: plan.ConfigurationProfile.ValueString(),
					Filename:             plan.Filename.ValueString(),
				},
			},
		},
	}

	configuration, err := r.client.CreateConfiguration(createCtx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Configuration", err.Error())
		return
	}

	state := plan
	if err := r.refreshConfigurationState(createCtx, configuration.ID, &state); err != nil {
		resp.Diagnostics.AddError("Failed to read Configuration", err.Error())
		return
	}

	if resp.Identity != nil {
		identity := configurationIdentityModel{
			ID: state.ID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	state.Timeouts = ensureConfigurationTimeouts(state.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the Configuration state.
func (r *ConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ConfigurationModel

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

	configuration, err := r.client.GetConfiguration(readCtx, state.ID.ValueString(), nil)
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Configuration", err.Error())
		return
	}

	if err := r.refreshConfigurationState(readCtx, configuration.ID, &state); err != nil {
		resp.Diagnostics.AddError("Failed to read Configuration", err.Error())
		return
	}

	if resp.Identity != nil {
		identity := configurationIdentityModel{
			ID: state.ID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	state.Timeouts = ensureConfigurationTimeouts(state.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates an existing Configuration.
func (r *ConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ConfigurationModel
	var state ConfigurationModel

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

	name := plan.Name.ValueString()
	customSettings := client.CustomSettingsValues{
		ConfigurationProfile: plan.ConfigurationProfile.ValueString(),
		Filename:             plan.Filename.ValueString(),
	}

	updateRequest := client.ConfigurationUpdateRequest{
		Data: client.ConfigurationUpdateRequestData{
			Type: configurationResourceType,
			ID:   state.ID.ValueString(),
			Attributes: client.ConfigurationUpdateRequestAttributes{
				Name:                   &name,
				ConfiguredForPlatforms: common.SetToStrings(plan.ConfiguredForPlatforms),
				CustomSettingsValues:   &customSettings,
			},
		},
	}

	if _, err := r.client.UpdateConfiguration(updateCtx, updateRequest); err != nil {
		resp.Diagnostics.AddError("Failed to update Configuration", err.Error())
		return
	}

	if err := r.refreshConfigurationState(updateCtx, state.ID.ValueString(), &state); err != nil {
		resp.Diagnostics.AddError("Failed to read Configuration", err.Error())
		return
	}

	if resp.Identity != nil {
		identity := configurationIdentityModel{
			ID: state.ID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	state.Timeouts = ensureConfigurationTimeouts(state.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete removes a Configuration.
func (r *ConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ConfigurationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteConfiguration(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete Configuration", err.Error())
		return
	}
}

func (r *ConfigurationResource) refreshConfigurationState(ctx context.Context, configurationID string, state *ConfigurationModel) error {
	configuration, err := r.client.GetConfiguration(ctx, configurationID, nil)
	if err != nil {
		return err
	}

	state.ID = types.StringValue(configuration.ID)
	state.Name = types.StringValue(configuration.Attributes.Name)
	state.Type = types.StringValue(configuration.Attributes.Type)
	state.CreatedDateTime = types.StringValue(configuration.Attributes.CreatedDateTime)
	state.UpdatedDateTime = types.StringValue(configuration.Attributes.UpdatedDateTime)

	platformSet, diags := common.StringsToSet(configuration.Attributes.ConfiguredForPlatforms)
	if diags.HasError() {
		return fmt.Errorf("failed to build configured_for_platforms set")
	}
	state.ConfiguredForPlatforms = platformSet

	if configuration.Attributes.CustomSettingsValues != nil {
		state.ConfigurationProfile = types.StringValue(configuration.Attributes.CustomSettingsValues.ConfigurationProfile)
		state.Filename = types.StringValue(configuration.Attributes.CustomSettingsValues.Filename)
	} else {
		state.ConfigurationProfile = types.StringNull()
		state.Filename = types.StringNull()
	}

	tflog.Debug(ctx, "Read configuration", map[string]any{
		"configuration_id": configurationID,
	})

	return nil
}
