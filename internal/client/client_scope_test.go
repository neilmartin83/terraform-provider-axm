// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import "testing"

func TestClientScopeAccessors(t *testing.T) {
	c := &Client{scope: "business.api"}
	if c.Scope() != "business.api" {
		t.Fatalf("expected scope business.api, got %s", c.Scope())
	}
	if !c.IsBusinessScope() {
		t.Fatal("expected IsBusinessScope to be true")
	}
}
