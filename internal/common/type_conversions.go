// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringsToTypesStrings converts a []T (where T's underlying type is string) to
// a []types.String suitable for Terraform list attributes in data source models.
func StringsToTypesStrings[T ~string](values []T) []types.String {
	result := make([]types.String, len(values))
	for i, v := range values {
		result[i] = types.StringValue(string(v))
	}
	return result
}

// StringsToList converts a []T (where T's underlying type is string) to a
// types.List suitable for Terraform list attributes in resource and data source models.
func StringsToList[T ~string](ctx context.Context, values []T) types.List {
	strs := make([]string, len(values))
	for i, v := range values {
		strs[i] = string(v)
	}
	list, _ := types.ListValueFrom(ctx, types.StringType, strs)
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

// BoolPointerToBoolValue converts a *bool to a types.Bool, handling nil.
func BoolPointerToBoolValue(b *bool) types.Bool {
	if b == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*b)
}

// StringPointerToTypesString converts a *T (where T's underlying type is string)
// to a types.String, handling nil.
func StringPointerToTypesString[T ~string](v *T) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*v))
}
