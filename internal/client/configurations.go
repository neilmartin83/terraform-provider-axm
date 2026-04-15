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

// ConfigurationResponse represents a response that contains a single Configuration resource.
type ConfigurationResponse struct {
	Data  Configuration `json:"data"`
	Links DocumentLinks `json:"links"`
}

// ConfigurationsResponse represents a response that contains a list of Configuration resources.
type ConfigurationsResponse struct {
	Data  []Configuration    `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// Configuration represents a Configuration resource.
type Configuration struct {
	Type       string                  `json:"type"`
	ID         string                  `json:"id"`
	Attributes ConfigurationAttributes `json:"attributes"`
	Links      ResourceLinks           `json:"links"`
}

// ConfigurationAttributes represents attributes that describe a Configuration resource.
type ConfigurationAttributes struct {
	Type                   string                `json:"type,omitempty"`
	Name                   string                `json:"name,omitempty"`
	ConfiguredForPlatforms []string              `json:"configuredForPlatforms,omitempty"`
	CustomSettingsValues   *CustomSettingsValues `json:"customSettingsValues,omitempty"`
	CreatedDateTime        string                `json:"createdDateTime,omitempty"`
	UpdatedDateTime        string                `json:"updatedDateTime,omitempty"`
}

// CustomSettingsValues represents the custom settings payload for a Configuration.
type CustomSettingsValues struct {
	ConfigurationProfile string `json:"configurationProfile,omitempty"`
	Filename             string `json:"filename,omitempty"`
}

// ConfigurationCreateRequest represents a request to create a Configuration.
type ConfigurationCreateRequest struct {
	Data ConfigurationCreateRequestData `json:"data"`
}

// ConfigurationCreateRequestData represents the data element of a create request.
type ConfigurationCreateRequestData struct {
	Type       string                               `json:"type"`
	Attributes ConfigurationCreateRequestAttributes `json:"attributes"`
}

// ConfigurationCreateRequestAttributes represents attributes for creating a Configuration.
type ConfigurationCreateRequestAttributes struct {
	Type                   string               `json:"type"`
	Name                   string               `json:"name"`
	ConfiguredForPlatforms []string             `json:"configuredForPlatforms,omitempty"`
	CustomSettingsValues   CustomSettingsValues `json:"customSettingsValues"`
}

// ConfigurationUpdateRequest represents a request to update a Configuration.
type ConfigurationUpdateRequest struct {
	Data ConfigurationUpdateRequestData `json:"data"`
}

// ConfigurationUpdateRequestData represents the data element of an update request.
type ConfigurationUpdateRequestData struct {
	Type       string                               `json:"type"`
	ID         string                               `json:"id"`
	Attributes ConfigurationUpdateRequestAttributes `json:"attributes"`
}

// ConfigurationUpdateRequestAttributes represents attributes for updating a Configuration.
type ConfigurationUpdateRequestAttributes struct {
	Name                   *string               `json:"name,omitempty"`
	ConfiguredForPlatforms []string              `json:"configuredForPlatforms,omitempty"`
	CustomSettingsValues   *CustomSettingsValues `json:"customSettingsValues,omitempty"`
}

// GetConfigurations retrieves all Configurations in the organization.
func (c *Client) GetConfigurations(ctx context.Context, queryParams url.Values) ([]Configuration, error) {
	var allConfigs []Configuration
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/configurations", c.baseURL)
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

			var response ConfigurationsResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allConfigs = append(allConfigs, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allConfigs, nil
}

// GetConfiguration retrieves a single Configuration by ID.
func (c *Client) GetConfiguration(ctx context.Context, id string, queryParams url.Values) (*Configuration, error) {
	baseURL := fmt.Sprintf("%s/v1/configurations/%s", c.baseURL, id)
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

	var response ConfigurationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// CreateConfiguration creates a new Configuration.
func (c *Client) CreateConfiguration(ctx context.Context, request ConfigurationCreateRequest) (*Configuration, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v1/configurations", c.baseURL), bytes.NewReader(jsonData))
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

	var response ConfigurationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// UpdateConfiguration updates an existing Configuration.
func (c *Client) UpdateConfiguration(ctx context.Context, request ConfigurationUpdateRequest) (*Configuration, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/v1/configurations/%s", c.baseURL, request.Data.ID), bytes.NewReader(jsonData))
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

	var response ConfigurationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// DeleteConfiguration deletes a Configuration by ID.
func (c *Client) DeleteConfiguration(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/v1/configurations/%s", c.baseURL, id), nil)
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
