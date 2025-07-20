package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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
	ServerName      string `json:"serverName"`
	ServerType      string `json:"serverType"`
	CreatedDateTime string `json:"createdDateTime"`
	UpdatedDateTime string `json:"updatedDateTime"`
}

// MdmServerRelationships represents the relationships you include in the request, and those that you can operate on.
type MdmServerRelationships struct {
	Devices MdmServerRelationshipsDevices `json:"devices"`
}

// MdmServerRelationshipsDevices represents the data and links that describe the relationship between the resources
type MdmServerRelationshipsDevices struct {
	Data  []Data            `json:"data,omitempty"`
	Links RelationshipLinks `json:"links"`
	Meta  Meta              `json:"meta,omitempty"`
}

// MdmServerDevicesLinkagesResponse represents the data and links that describe the relationship between the resources
type MdmServerDevicesLinkagesResponse struct {
	Data  []Data             `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// GetDeviceManagementServices retrieves all MDM servers configured in the organization
func (c *Client) GetDeviceManagementServices(ctx context.Context, queryParams url.Values) ([]MdmServer, error) {
	var allServers []MdmServer
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/mdmServers", c.baseURL)
		if len(queryParams) > 0 {
			baseURL += "?" + queryParams.Encode()
		}

		req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("limit", fmt.Sprintf("%d", limit))
		if nextCursor != "" {
			q.Add("cursor", nextCursor)
		}
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Accept", "application/json")

		resp, err := c.doRequest(ctx, req)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Printf("warning: failed to close response body: %v\n", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return nil, c.handleErrorResponse(resp)
		}

		var response MdmServersResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response JSON: %w", err)
		}

		allServers = append(allServers, response.Data...)

		nextCursor = response.Meta.Paging.NextCursor
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
		req, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("%s/v1/mdmServers/%s/relationships/devices", c.baseURL, serverID), nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("limit", fmt.Sprintf("%d", limit))
		if nextCursor != "" {
			q.Add("cursor", nextCursor)
		}
		req.URL.RawQuery = q.Encode()

		req.Header.Set("Accept", "application/json")

		resp, err := c.doRequest(ctx, req)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Printf("warning: failed to close response body: %v\n", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return nil, c.handleErrorResponse(resp)
		}

		var response MdmServerDevicesLinkagesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response JSON: %w", err)
		}

		for _, device := range response.Data {
			if device.Type == "orgDevices" {
				allSerialNumbers = append(allSerialNumbers, device.ID)
			}
		}

		nextCursor = response.Meta.Paging.NextCursor
		if nextCursor == "" {
			break
		}
	}

	return allSerialNumbers, nil
}
