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

// UserGroupResponse represents a response that contains a single user group resource.
type UserGroupResponse struct {
	Data  UserGroup     `json:"data"`
	Links DocumentLinks `json:"links"`
}

// UserGroupsResponse represents a response that contains a list of user group resources.
type UserGroupsResponse struct {
	Data  []UserGroup        `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// UserGroup represents a user group resource.
type UserGroup struct {
	Type          string                 `json:"type"`
	ID            string                 `json:"id"`
	Attributes    UserGroupAttributes    `json:"attributes"`
	Relationships UserGroupRelationships `json:"relationships"`
	Links         ResourceLinks          `json:"links"`
}

// UserGroupAttributes represents attributes that describe a user group resource.
type UserGroupAttributes struct {
	OuID             string `json:"ouId,omitempty"`
	Name             string `json:"name,omitempty"`
	Type             string `json:"type,omitempty"`
	TotalMemberCount int    `json:"totalMemberCount,omitempty"`
	Status           string `json:"status,omitempty"`
	CreatedDateTime  string `json:"createdDateTime,omitempty"`
	UpdatedDateTime  string `json:"updatedDateTime,omitempty"`
}

// UserGroupRelationships represents the relationships you include in the request.
type UserGroupRelationships struct {
	Users UserGroupRelationshipsUsers `json:"users"`
}

// UserGroupRelationshipsUsers represents the relationship between a user group and users.
type UserGroupRelationshipsUsers struct {
	Links RelationshipLinks `json:"links"`
}

// UserGroupUsersLinkagesResponse represents a list of user IDs for a user group.
type UserGroupUsersLinkagesResponse struct {
	Data  []Data             `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// GetUserGroups retrieves all user groups in the organization.
func (c *Client) GetUserGroups(ctx context.Context, queryParams url.Values) ([]UserGroup, error) {
	var allGroups []UserGroup
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/userGroups", c.baseURL)
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

			var response UserGroupsResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allGroups = append(allGroups, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allGroups, nil
}

// GetUserGroup retrieves a single user group by ID.
func (c *Client) GetUserGroup(ctx context.Context, id string, queryParams url.Values) (*UserGroup, error) {
	baseURL := fmt.Sprintf("%s/v1/userGroups/%s", c.baseURL, id)
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

	var response UserGroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// GetUserGroupUserIDs retrieves all user IDs for a user group.
func (c *Client) GetUserGroupUserIDs(ctx context.Context, groupID string) ([]string, error) {
	var allUserIDs []string
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/userGroups/%s/relationships/users", c.baseURL, groupID), nil)
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

			var response UserGroupUsersLinkagesResponse
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
