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

// OrganizationalUnitResponse represents a response that contains a single organizational unit resource.
type OrganizationalUnitResponse struct {
	Data  OrganizationalUnit `json:"data"`
	Links DocumentLinks      `json:"links"`
}

// OrganizationalUnitsResponse represents a response that contains a list of organizational unit resources.
type OrganizationalUnitsResponse struct {
	Data  []OrganizationalUnit `json:"data"`
	Links PagedDocumentLinks   `json:"links"`
	Meta  Meta                 `json:"meta"`
}

// OrganizationalUnit represents an organizational unit resource.
type OrganizationalUnit struct {
	Type          string                          `json:"type"`
	ID            string                          `json:"id"`
	Attributes    OrganizationalUnitAttributes    `json:"attributes"`
	Relationships OrganizationalUnitRelationships `json:"relationships"`
	Links         ResourceLinks                   `json:"links"`
}

// OrganizationalUnitAttributes represents attributes that describe an organizational unit resource.
type OrganizationalUnitAttributes struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	CreatedDateTime string `json:"createdDateTime,omitempty"`
	UpdatedDateTime string `json:"updatedDateTime,omitempty"`
}

// OrganizationalUnitRelationships represents the relationships of an organizational unit.
type OrganizationalUnitRelationships struct {
	Users OrganizationalUnitRelationshipsUsers `json:"users"`
}

// OrganizationalUnitRelationshipsUsers represents the relationship between an organizational unit and users.
type OrganizationalUnitRelationshipsUsers struct {
	Links RelationshipLinks `json:"links"`
}

// OrganizationalUnitUsersLinkagesResponse represents a list of user IDs for an organizational unit.
type OrganizationalUnitUsersLinkagesResponse struct {
	Data  []Data             `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// GetOrganizationalUnits retrieves all organizational units in the organization.
func (c *Client) GetOrganizationalUnits(ctx context.Context, queryParams url.Values) ([]OrganizationalUnit, error) {
	var allOUs []OrganizationalUnit
	nextCursor := ""
	limit := 1000

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/organizationalUnits", c.baseURL), nil)
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

			var response OrganizationalUnitsResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allOUs = append(allOUs, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allOUs, nil
}

// GetOrganizationalUnit retrieves a single organizational unit by ID.
func (c *Client) GetOrganizationalUnit(ctx context.Context, id string, queryParams url.Values) (*OrganizationalUnit, error) {
	baseURL := fmt.Sprintf("%s/v1/organizationalUnits/%s", c.baseURL, id)
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

	var response OrganizationalUnitResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// GetOrganizationalUnitUserIDs retrieves all user IDs for an organizational unit.
func (c *Client) GetOrganizationalUnitUserIDs(ctx context.Context, ouID string) ([]string, error) {
	var allUserIDs []string
	nextCursor := ""
	limit := 1000

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/organizationalUnits/%s/relationships/users", c.baseURL, ouID), nil)
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

			var response OrganizationalUnitUsersLinkagesResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			for _, user := range response.Data {
				if user.Type == "users" {
					allUserIDs = append(allUserIDs, user.ID)
				}
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

	return allUserIDs, nil
}
