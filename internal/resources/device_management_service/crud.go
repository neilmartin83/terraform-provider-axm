package device_management_service

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Create handles the creation of device assignments. Currently only supports
// assigning devices to an existing MDM server.
func (r *deviceManagementServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan mdmDeviceAssignmentModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceIDs := extractStrings(plan.DeviceIDs)

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

	_, err := r.client.AssignDevicesToMDMServer(ctx, plan.ID.ValueString(), deviceIDs, true)
	if err != nil {
		resp.Diagnostics.AddError("Failed to assign devices", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read retrieves the current state of device assignments from the MDM server
// and updates the Terraform state accordingly.
func (r *deviceManagementServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state mdmDeviceAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceIDs, err := r.client.GetDeviceManagementServiceSerialNumbers(ctx, state.ID.ValueString())
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

	currentStateDevices := make(map[string]struct{})
	for _, id := range extractStrings(state.DeviceIDs) {
		currentStateDevices[id] = struct{}{}
	}

	apiDevices := make(map[string]struct{})
	for _, id := range deviceIDs {
		apiDevices[id] = struct{}{}
	}

	if !setsEqual(currentStateDevices, apiDevices) {
		elements := make([]attr.Value, len(deviceIDs))
		for i, id := range deviceIDs {
			elements[i] = types.StringValue(id)
		}
		deviceList, diags := types.ListValue(types.StringType, elements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.DeviceIDs = deviceList
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	}
}

// Update handles changes to device assignments by calculating and applying
// the difference between current and desired state.
func (r *deviceManagementServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan mdmDeviceAssignmentModel
	var state mdmDeviceAssignmentModel

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

	currentSet := make(map[string]struct{})
	for _, id := range currentDeviceIDs {
		currentSet[id] = struct{}{}
	}

	plannedDevices := extractStrings(plan.DeviceIDs)
	desiredSet := make(map[string]struct{})
	for _, id := range plannedDevices {
		desiredSet[id] = struct{}{}
	}

	var toUnassign []string
	for id := range currentSet {
		if _, exists := desiredSet[id]; !exists {
			toUnassign = append(toUnassign, id)
		}
	}

	var toAssign []string
	for id := range desiredSet {
		if _, exists := currentSet[id]; !exists {
			toAssign = append(toAssign, id)
		}
	}

	if len(toAssign) > 0 {
		if errors := r.validateDevices(ctx, toAssign); len(errors) > 0 {
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

	if len(toUnassign) > 0 {
		_, err := r.client.AssignDevicesToMDMServer(ctx, plan.ID.ValueString(), toUnassign, false)
		if err != nil {
			resp.Diagnostics.AddError("Failed to unassign devices", err.Error())
			return
		}
	}

	if len(toAssign) > 0 {
		_, err := r.client.AssignDevicesToMDMServer(ctx, plan.ID.ValueString(), toAssign, true)
		if err != nil {
			resp.Diagnostics.AddError("Failed to assign devices", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete handles resource deletion. Currently only removes the resource from
// Terraform state as API doesn't support MDM server deletion.
func (r *deviceManagementServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state mdmDeviceAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}
