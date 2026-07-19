// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGetOrganizationalUnits_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organizationalUnits" {
			t.Fatalf("expected path /v1/organizationalUnits, got %s", r.URL.Path)
		}
		resp := OrganizationalUnitsResponse{
			Data: []OrganizationalUnit{
				{
					Type: "organizationalUnits",
					ID:   "OU789012",
					Attributes: OrganizationalUnitAttributes{
						Name:        "Engineering",
						Description: "Engineering organizational unit",
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
	ous, err := c.GetOrganizationalUnits(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ous) != 1 {
		t.Fatalf("expected 1 OU, got %d", len(ous))
	}
	if ous[0].Attributes.Name != "Engineering" {
		t.Errorf("expected name Engineering, got %s", ous[0].Attributes.Name)
	}
}

func TestGetOrganizationalUnit_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/organizationalUnits/OU789012" {
			t.Fatalf("expected path /v1/organizationalUnits/OU789012, got %s", r.URL.Path)
		}
		resp := OrganizationalUnitResponse{
			Data: OrganizationalUnit{
				Type: "organizationalUnits",
				ID:   "OU789012",
				Attributes: OrganizationalUnitAttributes{
					Name:        "Engineering",
					Description: "Engineering organizational unit",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ou, err := c.GetOrganizationalUnit(context.Background(), "OU789012", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ou.ID != "OU789012" {
		t.Errorf("expected OU ID OU789012, got %s", ou.ID)
	}
}

func TestGetOrganizationalUnitUserIDs_MultiPage(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/organizationalUnits/OU789012/relationships/users") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		count := requests.Add(1)
		resp := OrganizationalUnitUsersLinkagesResponse{
			Data: []Data{
				{Type: "users", ID: "USER-1"},
			},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		if count == 1 {
			resp.Meta.Paging.NextCursor = "next-page"
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ids, err := c.GetOrganizationalUnitUserIDs(context.Background(), "OU789012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 user IDs, got %d", len(ids))
	}
}

func TestGetOrganizationalUnits_Live(t *testing.T) {
	clientID := os.Getenv("AXM_CLIENT_ID")
	keyID := os.Getenv("AXM_KEY_ID")
	privateKey := os.Getenv("AXM_PRIVATE_KEY")
	scope := os.Getenv("AXM_SCOPE")
	if scope == "" {
		scope = "business.api"
	}
	if clientID == "" || keyID == "" || privateKey == "" {
		t.Skip("AXM_CLIENT_ID, AXM_KEY_ID, and AXM_PRIVATE_KEY must be set for live tests")
	}

	c, err := NewClient("https://api-business.apple.com", clientID, clientID, keyID, scope, privateKey)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	ous, err := c.GetOrganizationalUnits(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get organizational units: %v", err)
	}
	t.Logf("Found %d organizational units", len(ous))
	if len(ous) == 0 {
		t.Fatal("expected at least 1 organizational unit")
	}

	for _, ou := range ous {
		t.Logf("  ID=%s  Name=%s  Description=%s", ou.ID, ou.Attributes.Name, ou.Attributes.Description)
	}

	first := ous[0]

	single, err := c.GetOrganizationalUnit(ctx, first.ID, nil)
	if err != nil {
		t.Fatalf("failed to get organizational unit %s: %v", first.ID, err)
	}
	if single.ID != first.ID {
		t.Errorf("expected ID %s, got %s", first.ID, single.ID)
	}

	userIDs, err := c.GetOrganizationalUnitUserIDs(ctx, first.ID)
	if err != nil {
		t.Fatalf("failed to get user IDs for OU %s: %v", first.ID, err)
	}
	t.Logf("Found %d user IDs for OU %s", len(userIDs), first.ID)
	for _, uid := range userIDs {
		t.Logf("  User ID: %s", uid)
	}
}
