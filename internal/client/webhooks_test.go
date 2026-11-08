// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestGetWebhooks_MultiPage(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/webhooks" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.URL.Query().Get("limit") != "1000" {
			t.Errorf("expected limit 1000, got %q", r.URL.Query().Get("limit"))
		}
		if requests.Add(1) == 1 && r.URL.Query().Get("cursor") != "" {
			t.Errorf("first request unexpectedly has cursor %q", r.URL.Query().Get("cursor"))
		}
		if requests.Load() == 2 && r.URL.Query().Get("cursor") != "next-page" {
			t.Errorf("expected next-page cursor, got %q", r.URL.Query().Get("cursor"))
		}

		response := WebhooksResponse{
			Data: []Webhook{{ID: "webhook-" + string(rune('0'+requests.Load()))}},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		if requests.Load() == 1 {
			response.Meta.Paging.NextCursor = "next-page"
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, response))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	webhooks, err := c.GetWebhooks(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(webhooks) != 2 || webhooks[0].ID != "webhook-1" || webhooks[1].ID != "webhook-2" {
		t.Fatalf("unexpected webhooks: %#v", webhooks)
	}
}

func TestGetWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/webhooks/webhook-123" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"data":{"type":"webhooks","id":"webhook-123","attributes":{"name":"Device changes","enabled":true,"eventTypes":["ORG_DEVICE_ATTRIBUTE_CHANGED"],"url":"https://example.com/hook"}}}`))
	}))
	defer server.Close()

	webhook, err := newTestClient(t, server).GetWebhook(context.Background(), "webhook-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if webhook.ID != "webhook-123" || webhook.Attributes.Name != "Device changes" || len(webhook.Attributes.EventTypes) != 1 {
		t.Fatalf("unexpected webhook: %#v", webhook)
	}
}

func TestCreateWebhook(t *testing.T) {
	logger := &webhookRequestLogger{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/webhooks" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var request WebhookCreateRequest
		if err := json.UnmarshalRead(r.Body, &request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		attributes := request.Data.Attributes
		if request.Data.Type != "webhooks" || attributes.Name != "Device changes" ||
			!attributes.Enabled || attributes.Secret != "test-secret" ||
			attributes.URL != "https://example.com/hook" ||
			len(attributes.EventTypes) != 1 || attributes.EventTypes[0] != "ORG_DEVICE_ATTRIBUTE_CHANGED" {
			t.Fatalf("unexpected create request: %#v", request)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"data":{"type":"webhooks","id":"webhook-new","attributes":{"name":"Device changes","enabled":true,"eventTypes":["ORG_DEVICE_ATTRIBUTE_CHANGED"],"url":"https://example.com/hook"}}}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	c.SetLogger(logger)
	webhook, err := c.CreateWebhook(context.Background(), WebhookCreateRequest{
		Data: WebhookCreateRequestData{
			Type: "webhooks",
			Attributes: WebhookCreateRequestAttributes{
				Name:       "Device changes",
				Enabled:    true,
				EventTypes: []string{"ORG_DEVICE_ATTRIBUTE_CHANGED"},
				Secret:     "test-secret",
				URL:        "https://example.com/hook",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if webhook.ID != "webhook-new" {
		t.Fatalf("expected webhook-new, got %q", webhook.ID)
	}
	if strings.Contains(string(logger.requestBody), "test-secret") {
		t.Fatal("webhook secret was included in the logged request body")
	}
}

type webhookRequestLogger struct {
	requestBody []byte
}

func (l *webhookRequestLogger) LogRequest(_ context.Context, _, _ string, body []byte) {
	l.requestBody = append([]byte(nil), body...)
}

func (*webhookRequestLogger) LogResponse(context.Context, int, http.Header, []byte) {}

func (*webhookRequestLogger) LogAuth(context.Context, string, map[string]any) {}

func TestUpdateWebhook_SendsOnlyConfiguredAttributes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/v1/webhooks/webhook-123" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var request WebhookUpdateRequest
		if err := json.UnmarshalRead(r.Body, &request); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if request.Data.ID != "webhook-123" || request.Data.Type != "webhooks" {
			t.Fatalf("unexpected request data: %#v", request.Data)
		}
		if request.Data.Attributes.Enabled == nil || *request.Data.Attributes.Enabled {
			t.Fatal("expected enabled=false to be sent")
		}
		if request.Data.Attributes.Name != nil || request.Data.Attributes.Secret != nil {
			t.Fatalf("unexpected omitted fields: %#v", request.Data.Attributes)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"type":"webhooks","id":"webhook-123","attributes":{"enabled":false}}}`))
	}))
	defer server.Close()

	enabled := false
	_, err := newTestClient(t, server).UpdateWebhook(context.Background(), WebhookUpdateRequest{
		Data: WebhookUpdateRequestData{
			Type: "webhooks",
			ID:   "webhook-123",
			Attributes: WebhookUpdateRequestAttributes{
				Enabled: &enabled,
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteWebhook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/webhooks/webhook-123" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	if err := newTestClient(t, server).DeleteWebhook(context.Background(), "webhook-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetWebhookDeliveries_MultiPage(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/webhooks/webhook-123/deliveries" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		count := requests.Add(1)
		response := WebhookDeliveriesResponse{
			Data: []WebhookDelivery{{
				Type: "webhookDeliveries",
				ID:   "delivery-" + string(rune('0'+count)),
				Attributes: WebhookDeliveryAttributes{
					DeliveryState: "SUCCEEDED",
					Request:       WebhookDeliveryRequest{URL: "https://example.com/hook"},
					Response:      WebhookDeliveryResponseInfo{HTTPStatusCode: http.StatusOK, Body: "ok"},
				},
				Relationships: WebhookDeliveryRelations{
					Event: WebhookDeliveryEvent{Data: Data{Type: "webhookEvents", ID: "event-123"}},
				},
			}},
			Meta: Meta{Paging: Paging{Limit: 100}},
		}
		if count == 1 {
			response.Meta.Paging.NextCursor = "next-page"
		} else if r.URL.Query().Get("cursor") != "next-page" {
			t.Errorf("expected next-page cursor, got %q", r.URL.Query().Get("cursor"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(mustMarshalJSON(t, response))
	}))
	defer server.Close()

	deliveries, err := newTestClient(t, server).GetWebhookDeliveries(context.Background(), "webhook-123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deliveries) != 2 || deliveries[0].Attributes.DeliveryState != "SUCCEEDED" ||
		deliveries[0].Attributes.Response.HTTPStatusCode != http.StatusOK ||
		deliveries[0].Relationships.Event.Data.ID != "event-123" {
		t.Fatalf("unexpected deliveries: %#v", deliveries)
	}
}

func TestWebhookUpdateRequestOmitsNilAttributes(t *testing.T) {
	data, err := json.Marshal(WebhookUpdateRequest{
		Data: WebhookUpdateRequestData{
			Type:       "webhooks",
			ID:         "webhook-123",
			Attributes: WebhookUpdateRequestAttributes{},
		},
	})
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	if strings.Contains(string(data), `"name"`) || strings.Contains(string(data), `"secret"`) {
		t.Fatalf("nil attributes should be omitted: %s", data)
	}
}
