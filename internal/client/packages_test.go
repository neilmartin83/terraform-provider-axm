// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPackages_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/packages" {
			t.Fatalf("expected path /v1/packages, got %s", r.URL.Path)
		}
		resp := PackagesResponse{
			Data: []Package{
				{
					Type: "packages",
					ID:   "pkg-12345",
					Attributes: PackageAttributes{
						Name: "Enterprise Suite",
						Hash: "hash123",
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
	packages, err := c.GetPackages(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
	if packages[0].Attributes.Hash != "hash123" {
		t.Errorf("unexpected hash: %s", packages[0].Attributes.Hash)
	}
}

func TestGetPackage_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/packages/pkg-12345" {
			t.Fatalf("expected path /v1/packages/pkg-12345, got %s", r.URL.Path)
		}
		resp := PackageResponse{
			Data: Package{
				Type: "packages",
				ID:   "pkg-12345",
				Attributes: PackageAttributes{
					Name: "Enterprise Suite",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	pkg, err := c.GetPackage(context.Background(), "pkg-12345", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Attributes.Name != "Enterprise Suite" {
		t.Errorf("expected name Enterprise Suite, got %s", pkg.Attributes.Name)
	}
}
