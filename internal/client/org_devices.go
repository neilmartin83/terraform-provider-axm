package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OrgDevicesResponse represents a response that contains a list of organization device resources.
type OrgDevicesResponse struct {
	Data  []OrgDevice        `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// OrgDeviceResponse represents a response that contains a single organization device resource.
type OrgDeviceResponse struct {
	Data  OrgDevice     `json:"data"`
	Links DocumentLinks `json:"links"`
}

// OrgDevice represents an organization device resource.
type OrgDevice struct {
	Type          string                 `json:"type"`
	ID            string                 `json:"id"`
	Attributes    DeviceAttribute        `json:"attributes"`
	Relationships OrgDeviceRelationships `json:"relationships"`
	Links         ResourceLinks          `json:"links"`
}

// DeviceAttribute represents attributes that describe an organization device resource.
type DeviceAttribute struct {
	SerialNumber        string   `json:"serialNumber"`
	AddedToOrgDateTime  string   `json:"addedToOrgDateTime"`
	UpdatedDateTime     string   `json:"updatedDateTime"`
	DeviceModel         string   `json:"deviceModel"`
	ProductFamily       string   `json:"productFamily"`
	ProductType         string   `json:"productType"`
	DeviceCapacity      string   `json:"deviceCapacity"`
	PartNumber          string   `json:"partNumber"`
	OrderNumber         string   `json:"orderNumber"`
	Color               string   `json:"color"`
	Status              string   `json:"status"`
	OrderDateTime       string   `json:"orderDateTime"`
	IMEI                []string `json:"imei"`
	MEID                []string `json:"meid"`
	EID                 string   `json:"eid"`
	PurchaseSourceID    string   `json:"purchaseSourceId"`
	PurchaseSourceType  string   `json:"purchaseSourceType"`
	WifiMacAddress      string   `json:"wifiMacAddress"`
	BluetoothMacAddress string   `json:"bluetoothMacAddress"`
}

// OrgDeviceRelationships represents the relationships you include in the request, and those that you can operate on.
type OrgDeviceRelationships struct {
	AssignedServer OrgDeviceRelationshipsAssignedServer `json:"assignedServer"`
}

// OrgDeviceRelationshipsAssignedServer represents the relationship representing a device and its assigned device management service.
type OrgDeviceRelationshipsAssignedServer struct {
	Links RelationshipLinks `json:"links"`
}

// OrgDeviceAssignedServerLinkageResponse represents the data and links that describe the relationship between the resources
type OrgDeviceAssignedServerLinkageResponse struct {
	Data  Data          `json:"data"`
	Links DocumentLinks `json:"links"`
}

// GetOrgDevices retrieves all organization devices from the API.
func (c *Client) GetOrgDevices(ctx context.Context, queryParams url.Values) ([]OrgDevice, error) {
	var allDevices []OrgDevice
	nextCursor := ""
	limit := 100

	for {
		baseURL := fmt.Sprintf("%s/v1/orgDevices", c.baseURL)
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

		var response OrgDevicesResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode response JSON: %w", err)
		}

		allDevices = append(allDevices, response.Data...)

		nextCursor = response.Meta.Paging.NextCursor
		if nextCursor == "" {
			break
		}
	}

	return allDevices, nil
}

// GetOrgDevice retrieves a single organization device by its ID.
func (c *Client) GetOrgDevice(ctx context.Context, id string, queryParams url.Values) (*OrgDevice, error) {
	baseURL := fmt.Sprintf("%s/v1/orgDevices/%s", c.baseURL, id)
	if len(queryParams) > 0 {
		baseURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

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

	var response OrgDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// GetOrgDeviceAssignedServerID retrieves the MDM server ID assigned to a specific device.
func (c *Client) GetOrgDeviceAssignedServerID(ctx context.Context, deviceID string) (*Data, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/v1/orgDevices/%s/relationships/assignedServer", c.baseURL, deviceID), nil)
	if err != nil {
		return nil, err
	}

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

	var response OrgDeviceAssignedServerLinkageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// GetOrgDeviceAssignedServer retrieves the MDM server assigned to a specific device.
func (c *Client) GetOrgDeviceAssignedServer(ctx context.Context, deviceID string, queryParams url.Values) (*MdmServer, error) {
	baseURL := fmt.Sprintf("%s/v1/orgDevices/%s/assignedServer", c.baseURL, deviceID)
	if len(queryParams) > 0 {
		baseURL += "?" + queryParams.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

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

	var response MdmServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
