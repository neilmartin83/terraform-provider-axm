package device_management_service

import "github.com/hashicorp/terraform-plugin-framework/types"

// deviceManagementServiceIdentityModel captures the fields that make up the
// resource identity shared between resource CRUD and terraform query list
// support.
type deviceManagementServiceIdentityModel struct {
	ID types.String `tfsdk:"id"`
}
