package device_management_service

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Create handles the creation of device assignments. Currently only supports
// assigning devices to an existing MDM server.
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

	deviceIDs := extractStrings(data.DeviceIDs)

	activity, err := r.client.AssignDevicesToMDMServer(createCtx, data.ID.ValueString(), deviceIDs, true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to assign devices", err.Error())
		return
	}

	if err := r.waitForActivityCompletion(createCtx, activity.ID, &resp.Diagnostics); err != nil {
		resp.Diagnostics.AddError("Failed to complete device assignment", err.Error())
		return
	}

	service, found, err := r.fetchDeviceManagementService(createCtx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read device management service", err.Error())
		return
	}
	if !found {
		resp.Diagnostics.AddError(
			"Device management service not found",
			fmt.Sprintf("The device management service with ID %q could not be located via the GET_COLLECTION API.", data.ID.ValueString()),
		)
		return
	}

	data.Name = types.StringValue(service.Attributes.ServerName)
	data.Type = types.StringValue(service.Attributes.ServerType)

	if resp.Identity != nil {
		identity := deviceManagementServiceIdentityModel{
			ID: types.StringValue(data.ID.ValueString()),
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	tflog.Debug(ctx, "Assigned devices to MDM server", map[string]interface{}{
		"mdm_server_id": data.ID.ValueString(),
		"device_ids":    deviceIDs,
	})

	data.Timeouts = ensureDeviceManagementServiceTimeouts(data.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read retrieves the current state of device assignments from the MDM server
// and updates the Terraform state accordingly.
func (r *DeviceManagementServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MdmDeviceAssignmentModel

	if req.State.Raw.IsNull() {
		if req.Identity == nil {
			resp.Diagnostics.AddError(
				"Missing resource identity",
				"Terraform requested a refresh for this resource without any prior state or identity information, so the provider cannot determine which device management service to query.",
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
				"The resource identity did not include an 'id' attribute, so the provider cannot refresh the device management service.",
			)
			return
		}

		data.ID = identity.ID
		data.Timeouts = newDeviceManagementServiceTimeoutsNullValue()
	} else {
		resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	readTimeout := defaultReadTimeout
	if !data.Timeouts.IsNull() && !data.Timeouts.IsUnknown() {
		configuredTimeout, timeoutDiags := data.Timeouts.Read(ctx, defaultReadTimeout)
		resp.Diagnostics.Append(timeoutDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		readTimeout = configuredTimeout
	}

	readCtx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	deviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(readCtx, data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "NOT_FOUND") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading device assignments",
			fmt.Sprintf("Failed to get device assignments: %s", err),
		)
		return
	}

	elements := make([]attr.Value, len(deviceIDs))
	for i, id := range deviceIDs {
		elements[i] = types.StringValue(id)
	}
	deviceSet, diags := types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.DeviceIDs = deviceSet

	service, found, err := r.fetchDeviceManagementService(readCtx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading device management service",
			fmt.Sprintf("Failed to get device management service: %s", err),
		)
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Name = types.StringValue(service.Attributes.ServerName)
	data.Type = types.StringValue(service.Attributes.ServerType)

	if resp.Identity != nil {
		identity := deviceManagementServiceIdentityModel{
			ID: types.StringValue(data.ID.ValueString()),
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	data.Timeouts = ensureDeviceManagementServiceTimeouts(data.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update handles changes to device assignments by calculating and applying
// the difference between current and desired state.
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

	currentDeviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(updateCtx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get current device assignments", err.Error())
		return
	}

	currentElements := make([]attr.Value, len(currentDeviceIDs))
	for i, id := range currentDeviceIDs {
		currentElements[i] = types.StringValue(id)
	}
	currentSet, diags := types.SetValue(types.StringType, currentElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	toUnassign := extractStrings(currentSet)
	plannedDevices := extractStrings(plan.DeviceIDs)
	plannedMap := make(map[string]bool)
	for _, id := range plannedDevices {
		plannedMap[id] = true
	}

	var devicesToUnassign []string
	for _, id := range toUnassign {
		if !plannedMap[id] {
			devicesToUnassign = append(devicesToUnassign, id)
		}
	}

	currentMap := make(map[string]bool)
	for _, id := range currentDeviceIDs {
		currentMap[id] = true
	}

	var devicesToAssign []string
	for _, id := range plannedDevices {
		if !currentMap[id] {
			devicesToAssign = append(devicesToAssign, id)
		}
	}

	if len(devicesToUnassign) > 0 {
		activity, err := r.client.AssignDevicesToMDMServer(updateCtx, plan.ID.ValueString(), devicesToUnassign, false)
		if err != nil {
			resp.Diagnostics.AddError("Failed to unassign devices", err.Error())
			return
		}
		if err := r.waitForActivityCompletion(updateCtx, activity.ID, &resp.Diagnostics); err != nil {
			resp.Diagnostics.AddError("Failed to complete device unassignment", err.Error())
			return
		}
	}

	if len(devicesToAssign) > 0 {
		activity, err := r.client.AssignDevicesToMDMServer(updateCtx, plan.ID.ValueString(), devicesToAssign, true)
		if err != nil {
			resp.Diagnostics.AddError("Failed to assign devices", err.Error())
			return
		}
		if err := r.waitForActivityCompletion(updateCtx, activity.ID, &resp.Diagnostics); err != nil {
			resp.Diagnostics.AddError("Failed to complete device assignment", err.Error())
			return
		}
	}

	service, found, err := r.fetchDeviceManagementService(updateCtx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read device management service", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	plan.Name = types.StringValue(service.Attributes.ServerName)
	plan.Type = types.StringValue(service.Attributes.ServerType)

	if resp.Identity != nil {
		identity := deviceManagementServiceIdentityModel{
			ID: types.StringValue(plan.ID.ValueString()),
		}
		resp.Diagnostics.Append(resp.Identity.Set(ctx, identity)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	tflog.Debug(ctx, "Updated device assignments for MDM server", map[string]interface{}{
		"mdm_server_id": plan.ID.ValueString(),
		"assigned":      devicesToAssign,
		"unassigned":    devicesToUnassign,
	})

	plan.Timeouts = ensureDeviceManagementServiceTimeouts(plan.Timeouts)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete handles resource deletion. Currently only removes the resource from
// Terraform state as API doesn't support MDM server deletion.
func (r *DeviceManagementServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MdmDeviceAssignmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}
