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

func TestGetMdmDevices_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/mdmDevices" {
			t.Errorf("expected path /v1/mdmDevices, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := MdmDeviceResponse{
			Data: []MdmDevice{
				{Type: "mdmDevices", ID: "dev-1", Attributes: MdmDeviceAttribute{SerialNumber: "SN001", DeviceName: "Neil's MacBook Pro", ProductFamily: "Mac"}},
				{Type: "mdmDevices", ID: "dev-2", Attributes: MdmDeviceAttribute{SerialNumber: "SN002", DeviceName: "Classroom iPad", ProductFamily: "iPad"}},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	devices, err := c.GetMdmDevices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(devices))
	}
	if devices[0].Attributes.SerialNumber != "SN001" {
		t.Errorf("expected serial SN001, got %s", devices[0].Attributes.SerialNumber)
	}
	if devices[0].Attributes.DeviceName != "Neil's MacBook Pro" {
		t.Errorf("expected deviceName 'Neil\\'s MacBook Pro', got %s", devices[0].Attributes.DeviceName)
	}
	if devices[1].ID != "dev-2" {
		t.Errorf("expected ID dev-2, got %s", devices[1].ID)
	}
}

func TestGetMdmDevices_MultiPage(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requestCount.Add(1)
		q := r.URL.Query()

		if q.Get("limit") != "100" {
			t.Errorf("expected limit=100, got %s", q.Get("limit"))
		}

		w.Header().Set("Content-Type", "application/json")
		if count == 1 {
			resp := MdmDeviceResponse{
				Data: []MdmDevice{
					{Type: "mdmDevices", ID: "dev-1", Attributes: MdmDeviceAttribute{SerialNumber: "SN001", DeviceName: "Work Mac", ProductFamily: "Mac"}},
					{Type: "mdmDevices", ID: "dev-2", Attributes: MdmDeviceAttribute{SerialNumber: "SN002", DeviceName: "Lab iPad", ProductFamily: "iPad"}},
				},
				Meta: Meta{Paging: Paging{Limit: 100, NextCursor: "next-page"}},
			}
			_, _ = w.Write(mustMarshalJSON(t, resp))
			return
		}
		if q.Get("cursor") != "next-page" {
			t.Errorf("expected cursor=next-page, got %s", q.Get("cursor"))
		}
		resp := MdmDeviceResponse{
			Data: []MdmDevice{
				{Type: "mdmDevices", ID: "dev-3", Attributes: MdmDeviceAttribute{SerialNumber: "SN003", DeviceName: "Dev iPhone", ProductFamily: "iPhone"}},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	devices, err := c.GetMdmDevices(context.Background(), nil)
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

func TestGetMdmDevices_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := MdmDeviceResponse{
			Data: []MdmDevice{},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	devices, err := c.GetMdmDevices(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 0 {
		t.Fatalf("expected 0 devices, got %d", len(devices))
	}
}

func TestGetMdmDevices_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors":[{"id":"e1","status":"500","code":"INTERNAL","title":"Server Error","detail":"Internal failure"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetMdmDevices(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Server Error") {
		t.Errorf("expected 'Server Error' in error, got %q", err.Error())
	}
}

func TestGetMdmDeviceDetail_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/mdmDevices/dev-1/details" {
			t.Errorf("expected path /v1/mdmDevices/dev-1/details, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		detailAttrs := MdmDeviceDetailAttribute{
			SerialNumber:        "SN001",
			DeviceModel:         "MacBook Pro",
			DeviceName:          "Neil's MacBook Pro",
			OsVersion:           "15.4",
			Platform:            "mac",
			WifiMacAddress:      "00:11:22:33:44:55",
			BluetoothMacAddress: "66:77:88:99:AA:BB",
			IsFirewallEnabled:   boolPtr(true),
		}
		resp := MdmDeviceDetailResponse{
			Data: MdmDeviceDetail{
				Type:       "mdmDeviceDetails",
				ID:         "dev-1",
				Attributes: detailAttrs,
			},
			Links: DocumentLinks{Self: "http://example.com/v1/mdmDevices/dev-1/details"},
		}
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	detail, err := c.GetMdmDeviceDetail(context.Background(), "dev-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.ID != "dev-1" {
		t.Errorf("expected ID dev-1, got %s", detail.ID)
	}
	if detail.Attributes.SerialNumber != "SN001" {
		t.Errorf("expected serial SN001, got %s", detail.Attributes.SerialNumber)
	}
	if detail.Attributes.DeviceModel != "MacBook Pro" {
		t.Errorf("expected deviceModel 'MacBook Pro', got %s", detail.Attributes.DeviceModel)
	}
	if detail.Attributes.WifiMacAddress != "00:11:22:33:44:55" {
		t.Errorf("expected wifiMacAddress 00:11:22:33:44:55, got %s", detail.Attributes.WifiMacAddress)
	}
	if *detail.Attributes.IsFirewallEnabled != true {
		t.Errorf("expected isFirewallEnabled true, got %v", *detail.Attributes.IsFirewallEnabled)
	}
}

func TestGetMdmDeviceDetail_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors":[{"id":"e1","status":"404","code":"NOT_FOUND","title":"Not Found","detail":"Resource not found"}]}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetMdmDeviceDetail(context.Background(), "missing", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Not Found") {
		t.Errorf("expected 'Not Found' in error, got %q", err.Error())
	}
}
