// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGetOrgDevices_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/orgDevices") {
			t.Errorf("expected path /v1/orgDevices, got %s", r.URL.Path)
		}
		resp := OrgDevicesResponse{
			Data: []OrgDevice{
				{Type: "orgDevices", ID: "DEV001", Attributes: DeviceAttribute{SerialNumber: "SN001", DeviceModel: "iPad Pro", Status: "active", Color: "Silver", ProductFamily: "iPad", ProductType: "iPad13,4", DeviceCapacity: "256GB", PurchaseSourceID: "src-1", PurchaseSourceType: "DIRECT", AddedToOrgDateTime: "2024-01-15T10:00:00Z", UpdatedDateTime: "2024-02-01T12:00:00Z"}},
				{Type: "orgDevices", ID: "DEV002", Attributes: DeviceAttribute{SerialNumber: "SN002", DeviceModel: "MacBook Pro", Status: "active", Color: "Space Gray", ProductFamily: "Mac", IMEI: []string{"111111111111111", "222222222222222"}, MEID: []string{"AABBCCDD"}, EthernetMacAddress: []string{"00:11:22:33:44:55", "66:77:88:99:AA:BB"}}},
				{Type: "orgDevices", ID: "DEV003", Attributes: DeviceAttribute{SerialNumber: "SN003", DeviceModel: "iPhone 15", Status: "active", Color: "Blue", ProductFamily: "iPhone"}},
			},
			Meta: Meta{Paging: Paging{Limit: 1000}},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	devices, err := c.GetOrgDevices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devices))
	}
	if devices[0].Attributes.SerialNumber != "SN001" {
		t.Errorf("expected serial SN001, got %s", devices[0].Attributes.SerialNumber)
	}
	if devices[1].Attributes.DeviceModel != "MacBook Pro" {
		t.Errorf("expected model MacBook Pro, got %s", devices[1].Attributes.DeviceModel)
	}
	if len(devices[1].Attributes.IMEI) != 2 {
		t.Errorf("expected 2 IMEIs, got %d", len(devices[1].Attributes.IMEI))
	}
	if len(devices[1].Attributes.EthernetMacAddress) != 2 {
		t.Errorf("expected 2 ethernet MACs, got %d", len(devices[1].Attributes.EthernetMacAddress))
	}
}

func TestGetOrgDevices_MultiPage(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		q := r.URL.Query()

		if q.Get("limit") != "1000" {
			t.Errorf("expected limit=1000, got %s", q.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			if q.Get("cursor") != "" {
				t.Errorf("expected no cursor on first request, got %s", q.Get("cursor"))
			}
			resp := OrgDevicesResponse{
				Data: []OrgDevice{
					{Type: "orgDevices", ID: "DEV001", Attributes: DeviceAttribute{SerialNumber: "SN001"}},
					{Type: "orgDevices", ID: "DEV002", Attributes: DeviceAttribute{SerialNumber: "SN002"}},
				},
				Meta: Meta{Paging: Paging{Limit: 1000, NextCursor: "page2cursor"}},
			}
			_, _ = w.Write(mustMarshalJSON(t, resp))
			return
		}
		if q.Get("cursor") != "page2cursor" {
			t.Errorf("expected cursor=page2cursor on second request, got %s", q.Get("cursor"))
		}
		resp := OrgDevicesResponse{
			Data: []OrgDevice{
				{Type: "orgDevices", ID: "DEV003", Attributes: DeviceAttribute{SerialNumber: "SN003"}},
			},
			Meta: Meta{Paging: Paging{Limit: 1000}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	devices, err := c.GetOrgDevices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 3 {
		t.Fatalf("expected 3 devices, got %d", len(devices))
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("expected 2 requests, got %d", got)
	}
}

func TestGetOrgDevices_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := OrgDevicesResponse{
			Data: []OrgDevice{},
			Meta: Meta{Paging: Paging{Limit: 1000}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	devices, err := c.GetOrgDevices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(devices))
	}
}

func TestGetOrgDevices_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"id":"e1","status":"500","code":"INTERNAL","title":"Internal Error","detail":"Something broke"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetOrgDevices(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Internal Error") {
		t.Errorf("expected 'Internal Error' in error, got %q", err.Error())
	}
}

func TestGetOrgDevices_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "filter%5Bstatus%5D=active") && !strings.Contains(r.URL.RawQuery, "filter[status]=active") {
			t.Errorf("expected query param filter[status]=active, got %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := OrgDevicesResponse{Data: []OrgDevice{}, Meta: Meta{Paging: Paging{Limit: 1000}}}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	params := url.Values{"filter[status]": {"active"}}
	_, err := c.GetOrgDevices(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetOrgDevice_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v1/orgDevices/DEV001") {
			t.Errorf("expected path containing /v1/orgDevices/DEV001, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := OrgDeviceResponse{
			Data: OrgDevice{
				Type: "orgDevices",
				ID:   "DEV001",
				Attributes: DeviceAttribute{
					SerialNumber:        "SN001",
					AddedToOrgDateTime:  "2024-01-15T10:00:00Z",
					UpdatedDateTime:     "2024-02-01T12:00:00Z",
					DeviceModel:         "iPad Pro (11-inch)",
					ProductFamily:       "iPad",
					ProductType:         "iPad13,4",
					DeviceCapacity:      "256GB",
					PartNumber:          "MK893LL/A",
					OrderNumber:         "PO-12345",
					Color:               "Space Gray",
					Status:              "active",
					OrderDateTime:       "2024-01-10T00:00:00Z",
					IMEI:                []string{"111111111111111", "222222222222222"},
					MEID:                []string{"AABBCC"},
					EID:                 "eid-12345",
					PurchaseSourceID:    "src-1",
					PurchaseSourceType:  "DIRECT",
					WifiMacAddress:      "00:1A:2B:3C:4D:5E",
					BluetoothMacAddress: "00:1A:2B:3C:4D:5F",
					EthernetMacAddress:  []string{"AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66"},
				},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	device, err := c.GetOrgDevice(context.Background(), "DEV001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if device.ID != "DEV001" {
		t.Errorf("expected ID DEV001, got %s", device.ID)
	}
	if device.Attributes.SerialNumber != "SN001" {
		t.Errorf("expected serial SN001, got %s", device.Attributes.SerialNumber)
	}
	if device.Attributes.DeviceModel != "iPad Pro (11-inch)" {
		t.Errorf("expected model 'iPad Pro (11-inch)', got %s", device.Attributes.DeviceModel)
	}
	if len(device.Attributes.IMEI) != 2 {
		t.Errorf("expected 2 IMEIs, got %d", len(device.Attributes.IMEI))
	}
	if len(device.Attributes.MEID) != 1 {
		t.Errorf("expected 1 MEID, got %d", len(device.Attributes.MEID))
	}
	if device.Attributes.EID != "eid-12345" {
		t.Errorf("expected EID eid-12345, got %s", device.Attributes.EID)
	}
	if len(device.Attributes.EthernetMacAddress) != 2 {
		t.Errorf("expected 2 ethernet MACs, got %d", len(device.Attributes.EthernetMacAddress))
	}
	if device.Attributes.WifiMacAddress != "00:1A:2B:3C:4D:5E" {
		t.Errorf("expected wifi MAC 00:1A:2B:3C:4D:5E, got %s", device.Attributes.WifiMacAddress)
	}
}

func TestGetOrgDevice_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"id":"e1","status":"404","code":"NOT_FOUND","title":"Not Found","detail":"Device not found"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetOrgDevice(context.Background(), "NONEXISTENT", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Not Found") {
		t.Errorf("expected 'Not Found' in error, got %q", err.Error())
	}
}

func TestGetOrgDeviceAssignedServerID_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/relationships/assignedServer") {
			t.Errorf("expected path containing /relationships/assignedServer, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := OrgDeviceAssignedServerLinkageResponse{
			Data: Data{ID: "server-1", Type: "mdmServers"},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	data, err := c.GetOrgDeviceAssignedServerID(context.Background(), "DEV001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.ID != "server-1" {
		t.Errorf("expected server ID server-1, got %s", data.ID)
	}
	if data.Type != "mdmServers" {
		t.Errorf("expected type mdmServers, got %s", data.Type)
	}
}

func TestGetOrgDeviceAssignedServer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v1/orgDevices/DEV001/assignedServer") {
			t.Errorf("expected path /v1/orgDevices/DEV001/assignedServer, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := MdmServerResponse{
			Data: MdmServer{
				Type: "mdmServers",
				ID:   "server-1",
				Attributes: MdmServerAttribute{
					ServerName:      "Jamf Pro",
					ServerType:      "MDM",
					CreatedDateTime: "2023-06-15T08:00:00Z",
					UpdatedDateTime: "2024-02-20T12:45:30Z",
				},
			},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	srv, err := c.GetOrgDeviceAssignedServer(context.Background(), "DEV001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if srv.ID != "server-1" {
		t.Errorf("expected ID server-1, got %s", srv.ID)
	}
	if srv.Attributes.ServerName != "Jamf Pro" {
		t.Errorf("expected name 'Jamf Pro', got %s", srv.Attributes.ServerName)
	}
	if srv.Attributes.ServerType != "MDM" {
		t.Errorf("expected type MDM, got %s", srv.Attributes.ServerType)
	}
	if srv.Attributes.CreatedDateTime != "2023-06-15T08:00:00Z" {
		t.Errorf("expected created 2023-06-15T08:00:00Z, got %s", srv.Attributes.CreatedDateTime)
	}
}

func TestGetOrgDeviceAppleCareCoverage_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := AppleCareCoverageResponse{
			Data: []AppleCareCoverage{
				{
					Type: "appleCareCoverage",
					ID:   "ACC-001",
					Attributes: AppleCareCoverageAttribute{
						Status:          "ACTIVE",
						PaymentType:     "SUBSCRIPTION",
						Description:     "AppleCare+ for iPad",
						StartDateTime:   "2024-01-15T00:00:00Z",
						EndDateTime:     "2026-01-14T23:59:59Z",
						IsRenewable:     true,
						IsCanceled:      false,
						AgreementNumber: "AGR-001",
					},
				},
				{
					Type: "appleCareCoverage",
					ID:   "ACC-002",
					Attributes: AppleCareCoverageAttribute{
						Status:          "EXPIRED",
						PaymentType:     "NONE",
						Description:     "AppleCare for iPad (Original)",
						StartDateTime:   "2022-01-15T00:00:00Z",
						EndDateTime:     "2023-01-14T23:59:59Z",
						IsRenewable:     false,
						IsCanceled:      true,
						AgreementNumber: "AGR-002",
					},
				},
			},
			Meta: Meta{Paging: Paging{Limit: 1000}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	coverages, err := c.GetOrgDeviceAppleCareCoverage(context.Background(), "DEV001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(coverages) != 2 {
		t.Fatalf("expected 2 coverages, got %d", len(coverages))
	}
	if coverages[0].Attributes.Status != "ACTIVE" {
		t.Errorf("expected ACTIVE status, got %s", coverages[0].Attributes.Status)
	}
	if coverages[0].Attributes.IsRenewable != true {
		t.Errorf("expected IsRenewable true")
	}
	if coverages[1].Attributes.IsCanceled != true {
		t.Errorf("expected IsCanceled true")
	}
	if coverages[0].Attributes.AgreementNumber != "AGR-001" {
		t.Errorf("expected agreement AGR-001, got %s", coverages[0].Attributes.AgreementNumber)
	}
}

func TestGetOrgDeviceAppleCareCoverage_MultiPage(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			resp := AppleCareCoverageResponse{
				Data: []AppleCareCoverage{
					{Type: "appleCareCoverage", ID: "ACC-001", Attributes: AppleCareCoverageAttribute{Status: "ACTIVE"}},
					{Type: "appleCareCoverage", ID: "ACC-002", Attributes: AppleCareCoverageAttribute{Status: "ACTIVE"}},
				},
				Meta: Meta{Paging: Paging{Limit: 1000, NextCursor: "page2"}},
			}
			_, _ = w.Write(mustMarshalJSON(t, resp))
			return
		}
		resp := AppleCareCoverageResponse{
			Data: []AppleCareCoverage{
				{Type: "appleCareCoverage", ID: "ACC-003", Attributes: AppleCareCoverageAttribute{Status: "EXPIRED"}},
			},
			Meta: Meta{Paging: Paging{Limit: 1000}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	coverages, err := c.GetOrgDeviceAppleCareCoverage(context.Background(), "DEV001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(coverages) != 3 {
		t.Fatalf("expected 3 coverages, got %d", len(coverages))
	}
}

func TestGetOrgDeviceAppleCareCoverage_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := AppleCareCoverageResponse{
			Data: []AppleCareCoverage{},
			Meta: Meta{Paging: Paging{Limit: 1000}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	coverages, err := c.GetOrgDeviceAppleCareCoverage(context.Background(), "DEV001", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(coverages) != 0 {
		t.Fatalf("expected 0 coverages, got %d", len(coverages))
	}
}
