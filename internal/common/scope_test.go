// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func TestRequireBusinessScope(t *testing.T) {
	businessClient, err := client.NewClient("https://example.com", "team", "client", "key", "business.api", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var diags diag.Diagnostics
	if ok := RequireBusinessScope(businessClient, &diags, "axm_blueprint"); !ok {
		t.Fatal("expected business scope to be accepted")
	}

	schoolClient, err := client.NewClient("https://example.com", "team", "client", "key", "school.api", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diags = diag.Diagnostics{}
	if ok := RequireBusinessScope(schoolClient, &diags, "axm_blueprint"); ok {
		t.Fatal("expected school scope to be rejected")
	}
	if !diags.HasError() {
		t.Fatal("expected diagnostics to include an error")
	}
}
