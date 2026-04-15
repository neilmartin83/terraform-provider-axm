// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGetBlueprints_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/blueprints" {
			t.Fatalf("expected path /v1/blueprints, got %s", r.URL.Path)
		}
		resp := BlueprintsResponse{
			Data: []Blueprint{
				{
					Type: "blueprints",
					ID:   "blueprint-123",
					Attributes: BlueprintAttributes{
						Name:   "Engineering",
						Status: "ACTIVE",
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
	blueprints, err := c.GetBlueprints(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(blueprints) != 1 {
		t.Fatalf("expected 1 blueprint, got %d", len(blueprints))
	}
	if blueprints[0].Attributes.Name != "Engineering" {
		t.Errorf("unexpected name: %s", blueprints[0].Attributes.Name)
	}
}

func TestGetBlueprint_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/blueprints/blueprint-123" {
			t.Fatalf("expected path /v1/blueprints/blueprint-123, got %s", r.URL.Path)
		}
		resp := BlueprintResponse{
			Data: Blueprint{
				Type: "blueprints",
				ID:   "blueprint-123",
				Attributes: BlueprintAttributes{
					Name: "Engineering",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	blueprint, err := c.GetBlueprint(context.Background(), "blueprint-123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blueprint.ID != "blueprint-123" {
		t.Errorf("expected blueprint-123, got %s", blueprint.ID)
	}
}

func TestCreateBlueprint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/blueprints" {
			t.Fatalf("expected path /v1/blueprints, got %s", r.URL.Path)
		}

		var request BlueprintCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if request.Data.Attributes.Name != "Marketing" {
			t.Fatalf("unexpected name: %s", request.Data.Attributes.Name)
		}

		resp := BlueprintResponse{
			Data: Blueprint{
				Type: "blueprints",
				ID:   "blueprint-new",
				Attributes: BlueprintAttributes{
					Name: "Marketing",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	blueprint, err := c.CreateBlueprint(context.Background(), BlueprintCreateRequest{
		Data: BlueprintCreateRequestData{
			Type: "blueprints",
			Attributes: BlueprintCreateAttributes{
				Name: "Marketing",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if blueprint.ID != "blueprint-new" {
		t.Errorf("expected blueprint-new, got %s", blueprint.ID)
	}
}

func TestUpdateBlueprint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/v1/blueprints/blueprint-123" {
			t.Fatalf("expected path /v1/blueprints/blueprint-123, got %s", r.URL.Path)
		}

		var request BlueprintUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if request.Data.ID != "blueprint-123" {
			t.Fatalf("unexpected ID: %s", request.Data.ID)
		}

		resp := BlueprintResponse{
			Data: Blueprint{
				Type: "blueprints",
				ID:   "blueprint-123",
				Attributes: BlueprintAttributes{
					Name: "Updated",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	name := "Updated"
	_, err := c.UpdateBlueprint(context.Background(), BlueprintUpdateRequest{
		Data: BlueprintUpdateRequestData{
			Type: "blueprints",
			ID:   "blueprint-123",
			Attributes: &BlueprintUpdateAttributes{
				Name: &name,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteBlueprint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/blueprints/blueprint-123" {
			t.Fatalf("expected path /v1/blueprints/blueprint-123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	if err := c.DeleteBlueprint(context.Background(), "blueprint-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetBlueprintRelationshipIDs_MultiPage(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/blueprints/blueprint-123/relationships/apps") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		count := requests.Add(1)

		resp := BlueprintLinkagesResponse{
			Data: []Data{{Type: "apps", ID: "361309726"}},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		if count == 1 {
			resp.Meta.Paging.NextCursor = "next"
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ids, err := c.GetBlueprintRelationshipIDs(context.Background(), "blueprint-123", "apps")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 IDs, got %d", len(ids))
	}
}

func TestUpdateBlueprintRelationship(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blueprints/blueprint-123/relationships/apps" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	if err := c.UpdateBlueprintRelationship(context.Background(), "blueprint-123", "apps", "apps", http.MethodPost, []string{"361309726"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
