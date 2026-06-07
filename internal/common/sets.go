// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// SetToStrings converts a types.Set containing string values into a slice of strings.
func SetToStrings(set types.Set) []string {
	var result []string
	if set.IsNull() || set.IsUnknown() {
		return result
	}
	for _, v := range set.Elements() {
		if strVal, ok := v.(types.String); ok && !strVal.IsUnknown() && !strVal.IsNull() {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

// StringsFromSet converts a types.Set of string values into a []T (where T's underlying type is string).
func StringsFromSet[T ~string](set types.Set) []T {
	var result []T
	if set.IsNull() || set.IsUnknown() {
		return result
	}
	for _, v := range set.Elements() {
		if strVal, ok := v.(types.String); ok && !strVal.IsUnknown() && !strVal.IsNull() {
			result = append(result, T(strVal.ValueString()))
		}
	}
	return result
}

// StringsToSet converts a slice of T (where T's underlying type is string) into a types.Set of string values.
func StringsToSet[T ~string](values []T) (types.Set, diag.Diagnostics) {
	var diags diag.Diagnostics
	elements := make([]attr.Value, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(string(value))
	}

	set, setDiags := types.SetValue(types.StringType, elements)
	diags.Append(setDiags...)
	if diags.HasError() {
		return types.SetNull(types.StringType), diags
	}

	return set, diags
}
