// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUsers_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/users" {
			t.Fatalf("expected path /v1/users, got %s", r.URL.Path)
		}
		resp := UsersResponse{
			Data: []User{
				{
					Type: "users",
					ID:   "1234567890",
					Attributes: UserAttributes{
						FirstName:           "John",
						LastName:            "Doe",
						ManagedAppleAccount: "john.doe@example.com",
						Status:              "ACTIVE",
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
	users, err := c.GetUsers(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
	if users[0].Attributes.ManagedAppleAccount != "john.doe@example.com" {
		t.Errorf("unexpected managedAppleAccount: %s", users[0].Attributes.ManagedAppleAccount)
	}
}

func TestGetUser_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/users/1234567890" {
			t.Fatalf("expected path /v1/users/1234567890, got %s", r.URL.Path)
		}
		resp := UserResponse{
			Data: User{
				Type: "users",
				ID:   "1234567890",
				Attributes: UserAttributes{
					FirstName: "Jane",
					LastName:  "Doe",
					Status:    "ACTIVE",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	user, err := c.GetUser(context.Background(), "1234567890", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "1234567890" {
		t.Errorf("expected user ID 1234567890, got %s", user.ID)
	}
	if user.Attributes.FirstName != "Jane" {
		t.Errorf("expected first name Jane, got %s", user.Attributes.FirstName)
	}
}
