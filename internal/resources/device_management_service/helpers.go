package device_management_service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

// extractStrings converts a types.Set containing string values into a slice of strings,
// handling null and unknown values appropriately.
func extractStrings(set types.Set) []string {
	var result []string
	if set.IsNull() || set.IsUnknown() {
		return result
	}
	for _, v := range set.Elements() {
		if strVal, ok := v.(types.String); ok && !strVal.IsUnknown() && !strVal.IsNull() {
			result = append(result, strVal.ValueString())
		}
	}
	return result
}

func stringsToSet(values []string) (types.Set, diag.Diagnostics) {
	elements := make([]attr.Value, len(values))
	for i, value := range values {
		elements[i] = types.StringValue(value)
	}

	return types.SetValue(types.StringType, elements)
}

// downloadAndParseActivityLog downloads the CSV from a pre-signed URL and parses it into a summary.
// This is a standalone function (not a client method) because the URL is pre-signed and doesn't
// require authentication - it's a utility operation, not an API call.
func downloadAndParseActivityLog(ctx context.Context, downloadURL string) (string, error) {
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL provided")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download activity log: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download activity log: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read activity log: %w", err)
	}

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil {
		return "", fmt.Errorf("failed to parse CSV: %w", err)
	}

	var summary strings.Builder
	var inDataSection bool
	var headers []string
	var errorCount int
	var errors []map[string]string

	for _, record := range records {
		if len(record) == 0 {
			continue
		}

		allEmpty := true
		for _, field := range record {
			if strings.TrimSpace(field) != "" {
				allEmpty = false
				break
			}
		}
		if allEmpty {
			continue
		}

		if !inDataSection && len(record) > 0 && strings.Contains(strings.ToLower(record[0]), "serial_number") {
			inDataSection = true
			headers = record
			continue
		}

		if inDataSection && len(headers) > 0 {
			rowData := make(map[string]string)
			for i, header := range headers {
				if i < len(record) {
					rowData[strings.TrimSpace(header)] = strings.TrimSpace(record[i])
				}
			}

			status := rowData["operation_status"]
			if status != "" && status != "SUCCESS" {
				errorCount++
				errors = append(errors, rowData)
			}
		}
	}

	if errorCount == 0 {
		summary.WriteString("Activity completed but detailed results are available in the activity log.")
	} else {
		summary.WriteString(fmt.Sprintf("Activity completed with %d error(s):\n\n", errorCount))

		for i, errRow := range errors {
			if i >= 10 {
				summary.WriteString(fmt.Sprintf("... and %d more error(s). Check Apple Business Manager console for full details.\n", errorCount-10))
				break
			}

			serial := errRow["serial_number"]
			status := errRow["operation_status"]
			subStatus := errRow["operation_substatus"]

			summary.WriteString(fmt.Sprintf("  • Serial: %s - Status: %s", serial, status))
			if subStatus != "" {
				summary.WriteString(fmt.Sprintf(" (%s)", subStatus))
			}
			summary.WriteString("\n")
		}
	}

	return summary.String(), nil
}

// fetchDeviceManagementService retrieves metadata for a specific MDM server by scanning the available
// collection endpoint (GET_INSTANCE is not permitted by the upstream API).
func (r *DeviceManagementServiceResource) fetchDeviceManagementService(ctx context.Context, serverID string) (*client.MdmServer, bool, error) {
	servers, err := r.client.GetDeviceManagementServices(ctx, nil)
	if err != nil {
		return nil, false, err
	}

	for i := range servers {
		if servers[i].ID == serverID {
			server := servers[i]
			return &server, true, nil
		}
	}

	return nil, false, nil
}

// waitForActivityCompletion polls the activity status until it completes, fails, or times out
func (r *DeviceManagementServiceResource) waitForActivityCompletion(ctx context.Context, activityID string, diags *diag.Diagnostics) error {
	maxAttempts := 30
	retryInterval := 5 * time.Second

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(retryInterval):
		}

		activity, err := r.client.GetOrgDeviceActivity(ctx, activityID, nil)
		if err != nil {
			return fmt.Errorf("error checking activity status: %w", err)
		}

		switch activity.Attributes.Status {
		case "COMPLETED":
			if activity.Attributes.SubStatus != "COMPLETED_WITH_SUCCESS" {
				summary := fmt.Sprintf("Activity ID: %s\n\nCompleted with SubStatus: %s", activityID, activity.Attributes.SubStatus)

				if activity.Attributes.DownloadURL != "" {
					logSummary, err := downloadAndParseActivityLog(ctx, activity.Attributes.DownloadURL)
					if err == nil {
						summary = fmt.Sprintf("Activity ID: %s\n\n%s", activityID, logSummary)
					} else {
						summary = fmt.Sprintf("%s\n\nFailed to download activity log: %v\n\nActivity log available at: %s", summary, err, activity.Attributes.DownloadURL)
					}
				}

				diags.AddWarning(
					"Device operation completed with errors. Please check the Activity Log in the AxM portal for more details.",
					summary,
				)
			}
			return nil
		case "FAILED":
			return fmt.Errorf("activity failed with sub-status: %s", activity.Attributes.SubStatus)
		case "STOPPED":
			return fmt.Errorf("activity stopped with sub-status: %s", activity.Attributes.SubStatus)
		case "IN_PROGRESS":
			if attempt == maxAttempts {
				return fmt.Errorf("timed out waiting for activity to complete after %d attempts", maxAttempts)
			}
		default:
			return fmt.Errorf("unknown activity status: %s", activity.Attributes.Status)
		}
	}

	return fmt.Errorf("unexpected error monitoring activity status")
}
