package common

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// DefaultReadTimeout is the standard read timeout used by data sources and resources
// when no user-configured timeout is specified.
const DefaultReadTimeout = 90 * time.Second

// TimeoutReader abstracts both datasource/timeouts.Value and resource/timeouts.Value
// so that timeout resolution can be shared across construct types.
type TimeoutReader interface {
	IsNull() bool
	IsUnknown() bool
	Read(ctx context.Context, defaultTimeout time.Duration) (time.Duration, diag.Diagnostics)
}

// ResolveReadTimeout resolves a configured read timeout from a TimeoutReader or falls back
// to the provided default. It returns a child context with the resolved deadline and a
// cancel function.
func ResolveReadTimeout(ctx context.Context, tr TimeoutReader, defaultTimeout time.Duration) (context.Context, context.CancelFunc, diag.Diagnostics) {
	var diags diag.Diagnostics

	readTimeout := defaultTimeout
	if !tr.IsNull() && !tr.IsUnknown() {
		configuredTimeout, timeoutDiags := tr.Read(ctx, defaultTimeout)
		diags.Append(timeoutDiags...)
		if diags.HasError() {
			cancel := func() {}
			return ctx, cancel, diags
		}
		readTimeout = configuredTimeout
	}

	readCtx, cancel := context.WithTimeout(ctx, readTimeout)
	return readCtx, cancel, diags
}
