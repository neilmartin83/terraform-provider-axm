// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// OrgDeviceActivityType is the type of organization device activity to create.
type OrgDeviceActivityType string

// Organization device activity types supported by the Apple School and Business Manager APIs.
const (
	OrgDeviceActivityAssignDevices                  OrgDeviceActivityType = "ASSIGN_DEVICES"
	OrgDeviceActivityUnassignDevices                OrgDeviceActivityType = "UNASSIGN_DEVICES"
	OrgDeviceActivityAssignWithMDMMigrationDeadline OrgDeviceActivityType = "ASSIGN_DEVICES_WITH_MDM_MIGRATION_DEADLINE"
	OrgDeviceActivityUpdateMDMMigrationDeadline     OrgDeviceActivityType = "UPDATE_MDM_MIGRATION_DEADLINE"
	OrgDeviceActivityCancelMDMMigration             OrgDeviceActivityType = "CANCEL_MDM_MIGRATION"
	OrgDeviceActivityReleaseDevices                 OrgDeviceActivityType = "RELEASE_DEVICES"
)

// ActivityOption configures an organization device activity create request.
type ActivityOption func(*orgDeviceActivityOptions)

// orgDeviceActivityOptions holds the optional fields for an organization device activity create request.
type orgDeviceActivityOptions struct {
	serverID string
	deadline string
}

// WithMdmServer sets the mdmServer relationship of an organization device activity create request.
func WithMdmServer(serverID string) ActivityOption {
	return func(o *orgDeviceActivityOptions) {
		o.serverID = serverID
	}
}

// WithMigrationDeadline sets activityTypeMetadata.mdmMigrationDeadlineDateTime on an organization
// device activity create request.
func WithMigrationDeadline(deadline string) ActivityOption {
	return func(o *orgDeviceActivityOptions) {
		o.deadline = deadline
	}
}

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

// OrgDeviceActivityCreateRequestAttributes represents attributes with values that you're changing as part of the create request.
type OrgDeviceActivityCreateRequestAttributes struct {
	ActivityType         string                     `json:"activityType"`
	ActivityTypeMetadata *OrgDeviceActivityMetadata `json:"activityTypeMetadata,omitempty"`
}

// OrgDeviceActivityMetadata represents additional metadata for an organization device activity, used by device management service migration activity types.
type OrgDeviceActivityMetadata struct {
	MdmMigrationDeadlineDateTime string `json:"mdmMigrationDeadlineDateTime,omitempty"`
}

// OrgDeviceActivityCreateRequestRelationships represents the relationships you include in the request, and those that you can operate on.
type OrgDeviceActivityCreateRequestRelationships struct {
	MdmServer *OrgDeviceActivityCreateRequestDataRelationshipsMdmServer `json:"mdmServer,omitempty"`
	Devices   OrgDeviceActivityCreateRequestDataRelationships           `json:"devices"`
}

// OrgDeviceActivityCreateRequestDataRelationshipsMdmServer represents the data that describe the relationship between the resources.
type OrgDeviceActivityCreateRequestDataRelationshipsMdmServer struct {
	Data Data `json:"data"`
}

// OrgDeviceActivityCreateRequestDataRelationships represents the relationships you include in the request, and those that you can operate on
type OrgDeviceActivityCreateRequestDataRelationships struct {
	Data []Data `json:"data"`
}

// CreateOrgDeviceActivity creates an organization device activity for the given activity type and devices.
// Activity types that assign devices to a device management service include WithMdmServer, and activity
// types that schedule or manage a device management service migration include WithMigrationDeadline.
func (c *Client) CreateOrgDeviceActivity(ctx context.Context, activityType OrgDeviceActivityType, deviceIDs []string, opts ...ActivityOption) (*OrgDeviceActivity, error) {
	var options orgDeviceActivityOptions
	for _, opt := range opts {
		opt(&options)
	}

	var metadata *OrgDeviceActivityMetadata
	if options.deadline != "" {
		metadata = &OrgDeviceActivityMetadata{MdmMigrationDeadlineDateTime: options.deadline}
	}

	return c.postOrgDeviceActivity(ctx, string(activityType), metadata, options.serverID, deviceIDs)
}

// postOrgDeviceActivity posts an organization device activity. An empty serverID omits the mdmServer relationship from the request.
func (c *Client) postOrgDeviceActivity(ctx context.Context, activityType string, metadata *OrgDeviceActivityMetadata, serverID string, deviceIDs []string) (*OrgDeviceActivity, error) {
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
				ActivityType:         activityType,
				ActivityTypeMetadata: metadata,
			},
			Relationships: OrgDeviceActivityCreateRequestRelationships{
				Devices: OrgDeviceActivityCreateRequestDataRelationships{
					Data: devices,
				},
			},
		},
	}

	if serverID != "" {
		request.Data.Relationships.MdmServer = &OrgDeviceActivityCreateRequestDataRelationshipsMdmServer{
			Data: Data{
				Type: "mdmServers",
				ID:   serverID,
			},
		}
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
