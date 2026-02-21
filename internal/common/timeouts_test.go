// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

type mockTimeoutReader struct {
	isNull       bool
	isUnknown    bool
	readDuration time.Duration
	readDiags    diag.Diagnostics
}

func (m *mockTimeoutReader) IsNull() bool    { return m.isNull }
func (m *mockTimeoutReader) IsUnknown() bool { return m.isUnknown }
func (m *mockTimeoutReader) Read(_ context.Context, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics) {
	if m.readDiags.HasError() {
		return 0, m.readDiags
	}
	if m.readDuration > 0 {
		return m.readDuration, nil
	}
	return defaultTimeout, nil
}

func TestResolveReadTimeout_NullTimeout(t *testing.T) {
	tr := &mockTimeoutReader{isNull: true}
	ctx, cancel, diags := ResolveReadTimeout(context.Background(), tr, 90*time.Second)
	defer cancel()

	if diags.HasError() {
		t.Fatalf("expected no errors, got %v", diags.Errors())
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have deadline")
	}
	remaining := time.Until(deadline)
	if remaining < 85*time.Second || remaining > 95*time.Second {
		t.Fatalf("expected deadline ~90s from now, got %v", remaining)
	}
}

func TestResolveReadTimeout_UnknownTimeout(t *testing.T) {
	tr := &mockTimeoutReader{isUnknown: true}
	ctx, cancel, diags := ResolveReadTimeout(context.Background(), tr, 90*time.Second)
	defer cancel()

	if diags.HasError() {
		t.Fatalf("expected no errors, got %v", diags.Errors())
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have deadline")
	}
	remaining := time.Until(deadline)
	if remaining < 85*time.Second || remaining > 95*time.Second {
		t.Fatalf("expected deadline ~90s from now, got %v", remaining)
	}
}

func TestResolveReadTimeout_ConfiguredTimeout(t *testing.T) {
	tr := &mockTimeoutReader{readDuration: 120 * time.Second}
	ctx, cancel, diags := ResolveReadTimeout(context.Background(), tr, 90*time.Second)
	defer cancel()

	if diags.HasError() {
		t.Fatalf("expected no errors, got %v", diags.Errors())
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("expected context to have deadline")
	}
	remaining := time.Until(deadline)
	if remaining < 115*time.Second || remaining > 125*time.Second {
		t.Fatalf("expected deadline ~120s from now, got %v", remaining)
	}
}

func TestResolveReadTimeout_ReadError(t *testing.T) {
	errorDiags := diag.Diagnostics{}
	errorDiags.AddError("Timeout Error", "failed to parse timeout")

	tr := &mockTimeoutReader{readDiags: errorDiags}
	_, cancel, diags := ResolveReadTimeout(context.Background(), tr, 90*time.Second)
	defer cancel()

	if !diags.HasError() {
		t.Fatal("expected errors, got none")
	}
}
