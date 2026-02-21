// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

// ConfigureClient extracts a *client.Client from the provider data passed during
// Configure. The constructKind parameter is used in error messages (e.g. "Data Source",
// "Resource", "List").
func ConfigureClient(providerData any, constructKind string) (*client.Client, diag.Diagnostics) {
	var diags diag.Diagnostics

	if providerData == nil {
		return nil, diags
	}

	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError(
			fmt.Sprintf("Unexpected %s Configure Type", constructKind),
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", providerData),
		)
		return nil, diags
	}

	return c, diags
}
