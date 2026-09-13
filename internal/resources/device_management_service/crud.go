// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

// runActivity creates an org device activity and waits for completion.
func (r *DeviceManagementServiceResource) runActivity(ctx context.Context, activityType client.OrgDeviceActivityType, deadline, serverID string, deviceIDs []string, action string, diags *diag.Diagnostics) {
	opts := make([]client.ActivityOption, 0)
	if deadline != "" {
		opts = append(opts, client.WithMigrationDeadline(deadline))
	}
	if serverID != "" {
		opts = append(opts, client.WithMdmServer(serverID))
	}
	activity, err := r.client.CreateOrgDeviceActivity(ctx, activityType, deviceIDs, opts...)
	if err != nil {
		diags.AddError("Failed to "+action, err.Error())
		return
	}
	if err := r.waitForActivityCompletion(ctx, activity.ID, diags); err != nil {
		diags.AddError("Failed to complete "+action, err.Error())
	}
}

// Create creates a new MDM server (business scope) and optionally assigns devices to it.
func (r *DeviceManagementServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MdmDeviceAssignmentModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := defaultCreateTimeout
	if !data.Timeouts.IsNull() && !data.Timeouts.IsUnknown() {
		configuredTimeout, timeoutDiags := data.Timeouts.Create(ctx, defaultCreateTimeout)
		resp.Diagnostics.Append(timeoutDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createTimeout = configuredTimeout
	}

	createCtx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	if !r.client.IsBusinessScope() {
		resp.Diagnostics.AddError(
			"Business scope required",
			"Creating MDM servers requires business API scope. Use terraform import to manage existing servers in education scope.",
		)
		return
	}

	if data.ServerCertificate == nil {
		resp.Diagnostics.AddError(
			"Missing server_certificate",
			"server_certificate is required when creating a new MDM server.",
		)
		return
	}

	// Verify migration eligibility before creating anything, so a failed
	// eligibility check leaves no orphaned server behind.
	if len(data.MigratedDevices) > 0 {
		serials := make([]string, 0, len(data.MigratedDevices))
		for _, device := range data.MigratedDevices {
			if !device.ID.IsNull() && !device.ID.IsUnknown() {
				serials = append(serials, device.ID.ValueString())
			}
		}
		if err := verifyMigrationsCapable(createCtx, serials, func(ctx context.Context, id string) (*client.OrgDevice, error) {
			return r.client.GetOrgDevice(ctx, id, nil)
		}, migrationEligibilityTimeout); err != nil {
			resp.Diagnostics.AddError("Failed to verify MDM migration eligibility", err.Error())
			return
		}
	}

	enableDisown := data.AllowRelease.ValueBoolPointer()
	attrs := client.MdmServerCreateAttributes{
		ServerName: data.Name.ValueString(),
		ServerCertificate: client.MdmServerCertificate{
			Name: data.ServerCertificate.Name.ValueString(),
			Data: data.ServerCertificate.Data.ValueString(),
		},
		EnableMdmDisownFlag: enableDisown,
	}

	srv, err := r.client.CreateDeviceManagementService(createCtx, client.MdmServerCreateRequest{
		Data: client.MdmServerCreateRequestData{Type: "mdmServers", Attributes: attrs},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create MDM server", err.Error())
		return
	}

	data.ID = types.StringValue(srv.ID)
	data.Type = types.StringValue(srv.Attributes.ServerType)
	data.Status = common.StringPointerToTypesString(srv.Attributes.Status)
	data.DeviceCount = types.Int64PointerValue(srv.Attributes.DeviceCount)
	data.LastConnectedDateTime = types.StringPointerValue(srv.Attributes.LastConnectedDateTime)
	data.LastConnectedIp = types.StringPointerValue(srv.Attributes.LastConnectedIp)
	data.CreatedDateTime = types.StringValue(srv.Attributes.CreatedDateTime)
	data.UpdatedDateTime = types.StringValue(srv.Attributes.UpdatedDateTime)
	data.DefaultProductFamilies = common.StringsToList(ctx, srv.Attributes.DefaultProductFamilies)
	// AllowRelease is not reliably echoed by the create response; keep the
	// requested value so the field is known after apply. Read reconciles on the
	// next refresh if Apple silently ignored it.
	data.AllowRelease = types.BoolPointerValue(enableDisown)

	deviceIDs := extractStrings(data.DeviceIDs)
	if len(deviceIDs) > 0 {
		migrations, err := migrationMap(data.MigratedDevices)
		if err != nil {
			resp.Diagnostics.AddError("Invalid migrated_devices", err.Error())
			return
		}
		plain, byDeadline := partitionMigrations(deviceIDs, migrations)

		if len(plain) > 0 {
			r.runActivity(createCtx, client.OrgDeviceActivityAssignDevices, "", srv.ID, plain, "assign devices", &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		for _, group := range byDeadline {
			r.runActivity(createCtx, client.OrgDeviceActivityAssignWithMDMMigrationDeadline, group.deadline, srv.ID, group.devices, "assign devices with migration deadline", &resp.Diagnostics)
			if resp.Diagnostics.HasError() {
				return
			}
		}
	}

	// Resolve device_ids to a known value — required because it is Optional+Computed and
	// the plan value is Unknown on first create when the attribute is not in config.
	deviceSet, diags := stringsToSet(deviceIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DeviceIDs = deviceSet

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, deviceManagementServiceIdentityModel{
			ID: types.StringValue(srv.ID),
		})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Created MDM server", map[string]any{
		"mdm_server_id": srv.ID,
		"device_ids":    deviceIDs,
	})

	data.Timeouts = ensureDeviceManagementServiceTimeouts(data.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read retrieves the current state of the MDM server and its device assignments.
func (r *DeviceManagementServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MdmDeviceAssignmentModel

	if req.State.Raw.IsNull() {
		if req.Identity == nil {
			resp.Diagnostics.AddError(
				"Missing resource identity",
				"Terraform requested a refresh for this resource without any prior state or identity information.",
			)
			return
		}

		var identity deviceManagementServiceIdentityModel
		resp.Diagnostics.Append(req.Identity.Get(ctx, &identity)...)
		if resp.Diagnostics.HasError() {
			return
		}

		if identity.ID.IsNull() || identity.ID.IsUnknown() || identity.ID.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Missing device management service ID",
				"The resource identity did not include an 'id' attribute.",
			)
			return
		}

		data.ID = identity.ID
	} else {
		resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	readCtx, cancel, timeoutDiags := common.ResolveReadTimeout(ctx, data.Timeouts, common.DefaultReadTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	defer cancel()

	srv, err := r.client.GetDeviceManagementService(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read MDM server", err.Error())
		return
	}

	data.Name = types.StringValue(srv.Attributes.ServerName)
	data.Type = types.StringValue(srv.Attributes.ServerType)
	data.Status = common.StringPointerToTypesString(srv.Attributes.Status)
	data.DeviceCount = types.Int64PointerValue(srv.Attributes.DeviceCount)
	data.LastConnectedDateTime = types.StringPointerValue(srv.Attributes.LastConnectedDateTime)
	data.LastConnectedIp = types.StringPointerValue(srv.Attributes.LastConnectedIp)
	data.CreatedDateTime = types.StringValue(srv.Attributes.CreatedDateTime)
	data.UpdatedDateTime = types.StringValue(srv.Attributes.UpdatedDateTime)
	data.DefaultProductFamilies = common.StringsToList(ctx, srv.Attributes.DefaultProductFamilies)
	data.AllowRelease = types.BoolPointerValue(srv.Attributes.EnableMdmDisownFlag)

	deviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(readCtx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read device assignments", err.Error())
		return
	}

	deviceSet, diags := stringsToSet(deviceIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DeviceIDs = deviceSet

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, deviceManagementServiceIdentityModel{
			ID: types.StringValue(data.ID.ValueString()),
		})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	data.Timeouts = ensureDeviceManagementServiceTimeouts(data.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies changes to the MDM server attributes and device assignments.
func (r *DeviceManagementServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state MdmDeviceAssignmentModel

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

	if r.client.IsBusinessScope() {
		serverAttrs := client.MdmServerUpdateAttributes{}
		changed := false

		if !plan.Name.Equal(state.Name) {
			v := plan.Name.ValueString()
			serverAttrs.ServerName = &v
			changed = true
		}
		if !plan.AllowRelease.Equal(state.AllowRelease) {
			v := plan.AllowRelease.ValueBool()
			serverAttrs.EnableMdmDisownFlag = &v
			changed = true
		}
		if plan.ServerCertificate != nil && (state.ServerCertificate == nil ||
			!plan.ServerCertificate.Data.Equal(state.ServerCertificate.Data) ||
			!plan.ServerCertificate.Name.Equal(state.ServerCertificate.Name)) {
			serverAttrs.ServerCertificate = &client.MdmServerCertificate{
				Name: plan.ServerCertificate.Name.ValueString(),
				Data: plan.ServerCertificate.Data.ValueString(),
			}
			changed = true
		}

		if changed {
			srv, err := r.client.UpdateDeviceManagementService(updateCtx, client.MdmServerUpdateRequest{
				Data: client.MdmServerUpdateRequestData{
					Type:       "mdmServers",
					ID:         plan.ID.ValueString(),
					Attributes: serverAttrs,
				},
			})
			if err != nil {
				resp.Diagnostics.AddError("Failed to update MDM server", err.Error())
				return
			}
			plan.Type = types.StringValue(srv.Attributes.ServerType)
			plan.Status = common.StringPointerToTypesString(srv.Attributes.Status)
			plan.DeviceCount = types.Int64PointerValue(srv.Attributes.DeviceCount)
			plan.LastConnectedDateTime = types.StringPointerValue(srv.Attributes.LastConnectedDateTime)
			plan.LastConnectedIp = types.StringPointerValue(srv.Attributes.LastConnectedIp)
			plan.CreatedDateTime = types.StringValue(srv.Attributes.CreatedDateTime)
			plan.UpdatedDateTime = types.StringValue(srv.Attributes.UpdatedDateTime)
			plan.DefaultProductFamilies = common.StringsToList(ctx, srv.Attributes.DefaultProductFamilies)
			plan.AllowRelease = types.BoolPointerValue(srv.Attributes.EnableMdmDisownFlag)
		}
	}

	currentDeviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(updateCtx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get current device assignments", err.Error())
		return
	}

	plannedDevices := extractStrings(plan.DeviceIDs)
	plannedMap := make(map[string]bool, len(plannedDevices))
	for _, id := range plannedDevices {
		plannedMap[id] = true
	}
	currentMap := make(map[string]bool, len(currentDeviceIDs))
	for _, id := range currentDeviceIDs {
		currentMap[id] = true
	}

	planMigrations, err := migrationMap(plan.MigratedDevices)
	if err != nil {
		resp.Diagnostics.AddError("Invalid migrated_devices", err.Error())
		return
	}
	stateMigrations, err := migrationMap(state.MigratedDevices)
	if err != nil {
		resp.Diagnostics.AddError("Invalid migrated_devices", err.Error())
		return
	}

	var toUnassign []string
	for _, id := range currentDeviceIDs {
		if !plannedMap[id] {
			toUnassign = append(toUnassign, id)
		}
	}
	var toAssign []string
	for _, id := range plannedDevices {
		if !currentMap[id] {
			toAssign = append(toAssign, id)
		}
	}

	// Cancel migrations for devices dropped from migrated_devices but still assigned,
	// and for devices being unassigned that had an in-progress migration.
	var toCancel []string
	for serial := range stateMigrations {
		if plannedMap[serial] && planMigrations[serial] == "" {
			toCancel = append(toCancel, serial)
		}
	}
	for _, serial := range toUnassign {
		if _, ok := stateMigrations[serial]; ok {
			toCancel = append(toCancel, serial)
		}
	}

	// Update deadlines for devices already assigned where the deadline changed.
	toUpdate := make(map[string][]string)
	for serial, deadline := range planMigrations {
		if currentMap[serial] {
			if stateDeadline, ok := stateMigrations[serial]; ok && stateDeadline != deadline {
				toUpdate[deadline] = append(toUpdate[deadline], serial)
			}
		}
	}

	if len(toCancel) > 0 {
		r.runActivity(updateCtx, client.OrgDeviceActivityCancelMDMMigration, "", "", toCancel, "cancel MDM migration", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if len(toUnassign) > 0 {
		r.runActivity(updateCtx, client.OrgDeviceActivityUnassignDevices, "", plan.ID.ValueString(), toUnassign, "unassign devices", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	plain, byDeadline := partitionMigrations(toAssign, planMigrations)

	// Verify eligibility for devices being newly assigned with a migration
	// deadline before issuing any activity.
	var migrationSerials []string
	for _, group := range byDeadline {
		migrationSerials = append(migrationSerials, group.devices...)
	}
	if len(migrationSerials) > 0 {
		if err := verifyMigrationsCapable(updateCtx, migrationSerials, func(ctx context.Context, id string) (*client.OrgDevice, error) {
			return r.client.GetOrgDevice(ctx, id, nil)
		}, migrationEligibilityTimeout); err != nil {
			resp.Diagnostics.AddError("Failed to verify MDM migration eligibility", err.Error())
			return
		}
	}

	if len(plain) > 0 {
		r.runActivity(updateCtx, client.OrgDeviceActivityAssignDevices, "", plan.ID.ValueString(), plain, "assign devices", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	for _, group := range byDeadline {
		r.runActivity(updateCtx, client.OrgDeviceActivityAssignWithMDMMigrationDeadline, group.deadline, plan.ID.ValueString(), group.devices, "assign devices with migration deadline", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	for deadline, serials := range toUpdate {
		r.runActivity(updateCtx, client.OrgDeviceActivityUpdateMDMMigrationDeadline, deadline, "", serials, "update migration deadline", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if resp.Identity != nil {
		resp.Diagnostics.Append(resp.Identity.Set(ctx, deviceManagementServiceIdentityModel{
			ID: types.StringValue(plan.ID.ValueString()),
		})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Updated MDM server", map[string]any{
		"mdm_server_id": plan.ID.ValueString(),
		"assigned":      toAssign,
		"unassigned":    toUnassign,
	})

	plan.Timeouts = ensureDeviceManagementServiceTimeouts(plan.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes an MDM server (business scope only). In education scope it removes the resource from state.
func (r *DeviceManagementServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MdmDeviceAssignmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !r.client.IsBusinessScope() {
		return
	}

	deleteTimeout := defaultDeleteTimeout
	if !data.Timeouts.IsNull() && !data.Timeouts.IsUnknown() {
		configuredTimeout, timeoutDiags := data.Timeouts.Delete(ctx, defaultDeleteTimeout)
		resp.Diagnostics.Append(timeoutDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		deleteTimeout = configuredTimeout
	}

	deleteCtx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// GET the server first — confirms it exists and reveals current family assignments.
	srv, err := r.client.GetDeviceManagementService(deleteCtx, data.ID.ValueString(), nil)
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return
		}
		resp.Diagnostics.AddError("Failed to read MDM server before deletion", err.Error())
		return
	}

	// Clear default product family assignments only if any are set.
	// Apple returns 400 when PATCHing defaultProductFamilies: [] on a server that has none.
	if len(srv.Attributes.DefaultProductFamilies) > 0 {
		if _, err := r.client.ClearDeviceManagementServiceDefaultFamilies(deleteCtx, data.ID.ValueString()); err != nil {
			resp.Diagnostics.AddError("Failed to clear default product families before deletion", err.Error())
			return
		}
	}

	// Unassign all devices before deletion.
	currentDeviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(deleteCtx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get device assignments before deletion", err.Error())
		return
	}

	// Cancel any in-progress MDM migrations before unassigning their devices.
	stateMigrations, err := migrationMap(data.MigratedDevices)
	if err != nil {
		resp.Diagnostics.AddError("Invalid migrated_devices", err.Error())
		return
	}
	var toCancel []string
	for _, serial := range currentDeviceIDs {
		if _, ok := stateMigrations[serial]; ok {
			toCancel = append(toCancel, serial)
		}
	}
	if len(toCancel) > 0 {
		r.runActivity(deleteCtx, client.OrgDeviceActivityCancelMDMMigration, "", "", toCancel, "cancel MDM migration before deletion", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(currentDeviceIDs) > 0 {
		r.runActivity(deleteCtx, client.OrgDeviceActivityUnassignDevices, "", data.ID.ValueString(), currentDeviceIDs, "unassign devices before deletion", &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if err := r.client.DeleteDeviceManagementService(deleteCtx, data.ID.ValueString()); err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			return
		}
		resp.Diagnostics.AddError("Failed to delete MDM server", err.Error())
	}
}
