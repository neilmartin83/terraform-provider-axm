// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration

import (
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ConfigurationModel describes the Terraform state for a Configuration resource.
type ConfigurationModel struct {
	ID                     types.String   `tfsdk:"id"`
	Name                   types.String   `tfsdk:"name"`
	Type                   types.String   `tfsdk:"type"`
	ConfiguredForPlatforms types.Set      `tfsdk:"configured_for_platforms"`
	ConfigurationProfile   types.String   `tfsdk:"configuration_profile"`
	Filename               types.String   `tfsdk:"filename"`
	CreatedDateTime        types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime        types.String   `tfsdk:"updated_date_time"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}

// configurationIdentityModel captures the fields that make up the resource identity
// shared between resource CRUD and terraform query list support.
type configurationIdentityModel struct {
	ID types.String `tfsdk:"id"`
}

// ConfigurationListResourceModel captures filters supported by the list query.
type ConfigurationListResourceModel struct {
	Name         types.String `tfsdk:"name"`
	NameContains types.String `tfsdk:"name_contains"`
}
