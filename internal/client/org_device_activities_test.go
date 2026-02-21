// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAssignDevicesToMDMServer_Assign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/orgDeviceActivities") {
			t.Errorf("expected path /v1/orgDeviceActivities, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityType != "ASSIGN_DEVICES" {
			t.Errorf("expected ASSIGN_DEVICES, got %s", req.Data.Attributes.ActivityType)
		}
		if req.Data.Relationships.MdmServer.Data.ID != "srv-1" {
			t.Errorf("expected server ID srv-1, got %s", req.Data.Relationships.MdmServer.Data.ID)
		}
		if req.Data.Relationships.MdmServer.Data.Type != "mdmServers" {
			t.Errorf("expected type mdmServers, got %s", req.Data.Relationships.MdmServer.Data.Type)
		}
		if len(req.Data.Relationships.Devices.Data) != 3 {
			t.Fatalf("expected 3 devices, got %d", len(req.Data.Relationships.Devices.Data))
		}
		expectedIDs := []string{"DEV001", "DEV002", "DEV003"}
		for i, expected := range expectedIDs {
			if req.Data.Relationships.Devices.Data[i].ID != expected {
				t.Errorf("device[%d]: expected %s, got %s", i, expected, req.Data.Relationships.Devices.Data[i].ID)
			}
			if req.Data.Relationships.Devices.Data[i].Type != "orgDevices" {
				t.Errorf("device[%d]: expected type orgDevices, got %s", i, req.Data.Relationships.Devices.Data[i].Type)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type: "orgDeviceActivities",
				ID:   "activity-1",
				Attributes: OrgDeviceActivityAttributes{
					Status:          "IN_PROGRESS",
					SubStatus:       "",
					CreatedDateTime: "2024-02-21T10:00:00Z",
				},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.AssignDevicesToMDMServer(context.Background(), "srv-1", []string{"DEV001", "DEV002", "DEV003"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-1" {
		t.Errorf("expected activity ID activity-1, got %s", activity.ID)
	}
	if activity.Attributes.Status != "IN_PROGRESS" {
		t.Errorf("expected status IN_PROGRESS, got %s", activity.Attributes.Status)
	}
}

func TestAssignDevicesToMDMServer_Unassign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityType != "UNASSIGN_DEVICES" {
			t.Errorf("expected UNASSIGN_DEVICES, got %s", req.Data.Attributes.ActivityType)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-2",
				Attributes: OrgDeviceActivityAttributes{Status: "IN_PROGRESS"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.AssignDevicesToMDMServer(context.Background(), "srv-1", []string{"DEV001", "DEV002"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-2" {
		t.Errorf("expected activity ID activity-2, got %s", activity.ID)
	}
}

func TestAssignDevicesToMDMServer_SingleDevice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if len(req.Data.Relationships.Devices.Data) != 1 {
			t.Errorf("expected 1 device, got %d", len(req.Data.Relationships.Devices.Data))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-3",
				Attributes: OrgDeviceActivityAttributes{Status: "IN_PROGRESS"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.AssignDevicesToMDMServer(context.Background(), "srv-1", []string{"DEV001"}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-3" {
		t.Errorf("expected activity ID activity-3, got %s", activity.ID)
	}
}

func TestAssignDevicesToMDMServer_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"id":"e1","status":"400","code":"BAD_REQUEST","title":"Bad Request","detail":"Invalid device IDs"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.AssignDevicesToMDMServer(context.Background(), "srv-1", []string{"INVALID"}, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Bad Request") {
		t.Errorf("expected 'Bad Request' in error, got %q", err.Error())
	}
}

func TestGetOrgDeviceActivity_Completed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v1/orgDeviceActivities/activity-1") {
			t.Errorf("expected path containing /v1/orgDeviceActivities/activity-1, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type: "orgDeviceActivities",
				ID:   "activity-1",
				Attributes: OrgDeviceActivityAttributes{
					Status:            "COMPLETED",
					SubStatus:         "COMPLETED_WITH_SUCCESS",
					CreatedDateTime:   "2024-02-21T10:00:00Z",
					CompletedDateTime: "2024-02-21T10:02:00Z",
					DownloadURL:       "https://example.com/results.csv",
				},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.GetOrgDeviceActivity(context.Background(), "activity-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.Attributes.Status != "COMPLETED" {
		t.Errorf("expected COMPLETED, got %s", activity.Attributes.Status)
	}
	if activity.Attributes.SubStatus != "COMPLETED_WITH_SUCCESS" {
		t.Errorf("expected COMPLETED_WITH_SUCCESS, got %s", activity.Attributes.SubStatus)
	}
	if activity.Attributes.DownloadURL != "https://example.com/results.csv" {
		t.Errorf("expected download URL, got %s", activity.Attributes.DownloadURL)
	}
}

func TestGetOrgDeviceActivity_InProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type: "orgDeviceActivities",
				ID:   "activity-1",
				Attributes: OrgDeviceActivityAttributes{
					Status:          "IN_PROGRESS",
					CreatedDateTime: "2024-02-21T10:00:00Z",
				},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.GetOrgDeviceActivity(context.Background(), "activity-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.Attributes.Status != "IN_PROGRESS" {
		t.Errorf("expected IN_PROGRESS, got %s", activity.Attributes.Status)
	}
}

func TestGetOrgDeviceActivity_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "include=details") {
			t.Errorf("expected query param include=details, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-1",
				Attributes: OrgDeviceActivityAttributes{Status: "COMPLETED"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	params := url.Values{"include": {"details"}}
	_, err := c.GetOrgDeviceActivity(context.Background(), "activity-1", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
