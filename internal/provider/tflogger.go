package provider

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

// Ensure TerraformLogger implements client.Logger interface
var _ client.Logger = (*TerraformLogger)(nil)

// TerraformLogger implements the client.Logger interface using tflog
type TerraformLogger struct{}

// NewTerraformLogger creates a new TerraformLogger
func NewTerraformLogger() *TerraformLogger {
	return &TerraformLogger{}
}

// prettyPrintJSON attempts to format JSON for better readability
func prettyPrintJSON(data []byte) string {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return string(data)
	}

	pretty, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return string(data)
	}

	return string(pretty)
}

// LogRequest logs HTTP request details using tflog at DEBUG level
func (l *TerraformLogger) LogRequest(ctx context.Context, method, url string, body []byte) {
	fields := map[string]interface{}{
		"method": method,
		"url":    url,
	}

	if len(body) > 0 {
		fields["request_body"] = prettyPrintJSON(body)
	}

	tflog.Debug(ctx, "HTTP Request", fields)
}

// LogResponse logs HTTP response details using tflog at DEBUG level
func (l *TerraformLogger) LogResponse(ctx context.Context, statusCode int, body []byte) {
	fields := map[string]interface{}{
		"status_code": statusCode,
	}

	if len(body) > 0 {
		bodyStr := string(body)
		if len(bodyStr) > 5000 {
			// For large responses, truncate before pretty printing
			bodyStr = bodyStr[:5000] + "... (truncated)"
			fields["response_body"] = bodyStr
		} else {
			fields["response_body"] = prettyPrintJSON(body)
		}
	}

	tflog.Debug(ctx, "HTTP Response", fields)
}

// LogAuth logs authentication-related events using tflog at DEBUG level
func (l *TerraformLogger) LogAuth(ctx context.Context, message string, fields map[string]interface{}) {
	tflog.Debug(ctx, message, fields)
}
