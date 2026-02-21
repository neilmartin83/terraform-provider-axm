package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OrgDeviceActivity represents the data structure that represents an organization device activity resource.
type OrgDeviceActivity struct {
	Type       string                      `json:"type"`
	ID         string                      `json:"id"`
	Attributes OrgDeviceActivityAttributes `json:"attributes"`
	Links      ResourceLinks               `json:"links"`
}

// OrgDeviceActivityAttributes represents attributes that describe an organization device activity resource.
type OrgDeviceActivityAttributes struct {
	Status            string `json:"status"`
	SubStatus         string `json:"subStatus"`
	CreatedDateTime   string `json:"createdDateTime"`
	CompletedDateTime string `json:"completedDateTime,omitempty"`
	DownloadURL       string `json:"downloadUrl,omitempty"`
}

// OrgDeviceActivityResponse represents a response that contains a single organization device activity resource.
type OrgDeviceActivityResponse struct {
	Data  OrgDeviceActivity `json:"data"`
	Links DocumentLinks     `json:"links"`
}

// OrgDeviceActivityCreateRequest represents the request body you use to update the device management service for a device.
type OrgDeviceActivityCreateRequest struct {
	Data OrgDeviceActivityCreateRequestData `json:"data"`
}

// OrgDeviceActivityCreateRequestData represents the data element of the request body.
type OrgDeviceActivityCreateRequestData struct {
	Type          string                                      `json:"type"`
	Attributes    OrgDeviceActivityCreateRequestAttributes    `json:"attributes"`
	Relationships OrgDeviceActivityCreateRequestRelationships `json:"relationships"`
}

// OrgDeviceActivityCreateRequestAttributes represents attributes with values that you’re changing as part of the create request.
type OrgDeviceActivityCreateRequestAttributes struct {
	ActivityType string `json:"activityType"`
}

// OrgDeviceActivityCreateRequestRelationships represents the relationships you include in the request, and those that you can operate on.
type OrgDeviceActivityCreateRequestRelationships struct {
	MdmServer OrgDeviceActivityCreateRequestDataRelationshipsMdmServer `json:"mdmServer"`
	Devices   OrgDeviceActivityCreateRequestDataRelationships          `json:"devices"`
}

// OrgDeviceActivityCreateRequestDataRelationshipsMdmServer represents the data that describe the relationship between the resources.
type OrgDeviceActivityCreateRequestDataRelationshipsMdmServer struct {
	Data Data `json:"data"`
}

// OrgDeviceActivityCreateRequestDataRelationships represents the relationships you include in the request, and those that you can operate on
type OrgDeviceActivityCreateRequestDataRelationships struct {
	Data []Data `json:"data"`
}

// AssignDevicesToMDMServer assigns or unassigns devices to/from an MDM server
// Returns the created activity. Caller is responsible for polling activity status if needed.
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/v1/orgDeviceActivities", c.baseURL), bytes.NewReader(jsonData))
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

	var response OrgDeviceActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}

// GetOrgDeviceActivity retrieves information about a specific organization device activity.
func (c *Client) GetOrgDeviceActivity(ctx context.Context, activityID string, queryParams url.Values) (*OrgDeviceActivity, error) {
	baseURL := fmt.Sprintf("%s/v1/orgDeviceActivities/%s", c.baseURL, activityID)
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

	var response OrgDeviceActivityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	return &response.Data, nil
}
