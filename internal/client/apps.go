// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strconv"
)

// SupportedOS represents an operating system supported by an app.
type SupportedOS string

const (
	SupportedOSUnspecified SupportedOS = "SUPPORTED_OS_UNSPECIFIED"
	SupportedOSIpadOS      SupportedOS = "SUPPORTED_OS_IPADOS"
	SupportedOSIOS         SupportedOS = "SUPPORTED_OS_IOS"
	SupportedOSMacOS       SupportedOS = "SUPPORTED_OS_MACOS"
	SupportedOSTvOS        SupportedOS = "SUPPORTED_OS_TVOS"
	SupportedOSWatchOS     SupportedOS = "SUPPORTED_OS_WATCHOS"
	SupportedOSVisionOS    SupportedOS = "SUPPORTED_OS_VISIONOS"
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
	Name        string        `json:"name,omitempty"`
	BundleID    string        `json:"bundleId,omitempty"`
	WebsiteURL  string        `json:"websiteUrl,omitempty"`
	Version     string        `json:"version,omitempty"`
	SupportedOS []SupportedOS `json:"supportedOS,omitempty"`
	IsCustomApp bool          `json:"isCustomApp,omitempty"`
	AppStoreURL string        `json:"appStoreUrl,omitempty"`
}

// GetApps retrieves all apps in the organization.
func (c *Client) GetApps(ctx context.Context, queryParams url.Values) ([]App, error) {
	var allApps []App
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/apps", c.baseURL), nil)
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
