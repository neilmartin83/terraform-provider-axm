// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

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

// Create creates a new Blueprint.
func (r *BlueprintResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BlueprintModel

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

	request := client.BlueprintCreateRequest{
		Data: client.BlueprintCreateRequestData{
			Type: blueprintResourceType,
			Attributes: client.BlueprintCreateAttributes{
				Name:        plan.Name.ValueString(),
				Description: plan.Description.ValueString(),
			},
			Relationships: buildBlueprintRelationshipsRequest(plan),
		},
	}

	blueprint, err := r.client.CreateBlueprint(createCtx, request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Blueprint", err.Error())
		return
	}

	state := plan
	if err := r.refreshBlueprintAttributes(createCtx, blueprint.ID, &state); err != nil {
		resp.Diagnostics.AddError("Failed to read Blueprint", err.Error())
		return
	}

	if err := r.populateRelationshipSets(createCtx, blueprint.ID, &state); err != nil {
		resp.Diagnostics.AddError("Failed to read Blueprint relationships", err.Error())
		return
	}

	if resp.Identity != nil {
		identity := blueprintIdentityModel{
			ID: state.ID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	state.Timeouts = ensureBlueprintTimeouts(state.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the Blueprint state.
func (r *BlueprintResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BlueprintModel

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

	if err := r.refreshBlueprintAttributes(readCtx, state.ID.ValueString(), &state); err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read Blueprint", err.Error())
		return
	}

	if err := r.populateRelationshipSets(readCtx, state.ID.ValueString(), &state); err != nil {
		resp.Diagnostics.AddError("Failed to read Blueprint relationships", err.Error())
		return
	}

	if resp.Identity != nil {
		identity := blueprintIdentityModel{
			ID: state.ID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	state.Timeouts = ensureBlueprintTimeouts(state.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update applies changes to a Blueprint.
func (r *BlueprintResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BlueprintModel
	var state BlueprintModel

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

	if plan.Name.ValueString() != state.Name.ValueString() || plan.Description.ValueString() != state.Description.ValueString() {
		name := plan.Name.ValueString()
		description := plan.Description.ValueString()

		updateRequest := client.BlueprintUpdateRequest{
			Data: client.BlueprintUpdateRequestData{
				Type: blueprintResourceType,
				ID:   state.ID.ValueString(),
				Attributes: &client.BlueprintUpdateAttributes{
					Name:        &name,
					Description: &description,
				},
			},
		}

		if _, err := r.client.UpdateBlueprint(updateCtx, updateRequest); err != nil {
			resp.Diagnostics.AddError("Failed to update Blueprint", err.Error())
			return
		}
	}

	if err := r.updateBlueprintRelationships(updateCtx, state, plan); err != nil {
		resp.Diagnostics.AddError("Failed to update Blueprint relationships", err.Error())
		return
	}

	newState := plan
	if err := r.refreshBlueprintAttributes(updateCtx, state.ID.ValueString(), &newState); err != nil {
		resp.Diagnostics.AddError("Failed to read Blueprint", err.Error())
		return
	}

	if err := r.populateRelationshipSets(updateCtx, state.ID.ValueString(), &newState); err != nil {
		resp.Diagnostics.AddError("Failed to read Blueprint relationships", err.Error())
		return
	}

	if resp.Identity != nil {
		identity := blueprintIdentityModel{
			ID: newState.ID,
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	newState.Timeouts = ensureBlueprintTimeouts(newState.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// Delete removes a Blueprint.
func (r *BlueprintResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BlueprintModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBlueprint(ctx, state.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete Blueprint", err.Error())
		return
	}
}

// refreshBlueprintAttributes reads the Blueprint from the API and sets computed
// attribute fields on state. Relationship sets are NOT refreshed here — call
// populateRelationshipSets separately.
func (r *BlueprintResource) refreshBlueprintAttributes(ctx context.Context, blueprintID string, state *BlueprintModel) error {
	blueprint, err := r.client.GetBlueprint(ctx, blueprintID, nil)
	if err != nil {
		return err
	}

	state.ID = types.StringValue(blueprint.ID)
	state.Name = types.StringValue(blueprint.Attributes.Name)
	state.Description = types.StringPointerValue(common.StringPointerOrNil(blueprint.Attributes.Description))
	state.Status = types.StringValue(blueprint.Attributes.Status)
	state.AppLicenseDeficient = types.BoolValue(blueprint.Attributes.AppLicenseDeficient)
	state.CreatedDateTime = types.StringValue(blueprint.Attributes.CreatedDateTime)
	state.UpdatedDateTime = types.StringValue(blueprint.Attributes.UpdatedDateTime)

	tflog.Debug(ctx, "Read blueprint attributes", map[string]any{
		"blueprint_id": blueprintID,
	})

	return nil
}

// populateRelationshipSets reads every Blueprint relationship from the API and
// updates the corresponding set on state.
func (r *BlueprintResource) populateRelationshipSets(ctx context.Context, blueprintID string, state *BlueprintModel) error {
	type relTarget struct {
		name string
		dest *types.Set
	}

	targets := []relTarget{
		{relationshipApps, &state.AppIDs},
		{relationshipConfigurations, &state.ConfigurationIDs},
		{relationshipPackages, &state.PackageIDs},
		{relationshipOrgDevices, &state.DeviceIDs},
		{relationshipUsers, &state.UserIDs},
		{relationshipUserGroups, &state.UserGroupIDs},
	}

	for _, t := range targets {
		ids, err := r.readBlueprintRelationshipIDs(ctx, blueprintID, t.name)
		if err != nil {
			return fmt.Errorf("reading %s: %w", t.name, err)
		}

		set, diags := common.StringsToSet(ids)
		if diags.HasError() {
			return fmt.Errorf("building %s set", t.name)
		}
		*t.dest = set
	}

	return nil
}

func buildBlueprintRelationshipsRequest(plan BlueprintModel) *client.BlueprintRelationshipsRequest {
	relationships := &client.BlueprintRelationshipsRequest{}
	used := false

	if !plan.AppIDs.IsNull() && !plan.AppIDs.IsUnknown() {
		appIDs := common.SetToStrings(plan.AppIDs)
		if len(appIDs) > 0 {
			relationships.Apps = &client.BlueprintRelationshipData{
				Data: buildRelationshipData(relationshipApps, appIDs),
			}
			used = true
		}
	}

	if !plan.ConfigurationIDs.IsNull() && !plan.ConfigurationIDs.IsUnknown() {
		configurationIDs := common.SetToStrings(plan.ConfigurationIDs)
		if len(configurationIDs) > 0 {
			relationships.Configurations = &client.BlueprintRelationshipData{
				Data: buildRelationshipData(relationshipConfigurations, configurationIDs),
			}
			used = true
		}
	}

	if !plan.PackageIDs.IsNull() && !plan.PackageIDs.IsUnknown() {
		packageIDs := common.SetToStrings(plan.PackageIDs)
		if len(packageIDs) > 0 {
			relationships.Packages = &client.BlueprintRelationshipData{
				Data: buildRelationshipData(relationshipPackages, packageIDs),
			}
			used = true
		}
	}

	if !plan.DeviceIDs.IsNull() && !plan.DeviceIDs.IsUnknown() {
		deviceIDs := common.SetToStrings(plan.DeviceIDs)
		if len(deviceIDs) > 0 {
			relationships.OrgDevices = &client.BlueprintRelationshipData{
				Data: buildRelationshipData(relationshipOrgDevices, deviceIDs),
			}
			used = true
		}
	}

	if !plan.UserIDs.IsNull() && !plan.UserIDs.IsUnknown() {
		userIDs := common.SetToStrings(plan.UserIDs)
		if len(userIDs) > 0 {
			relationships.Users = &client.BlueprintRelationshipData{
				Data: buildRelationshipData(relationshipUsers, userIDs),
			}
			used = true
		}
	}

	if !plan.UserGroupIDs.IsNull() && !plan.UserGroupIDs.IsUnknown() {
		userGroupIDs := common.SetToStrings(plan.UserGroupIDs)
		if len(userGroupIDs) > 0 {
			relationships.UserGroups = &client.BlueprintRelationshipData{
				Data: buildRelationshipData(relationshipUserGroups, userGroupIDs),
			}
			used = true
		}
	}

	if !used {
		return nil
	}

	return relationships
}

// updateBlueprintRelationships diffs old state against the plan and applies
// add/remove operations for each relationship. Uses Terraform state for the
// current values instead of reading from the API, avoiding potential 403 errors
// on GET_RELATIONSHIP endpoints.
func (r *BlueprintResource) updateBlueprintRelationships(ctx context.Context, oldState, plan BlueprintModel) error {
	blueprintID := oldState.ID.ValueString()

	type relUpdate struct {
		name    string
		current types.Set
		desired types.Set
	}

	updates := []relUpdate{
		{relationshipApps, oldState.AppIDs, plan.AppIDs},
		{relationshipConfigurations, oldState.ConfigurationIDs, plan.ConfigurationIDs},
		{relationshipPackages, oldState.PackageIDs, plan.PackageIDs},
		{relationshipOrgDevices, oldState.DeviceIDs, plan.DeviceIDs},
		{relationshipUsers, oldState.UserIDs, plan.UserIDs},
		{relationshipUserGroups, oldState.UserGroupIDs, plan.UserGroupIDs},
	}

	for _, u := range updates {
		current := common.SetToStrings(u.current)
		desired := common.SetToStrings(u.desired)
		toAdd, toRemove := diffIDs(current, desired)
		if err := r.updateBlueprintRelationship(ctx, blueprintID, u.name, u.name, toAdd, toRemove); err != nil {
			return err
		}
	}

	return nil
}
