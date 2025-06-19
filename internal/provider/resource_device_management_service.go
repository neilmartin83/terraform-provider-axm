package axm

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &deviceManagementServiceResource{}

type deviceManagementServiceResource struct {
	client *Client
}

type mdmDeviceAssignmentModel struct {
	ID        types.String `tfsdk:"id"`
	DeviceIDs types.List   `tfsdk:"device_ids"`
}

// NewDeviceManagementServiceResource creates a new instance of deviceManagementServiceResource
// with the provided client for managing device assignments.
func NewDeviceManagementServiceResource(client *Client) resource.Resource {
	return &deviceManagementServiceResource{
		client: client,
	}
}

// Metadata sets the provider type name for the resource.
func (r *deviceManagementServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_management_service"
}

// Schema defines the schema for the resource.
func (r *deviceManagementServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages device assignments to a specific Apple Business Manager MDM server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "MDM server ID. This is a unique ID for the server and is visible in the browser address bar when navigating to Preferences and selecting the desired 'Device Management Service'. Required until creation is supported.",
			},
			"device_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
				Description: "A list of device IDs to assign to the MDM server. These are device serial numbers.",
			},
		},
	}
}

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

// extractStrings converts a types.List containing string values into a slice of strings,
// handling null and unknown values appropriately.
func extractStrings(list types.List) []string {
	var result []string
	if list.IsNull() || list.IsUnknown() {
		return result
	}
	for _, v := range list.Elements() {
		if strVal, ok := v.(types.String); ok && !strVal.IsUnknown() && !strVal.IsNull() {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

// setsEqual compares two sets represented as maps and returns true if they contain
// exactly the same elements, false otherwise.
func setsEqual(a, b map[string]struct{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if _, exists := b[k]; !exists {
			return false
		}
	}
	return true
}

// validateDevices checks all devices and returns a list of validation errors
func (r *deviceManagementServiceResource) validateDevices(ctx context.Context, deviceIDs []string) []error {
	queryParams := url.Values{}
	queryParams.Add("fields[orgDevices]", "serialNumber")

	var errors []error
	for _, id := range deviceIDs {
		device, err := r.client.GetOrgDevice(ctx, id, queryParams)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to validate device %s: %s", id, err))
			continue
		}
		if device == nil {
			errors = append(errors, fmt.Errorf("device %s not found in Apple Business Manager", id))
		}
	}
	return errors
}
