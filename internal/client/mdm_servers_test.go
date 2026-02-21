// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGetDeviceManagementServices_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := MdmServersResponse{
			Data: []MdmServer{
				{Type: "mdmServers", ID: "srv-1", Attributes: MdmServerAttribute{ServerName: "Jamf Pro", ServerType: "MDM", CreatedDateTime: "2023-06-15T08:00:00Z", UpdatedDateTime: "2024-02-20T12:45:30Z"}},
				{Type: "mdmServers", ID: "srv-2", Attributes: MdmServerAttribute{ServerName: "Apple Configurator", ServerType: "APPLE_CONFIGURATOR", CreatedDateTime: "2023-07-01T09:00:00Z", UpdatedDateTime: "2024-01-15T10:00:00Z"}},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	servers, err := c.GetDeviceManagementServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}
	if servers[0].Attributes.ServerName != "Jamf Pro" {
		t.Errorf("expected 'Jamf Pro', got %s", servers[0].Attributes.ServerName)
	}
	if servers[0].Attributes.ServerType != "MDM" {
		t.Errorf("expected type MDM, got %s", servers[0].Attributes.ServerType)
	}
	if servers[1].ID != "srv-2" {
		t.Errorf("expected ID srv-2, got %s", servers[1].ID)
	}
}

func TestGetDeviceManagementServices_MultiPage(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		q := r.URL.Query()

		if q.Get("limit") != "100" {
			t.Errorf("expected limit=100, got %s", q.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			resp := MdmServersResponse{
				Data: []MdmServer{
					{Type: "mdmServers", ID: "srv-1", Attributes: MdmServerAttribute{ServerName: "Server 1", ServerType: "MDM"}},
					{Type: "mdmServers", ID: "srv-2", Attributes: MdmServerAttribute{ServerName: "Server 2", ServerType: "MDM"}},
				},
				Meta: Meta{Paging: Paging{Limit: 100, NextCursor: "next-page"}},
			}
			_, _ = w.Write(mustMarshalJSON(t, resp))
			return
		}
		if q.Get("cursor") != "next-page" {
			t.Errorf("expected cursor=next-page, got %s", q.Get("cursor"))
		}
		resp := MdmServersResponse{
			Data: []MdmServer{
				{Type: "mdmServers", ID: "srv-3", Attributes: MdmServerAttribute{ServerName: "Server 3", ServerType: "APPLE_CONFIGURATOR"}},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	servers, err := c.GetDeviceManagementServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(servers))
	}
	if got := requestCount.Load(); got != 2 {
		t.Fatalf("expected 2 requests, got %d", got)
	}
}

func TestGetDeviceManagementServices_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := MdmServersResponse{
			Data: []MdmServer{},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	servers, err := c.GetDeviceManagementServices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(servers) != 0 {
		t.Fatalf("expected 0 servers, got %d", len(servers))
	}
}

func TestGetDeviceManagementServices_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"id":"e1","status":"500","code":"INTERNAL","title":"Server Error","detail":"Internal failure"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetDeviceManagementServices(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Server Error") {
		t.Errorf("expected 'Server Error' in error, got %q", err.Error())
	}
}

func TestGetDeviceManagementServiceSerialNumbers_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v1/mdmServers/srv-1/relationships/devices") {
			t.Errorf("expected path /v1/mdmServers/srv-1/relationships/devices, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := MdmServerDevicesLinkagesResponse{
			Data: []Data{
				{ID: "SN001", Type: "orgDevices"},
				{ID: "SN002", Type: "orgDevices"},
				{ID: "SN003", Type: "orgDevices"},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	serials, err := c.GetDeviceManagementServiceSerialNumbers(context.Background(), "srv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(serials) != 3 {
		t.Fatalf("expected 3 serials, got %d", len(serials))
	}
	expected := []string{"SN001", "SN002", "SN003"}
	for i, want := range expected {
		if serials[i] != want {
			t.Errorf("serial[%d]: expected %s, got %s", i, want, serials[i])
		}
	}
}

func TestGetDeviceManagementServiceSerialNumbers_MultiPage(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			resp := MdmServerDevicesLinkagesResponse{
				Data: []Data{
					{ID: "SN001", Type: "orgDevices"},
					{ID: "SN002", Type: "orgDevices"},
				},
				Meta: Meta{Paging: Paging{Limit: 100, NextCursor: "cursor2"}},
			}
			_, _ = w.Write(mustMarshalJSON(t, resp))
			return
		}
		resp := MdmServerDevicesLinkagesResponse{
			Data: []Data{
				{ID: "SN003", Type: "orgDevices"},
				{ID: "SN004", Type: "orgDevices"},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	serials, err := c.GetDeviceManagementServiceSerialNumbers(context.Background(), "srv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(serials) != 4 {
		t.Fatalf("expected 4 serials, got %d", len(serials))
	}
}

func TestGetDeviceManagementServiceSerialNumbers_FiltersNonOrgDevices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := MdmServerDevicesLinkagesResponse{
			Data: []Data{
				{ID: "SN001", Type: "orgDevices"},
				{ID: "OTHER-1", Type: "someOtherType"},
				{ID: "SN002", Type: "orgDevices"},
				{ID: "OTHER-2", Type: "differentType"},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	serials, err := c.GetDeviceManagementServiceSerialNumbers(context.Background(), "srv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(serials) != 2 {
		t.Fatalf("expected 2 serials (filtered), got %d", len(serials))
	}
	if serials[0] != "SN001" || serials[1] != "SN002" {
		t.Errorf("expected [SN001, SN002], got %v", serials)
	}
}

func TestGetDeviceManagementServiceSerialNumbers_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := MdmServerDevicesLinkagesResponse{
			Data: []Data{},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	serials, err := c.GetDeviceManagementServiceSerialNumbers(context.Background(), "srv-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(serials) != 0 {
		t.Fatalf("expected 0 serials, got %d", len(serials))
	}
}
