// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// AuditEventsResponse represents a response that contains a list of audit event resources.
type AuditEventsResponse struct {
	Data  []AuditEvent       `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// AuditEvent represents an audit event resource.
type AuditEvent struct {
	Type       string               `json:"type"`
	ID         string               `json:"id"`
	Attributes AuditEventAttributes `json:"attributes"`
}

// AuditEventAttributes represents common audit event attributes and captures event-specific data.
type AuditEventAttributes struct {
	EventDateTime        string         `json:"eventDateTime,omitempty"`
	Type                 string         `json:"type,omitempty"`
	Category             string         `json:"category,omitempty"`
	ActorType            string         `json:"actorType,omitempty"`
	ActorID              string         `json:"actorId,omitempty"`
	ActorName            string         `json:"actorName,omitempty"`
	SubjectType          string         `json:"subjectType,omitempty"`
	SubjectID            string         `json:"subjectId,omitempty"`
	SubjectName          string         `json:"subjectName,omitempty"`
	Outcome              string         `json:"outcome,omitempty"`
	GroupID              string         `json:"groupId,omitempty"`
	EventDataPropertyKey string         `json:"eventDataPropertyKey,omitempty"`
	Additional           map[string]any `json:"-"`
}

// UnmarshalJSON captures common audit event fields and preserves event-specific payloads.
func (a *AuditEventAttributes) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["eventDateTime"].(string); ok {
		a.EventDateTime = v
		delete(raw, "eventDateTime")
	}
	if v, ok := raw["type"].(string); ok {
		a.Type = v
		delete(raw, "type")
	}
	if v, ok := raw["category"].(string); ok {
		a.Category = v
		delete(raw, "category")
	}
	if v, ok := raw["actorType"].(string); ok {
		a.ActorType = v
		delete(raw, "actorType")
	}
	if v, ok := raw["actorId"].(string); ok {
		a.ActorID = v
		delete(raw, "actorId")
	}
	if v, ok := raw["actorName"].(string); ok {
		a.ActorName = v
		delete(raw, "actorName")
	}
	if v, ok := raw["subjectType"].(string); ok {
		a.SubjectType = v
		delete(raw, "subjectType")
	}
	if v, ok := raw["subjectId"].(string); ok {
		a.SubjectID = v
		delete(raw, "subjectId")
	}
	if v, ok := raw["subjectName"].(string); ok {
		a.SubjectName = v
		delete(raw, "subjectName")
	}
	if v, ok := raw["outcome"].(string); ok {
		a.Outcome = v
		delete(raw, "outcome")
	}
	if v, ok := raw["groupId"].(string); ok {
		a.GroupID = v
		delete(raw, "groupId")
	}
	if v, ok := raw["eventDataPropertyKey"].(string); ok {
		a.EventDataPropertyKey = v
		delete(raw, "eventDataPropertyKey")
	}

	a.Additional = raw
	return nil
}

// GetAuditEvents retrieves audit events based on the provided query parameters.
func (c *Client) GetAuditEvents(ctx context.Context, queryParams url.Values) ([]AuditEvent, error) {
	var allEvents []AuditEvent
	nextCursor := ""
	limit := 100

	if queryParams.Has("limit") {
		if parsed, err := strconv.Atoi(queryParams.Get("limit")); err == nil {
			limit = parsed
		}
	}

	for {
		baseURL := fmt.Sprintf("%s/v1/auditEvents", c.baseURL)
		if len(queryParams) > 0 {
			baseURL += "?" + queryParams.Encode()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		if !q.Has("limit") {
			q.Add("limit", strconv.Itoa(limit))
		}
		if nextCursor != "" {
			q.Set("cursor", nextCursor)
		}
		req.URL.RawQuery = q.Encode()
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

			var response AuditEventsResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allEvents = append(allEvents, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allEvents, nil
}
