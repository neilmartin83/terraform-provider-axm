package device_management_service

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// deviceManagementServiceIdentityModel captures the fields that make up the
// resource identity shared between resource CRUD and terraform query list
// support.
type deviceManagementServiceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

// MdmDeviceAssignmentModel describes the Terraform state for device assignments.
type MdmDeviceAssignmentModel struct {
	ID        types.String   `tfsdk:"id"`
	Name      types.String   `tfsdk:"name"`
	Type      types.String   `tfsdk:"type"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
	DeviceIDs types.Set      `tfsdk:"device_ids"`
}

// DeviceManagementServiceListResourceModel captures filters supported by the list query.
type DeviceManagementServiceListResourceModel struct {
	Name         types.String `tfsdk:"name"`
	NameContains types.String `tfsdk:"name_contains"`
	ServerType   types.String `tfsdk:"server_type"`
}
