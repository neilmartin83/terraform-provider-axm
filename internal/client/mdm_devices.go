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

// MdmDeviceResponse represents a response that contains a list of Apple devices
// enrolled in a device management service.
type MdmDeviceResponse struct {
	Data  []MdmDevice       `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

// MdmDevice represents a device management service enrolled device resource.
type MdmDevice struct {
	Type       string              `json:"type"`
	ID         string              `json:"id"`
	Attributes MdmDeviceAttribute  `json:"attributes"`
	Links      ResourceLinks       `json:"links"`
}

// MdmDeviceAttribute represents attributes that describe a device management
// service enrolled device resource.
type MdmDeviceAttribute struct {
	DeviceName     string `json:"deviceName,omitempty"`
	EnrolledUserID string `json:"enrolledUserId,omitempty"`
	ProductFamily  string `json:"productFamily,omitempty"`
	SerialNumber   string `json:"serialNumber,omitempty"`
}

// MdmDeviceDetailResponse represents a response that contains the detailed
// information for an Apple device enrolled in a device management service.
type MdmDeviceDetailResponse struct {
	Data  MdmDeviceDetail `json:"data"`
	Links DocumentLinks   `json:"links"`
}

// MdmDeviceDetail represents the detailed information for a device management
// service enrolled device resource.
type MdmDeviceDetail struct {
	Type       string                   `json:"type"`
	ID         string                   `json:"id"`
	Attributes MdmDeviceDetailAttribute `json:"attributes"`
	Links      ResourceLinks            `json:"links"`
}

// MdmDeviceDetailAttribute represents detailed attributes that describe a device
// management service enrolled device resource.
type MdmDeviceDetailAttribute struct {
	BluetoothMacAddress  string   `json:"bluetoothMacAddress,omitempty"`
	DeviceEraseStatus    string   `json:"deviceEraseStatus,omitempty"`
	DeviceLockStatus     string   `json:"deviceLockStatus,omitempty"`
	DeviceModel          string   `json:"deviceModel,omitempty"`
	DeviceName           string   `json:"deviceName,omitempty"`
	EthernetMacAddress   string   `json:"ethernetMacAddress,omitempty"`
	IMEI                 []string `json:"imei,omitempty"`
	IsFileVaultEnabled   *bool    `json:"isFileVaultEnabled,omitempty"`
	IsFirewallEnabled    *bool    `json:"isFirewallEnabled,omitempty"`
	LastCheckInDateTime  string   `json:"lastCheckInDateTime,omitempty"`
	LostModeStatus       string   `json:"lostModeStatus,omitempty"`
	MEID                 []string `json:"meid,omitempty"`
	OsVersion            string   `json:"osVersion,omitempty"`
	Platform             string   `json:"platform,omitempty"`
	SerialNumber         string   `json:"serialNumber,omitempty"`
	StorageFreeCapacity  *int64   `json:"storageFreeCapacity,omitempty"`
	StorageTotalCapacity *int64   `json:"storageTotalCapacity,omitempty"`
	WifiMacAddress       string   `json:"wifiMacAddress,omitempty"`
}

// GetMdmDevices retrieves all Apple devices enrolled in a device management
// service.
func (c *Client) GetMdmDevices(ctx context.Context, queryParams url.Values) ([]MdmDevice, error) {
	var allDevices []MdmDevice
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/v1/mdmDevices", c.baseURL), nil)
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

			var response MdmDeviceResponse
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

// GetMdmDeviceDetail retrieves detailed information about a specific Apple
// device enrolled in a device management service.
func (c *Client) GetMdmDeviceDetail(ctx context.Context, id string, queryParams url.Values) (*MdmDeviceDetail, error) {
	baseURL := fmt.Sprintf("%s/v1/mdmDevices/%s/details", c.baseURL, id)
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

	var response MdmDeviceDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
