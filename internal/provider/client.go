package axm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	auth    *AppleOAuthClient
	baseURL string
}

type OrgDevicesResponse struct {
	Data  []OrgDevice `json:"data"`
	Links Links       `json:"links"`
	Meta  Meta        `json:"meta"`
}

type OrgDeviceResponse struct {
	Data  OrgDevice `json:"data"`
	Links Links     `json:"links"`
}

type OrgDevice struct {
	Type          string              `json:"type"`
	ID            string              `json:"id"`
	Attributes    DeviceAttribute     `json:"attributes"`
	Relationships DeviceRelationships `json:"relationships"`
	Links         Links               `json:"links"`
}

type OrgDeviceAssignedServerResponse struct {
	Data  MdmServer `json:"data"`
	Links Links     `json:"links"`
}

type DeviceAttribute struct {
	SerialNumber       string   `json:"serialNumber"`
	AddedToOrgDateTime string   `json:"addedToOrgDateTime"`
	UpdatedDateTime    string   `json:"updatedDateTime"`
	DeviceModel        string   `json:"deviceModel"`
	ProductFamily      string   `json:"productFamily"`
	ProductType        string   `json:"productType"`
	DeviceCapacity     string   `json:"deviceCapacity"`
	PartNumber         string   `json:"partNumber"`
	OrderNumber        string   `json:"orderNumber"`
	Color              string   `json:"color"`
	Status             string   `json:"status"`
	OrderDateTime      string   `json:"orderDateTime"`
	IMEI               []string `json:"imei"`
	MEID               []string `json:"meid"`
	EID                string   `json:"eid"`
	PurchaseSourceID   string   `json:"purchaseSourceId"`
	PurchaseSourceType string   `json:"purchaseSourceType"`
}

type DeviceRelationships struct {
	AssignedServer AssignedServer `json:"assignedServer"`
}

type AssignedServer struct {
	Links Links `json:"links"`
}

type Links struct {
	Self    string `json:"self"`
	Next    string `json:"next,omitempty"`
	Related string `json:"related,omitempty"`
}

type Meta struct {
	Paging Paging `json:"paging"`
}

type Paging struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"nextCursor,omitempty"`
	Total      int    `json:"total,omitempty"`
}

type MdmServersResponse struct {
	Data     []MdmServer `json:"data"`
	Included []OrgDevice `json:"included,omitempty"`
	Links    Links       `json:"links"`
	Meta     Meta        `json:"meta"`
}

type MdmServer struct {
	Type          string             `json:"type"`
	ID            string             `json:"id"`
	Attributes    MdmServerAttribute `json:"attributes"`
	Relationships MdmRelationships   `json:"relationships"`
}

type MdmServerAttribute struct {
	ServerName      string `json:"serverName"`
	ServerType      string `json:"serverType"`
	CreatedDateTime string `json:"createdDateTime"`
	UpdatedDateTime string `json:"updatedDateTime"`
}

type MdmRelationships struct {
	Devices MdmDevicesRelationship `json:"devices"`
}

type MdmDevicesRelationship struct {
	Data  []MdmDeviceData `json:"data,omitempty"`
	Links Links           `json:"links"`
	Meta  Meta            `json:"meta,omitempty"`
}

type MdmDeviceData struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type DeviceRelationshipsResponse struct {
	Data  []MdmDeviceData `json:"data"`
	Links Links           `json:"links"`
	Meta  Meta            `json:"meta"`
}

func NewClient(baseURL, teamID, clientID, keyID, scope, p8Key string) (*Client, error) {
	config := &ClientConfig{
		TeamID:     teamID,
		ClientID:   clientID,
		KeyID:      keyID,
		Scope:      scope,
		PrivateKey: []byte(p8Key),
	}

	auth, err := NewAppleOAuthClient(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		auth:    auth,
		baseURL: baseURL,
	}, nil
}

func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := c.auth.Authenticate(ctx, req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return http.DefaultClient.Do(req)
}

func (c *Client) GetOrgDevices(ctx context.Context) ([]OrgDevice, error) {
	var allDevices []OrgDevice
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("%s/v1/orgDevices", c.baseURL), nil)
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
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("error %d: %s", resp.StatusCode, bodyBytes)
		}

		var response OrgDevicesResponse
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
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

func (c *Client) GetOrgDevice(ctx context.Context, id string) (*OrgDevice, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/v1/orgDevices/%s", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, bodyBytes)
	}

	var response OrgDeviceResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

func (c *Client) GetDeviceManagementServices(ctx context.Context) ([]MdmServer, error) {
	var allServers []MdmServer
	nextCursor := ""
	limit := 100

	for {
		req, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("%s/v1/mdmServers", c.baseURL), nil)
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
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("error %d: %s", resp.StatusCode, bodyBytes)
		}

		var response MdmServersResponse
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
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
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("error %d: %s", resp.StatusCode, bodyBytes)
		}

		var response DeviceRelationshipsResponse
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
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

func (c *Client) GetOrgDeviceAssignedServer(ctx context.Context, deviceID string) (*MdmServer, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/v1/orgDevices/%s/assignedServer", c.baseURL, deviceID), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error %d: %s", resp.StatusCode, bodyBytes)
	}

	var response OrgDeviceAssignedServerResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
