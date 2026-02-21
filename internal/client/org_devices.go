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
	SerialNumber            string   `json:"serialNumber"`
	AddedToOrgDateTime      string   `json:"addedToOrgDateTime"`
	ReleasedFromOrgDateTime string   `json:"releasedFromOrgDateTime,omitempty"`
	UpdatedDateTime         string   `json:"updatedDateTime"`
	DeviceModel             string   `json:"deviceModel"`
	ProductFamily           string   `json:"productFamily"`
	ProductType             string   `json:"productType"`
	DeviceCapacity          string   `json:"deviceCapacity"`
	PartNumber              string   `json:"partNumber,omitempty"`
	OrderNumber             string   `json:"orderNumber,omitempty"`
	Color                   string   `json:"color"`
	Status                  string   `json:"status"`
	OrderDateTime           string   `json:"orderDateTime,omitempty"`
	IMEI                    []string `json:"imei,omitempty"`
	MEID                    []string `json:"meid,omitempty"`
	EID                     string   `json:"eid,omitempty"`
	PurchaseSourceID        string   `json:"purchaseSourceId"`
	PurchaseSourceType      string   `json:"purchaseSourceType"`
	WifiMacAddress          string   `json:"wifiMacAddress,omitempty"`
	BluetoothMacAddress     string   `json:"bluetoothMacAddress,omitempty"`
	EthernetMacAddress      []string `json:"ethernetMacAddress,omitempty"`
}

// AppleCareCoverageResponse represents a response that contains AppleCare Coverage for an organization device.
type AppleCareCoverageResponse struct {
	Data  []AppleCareCoverage `json:"data"`
	Links DocumentLinks       `json:"links"`
	Meta  Meta                `json:"meta"`
}

// AppleCareCoverage represents AppleCare Coverage for an organization device.
type AppleCareCoverage struct {
	Attributes AppleCareCoverageAttribute `json:"attributes"`
	ID         string                     `json:"id"`
	Type       string                     `json:"type"`
}

// AppleCareCoverageAttribute represents AppleCare Coverage resources for an organization device.
type AppleCareCoverageAttribute struct {
	Status                 string `json:"status"`
	PaymentType            string `json:"paymentType"`
	Description            string `json:"description"`
	StartDateTime          string `json:"startDateTime"`
	EndDateTime            string `json:"endDateTime"`
	IsRenewable            bool   `json:"isRenewable"`
	IsCanceled             bool   `json:"isCanceled"`
	ContractCancelDateTime string `json:"contractCancelDateTime"`
	AgreementNumber        string `json:"agreementNumber"`
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
	limit := 1000

	for {
		baseURL := fmt.Sprintf("%s/v1/orgDevices", c.baseURL)
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

			var response OrgDevicesResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allDevices = append(allDevices, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

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

	var response OrgDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// GetOrgDeviceAssignedServerID retrieves the MDM server ID assigned to a specific device.
func (c *Client) GetOrgDeviceAssignedServerID(ctx context.Context, deviceID string) (*Data, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/v1/orgDevices/%s/relationships/assignedServer", c.baseURL, deviceID), nil)
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

// GetOrgDeviceAppleCareCoverage retrieves the AppleCare coverage details for a specific device.
func (c *Client) GetOrgDeviceAppleCareCoverage(ctx context.Context, deviceID string, queryParams url.Values) ([]AppleCareCoverage, error) {
	var allCoverages []AppleCareCoverage
	nextCursor := ""
	limit := 1000

	for {
		baseURL := fmt.Sprintf("%s/v1/orgDevices/%s/appleCareCoverage", c.baseURL, deviceID)
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

			var response AppleCareCoverageResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				return fmt.Errorf("failed to decode response JSON: %w", err)
			}

			allCoverages = append(allCoverages, response.Data...)
			nextCursor = response.Meta.Paging.NextCursor
			return nil
		}(); err != nil {
			return nil, err
		}

		if nextCursor == "" {
			break
		}
	}

	return allCoverages, nil
}
