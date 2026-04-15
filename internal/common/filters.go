// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NormalizedFilterString returns a trimmed string and whether it should be used as a filter.
func NormalizedFilterString(value types.String) (string, bool) {
	if value.IsNull() || value.IsUnknown() {
		return "", false
	}

	trimmed := strings.TrimSpace(value.ValueString())
	if trimmed == "" {
		return "", false
	}

	return trimmed, true
}
