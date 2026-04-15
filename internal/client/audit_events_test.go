// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetAuditEvents_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/auditEvents" {
			t.Fatalf("expected path /v1/auditEvents, got %s", r.URL.Path)
		}
		if r.URL.Query().Get("filter[startTimestamp]") == "" {
			t.Fatal("expected filter[startTimestamp] to be set")
		}
		if r.URL.Query().Get("filter[endTimestamp]") == "" {
			t.Fatal("expected filter[endTimestamp] to be set")
		}

		resp := AuditEventsResponse{
			Data: []AuditEvent{
				{
					Type: "auditEvents",
					ID:   "event-123",
					Attributes: AuditEventAttributes{
						EventDateTime: "2026-02-14T12:00:00Z",
						Type:          "DEVICE_ADDED_TO_ORG",
						Category:      "DEVICE_INVENTORY",
					},
				},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	queryParams := url.Values{}
	queryParams.Set("filter[startTimestamp]", "2026-03-01T00:00:00Z")
	queryParams.Set("filter[endTimestamp]", "2026-03-02T23:59:59Z")
	queryParams.Set("limit", "50")

	events, err := c.GetAuditEvents(context.Background(), queryParams)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Attributes.Type != "DEVICE_ADDED_TO_ORG" {
		t.Errorf("unexpected event type: %s", events[0].Attributes.Type)
	}
}

func TestAuditEventAttributes_UnmarshalAdditional(t *testing.T) {
	payload := `{
		"eventDateTime": "2026-02-14T12:00:00Z",
		"type": "DEVICE_ADDED_TO_ORG",
		"eventDataPropertyKey": "eventDataDeviceAddedToOrg",
		"eventDataDeviceAddedToOrg": {
			"serialNumber": "C02X1234ABCD"
		}
	}`

	var attrs AuditEventAttributes
	if err := attrs.UnmarshalJSON([]byte(payload)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if attrs.EventDataPropertyKey != "eventDataDeviceAddedToOrg" {
		t.Fatalf("unexpected property key: %s", attrs.EventDataPropertyKey)
	}
	if attrs.Additional["eventDataDeviceAddedToOrg"] == nil {
		t.Fatal("expected eventDataDeviceAddedToOrg to be captured")
	}
}
