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

func TestCreateOrgDeviceActivity_Assign(t *testing.T) {
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
		if req.Data.Attributes.ActivityTypeMetadata != nil {
			t.Error("expected no activityTypeMetadata for ASSIGN_DEVICES")
		}
		if req.Data.Relationships.MdmServer == nil {
			t.Fatal("expected mdmServer relationship, got nil")
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
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityAssignDevices, []string{"DEV001", "DEV002", "DEV003"}, WithMdmServer("srv-1"))
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

func TestCreateOrgDeviceActivity_Unassign(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityType != "UNASSIGN_DEVICES" {
			t.Errorf("expected UNASSIGN_DEVICES, got %s", req.Data.Attributes.ActivityType)
		}
		if req.Data.Relationships.MdmServer == nil {
			t.Fatal("expected mdmServer relationship, got nil")
		}
		if req.Data.Relationships.MdmServer.Data.ID != "srv-1" {
			t.Errorf("expected server ID srv-1, got %s", req.Data.Relationships.MdmServer.Data.ID)
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
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityUnassignDevices, []string{"DEV001", "DEV002"}, WithMdmServer("srv-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-2" {
		t.Errorf("expected activity ID activity-2, got %s", activity.ID)
	}
}

func TestCreateOrgDeviceActivity_SingleDevice(t *testing.T) {
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
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityAssignDevices, []string{"DEV001"}, WithMdmServer("srv-1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-3" {
		t.Errorf("expected activity ID activity-3, got %s", activity.ID)
	}
}

func TestCreateOrgDeviceActivity_NoOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityTypeMetadata != nil {
			t.Error("expected no activityTypeMetadata without WithMigrationDeadline")
		}
		if req.Data.Relationships.MdmServer != nil {
			t.Error("expected no mdmServer relationship without WithMdmServer")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-8",
				Attributes: OrgDeviceActivityAttributes{Status: "IN_PROGRESS"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityReleaseDevices, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-8" {
		t.Errorf("expected activity ID activity-8, got %s", activity.ID)
	}
}

func TestCreateOrgDeviceActivity_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"errors":[{"id":"e1","status":"400","code":"BAD_REQUEST","title":"Bad Request","detail":"Invalid device IDs"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityAssignDevices, []string{"INVALID"}, WithMdmServer("srv-1"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Bad Request") {
		t.Errorf("expected 'Bad Request' in error, got %q", err.Error())
	}
}

func TestCreateOrgDeviceActivity_AssignWithMigrationDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityType != "ASSIGN_DEVICES_WITH_MDM_MIGRATION_DEADLINE" {
			t.Errorf("expected ASSIGN_DEVICES_WITH_MDM_MIGRATION_DEADLINE, got %s", req.Data.Attributes.ActivityType)
		}
		if req.Data.Attributes.ActivityTypeMetadata == nil {
			t.Fatal("expected activityTypeMetadata, got nil")
		}
		if req.Data.Attributes.ActivityTypeMetadata.MdmMigrationDeadlineDateTime != "2026-09-15T17:00:00.000Z" {
			t.Errorf("expected deadline 2026-09-15T17:00:00.000Z, got %s", req.Data.Attributes.ActivityTypeMetadata.MdmMigrationDeadlineDateTime)
		}
		if req.Data.Relationships.MdmServer == nil {
			t.Fatal("expected mdmServer relationship, got nil")
		}
		if req.Data.Relationships.MdmServer.Data.ID != "srv-9" {
			t.Errorf("expected server ID srv-9, got %s", req.Data.Relationships.MdmServer.Data.ID)
		}
		if len(req.Data.Relationships.Devices.Data) != 2 {
			t.Fatalf("expected 2 devices, got %d", len(req.Data.Relationships.Devices.Data))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-4",
				Attributes: OrgDeviceActivityAttributes{Status: "IN_PROGRESS"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityAssignWithMDMMigrationDeadline, []string{"DEV001", "DEV002"},
		WithMdmServer("srv-9"), WithMigrationDeadline("2026-09-15T17:00:00.000Z"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-4" {
		t.Errorf("expected activity ID activity-4, got %s", activity.ID)
	}
}

func TestCreateOrgDeviceActivity_UpdateMigrationDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityType != "UPDATE_MDM_MIGRATION_DEADLINE" {
			t.Errorf("expected UPDATE_MDM_MIGRATION_DEADLINE, got %s", req.Data.Attributes.ActivityType)
		}
		if req.Data.Attributes.ActivityTypeMetadata == nil {
			t.Fatal("expected activityTypeMetadata, got nil")
		}
		if req.Data.Attributes.ActivityTypeMetadata.MdmMigrationDeadlineDateTime != "2026-09-16T09:00:00.000Z" {
			t.Errorf("expected deadline 2026-09-16T09:00:00.000Z, got %s", req.Data.Attributes.ActivityTypeMetadata.MdmMigrationDeadlineDateTime)
		}
		if req.Data.Relationships.MdmServer != nil {
			t.Error("expected no mdmServer relationship for UPDATE_MDM_MIGRATION_DEADLINE")
		}
		if len(req.Data.Relationships.Devices.Data) != 2 {
			t.Fatalf("expected 2 devices, got %d", len(req.Data.Relationships.Devices.Data))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-5",
				Attributes: OrgDeviceActivityAttributes{Status: "IN_PROGRESS"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityUpdateMDMMigrationDeadline, []string{"DEV001", "DEV002"}, WithMigrationDeadline("2026-09-16T09:00:00.000Z"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-5" {
		t.Errorf("expected activity ID activity-5, got %s", activity.ID)
	}
}

func TestCreateOrgDeviceActivity_CancelMigration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityType != "CANCEL_MDM_MIGRATION" {
			t.Errorf("expected CANCEL_MDM_MIGRATION, got %s", req.Data.Attributes.ActivityType)
		}
		if req.Data.Attributes.ActivityTypeMetadata != nil {
			t.Error("expected no activityTypeMetadata for CANCEL_MDM_MIGRATION")
		}
		if req.Data.Relationships.MdmServer != nil {
			t.Error("expected no mdmServer relationship for CANCEL_MDM_MIGRATION")
		}
		if len(req.Data.Relationships.Devices.Data) != 1 {
			t.Fatalf("expected 1 device, got %d", len(req.Data.Relationships.Devices.Data))
		}
		if req.Data.Relationships.Devices.Data[0].ID != "DEV001" {
			t.Errorf("expected device DEV001, got %s", req.Data.Relationships.Devices.Data[0].ID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-6",
				Attributes: OrgDeviceActivityAttributes{Status: "IN_PROGRESS"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityCancelMDMMigration, []string{"DEV001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-6" {
		t.Errorf("expected activity ID activity-6, got %s", activity.ID)
	}
}

func TestCreateOrgDeviceActivity_ReleaseDevices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req OrgDeviceActivityCreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse request body: %v", err)
		}

		if req.Data.Attributes.ActivityType != "RELEASE_DEVICES" {
			t.Errorf("expected RELEASE_DEVICES, got %s", req.Data.Attributes.ActivityType)
		}
		if req.Data.Attributes.ActivityTypeMetadata != nil {
			t.Error("expected no activityTypeMetadata for RELEASE_DEVICES")
		}
		if req.Data.Relationships.MdmServer != nil {
			t.Error("expected no mdmServer relationship for RELEASE_DEVICES")
		}
		if len(req.Data.Relationships.Devices.Data) != 2 {
			t.Fatalf("expected 2 devices, got %d", len(req.Data.Relationships.Devices.Data))
		}
		for _, d := range req.Data.Relationships.Devices.Data {
			if d.Type != "orgDevices" {
				t.Errorf("expected type orgDevices, got %s", d.Type)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		resp := OrgDeviceActivityResponse{
			Data: OrgDeviceActivity{
				Type:       "orgDeviceActivities",
				ID:         "activity-7",
				Attributes: OrgDeviceActivityAttributes{Status: "IN_PROGRESS"},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	activity, err := c.CreateOrgDeviceActivity(context.Background(), OrgDeviceActivityReleaseDevices, []string{"DEV001", "DEV002"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if activity.ID != "activity-7" {
		t.Errorf("expected activity ID activity-7, got %s", activity.ID)
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
