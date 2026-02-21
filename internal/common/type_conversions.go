package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringsToTypesStrings converts a []string to a []types.String suitable for
// Terraform list attributes in data source models.
func StringsToTypesStrings(values []string) []types.String {
	result := make([]types.String, len(values))
	for i, v := range values {
		result[i] = types.StringValue(v)
	}
	return result
}

// StringPointerOrNil returns a pointer to the string if it is non-empty,
// otherwise nil. This is useful for optional API fields that should map
// to null Terraform attributes.
func StringPointerOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
