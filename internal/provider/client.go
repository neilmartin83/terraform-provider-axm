package axm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	auth    *AppleOAuthClient
	baseURL string
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	ID     string       `json:"id"`
	Status string       `json:"status"`
	Code   string       `json:"code"`
	Title  string       `json:"title"`
	Detail string       `json:"detail"`
	Source *ErrorSource `json:"source,omitempty"`
	Links  *ErrorLinks  `json:"links,omitempty"`
	Meta   interface{}  `json:"meta,omitempty"`
}

type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}

type ErrorLinks struct {
	About      string                `json:"about,omitempty"`
	Associated *ErrorLinksAssociated `json:"associated,omitempty"`
}

type ErrorLinksAssociated struct {
	Href string                 `json:"href"`
	Meta map[string]interface{} `json:"meta,omitempty"`
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

// NewClient creates a new Apple Business/School Manager API client.
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

// handleErrorResponse processes error responses from the API.
func (c *Client) handleErrorResponse(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return fmt.Errorf("failed to decode error response: %w", err)
	}

	if len(errResp.Errors) > 0 {
		err := errResp.Errors[0]
		return fmt.Errorf("%s: %s (code: %s, status: %s, id: %s)",
			err.Title, err.Detail, err.Code, err.Status, err.ID)
	}

	return fmt.Errorf("unknown error occurred with status %d", resp.StatusCode)
}

// doRequest performs an authenticated HTTP request.
func (c *Client) doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := c.auth.Authenticate(ctx, req); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return http.DefaultClient.Do(req)
}

// GetOrgDevices retrieves all organization devices from the API.
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

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var response OrgDeviceResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// GetDeviceManagementServices retrieves all MDM servers configured in the organization
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
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, c.handleErrorResponse(resp)
		}

		var response DeviceRelationshipsResponse
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

// GetOrgDeviceAssignedServer retrieves the MDM server assigned to a specific device.
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

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var response OrgDeviceAssignedServerResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
