// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strconv"
)

// mdmServersFields is the complete list of server attributes the provider needs.
// Apple's API uses JSON:API sparse fieldsets — omitting this param causes it to
// return a default subset that may exclude lastConnectedDateTime and lastConnectedIp.
const mdmServersFields = "serverName,serverType,status,deviceCount,enableMdmDisownFlag,defaultProductFamilies,lastConnectedDateTime,lastConnectedIp,createdDateTime,updatedDateTime"

// MdmServerResponse represents a response that contains a single device management service resource.
type MdmServerResponse struct {
	Data     MdmServer     `json:"data"`
	Included []OrgDevice   `json:"included,omitempty"`
	Links    DocumentLinks `json:"links"`
}

// MdmServersResponse represents a response that contains a list of device management service resources.
type MdmServersResponse struct {
	Data     []MdmServer        `json:"data"`
	Included []OrgDevice        `json:"included,omitempty"`
	Links    PagedDocumentLinks `json:"links"`
	Meta     Meta               `json:"meta"`
}

// MdmServer represents the data structure that represents a device management service resource in an organization
type MdmServer struct {
	Type          string                 `json:"type"`
	ID            string                 `json:"id"`
	Attributes    MdmServerAttribute     `json:"attributes"`
	Relationships MdmServerRelationships `json:"relationships"`
}

// MdmServerAttribute represents attributes that describe a device management service resource
type MdmServerAttribute struct {
	ServerName             string   `json:"serverName"`
	ServerType             string   `json:"serverType"`
	Status                 *string  `json:"status"`
	DeviceCount            *int64   `json:"deviceCount"`
	EnableMdmDisownFlag    *bool    `json:"enableMdmDisownFlag"`
	DefaultProductFamilies []string `json:"defaultProductFamilies"`
	LastConnectedDateTime  *string  `json:"lastConnectedDateTime"`
	LastConnectedIp        *string  `json:"lastConnectedIp"`
	CreatedDateTime        string   `json:"createdDateTime"`
	UpdatedDateTime        string   `json:"updatedDateTime"`
}

// MdmServerRelationships represents the relationships you include in the request, and those that you can operate on.
type MdmServerRelationships struct {
	Devices MdmServerRelationshipsDevices `json:"devices"`
}

// MdmServerRelationshipsDevices represents the data and links that describe the relationship between the resources
type MdmServerRelationshipsDevices struct {
	Data  []Data            `json:"data,omitempty"`
	Links RelationshipLinks `json:"links"`
	Meta  Meta              `json:"meta"`
}

// MdmServerDevicesLinkagesResponse represents the data and links that describe the relationship between the resources
type MdmServerDevicesLinkagesResponse struct {
	Data  []Data             `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// MdmServerCertificate represents an X.509 certificate to associate with a device management service.
type MdmServerCertificate struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// MdmServerCreateRequest represents a request to create a new device management service.
type MdmServerCreateRequest struct {
	Data MdmServerCreateRequestData `json:"data"`
}

// MdmServerCreateRequestData represents the data element of a create request.
type MdmServerCreateRequestData struct {
	Type       string                    `json:"type"`
	Attributes MdmServerCreateAttributes `json:"attributes"`
}

// MdmServerCreateAttributes represents attributes for creating a device management service.
type MdmServerCreateAttributes struct {
	ServerName          string               `json:"serverName"`
	ServerCertificate   MdmServerCertificate `json:"serverCertificate"`
	EnableMdmDisownFlag *bool                `json:"enableMdmDisownFlag,omitempty"`
}

// MdmServerUpdateRequest represents a request to update an existing device management service.
type MdmServerUpdateRequest struct {
	Data MdmServerUpdateRequestData `json:"data"`
}

// MdmServerUpdateRequestData represents the data element of an update request.
type MdmServerUpdateRequestData struct {
	Type       string                    `json:"type"`
	ID         string                    `json:"id"`
	Attributes MdmServerUpdateAttributes `json:"attributes"`
}

// MdmServerUpdateAttributes represents attributes for updating a device management service.
type MdmServerUpdateAttributes struct {
	ServerName             *string               `json:"serverName,omitempty"`
	ServerCertificate      *MdmServerCertificate `json:"serverCertificate,omitempty"`
	DefaultProductFamilies []string              `json:"defaultProductFamilies,omitempty"`
	EnableMdmDisownFlag    *bool                 `json:"enableMdmDisownFlag,omitempty"`
}

// GetDeviceManagementServices retrieves all MDM servers configured in the organization.
func (c *Client) GetDeviceManagementServices(ctx context.Context, queryParams url.Values) ([]MdmServer, error) {
	var allServers []MdmServer
	nextCursor := ""

	for {
		params := make(url.Values)
		maps.Copy(params, queryParams)
		params.Set("limit", strconv.Itoa(1000))
		params.Set("fields[mdmServers]", mdmServersFields)
		if nextCursor != "" {
			params.Set("cursor", nextCursor)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/mdmServers", c.baseURL), nil)
		if err != nil {
			return nil, err
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

			var response MdmServersResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allServers = append(allServers, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allServers, nil
}

// GetDeviceManagementServiceSerialNumbers retrieves all device serial numbers assigned to a specific MDM server identified by serverID.
func (c *Client) GetDeviceManagementServiceSerialNumbers(ctx context.Context, serverID string) ([]string, error) {
	var allSerialNumbers []string
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/mdmServers/%s/relationships/devices", c.baseURL, serverID), nil)
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

			var response MdmServerDevicesLinkagesResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			for _, device := range response.Data {
				if device.Type == "orgDevices" {
					allSerialNumbers = append(allSerialNumbers, device.ID)
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

	return allSerialNumbers, nil
}

// GetDeviceManagementService retrieves a single MDM server by ID.
func (c *Client) GetDeviceManagementService(ctx context.Context, id string, queryParams url.Values) (*MdmServer, error) {
	if queryParams == nil {
		queryParams = url.Values{}
	}
	queryParams.Set("fields[mdmServers]", mdmServersFields)

	baseURL := fmt.Sprintf("%s/v1/mdmServers/%s", c.baseURL, id)
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

	var response MdmServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// CreateDeviceManagementService creates a new device management service.
func (c *Client) CreateDeviceManagementService(ctx context.Context, request MdmServerCreateRequest) (*MdmServer, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v1/mdmServers", c.baseURL), bytes.NewReader(jsonData))
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

	var response MdmServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// UpdateDeviceManagementService updates an existing device management service.
func (c *Client) UpdateDeviceManagementService(ctx context.Context, request MdmServerUpdateRequest) (*MdmServer, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/v1/mdmServers/%s", c.baseURL, request.Data.ID), bytes.NewReader(jsonData))
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

	var response MdmServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// ClearDeviceManagementServiceDefaultFamilies removes all default product family assignments
// from an MDM server by sending defaultProductFamilies: null explicitly.
func (c *Client) ClearDeviceManagementServiceDefaultFamilies(ctx context.Context, id string) (*MdmServer, error) {
	payload := map[string]any{
		"data": map[string]any{
			"type": "mdmServers",
			"id":   id,
			"attributes": map[string]any{
				"defaultProductFamilies": []any{}, // explicit empty array, not null
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch,
		fmt.Sprintf("%s/v1/mdmServers/%s", c.baseURL, id), bytes.NewReader(jsonData))
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

	var response MdmServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// DeleteDeviceManagementService deletes a device management service by ID.
func (c *Client) DeleteDeviceManagementService(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete,
		fmt.Sprintf("%s/v1/mdmServers/%s", c.baseURL, id), nil)
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
