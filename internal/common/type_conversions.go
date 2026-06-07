// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"

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

// StringsToList converts a []string to a types.List suitable for Terraform list
// attributes in resource and data source models.
func StringsToList(ctx context.Context, values []string) types.List {
	list, _ := types.ListValueFrom(ctx, types.StringType, values)
	return list
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
