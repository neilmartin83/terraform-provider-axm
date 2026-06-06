// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package default_device_assignment

import "github.com/hashicorp/terraform-plugin-framework/types"

// DefaultDeviceAssignmentModel describes the Terraform state for the org-wide default device assignment.
type DefaultDeviceAssignmentModel struct {
	ID             types.String `tfsdk:"id"`
	AppleTV        types.String `tfsdk:"apple_tv"`
	AppleVisionPro types.String `tfsdk:"apple_vision_pro"`
	IPad           types.String `tfsdk:"ipad"`
	IPhone         types.String `tfsdk:"iphone"`
	IPod           types.String `tfsdk:"ipod"`
	Mac            types.String `tfsdk:"mac"`
}
