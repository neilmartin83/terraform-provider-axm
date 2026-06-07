// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

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

// MdmServerCertificateModel holds the certificate name and base64-encoded data.
type MdmServerCertificateModel struct {
	Name types.String `tfsdk:"name"`
	Data types.String `tfsdk:"data"`
}

// MdmDeviceAssignmentModel describes the Terraform state for an MDM server and its device assignments.
type MdmDeviceAssignmentModel struct {
	ID                     types.String               `tfsdk:"id"`
	Name                   types.String               `tfsdk:"name"`
	Type                   types.String               `tfsdk:"type"`
	Status                 types.String               `tfsdk:"status"`
	DeviceCount            types.Int64                `tfsdk:"device_count"`
	DefaultProductFamilies types.List                 `tfsdk:"default_product_families"`
	LastConnectedDateTime  types.String               `tfsdk:"last_connected_date_time"`
	LastConnectedIp        types.String               `tfsdk:"last_connected_ip"`
	CreatedDateTime        types.String               `tfsdk:"created_date_time"`
	UpdatedDateTime        types.String               `tfsdk:"updated_date_time"`
	AllowRelease           types.Bool                 `tfsdk:"allow_release"`
	ServerCertificate      *MdmServerCertificateModel `tfsdk:"server_certificate"`
	Timeouts               timeouts.Value             `tfsdk:"timeouts"`
	DeviceIDs              types.Set                  `tfsdk:"device_ids"`
}

// DeviceManagementServiceListResourceModel captures filters supported by the list query.
type DeviceManagementServiceListResourceModel struct {
	Name         types.String `tfsdk:"name"`
	NameContains types.String `tfsdk:"name_contains"`
	ServerType   types.String `tfsdk:"server_type"`
}
