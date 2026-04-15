// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// BlueprintResponse represents a response that contains a single Blueprint resource.
type BlueprintResponse struct {
	Data  Blueprint     `json:"data"`
	Links DocumentLinks `json:"links"`
}

// BlueprintsResponse represents a response that contains a list of Blueprint resources.
type BlueprintsResponse struct {
	Data  []Blueprint        `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// Blueprint represents a Blueprint resource.
type Blueprint struct {
	Type          string                 `json:"type"`
	ID            string                 `json:"id"`
	Attributes    BlueprintAttributes    `json:"attributes"`
	Relationships BlueprintRelationships `json:"relationships"`
	Links         ResourceLinks          `json:"links"`
}

// BlueprintAttributes represents attributes that describe a Blueprint resource.
type BlueprintAttributes struct {
	Name                string `json:"name,omitempty"`
	Description         string `json:"description,omitempty"`
	Status              string `json:"status,omitempty"`
	AppLicenseDeficient bool   `json:"appLicenseDeficient,omitempty"`
	CreatedDateTime     string `json:"createdDateTime,omitempty"`
	UpdatedDateTime     string `json:"updatedDateTime,omitempty"`
}

// BlueprintRelationships represents link relationships for a Blueprint.
type BlueprintRelationships struct {
	Apps           BlueprintRelationshipLinks `json:"apps"`
	Configurations BlueprintRelationshipLinks `json:"configurations"`
	Packages       BlueprintRelationshipLinks `json:"packages"`
	OrgDevices     BlueprintRelationshipLinks `json:"orgDevices"`
	Users          BlueprintRelationshipLinks `json:"users"`
	UserGroups     BlueprintRelationshipLinks `json:"userGroups"`
}

// BlueprintRelationshipLinks represents relationship links for a Blueprint.
type BlueprintRelationshipLinks struct {
	Links RelationshipLinks `json:"links"`
}

// BlueprintCreateRequest represents a request to create a Blueprint.
type BlueprintCreateRequest struct {
	Data BlueprintCreateRequestData `json:"data"`
}

// BlueprintCreateRequestData represents the data element of a create request.
type BlueprintCreateRequestData struct {
	Type          string                         `json:"type"`
	Attributes    BlueprintCreateAttributes      `json:"attributes"`
	Relationships *BlueprintRelationshipsRequest `json:"relationships,omitempty"`
}

// BlueprintCreateAttributes represents attributes for creating a Blueprint.
type BlueprintCreateAttributes struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// BlueprintUpdateRequest represents a request to update a Blueprint.
type BlueprintUpdateRequest struct {
	Data BlueprintUpdateRequestData `json:"data"`
}

// BlueprintUpdateRequestData represents the data element of an update request.
type BlueprintUpdateRequestData struct {
	Type          string                         `json:"type"`
	ID            string                         `json:"id"`
	Attributes    *BlueprintUpdateAttributes     `json:"attributes,omitempty"`
	Relationships *BlueprintRelationshipsRequest `json:"relationships,omitempty"`
}

// BlueprintUpdateAttributes represents attributes for updating a Blueprint.
type BlueprintUpdateAttributes struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// BlueprintRelationshipsRequest represents relationships included in a Blueprint request.
type BlueprintRelationshipsRequest struct {
	Apps           *BlueprintRelationshipData `json:"apps,omitempty"`
	Configurations *BlueprintRelationshipData `json:"configurations,omitempty"`
	Packages       *BlueprintRelationshipData `json:"packages,omitempty"`
	OrgDevices     *BlueprintRelationshipData `json:"orgDevices,omitempty"`
	Users          *BlueprintRelationshipData `json:"users,omitempty"`
	UserGroups     *BlueprintRelationshipData `json:"userGroups,omitempty"`
}

// BlueprintRelationshipData represents related resource IDs in a Blueprint request.
type BlueprintRelationshipData struct {
	Data []Data `json:"data,omitempty"`
}

// BlueprintLinkagesResponse represents a list of relationship IDs for a Blueprint.
type BlueprintLinkagesResponse struct {
	Data  []Data             `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// BlueprintRelationshipRequest represents a relationship update payload.
type BlueprintRelationshipRequest struct {
	Data []Data `json:"data"`
}

// GetBlueprints retrieves all Blueprints in the organization.
func (c *Client) GetBlueprints(ctx context.Context, queryParams url.Values) ([]Blueprint, error) {
	var allBlueprints []Blueprint
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/blueprints", c.baseURL)
		if len(queryParams) > 0 {
			baseURL += "?" + queryParams.Encode()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("limit", strconv.Itoa(limit))
		if nextCursor != "" {
			q.Add("cursor", nextCursor)
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

			var response BlueprintsResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allBlueprints = append(allBlueprints, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allBlueprints, nil
}

// GetBlueprint retrieves a single Blueprint by ID.
func (c *Client) GetBlueprint(ctx context.Context, id string, queryParams url.Values) (*Blueprint, error) {
	baseURL := fmt.Sprintf("%s/v1/blueprints/%s", c.baseURL, id)
	if len(queryParams) > 0 {
		baseURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL, nil)
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

	var response BlueprintResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// CreateBlueprint creates a new Blueprint.
func (c *Client) CreateBlueprint(ctx context.Context, request BlueprintCreateRequest) (*Blueprint, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v1/blueprints", c.baseURL), bytes.NewReader(jsonData))
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

	var response BlueprintResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// UpdateBlueprint updates an existing Blueprint.
func (c *Client) UpdateBlueprint(ctx context.Context, request BlueprintUpdateRequest) (*Blueprint, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/v1/blueprints/%s", c.baseURL, request.Data.ID), bytes.NewReader(jsonData))
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

	var response BlueprintResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// DeleteBlueprint deletes a Blueprint by ID.
func (c *Client) DeleteBlueprint(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/v1/blueprints/%s", c.baseURL, id), nil)
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

// GetBlueprintRelationshipIDs retrieves related resource IDs for a Blueprint relationship.
func (c *Client) GetBlueprintRelationshipIDs(ctx context.Context, blueprintID, relationship string) ([]string, error) {
	var allIDs []string
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/blueprints/%s/relationships/%s", c.baseURL, blueprintID, relationship), nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("limit", strconv.Itoa(limit))
		if nextCursor != "" {
			q.Add("cursor", nextCursor)
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

			var response BlueprintLinkagesResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			for _, entry := range response.Data {
				allIDs = append(allIDs, entry.ID)
			}
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allIDs, nil
}

// UpdateBlueprintRelationship updates related resources for a Blueprint relationship.
func (c *Client) UpdateBlueprintRelationship(ctx context.Context, blueprintID, relationship, resourceType, method string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	data := make([]Data, len(ids))
	for i, id := range ids {
		data[i] = Data{
			Type: resourceType,
			ID:   id,
		}
	}

	request := BlueprintRelationshipRequest{
		Data: data,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method,
		fmt.Sprintf("%s/v1/blueprints/%s/relationships/%s", c.baseURL, blueprintID, relationship), bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return c.handleErrorResponse(resp)
	}

	return nil
}
