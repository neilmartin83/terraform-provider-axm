// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func TestExtractStrings(t *testing.T) {
	tests := []struct {
		name string
		set  types.Set
		want []string
	}{
		{
			name: "null_set",
			set:  types.SetNull(types.StringType),
			want: nil,
		},
		{
			name: "unknown_set",
			set:  types.SetUnknown(types.StringType),
			want: nil,
		},
		{
			name: "empty_set",
			set:  types.SetValueMust(types.StringType, []attr.Value{}),
			want: nil,
		},
		{
			name: "single_element",
			set:  types.SetValueMust(types.StringType, []attr.Value{types.StringValue("ABC123")}),
			want: []string{"ABC123"},
		},
		{
			name: "multiple_elements",
			set: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue("SN001"),
				types.StringValue("SN002"),
				types.StringValue("SN003"),
			}),
			want: []string{"SN001", "SN002", "SN003"},
		},
		{
			name: "mixed_null_unknown_elements",
			set: types.SetValueMust(types.StringType, []attr.Value{
				types.StringNull(),
				types.StringValue("valid"),
				types.StringUnknown(),
			}),
			want: []string{"valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStrings(tt.set)
			if len(result) != len(tt.want) {
				t.Fatalf("expected %d elements, got %d: %v", len(tt.want), len(result), result)
			}
			for i, want := range tt.want {
				if result[i] != want {
					t.Errorf("element[%d]: expected %q, got %q", i, want, result[i])
				}
			}
		})
	}
}

func TestStringsToSet(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  int
	}{
		{name: "empty_slice", input: []string{}, want: 0},
		{name: "single_element", input: []string{"ABC123"}, want: 1},
		{name: "multiple_elements", input: []string{"SN001", "SN002", "SN003"}, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set, diags := stringsToSet(tt.input)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags.Errors())
			}
			if set.IsNull() || set.IsUnknown() {
				t.Fatal("expected non-null, non-unknown set")
			}
			elements := set.Elements()
			if len(elements) != tt.want {
				t.Fatalf("expected %d elements, got %d", tt.want, len(elements))
			}

			roundTrip := extractStrings(set)
			if len(roundTrip) != tt.want {
				t.Fatalf("round-trip: expected %d elements, got %d", tt.want, len(roundTrip))
			}
			inputMap := make(map[string]bool)
			for _, s := range tt.input {
				inputMap[s] = true
			}
			for _, s := range roundTrip {
				if !inputMap[s] {
					t.Errorf("round-trip: unexpected element %q", s)
				}
			}
		})
	}
}

func TestDownloadAndParseActivityLog(t *testing.T) {
	t.Run("empty_url", func(t *testing.T) {
		_, err := downloadAndParseActivityLog(context.Background(), "")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "no download URL provided") {
			t.Errorf("expected 'no download URL provided', got %q", err.Error())
		}
	})

	t.Run("all_success_csv", func(t *testing.T) {
		csvData := "serial_number,operation_status,operation_substatus\nSN001,SUCCESS,\nSN002,SUCCESS,\nSN003,SUCCESS,\n"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(csvData))
		}))
		defer server.Close()

		summary, err := downloadAndParseActivityLog(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(summary, "detailed results are available") {
			t.Errorf("expected success summary, got %q", summary)
		}
	})

	t.Run("csv_with_errors", func(t *testing.T) {
		csvData := "serial_number,operation_status,operation_substatus\nSN001,SUCCESS,\nSN002,FAILED,DEVICE_NOT_FOUND\nSN003,FAILED,ALREADY_ASSIGNED\n"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(csvData))
		}))
		defer server.Close()

		summary, err := downloadAndParseActivityLog(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(summary, "2 error(s)") {
			t.Errorf("expected '2 error(s)' in summary, got %q", summary)
		}
		if !strings.Contains(summary, "SN002") {
			t.Errorf("expected SN002 in summary, got %q", summary)
		}
		if !strings.Contains(summary, "SN003") {
			t.Errorf("expected SN003 in summary, got %q", summary)
		}
	})

	t.Run("csv_with_many_errors", func(t *testing.T) {
		var b strings.Builder
		b.WriteString("serial_number,operation_status,operation_substatus\n")
		for i := 0; i < 15; i++ {
			b.WriteString("SN" + strings.Repeat("0", 3) + ",FAILED,ERROR\n")
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(b.String()))
		}))
		defer server.Close()

		summary, err := downloadAndParseActivityLog(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(summary, "15 error(s)") {
			t.Errorf("expected '15 error(s)' in summary, got %q", summary)
		}
		if !strings.Contains(summary, "5 more error(s)") {
			t.Errorf("expected truncation message, got %q", summary)
		}
	})

	t.Run("http_error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		_, err := downloadAndParseActivityLog(context.Background(), server.URL)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "HTTP 500") {
			t.Errorf("expected 'HTTP 500' in error, got %q", err.Error())
		}
	})
}

func TestFilterDeviceManagementServiceList(t *testing.T) {
	servers := []client.MdmServer{
		{ID: "srv-1", Attributes: client.MdmServerAttribute{ServerName: "Jamf Pro", ServerType: "MDM"}},
		{ID: "srv-2", Attributes: client.MdmServerAttribute{ServerName: "Mosyle", ServerType: "MDM"}},
		{ID: "srv-3", Attributes: client.MdmServerAttribute{ServerName: "Apple Configurator", ServerType: "APPLE_CONFIGURATOR"}},
		{ID: "srv-4", Attributes: client.MdmServerAttribute{ServerName: "  Jamf Pro Cloud  ", ServerType: "MDM"}},
	}

	tests := []struct {
		name     string
		config   DeviceManagementServiceListResourceModel
		wantIDs  []string
	}{
		{
			name:    "no_filters",
			config:  DeviceManagementServiceListResourceModel{},
			wantIDs: []string{"srv-1", "srv-2", "srv-3", "srv-4"},
		},
		{
			name:    "exact_name_match",
			config:  DeviceManagementServiceListResourceModel{Name: types.StringValue("Jamf Pro")},
			wantIDs: []string{"srv-1"},
		},
		{
			name:    "name_contains_match",
			config:  DeviceManagementServiceListResourceModel{NameContains: types.StringValue("jamf")},
			wantIDs: []string{"srv-1", "srv-4"},
		},
		{
			name:    "server_type_filter",
			config:  DeviceManagementServiceListResourceModel{ServerType: types.StringValue("APPLE_CONFIGURATOR")},
			wantIDs: []string{"srv-3"},
		},
		{
			name:    "combined_name_and_type",
			config:  DeviceManagementServiceListResourceModel{NameContains: types.StringValue("jamf"), ServerType: types.StringValue("MDM")},
			wantIDs: []string{"srv-1", "srv-4"},
		},
		{
			name:    "no_match",
			config:  DeviceManagementServiceListResourceModel{Name: types.StringValue("nonexistent")},
			wantIDs: []string{},
		},
		{
			name:    "case_insensitive",
			config:  DeviceManagementServiceListResourceModel{Name: types.StringValue("JAMF PRO")},
			wantIDs: []string{"srv-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterDeviceManagementServiceList(servers, tt.config)
			if len(filtered) != len(tt.wantIDs) {
				t.Fatalf("expected %d results, got %d", len(tt.wantIDs), len(filtered))
			}
			for i, wantID := range tt.wantIDs {
				if filtered[i].ID != wantID {
					t.Errorf("result[%d]: expected ID %s, got %s", i, wantID, filtered[i].ID)
				}
			}
		})
	}
}

func TestNormalizedFilterString(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		wantStr   string
		wantValid bool
	}{
		{name: "null_value", value: types.StringNull(), wantStr: "", wantValid: false},
		{name: "unknown_value", value: types.StringUnknown(), wantStr: "", wantValid: false},
		{name: "empty_string", value: types.StringValue(""), wantStr: "", wantValid: false},
		{name: "whitespace_only", value: types.StringValue("   "), wantStr: "", wantValid: false},
		{name: "normal_value", value: types.StringValue("test"), wantStr: "test", wantValid: true},
		{name: "value_with_whitespace", value: types.StringValue("  test  "), wantStr: "test", wantValid: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str, valid := normalizedFilterString(tt.value)
			if valid != tt.wantValid {
				t.Errorf("expected valid=%v, got %v", tt.wantValid, valid)
			}
			if str != tt.wantStr {
				t.Errorf("expected string %q, got %q", tt.wantStr, str)
			}
		})
	}
}
