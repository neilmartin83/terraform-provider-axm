// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetConfigurations_SinglePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/configurations" {
			t.Fatalf("expected path /v1/configurations, got %s", r.URL.Path)
		}
		resp := ConfigurationsResponse{
			Data: []Configuration{
				{
					Type: "configurations",
					ID:   "config-123",
					Attributes: ConfigurationAttributes{
						Type: "CUSTOM_SETTING",
						Name: "Wi-Fi",
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
	configs, err := c.GetConfigurations(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 configuration, got %d", len(configs))
	}
}

func TestGetConfiguration_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/configurations/config-123" {
			t.Fatalf("expected path /v1/configurations/config-123, got %s", r.URL.Path)
		}
		resp := ConfigurationResponse{
			Data: Configuration{
				Type: "configurations",
				ID:   "config-123",
				Attributes: ConfigurationAttributes{
					Name: "Wi-Fi",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	config, err := c.GetConfiguration(context.Background(), "config-123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.ID != "config-123" {
		t.Errorf("expected config-123, got %s", config.ID)
	}
}

func TestCreateConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/configurations" {
			t.Fatalf("expected path /v1/configurations, got %s", r.URL.Path)
		}

		var request ConfigurationCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if request.Data.Attributes.Name != "Wi-Fi" {
			t.Fatalf("unexpected name: %s", request.Data.Attributes.Name)
		}
		if request.Data.Attributes.CustomSettingsValues.Filename != "WiFi.mobileconfig" {
			t.Fatalf("unexpected filename: %s", request.Data.Attributes.CustomSettingsValues.Filename)
		}

		resp := ConfigurationResponse{
			Data: Configuration{
				Type: "configurations",
				ID:   "config-123",
				Attributes: ConfigurationAttributes{
					Name: "Wi-Fi",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	configuration, err := c.CreateConfiguration(context.Background(), ConfigurationCreateRequest{
		Data: ConfigurationCreateRequestData{
			Type: "configurations",
			Attributes: ConfigurationCreateRequestAttributes{
				Type: "CUSTOM_SETTING",
				Name: "Wi-Fi",
				CustomSettingsValues: CustomSettingsValues{
					ConfigurationProfile: "payload",
					Filename:             "WiFi.mobileconfig",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if configuration.ID != "config-123" {
		t.Errorf("expected config-123, got %s", configuration.ID)
	}
}

func TestUpdateConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/v1/configurations/config-123" {
			t.Fatalf("expected path /v1/configurations/config-123, got %s", r.URL.Path)
		}

		var request ConfigurationUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if request.Data.ID != "config-123" {
			t.Fatalf("unexpected ID: %s", request.Data.ID)
		}

		resp := ConfigurationResponse{
			Data: Configuration{
				Type: "configurations",
				ID:   "config-123",
				Attributes: ConfigurationAttributes{
					Name: "Wi-Fi Updated",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, resp))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	name := "Wi-Fi Updated"
	_, err := c.UpdateConfiguration(context.Background(), ConfigurationUpdateRequest{
		Data: ConfigurationUpdateRequestData{
			Type: "configurations",
			ID:   "config-123",
			Attributes: ConfigurationUpdateRequestAttributes{
				Name: &name,
				CustomSettingsValues: &CustomSettingsValues{
					ConfigurationProfile: "payload",
					Filename:             "WiFi.mobileconfig",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/v1/configurations/config-123" {
			t.Fatalf("expected path /v1/configurations/config-123, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	if err := c.DeleteConfiguration(context.Background(), "config-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
