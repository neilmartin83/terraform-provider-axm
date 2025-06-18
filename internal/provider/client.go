package axm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

type Client struct {
	auth    *AppleOAuthClient
	baseURL string
	limiter *rate.Limiter
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
	Data  []OrgDevice        `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

type OrgDeviceResponse struct {
	Data  OrgDevice     `json:"data"`
	Links DocumentLinks `json:"links"`
}

type OrgDevice struct {
	Type          string                 `json:"type"`
	ID            string                 `json:"id"`
	Attributes    DeviceAttribute        `json:"attributes"`
	Relationships OrgDeviceRelationships `json:"relationships"`
	Links         ResourceLinks          `json:"links"`
}

type MdmServerResponse struct {
	Data     MdmServer     `json:"data"`
	Included []OrgDevice   `json:"included,omitempty"`
	Links    DocumentLinks `json:"links"`
}

type OrgDeviceAssignedServerLinkageResponse struct {
	Data  Data          `json:"data"`
	Links DocumentLinks `json:"links"`
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

type OrgDeviceRelationships struct {
	AssignedServer OrgDeviceRelationshipsAssignedServer `json:"assignedServer"`
}

type OrgDeviceRelationshipsAssignedServer struct {
	Links RelationshipLinks `json:"links"`
}

type PagedDocumentLinks struct {
	First string `json:"first,omitempty"`
	Next  string `json:"next,omitempty"`
	Self  string `json:"self"`
}

type RelationshipLinks struct {
	Include string `json:"include,omitempty"`
	Related string `json:"related,omitempty"`
	Self    string `json:"self,omitempty"`
}

type ResourceLinks struct {
	Self string `json:"self,omitempty"`
}

type DocumentLinks struct {
	Self string `json:"self"`
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
	Data     []MdmServer        `json:"data"`
	Included []OrgDevice        `json:"included,omitempty"`
	Links    PagedDocumentLinks `json:"links"`
	Meta     Meta               `json:"meta"`
}

type MdmServer struct {
	Type          string                 `json:"type"`
	ID            string                 `json:"id"`
	Attributes    MdmServerAttribute     `json:"attributes"`
	Relationships MdmServerRelationships `json:"relationships"`
}

type MdmServerAttribute struct {
	ServerName      string `json:"serverName"`
	ServerType      string `json:"serverType"`
	CreatedDateTime string `json:"createdDateTime"`
	UpdatedDateTime string `json:"updatedDateTime"`
}

type MdmServerRelationships struct {
	Devices MdmServerRelationshipsDevices `json:"devices"`
}

type MdmServerRelationshipsDevices struct {
	Data  []Data            `json:"data,omitempty"`
	Links RelationshipLinks `json:"links"`
	Meta  Meta              `json:"meta,omitempty"`
}

type Data struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type MdmServerDevicesLinkagesResponse struct {
	Data  []Data             `json:"data"`
	Links PagedDocumentLinks `json:"links"`
	Meta  Meta               `json:"meta"`
}

type OrgDeviceActivity struct {
	Type       string                      `json:"type"`
	ID         string                      `json:"id"`
	Attributes OrgDeviceActivityAttributes `json:"attributes"`
	Links      ResourceLinks               `json:"links"`
}

type OrgDeviceActivityAttributes struct {
	Status            string `json:"status"`
	SubStatus         string `json:"subStatus"`
	CreatedDateTime   string `json:"createdDateTime"`
	CompletedDateTime string `json:"completedDateTime,omitempty"`
	DownloadURL       string `json:"downloadUrl,omitempty"`
}

type OrgDeviceActivityResponse struct {
	Data  OrgDeviceActivity `json:"data"`
	Links DocumentLinks     `json:"links"`
}

type OrgDeviceActivityCreateRequest struct {
	Data OrgDeviceActivityCreateRequestData `json:"data"`
}

type OrgDeviceActivityCreateRequestData struct {
	Type          string                                      `json:"type"`
	Attributes    OrgDeviceActivityCreateRequestAttributes    `json:"attributes"`
	Relationships OrgDeviceActivityCreateRequestRelationships `json:"relationships"`
}

type OrgDeviceActivityCreateRequestAttributes struct {
	ActivityType string `json:"activityType"`
}

type OrgDeviceActivityCreateRequestRelationships struct {
	MdmServer OrgDeviceActivityCreateRequestDataRelationshipsMdmServer `json:"mdmServer"`
	Devices   OrgDeviceActivityCreateRequestDataRelationships          `json:"devices"`
}

type OrgDeviceActivityCreateRequestDataRelationshipsMdmServer struct {
	Data Data `json:"data"`
}

type OrgDeviceActivityCreateRequestDataRelationships struct {
	Data []Data `json:"data"`
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
		limiter: rate.NewLimiter(rate.Every(3500*time.Millisecond), 1),
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
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

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

// AssignDevicesToMDMServer assigns or unassigns devices to/from an MDM server and monitors the operation until completion
func (c *Client) AssignDevicesToMDMServer(ctx context.Context, serverID string, deviceIDs []string, assign bool) (*OrgDeviceActivity, error) {
	activityType := "ASSIGN_DEVICES"
	if !assign {
		activityType = "UNASSIGN_DEVICES"
	}

	devices := make([]Data, len(deviceIDs))
	for i, id := range deviceIDs {
		devices[i] = Data{
			Type: "orgDevices",
			ID:   id,
		}
	}

	request := OrgDeviceActivityCreateRequest{
		Data: OrgDeviceActivityCreateRequestData{
			Type: "orgDeviceActivities",
			Attributes: OrgDeviceActivityCreateRequestAttributes{
				ActivityType: activityType,
			},
			Relationships: OrgDeviceActivityCreateRequestRelationships{
				MdmServer: OrgDeviceActivityCreateRequestDataRelationshipsMdmServer{
					Data: Data{
						Type: "mdmServers",
						ID:   serverID,
					},
				},
				Devices: OrgDeviceActivityCreateRequestDataRelationships{
					Data: devices,
				},
			},
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/v1/orgDeviceActivities", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
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

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, c.handleErrorResponse(resp)
	}

	var response OrgDeviceActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	if response.Data.Attributes.Status == "COMPLETED" {
		return &response.Data, nil
	}

	switch response.Data.Attributes.Status {
	case "FAILED":
		return nil, fmt.Errorf("activity failed with sub-status: %s", response.Data.Attributes.SubStatus)
	case "STOPPED":
		return nil, fmt.Errorf("activity stopped with sub-status: %s", response.Data.Attributes.SubStatus)
	}

	maxAttempts := 30
	retryInterval := time.Second * 5

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		status, err := c.GetOrgDeviceActivity(ctx, response.Data.ID)
		if err != nil {
			return nil, fmt.Errorf("error checking activity status: %w", err)
		}

		switch status.Attributes.Status {
		case "COMPLETED":
			return status, nil
		case "FAILED":
			return nil, fmt.Errorf("activity failed with sub-status: %s", status.Attributes.SubStatus)
		case "STOPPED":
			return nil, fmt.Errorf("activity stopped with sub-status: %s", status.Attributes.SubStatus)
		}

		if attempt == maxAttempts {
			return nil, fmt.Errorf("timed out waiting for activity to complete after %d attempts", maxAttempts)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(retryInterval):
			continue
		}
	}

	return nil, fmt.Errorf("unexpected error monitoring activity status")
}

// GetOrgDeviceActivity retrieves information about a specific organization device activity
func (c *Client) GetOrgDeviceActivity(ctx context.Context, activityID string) (*OrgDeviceActivity, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/v1/orgDeviceActivities/%s", c.baseURL, activityID), nil)
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

	var response OrgDeviceActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
