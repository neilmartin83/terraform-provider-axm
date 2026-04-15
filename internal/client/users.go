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

// UserResponse represents a response that contains a single user resource.
type UserResponse struct {
	Data  User          `json:"data"`
	Links DocumentLinks `json:"links"`
}

// UsersResponse represents a response that contains a list of user resources.
type UsersResponse struct {
	Data  []User             `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// User represents a user resource.
type User struct {
	Type       string         `json:"type"`
	ID         string         `json:"id"`
	Attributes UserAttributes `json:"attributes"`
	Links      ResourceLinks  `json:"links"`
}

// UserAttributes represents attributes that describe a user resource.
type UserAttributes struct {
	FirstName           string              `json:"firstName,omitempty"`
	LastName            string              `json:"lastName,omitempty"`
	MiddleName          string              `json:"middleName,omitempty"`
	Status              string              `json:"status,omitempty"`
	ManagedAppleAccount string              `json:"managedAppleAccount,omitempty"`
	IsExternalUser      bool                `json:"isExternalUser,omitempty"`
	RoleOuList          []UserRoleOuMapping `json:"roleOuList,omitempty"`
	Email               string              `json:"email,omitempty"`
	EmployeeNumber      string              `json:"employeeNumber,omitempty"`
	CostCenter          string              `json:"costCenter,omitempty"`
	Division            string              `json:"division,omitempty"`
	Department          string              `json:"department,omitempty"`
	JobTitle            string              `json:"jobTitle,omitempty"`
	StartDateTime       string              `json:"startDateTime,omitempty"`
	CreatedDateTime     string              `json:"createdDateTime,omitempty"`
	UpdatedDateTime     string              `json:"updatedDateTime,omitempty"`
	PhoneNumbers        []UserPhoneNumber   `json:"phoneNumbers,omitempty"`
}

// UserRoleOuMapping represents a user's role and organizational unit mapping.
type UserRoleOuMapping struct {
	RoleName string `json:"roleName,omitempty"`
	OuID     string `json:"ouId,omitempty"`
}

// UserPhoneNumber represents a user's phone number.
type UserPhoneNumber struct {
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Type        string `json:"type,omitempty"`
}

// GetUsers retrieves the list of users in the organization.
func (c *Client) GetUsers(ctx context.Context, queryParams url.Values) ([]User, error) {
	var allUsers []User
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/users", c.baseURL)
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

			var response UsersResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allUsers = append(allUsers, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allUsers, nil
}

// GetUser retrieves a single user by ID.
func (c *Client) GetUser(ctx context.Context, id string, queryParams url.Values) (*User, error) {
	baseURL := fmt.Sprintf("%s/v1/users/%s", c.baseURL, id)
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

	var response UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
