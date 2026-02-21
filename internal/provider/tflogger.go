package provider

import (
	"context"
	"encoding/json"
	"net/http"

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
	var obj any
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
	fields := map[string]any{
		"method": method,
		"url":    url,
	}

	if len(body) > 0 {
		fields["request_body"] = prettyPrintJSON(body)
	}

	tflog.Debug(ctx, "HTTP Request", fields)
}

// LogResponse logs HTTP response details using tflog at DEBUG level
func (l *TerraformLogger) LogResponse(ctx context.Context, statusCode int, headers http.Header, body []byte) {
	fields := map[string]any{
		"status_code": statusCode,
	}

	if len(headers) > 0 {
		headerMap := make(map[string]any)
		for key, values := range headers {
			if len(values) == 1 {
				headerMap[key] = values[0]
			} else {
				headerMap[key] = values
			}
		}
		fields["headers"] = headerMap
	}

	if len(body) > 0 {
		bodyStr := string(body)
		if len(bodyStr) > 5000 {
			bodyStr = bodyStr[:5000] + "... (truncated)"
			fields["response_body"] = bodyStr
		} else {
			fields["response_body"] = prettyPrintJSON(body)
		}
	}

	tflog.Debug(ctx, "HTTP Response", fields)
}

// LogAuth logs authentication-related events using tflog at DEBUG level
func (l *TerraformLogger) LogAuth(ctx context.Context, message string, fields map[string]any) {
	tflog.Debug(ctx, message, fields)
}
