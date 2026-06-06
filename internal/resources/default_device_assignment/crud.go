// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package default_device_assignment

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

// Create sets the singleton ID and applies desired assignments.
func (r *DefaultDeviceAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DefaultDeviceAssignmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("default")

	if err := r.applyAssignments(ctx, DefaultDeviceAssignmentModel{}, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply default device assignments", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes state by doing a single GET for each server ID tracked in the 6 family fields.
func (r *DefaultDeviceAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DefaultDeviceAssignmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build set of unique server IDs to look up.
	serverIDs := uniqueServerIDs(data)
	if len(serverIDs) == 0 {
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	// GET each referenced server and build the current family → serverID map.
	current := make(map[string]string) // family constant → server ID
	for id := range serverIDs {
		srv, err := r.client.GetDeviceManagementService(ctx, id, url.Values{})
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Failed to read MDM server %s", id), err.Error())
			return
		}
		for _, family := range srv.Attributes.DefaultProductFamilies {
			current[family] = id
		}
	}

	// Reconcile state: if a family field has a server ID but that server no longer holds
	// that family, clear the field.
	data.AppleTV = reconcileFamily(data.AppleTV, "APPLE_TV", current)
	data.AppleVisionPro = reconcileFamily(data.AppleVisionPro, "APPLE_VISION_PRO", current)
	data.IPad = reconcileFamily(data.IPad, "IPAD", current)
	data.IPhone = reconcileFamily(data.IPhone, "IPHONE", current)
	data.IPod = reconcileFamily(data.IPod, "IPOD", current)
	data.Mac = reconcileFamily(data.Mac, "MAC", current)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies the two-PATCH strategy: clear/reduce servers losing families first,
// then set servers gaining families.
func (r *DefaultDeviceAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state DefaultDeviceAssignmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyAssignments(ctx, state, plan); err != nil {
		resp.Diagnostics.AddError("Failed to apply default device assignments", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the resource from state without changing ABM assignments.
func (r *DefaultDeviceAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// applyAssignments computes the required PATCHes and executes them in the correct order.
// Phase A: reduce/clear servers that are losing families.
// Phase B: set servers that are gaining families.
func (r *DefaultDeviceAssignmentResource) applyAssignments(ctx context.Context, state, plan DefaultDeviceAssignmentModel) error {
	// Build family → server maps for state and plan.
	stateMap := familyServerMap(state)
	planMap := familyServerMap(plan)

	// Per-server family sets for phase A (servers that need to be reduced/cleared).
	// A server enters phase A when it loses at least one family.
	// We compute its full desired set (which may be empty or smaller) and send one PATCH.
	phaseA := make(map[string][]string) // serverID → new desired families (after removals)
	phaseB := make(map[string][]string) // serverID → new desired families (gains)

	allFamilies := []string{"APPLE_TV", "APPLE_VISION_PRO", "IPAD", "IPHONE", "IPOD", "MAC"}

	// For each server currently holding families, compute what it should hold after the plan.
	serversLosingFamilies := make(map[string]bool)
	for _, family := range allFamilies {
		oldServer := stateMap[family]
		newServer := planMap[family]
		if oldServer != "" && oldServer != newServer {
			serversLosingFamilies[oldServer] = true
		}
	}

	for server := range serversLosingFamilies {
		var keep []string
		for _, family := range allFamilies {
			if planMap[family] == server {
				keep = append(keep, family)
			}
		}
		phaseA[server] = keep // may be nil (clear all)
	}

	// For each server gaining families, collect its full desired set.
	serversGainingFamilies := make(map[string]bool)
	for _, family := range allFamilies {
		oldServer := stateMap[family]
		newServer := planMap[family]
		if newServer != "" && newServer != oldServer {
			serversGainingFamilies[newServer] = true
		}
	}

	for server := range serversGainingFamilies {
		var gain []string
		for _, family := range allFamilies {
			if planMap[family] == server {
				gain = append(gain, family)
			}
		}
		phaseB[server] = gain
	}

	// Phase A: clear or reduce servers losing families.
	for serverID, families := range phaseA {
		if err := r.setFamilies(ctx, serverID, families); err != nil {
			return fmt.Errorf("phase A PATCH server %s: %w", serverID, err)
		}
	}

	// Phase B: set servers gaining families.
	for serverID, families := range phaseB {
		if err := r.setFamilies(ctx, serverID, families); err != nil {
			return fmt.Errorf("phase B PATCH server %s: %w", serverID, err)
		}
	}

	return nil
}

// setFamilies sends a single PATCH for a server. Empty slice clears all families.
func (r *DefaultDeviceAssignmentResource) setFamilies(ctx context.Context, serverID string, families []string) error {
	if len(families) == 0 {
		_, err := r.client.ClearDeviceManagementServiceDefaultFamilies(ctx, serverID)
		return err
	}
	_, err := r.client.UpdateDeviceManagementService(ctx, client.MdmServerUpdateRequest{
		Data: client.MdmServerUpdateRequestData{
			Type: "mdmServers",
			ID:   serverID,
			Attributes: client.MdmServerUpdateAttributes{
				DefaultProductFamilies: families,
			},
		},
	})
	return err
}

// familyServerMap builds a map from Apple family constant to MDM server ID from state/plan.
// Empty-string values (sentinel for "ensure unassigned") map to the empty string key.
func familyServerMap(data DefaultDeviceAssignmentModel) map[string]string {
	m := make(map[string]string)
	set := func(family string, v types.String) {
		if !v.IsNull() && !v.IsUnknown() {
			m[family] = v.ValueString()
		}
	}
	set("APPLE_TV", data.AppleTV)
	set("APPLE_VISION_PRO", data.AppleVisionPro)
	set("IPAD", data.IPad)
	set("IPHONE", data.IPhone)
	set("IPOD", data.IPod)
	set("MAC", data.Mac)
	return m
}

// uniqueServerIDs collects unique non-empty server IDs from the 6 family fields.
func uniqueServerIDs(data DefaultDeviceAssignmentModel) map[string]struct{} {
	ids := make(map[string]struct{})
	for _, v := range []types.String{data.AppleTV, data.AppleVisionPro, data.IPad, data.IPhone, data.IPod, data.Mac} {
		if !v.IsNull() && !v.IsUnknown() && v.ValueString() != "" {
			ids[v.ValueString()] = struct{}{}
		}
	}
	return ids
}

// reconcileFamily returns the field value to store in state after comparing the tracked
// server assignment against what the API currently reports.
func reconcileFamily(field types.String, family string, current map[string]string) types.String {
	if field.IsNull() || field.IsUnknown() {
		return field
	}
	serverID := field.ValueString()
	if serverID == "" {
		return field // explicit unassigned sentinel — preserve
	}
	// If the server still holds this family, keep the field as-is.
	if current[family] == serverID {
		return field
	}
	// Family is no longer on this server — clear the field.
	return types.StringNull()
}
