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

// PackageResponse represents a response that contains a single package resource.
type PackageResponse struct {
	Data  Package       `json:"data"`
	Links DocumentLinks `json:"links"`
}

// PackagesResponse represents a response that contains a list of package resources.
type PackagesResponse struct {
	Data  []Package          `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// Package represents a package resource.
type Package struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Attributes PackageAttributes `json:"attributes"`
	Links      ResourceLinks     `json:"links"`
}

// PackageAttributes represents attributes that describe a package resource.
type PackageAttributes struct {
	Name            string   `json:"name,omitempty"`
	URL             string   `json:"url,omitempty"`
	Hash            string   `json:"hash,omitempty"`
	BundleIDs       []string `json:"bundleIds,omitempty"`
	Description     string   `json:"description,omitempty"`
	Version         string   `json:"version,omitempty"`
	CreatedDateTime string   `json:"createdDateTime,omitempty"`
	UpdatedDateTime string   `json:"updatedDateTime,omitempty"`
}

// GetPackages retrieves all packages in the organization.
func (c *Client) GetPackages(ctx context.Context, queryParams url.Values) ([]Package, error) {
	var allPackages []Package
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/packages", c.baseURL)
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

			var response PackagesResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allPackages = append(allPackages, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allPackages, nil
}

// GetPackage retrieves a single package by ID.
func (c *Client) GetPackage(ctx context.Context, id string, queryParams url.Values) (*Package, error) {
	baseURL := fmt.Sprintf("%s/v1/packages/%s", c.baseURL, id)
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

	var response PackageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
