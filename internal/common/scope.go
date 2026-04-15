// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

const businessScope = "business.api"

// RequireBusinessScope validates that the configured client scope is business.api.
func RequireBusinessScope(c *client.Client, diags *diag.Diagnostics, constructName string) bool {
	if c == nil {
		return false
	}
	if c.Scope() == businessScope {
		return true
	}
	diags.AddError(
		"Unsupported Scope",
		fmt.Sprintf("%s is only available when scope is set to %q.", constructName, businessScope),
	)
	return false
}
