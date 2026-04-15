// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package blueprint

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BlueprintModel describes the Terraform state for a Blueprint resource.
type BlueprintModel struct {
	ID                  types.String   `tfsdk:"id"`
	Name                types.String   `tfsdk:"name"`
	Description         types.String   `tfsdk:"description"`
	Status              types.String   `tfsdk:"status"`
	AppLicenseDeficient types.Bool     `tfsdk:"app_license_deficient"`
	CreatedDateTime     types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime     types.String   `tfsdk:"updated_date_time"`
	AppIDs              types.Set      `tfsdk:"app_ids"`
	ConfigurationIDs    types.Set      `tfsdk:"configuration_ids"`
	PackageIDs          types.Set      `tfsdk:"package_ids"`
	DeviceIDs           types.Set      `tfsdk:"device_ids"`
	UserIDs             types.Set      `tfsdk:"user_ids"`
	UserGroupIDs        types.Set      `tfsdk:"user_group_ids"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}

// blueprintIdentityModel captures the fields that make up the resource identity
// shared between resource CRUD and terraform query list support.
type blueprintIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

// BlueprintListResourceModel captures filters supported by the list query.
type BlueprintListResourceModel struct {
	Name         types.String `tfsdk:"name"`
	NameContains types.String `tfsdk:"name_contains"`
	Status       types.String `tfsdk:"status"`
}
