// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strconv"
)

// WebhookResponse represents a response that contains a single webhook resource.
type WebhookResponse struct {
	Data  Webhook       `json:"data"`
	Links DocumentLinks `json:"links"`
}

// WebhooksResponse represents a response that contains a list of webhook resources.
type WebhooksResponse struct {
	Data  []Webhook          `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// Webhook represents a webhook resource.
type Webhook struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Attributes WebhookAttributes `json:"attributes"`
	Links      ResourceLinks     `json:"links"`
}

// WebhookAttributes describes a webhook resource.
type WebhookAttributes struct {
	Name            string   `json:"name,omitempty"`
	Enabled         bool     `json:"enabled"`
	EventTypes      []string `json:"eventTypes,omitempty"`
	URL             string   `json:"url,omitempty"`
	CreatedDateTime string   `json:"createdDateTime,omitempty"`
	UpdatedDateTime string   `json:"updatedDateTime,omitempty"`
}

// WebhookDeliveryResponse represents a response that contains a single webhook delivery.
type WebhookDeliveryResponse struct {
	Data  WebhookDelivery `json:"data"`
	Links DocumentLinks   `json:"links"`
}

// WebhookDeliveriesResponse represents a response that contains webhook deliveries.
type WebhookDeliveriesResponse struct {
	Data  []WebhookDelivery  `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// WebhookDelivery represents a webhook delivery resource.
type WebhookDelivery struct {
	Type          string                    `json:"type"`
	ID            string                    `json:"id"`
	Attributes    WebhookDeliveryAttributes `json:"attributes"`
	Relationships WebhookDeliveryRelations  `json:"relationships"`
}

// WebhookDeliveryAttributes describes a webhook delivery.
type WebhookDeliveryAttributes struct {
	CreatedDateTime string                      `json:"createdDateTime,omitempty"`
	DeliveryState   string                      `json:"deliveryState,omitempty"`
	ErrorMessage    string                      `json:"errorMessage,omitempty"`
	Redelivery      bool                        `json:"redelivery"`
	SentDateTime    string                      `json:"sentDateTime,omitempty"`
	Request         WebhookDeliveryRequest      `json:"request"`
	Response        WebhookDeliveryResponseInfo `json:"response"`
}

// WebhookDeliveryRequest contains information about the delivery request.
type WebhookDeliveryRequest struct {
	URL string `json:"url,omitempty"`
}

// WebhookDeliveryResponseInfo contains information about the delivery response.
type WebhookDeliveryResponseInfo struct {
	HTTPStatusCode int    `json:"httpStatusCode"`
	Body           string `json:"body,omitempty"`
}

// WebhookDeliveryRelations describes the event associated with a delivery.
type WebhookDeliveryRelations struct {
	Event WebhookDeliveryEvent `json:"event"`
}

// WebhookDeliveryEvent describes the event associated with a delivery.
type WebhookDeliveryEvent struct {
	Data Data `json:"data"`
}

// WebhookCreateRequest represents a request to create a webhook.
type WebhookCreateRequest struct {
	Data WebhookCreateRequestData `json:"data"`
}

// WebhookCreateRequestData is the data element of a webhook create request.
type WebhookCreateRequestData struct {
	Type       string                         `json:"type"`
	Attributes WebhookCreateRequestAttributes `json:"attributes"`
}

// WebhookCreateRequestAttributes contains the required webhook fields.
type WebhookCreateRequestAttributes struct {
	Name       string   `json:"name"`
	Enabled    bool     `json:"enabled"`
	EventTypes []string `json:"eventTypes"`
	Secret     string   `json:"secret"`
	URL        string   `json:"url"`
}

// WebhookUpdateRequest represents a request to update a webhook.
type WebhookUpdateRequest struct {
	Data WebhookUpdateRequestData `json:"data"`
}

// WebhookUpdateRequestData is the data element of a webhook update request.
type WebhookUpdateRequestData struct {
	Type       string                         `json:"type"`
	ID         string                         `json:"id"`
	Attributes WebhookUpdateRequestAttributes `json:"attributes"`
}

// WebhookUpdateRequestAttributes contains the webhook fields being changed.
type WebhookUpdateRequestAttributes struct {
	Name       *string   `json:"name,omitzero"`
	Enabled    *bool     `json:"enabled,omitzero"`
	EventTypes *[]string `json:"eventTypes,omitzero"`
	Secret     *string   `json:"secret,omitzero"`
	URL        *string   `json:"url,omitzero"`
}

// GetWebhooks retrieves all webhooks in the organization.
func (c *Client) GetWebhooks(ctx context.Context, queryParams url.Values) ([]Webhook, error) {
	var allWebhooks []Webhook
	nextCursor := ""
	limit := 1000

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/webhooks", c.baseURL), nil)
		if err != nil {
			return nil, err
		}

		params := make(url.Values)
		maps.Copy(params, queryParams)
		params.Set("limit", strconv.Itoa(limit))
		if nextCursor != "" {
			params.Set("cursor", nextCursor)
		}
		req.URL.RawQuery = params.Encode()
		req.Header.Set("Accept", "application/json")

		resp, err := c.doRequest(ctx, req)
		if err != nil {
			return nil, err
		}

		if err := func() error {
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return c.handleErrorResponse(resp)
			}

			var response WebhooksResponse
			if err := json.UnmarshalRead(resp.Body, &response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}
			allWebhooks = append(allWebhooks, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allWebhooks, nil
}

// GetWebhook retrieves a webhook by ID.
func (c *Client) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v1/webhooks/%s", c.baseURL, url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var response WebhookResponse
	if err := json.UnmarshalRead(resp.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	return &response.Data, nil
}

// CreateWebhook creates a webhook.
func (c *Client) CreateWebhook(ctx context.Context, request WebhookCreateRequest) (*Webhook, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v1/webhooks", c.baseURL), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var response WebhookResponse
	if err := json.UnmarshalRead(resp.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	return &response.Data, nil
}

// UpdateWebhook updates an existing webhook.
func (c *Client) UpdateWebhook(ctx context.Context, request WebhookUpdateRequest) (*Webhook, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/v1/webhooks/%s", c.baseURL, url.PathEscape(request.Data.ID)), bytes.NewReader(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var response WebhookResponse
	if err := json.UnmarshalRead(resp.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	return &response.Data, nil
}

// DeleteWebhook deletes a webhook by ID.
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/v1/webhooks/%s", c.baseURL, url.PathEscape(id)), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent {
		return c.handleErrorResponse(resp)
	}
	return nil
}

// GetWebhookDeliveries retrieves all deliveries for a webhook.
func (c *Client) GetWebhookDeliveries(ctx context.Context, webhookID string, queryParams url.Values) ([]WebhookDelivery, error) {
	var allDeliveries []WebhookDelivery
	nextCursor := ""
	limit := 1000

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/webhooks/%s/deliveries", c.baseURL, url.PathEscape(webhookID)), nil)
		if err != nil {
			return nil, err
		}

		params := make(url.Values)
		maps.Copy(params, queryParams)
		params.Set("limit", strconv.Itoa(limit))
		if nextCursor != "" {
			params.Set("cursor", nextCursor)
		}
		req.URL.RawQuery = params.Encode()
		req.Header.Set("Accept", "application/json")

		resp, err := c.doRequest(ctx, req)
		if err != nil {
			return nil, err
		}

		if err := func() error {
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				return c.handleErrorResponse(resp)
			}

			var response WebhookDeliveriesResponse
			if err := json.UnmarshalRead(resp.Body, &response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}
			allDeliveries = append(allDeliveries, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allDeliveries, nil
}
