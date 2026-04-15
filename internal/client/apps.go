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

// AppResponse represents a response that contains a single app resource.
type AppResponse struct {
	Data  App           `json:"data"`
	Links DocumentLinks `json:"links"`
}

// AppsResponse represents a response that contains a list of app resources.
type AppsResponse struct {
	Data  []App              `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// App represents an app resource.
type App struct {
	Type       string        `json:"type"`
	ID         string        `json:"id"`
	Attributes AppAttributes `json:"attributes"`
	Links      ResourceLinks `json:"links"`
}

// AppAttributes represents attributes that describe an app resource.
type AppAttributes struct {
	Name        string   `json:"name,omitempty"`
	BundleID    string   `json:"bundleId,omitempty"`
	WebsiteURL  string   `json:"websiteUrl,omitempty"`
	Version     string   `json:"version,omitempty"`
	SupportedOS []string `json:"supportedOS,omitempty"`
	IsCustomApp bool     `json:"isCustomApp,omitempty"`
	AppStoreURL string   `json:"appStoreUrl,omitempty"`
}

// GetApps retrieves all apps in the organization.
func (c *Client) GetApps(ctx context.Context, queryParams url.Values) ([]App, error) {
	var allApps []App
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/apps", c.baseURL)
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

			var response AppsResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allApps = append(allApps, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allApps, nil
}

// GetApp retrieves a single app by ID.
func (c *Client) GetApp(ctx context.Context, id string, queryParams url.Values) (*App, error) {
	baseURL := fmt.Sprintf("%s/v1/apps/%s", c.baseURL, id)
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

	var response AppResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
