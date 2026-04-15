// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetApps_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/apps" {
			t.Fatalf("expected path /v1/apps, got %s", r.URL.Path)
		}
		resp := AppsResponse{
			Data: []App{
				{
					Type: "apps",
					ID:   "361309726",
					Attributes: AppAttributes{
						Name:     "Pages",
						BundleID: "com.apple.Pages",
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
	apps, err := c.GetApps(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("expected 1 app, got %d", len(apps))
	}
	if apps[0].Attributes.BundleID != "com.apple.Pages" {
		t.Errorf("unexpected bundleId: %s", apps[0].Attributes.BundleID)
	}
}

func TestGetApp_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/apps/361309726" {
			t.Fatalf("expected path /v1/apps/361309726, got %s", r.URL.Path)
		}
		resp := AppResponse{
			Data: App{
				Type: "apps",
				ID:   "361309726",
				Attributes: AppAttributes{
					Name: "Pages",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	app, err := c.GetApp(context.Background(), "361309726", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if app.Attributes.Name != "Pages" {
		t.Errorf("expected name Pages, got %s", app.Attributes.Name)
	}
}
