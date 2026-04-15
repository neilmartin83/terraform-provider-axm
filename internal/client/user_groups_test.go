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

func TestGetUserGroups_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/userGroups" {
			t.Fatalf("expected path /v1/userGroups, got %s", r.URL.Path)
		}
		resp := UserGroupsResponse{
			Data: []UserGroup{
				{
					Type: "userGroups",
					ID:   "UG123456",
					Attributes: UserGroupAttributes{
						Name:             "Engineering",
						Type:             "STANDARD",
						TotalMemberCount: 25,
						Status:           "ACTIVE",
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
	groups, err := c.GetUserGroups(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Attributes.Name != "Engineering" {
		t.Errorf("expected name Engineering, got %s", groups[0].Attributes.Name)
	}
}

func TestGetUserGroup_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/userGroups/UG123456" {
			t.Fatalf("expected path /v1/userGroups/UG123456, got %s", r.URL.Path)
		}
		resp := UserGroupResponse{
			Data: UserGroup{
				Type: "userGroups",
				ID:   "UG123456",
				Attributes: UserGroupAttributes{
					Name:   "Engineering",
					Status: "ACTIVE",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	group, err := c.GetUserGroup(context.Background(), "UG123456", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if group.ID != "UG123456" {
		t.Errorf("expected group ID UG123456, got %s", group.ID)
	}
}

func TestGetUserGroupUserIDs_MultiPage(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/userGroups/UG123456/relationships/users") {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		count := requests.Add(1)
		resp := UserGroupUsersLinkagesResponse{
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
	ids, err := c.GetUserGroupUserIDs(context.Background(), "UG123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 user IDs, got %d", len(ids))
	}
}
