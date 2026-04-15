// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringsToSetAndSetToStrings(t *testing.T) {
	values := []string{"a", "b", "c"}
	set, diags := StringsToSet(values)
	if diags.HasError() {
		t.Fatal("expected no diagnostics errors")
	}

	if set.IsNull() || set.IsUnknown() {
		t.Fatal("expected non-null set")
	}

	out := SetToStrings(set)
	if len(out) != 3 {
		t.Fatalf("expected 3 values, got %d", len(out))
	}

	emptySet := types.SetNull(types.StringType)
	if got := SetToStrings(emptySet); len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}
