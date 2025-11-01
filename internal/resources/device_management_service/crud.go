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

	deviceIDs := extractStrings(data.DeviceIDs)

	if errors := r.validateDevices(ctx, deviceIDs); len(errors) > 0 {
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		resp.Diagnostics.AddError(
			"Device Validation Errors",
			fmt.Sprintf("Multiple devices failed validation:\n- %s",
				strings.Join(errorMessages, "\n- ")),
		)
		return
	}

	_, err := r.client.AssignDevicesToMDMServer(ctx, data.ID.ValueString(), deviceIDs, true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to assign devices", err.Error())
		return
	}
	tflog.Trace(ctx, "Assigned devices to MDM server", map[string]interface{}{
		"mdm_server_id": data.ID.ValueString(),
		"device_ids":    deviceIDs,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read retrieves the current state of device assignments from the MDM server
// and updates the Terraform state accordingly.
func (r *DeviceManagementServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MdmDeviceAssignmentModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(ctx, data.ID.ValueString())
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

	currentDeviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(ctx, plan.ID.ValueString())
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

	if len(devicesToAssign) > 0 {
		if errors := r.validateDevices(ctx, devicesToAssign); len(errors) > 0 {
			var errorMessages []string
			for _, err := range errors {
				errorMessages = append(errorMessages, err.Error())
			}
			resp.Diagnostics.AddError(
				"Device Validation Errors",
				fmt.Sprintf("Multiple devices failed validation:\n- %s",
					strings.Join(errorMessages, "\n- ")),
			)
			return
		}
	}

	if len(devicesToUnassign) > 0 {
		_, err := r.client.AssignDevicesToMDMServer(ctx, plan.ID.ValueString(), devicesToUnassign, false)
		if err != nil {
			resp.Diagnostics.AddError("Failed to unassign devices", err.Error())
			return
		}
	}

	if len(devicesToAssign) > 0 {
		_, err := r.client.AssignDevicesToMDMServer(ctx, plan.ID.ValueString(), devicesToAssign, true)
		if err != nil {
			resp.Diagnostics.AddError("Failed to assign devices", err.Error())
			return
		}
	}
	tflog.Trace(ctx, "Updated device assignments for MDM server", map[string]interface{}{
		"mdm_server_id": plan.ID.ValueString(),
		"assigned":      devicesToAssign,
		"unassigned":    devicesToUnassign,
	})

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
