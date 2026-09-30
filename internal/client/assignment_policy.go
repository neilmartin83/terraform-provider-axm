// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

const assignmentPolicyPageLimit = 10000
const assignmentPolicyBodyLimit = 16 << 20

var assignmentPolicyID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

// AssignmentPolicySnapshot is a complete, cross-checked observation, not a desired configuration.
type AssignmentPolicySnapshot struct {
	Servers []AssignmentPolicyServer
	Devices []AssignmentPolicyDevice
}

// AssignmentPolicyServer identifies a live assignment destination.
type AssignmentPolicyServer struct{ ID, Name, Type string }

// AssignmentPolicyDevice separates the API identifier from the hardware serial.
type AssignmentPolicyDevice struct{ ID, SerialNumber, Status, ServerID string }

type assignmentPolicyPage struct {
	Data  json.RawMessage `json:"data"`
	Links struct {
		Next string `json:"next"`
	} `json:"links"`
	Meta struct {
		Paging *struct {
			NextCursor string `json:"nextCursor"`
			Total      *int   `json:"total"`
		} `json:"paging"`
	} `json:"meta"`
}

// assignmentPolicyRequest retries GETs with the standard bounded backoff, but never
// retries writes. Both paths suppress request logging and redact API error bodies.
func (c *Client) assignmentPolicyRequest(ctx context.Context, method, path string, body []byte, out any) error {
	if c.httpClient == nil {
		return errors.New("assignment policy client is not configured")
	}
	base, err := url.Parse(c.baseURL)
	if err != nil || base.Host == "" || (base.Scheme != "https" && base.Scheme != "http") || base.User != nil {
		return errors.New("assignment policy API origin is invalid")
	}
	rel, err := url.Parse(path)
	if err != nil || rel.IsAbs() || rel.Host != "" || !strings.HasPrefix(rel.Path, "/v1/") {
		return errors.New("assignment policy API path is invalid")
	}
	var reader io.Reader
	if body != nil {
		reader = io.NopCloser(bytes.NewReader(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, base.ResolveReference(rel).String(), reader)
	if err != nil {
		return errors.New("assignment policy request could not be prepared")
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))
	}
	hc := *c.httpClient
	hc.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	var resp *http.Response
	if method == http.MethodGet {
		readClient := *c
		readClient.logger = nil
		readClient.httpClient = &hc
		resp, err = readClient.doRequest(ctx, req)
	} else {
		resp, err = hc.Do(req)
	}
	if err != nil {
		return errors.New("assignment policy request failed; outcome may be unknown; inspect before retrying")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && (method != http.MethodPost || resp.StatusCode != http.StatusCreated) {
		return errors.New("assignment policy API rejected the request; inspect before retrying")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, assignmentPolicyBodyLimit+1))
	if err != nil || len(data) > assignmentPolicyBodyLimit {
		return errors.New("assignment policy API response is unreadable or oversized")
	}
	if err := json.Unmarshal(data, out); err != nil {
		return errors.New("assignment policy API response is invalid")
	}
	return nil
}

func (c *Client) assignmentPolicyPages(ctx context.Context, path string, visit func(json.RawMessage) error) error {
	seen := map[string]bool{}
	cursor := ""
	count := 0
	var total *int
	for page := 0; page < assignmentPolicyPageLimit; page++ {
		u, err := url.Parse(path)
		if err != nil {
			return errors.New("assignment policy pagination path is invalid")
		}
		q := u.Query()
		q.Set("limit", "1000")
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		u.RawQuery = q.Encode()
		var response assignmentPolicyPage
		if err := c.assignmentPolicyRequest(ctx, http.MethodGet, u.String(), nil, &response); err != nil {
			return err
		}
		if response.Meta.Paging == nil || len(response.Data) == 0 || response.Data[0] != '[' {
			return errors.New("assignment policy pagination metadata or records are missing")
		}
		var records []json.RawMessage
		if json.Unmarshal(response.Data, &records) != nil {
			return errors.New("assignment policy page records are invalid")
		}
		if len(records) > 1000 {
			return errors.New("assignment policy page exceeds its requested limit")
		}
		if n := response.Meta.Paging.Total; n != nil {
			if *n < 0 || (total != nil && *n != *total) {
				return errors.New("assignment policy pagination total changed")
			}
			value := *n
			total = &value
		}
		for _, record := range records {
			if err := visit(record); err != nil {
				return err
			}
		}
		count += len(records)
		next := response.Meta.Paging.NextCursor
		if response.Links.Next != "" {
			link, err := url.Parse(response.Links.Next)
			base, _ := url.Parse(c.baseURL)
			if err != nil || link.User != nil || (link.Host != "" && link.Host != base.Host) || (link.Scheme != "" && link.Scheme != base.Scheme) || link.Path != u.Path || next == "" || link.Query().Get("cursor") != next {
				return errors.New("assignment policy next-page link is inconsistent")
			}
		}
		if next == "" {
			if total != nil && count != *total {
				return errors.New("assignment policy pagination is incomplete")
			}
			return nil
		}
		if len(records) == 0 || seen[next] || next == cursor {
			return errors.New("assignment policy pagination did not advance")
		}
		seen[next] = true
		cursor = next
	}
	return errors.New("assignment policy pagination exceeded its bound")
}

func assignmentPolicyDecodeDevice(raw json.RawMessage) (AssignmentPolicyDevice, error) {
	var d OrgDevice
	if json.Unmarshal(raw, &d) != nil || d.Type != "orgDevices" || !assignmentPolicyID.MatchString(d.ID) || strings.TrimSpace(d.Attributes.SerialNumber) == "" {
		return AssignmentPolicyDevice{}, errors.New("assignment policy device identity is invalid")
	}
	switch d.Attributes.Status {
	case "ASSIGNED", "UNASSIGNED":
	default:
		return AssignmentPolicyDevice{}, errors.New("assignment policy device status is missing or unsupported")
	}
	status := d.Attributes.Status
	if d.Attributes.ReleasedFromOrgDateTime != "" {
		if _, err := time.Parse(time.RFC3339Nano, d.Attributes.ReleasedFromOrgDateTime); err != nil {
			return AssignmentPolicyDevice{}, errors.New("assignment policy device release timestamp is invalid")
		}
		status = "RELEASED"
	}
	switch d.Attributes.MdmMigrationStatus {
	case "", "SUCCESS", "FAILED":
	default:
		return AssignmentPolicyDevice{}, errors.New("assignment policy device has an active or unsupported migration state")
	}
	return AssignmentPolicyDevice{ID: d.ID, SerialNumber: d.Attributes.SerialNumber, Status: status}, nil
}

// GetAssignmentPolicySnapshot reads all pages and cross-checks status against explicit server membership.
// Active or unknown migration states stop this complete snapshot, even outside the
// configured policy scope. Missing batch release fields are not release evidence.
func (c *Client) GetAssignmentPolicySnapshot(ctx context.Context) (*AssignmentPolicySnapshot, error) {
	result := &AssignmentPolicySnapshot{}
	servers := map[string]bool{}
	err := c.assignmentPolicyPages(ctx, "/v1/mdmServers?fields%5BmdmServers%5D=serverName%2CserverType%2Cstatus", func(raw json.RawMessage) error {
		var s MdmServer
		if json.Unmarshal(raw, &s) != nil || s.Type != "mdmServers" || !assignmentPolicyID.MatchString(s.ID) || strings.TrimSpace(s.Attributes.ServerName) == "" || servers[s.ID] {
			return errors.New("assignment policy server identity is invalid or duplicated")
		}
		switch s.Attributes.ServerType {
		case "MDM", "APPLE_CONFIGURATOR", "APPLE_MDM":
		default:
			return errors.New("assignment policy server type is missing or unsupported")
		}
		if s.Attributes.Status != nil && *s.Attributes.Status != MdmServerStatusActive {
			return errors.New("assignment policy server is not active")
		}
		servers[s.ID] = true
		result.Servers = append(result.Servers, AssignmentPolicyServer{ID: s.ID, Name: s.Attributes.ServerName, Type: s.Attributes.ServerType})
		return nil
	})
	if err != nil {
		return nil, err
	}
	devices := map[string]int{}
	serials := map[string]bool{}
	err = c.assignmentPolicyPages(ctx, "/v1/orgDevices", func(raw json.RawMessage) error {
		d, err := assignmentPolicyDecodeDevice(raw)
		if err != nil {
			return err
		}
		if _, exists := devices[d.ID]; exists || serials[d.SerialNumber] {
			return errors.New("assignment policy device identity is duplicated")
		}
		devices[d.ID] = len(result.Devices)
		serials[d.SerialNumber] = true
		result.Devices = append(result.Devices, d)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, server := range result.Servers {
		err = c.assignmentPolicyPages(ctx, "/v1/mdmServers/"+url.PathEscape(server.ID)+"/relationships/devices", func(raw json.RawMessage) error {
			var link Data
			if json.Unmarshal(raw, &link) != nil || link.Type != "orgDevices" || link.ID == "" {
				return errors.New("assignment policy device linkage is invalid")
			}
			i, exists := devices[link.ID]
			if !exists {
				return errors.New("assignment policy linkage references an unknown device")
			}
			d := &result.Devices[i]
			if d.ServerID != "" || d.Status != "ASSIGNED" {
				return errors.New("assignment policy memberships conflict with device status")
			}
			d.ServerID = server.ID
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	for _, d := range result.Devices {
		if d.Status == "ASSIGNED" && d.ServerID == "" {
			return nil, errors.New("assignment policy assigned device has no known server")
		}
	}
	sort.Slice(result.Servers, func(i, j int) bool { return result.Servers[i].ID < result.Servers[j].ID })
	sort.Slice(result.Devices, func(i, j int) bool { return result.Devices[i].ID < result.Devices[j].ID })
	return result, nil
}

// GetAssignmentPolicyDevice verifies the explicit device status and, for an assigned
// device, its server relationship. Release timestamps are available on this endpoint.
// Unassigned status is never inferred from a missing resource or an HTTP error.
// The full response avoids a sparse-field selector that excludes migration state;
// Apple documents that attribute but does not list it among selectable fields.
func (c *Client) GetAssignmentPolicyDevice(ctx context.Context, id string) (*AssignmentPolicyDevice, error) {
	if !assignmentPolicyID.MatchString(id) {
		return nil, errors.New("assignment policy device identifier is empty")
	}
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := c.assignmentPolicyRequest(ctx, http.MethodGet, "/v1/orgDevices/"+url.PathEscape(id), nil, &envelope); err != nil {
		return nil, err
	}
	d, err := assignmentPolicyDecodeDevice(envelope.Data)
	if err != nil {
		return nil, err
	}
	if d.ID != id {
		return nil, errors.New("assignment policy device identity changed")
	}
	var evidence struct {
		Attributes map[string]json.RawMessage `json:"attributes"`
	}
	if json.Unmarshal(envelope.Data, &evidence) != nil {
		return nil, errors.New("assignment policy release evidence is invalid")
	}
	release, present := evidence.Attributes["releasedFromOrgDateTime"]
	if !present {
		return nil, errors.New("assignment policy release evidence is missing from the device response")
	}
	if !bytes.Equal(bytes.TrimSpace(release), []byte("null")) {
		var timestamp string
		if json.Unmarshal(release, &timestamp) != nil || timestamp == "" {
			return nil, errors.New("assignment policy release evidence is invalid")
		}
	}
	if d.Status == "RELEASED" || d.Status == "UNASSIGNED" {
		return &d, nil
	}
	envelope.Data = nil
	if err := c.assignmentPolicyRequest(ctx, http.MethodGet, "/v1/orgDevices/"+url.PathEscape(id)+"/relationships/assignedServer", nil, &envelope); err != nil {
		return nil, err
	}
	var link Data
	if len(envelope.Data) == 0 || json.Unmarshal(envelope.Data, &link) != nil || link.Type != "mdmServers" || !assignmentPolicyID.MatchString(link.ID) || d.Status != "ASSIGNED" {
		return nil, errors.New("assignment policy relationship is missing or inconsistent")
	}
	d.ServerID = link.ID
	return &d, nil
}

// AssignPolicyDevice sends exactly one assignment request, then requires explicit completion and exact read-back.
// Callers must revalidate their saved transition before invoking this write. An error never authorizes a retry.
func (c *Client) AssignPolicyDevice(ctx context.Context, deviceID, destinationID string) error {
	if !assignmentPolicyID.MatchString(deviceID) || (destinationID != "" && !assignmentPolicyID.MatchString(destinationID)) {
		return errors.New("assignment policy device identifier is empty")
	}
	activityType := string(OrgDeviceActivityAssignDevices)
	relationships := OrgDeviceActivityCreateRequestRelationships{Devices: OrgDeviceActivityCreateRequestDataRelationships{Data: []Data{{Type: "orgDevices", ID: deviceID}}}}
	if destinationID == "" {
		activityType = string(OrgDeviceActivityUnassignDevices)
	} else {
		relationships.MdmServer = &OrgDeviceActivityCreateRequestDataRelationshipsMdmServer{Data: Data{Type: "mdmServers", ID: destinationID}}
	}
	body, err := json.Marshal(OrgDeviceActivityCreateRequest{Data: OrgDeviceActivityCreateRequestData{Type: "orgDeviceActivities", Attributes: OrgDeviceActivityCreateRequestAttributes{ActivityType: activityType}, Relationships: relationships}})
	if err != nil {
		return errors.New("assignment policy activity could not be prepared")
	}
	var response OrgDeviceActivityResponse
	if err := c.assignmentPolicyRequest(ctx, http.MethodPost, "/v1/orgDeviceActivities", body, &response); err != nil {
		return err
	}
	activityID := response.Data.ID
	if response.Data.Type != "orgDeviceActivities" || !assignmentPolicyID.MatchString(activityID) {
		return errors.New("assignment policy activity identity is missing")
	}
	for attempt := 0; attempt < 120; attempt++ {
		if response.Data.ID != activityID || response.Data.Type != "orgDeviceActivities" {
			return errors.New("assignment policy activity identity changed")
		}
		switch response.Data.Attributes.Status {
		case "COMPLETED":
			if response.Data.Attributes.SubStatus != "COMPLETED_WITH_SUCCESS" {
				return errors.New("assignment policy activity did not completely succeed; inspect before retrying")
			}
			d, err := c.GetAssignmentPolicyDevice(ctx, deviceID)
			if err != nil {
				return err
			}
			if d.Status == "RELEASED" || d.ServerID != destinationID || (destinationID == "" && d.Status != "UNASSIGNED") || (destinationID != "" && d.Status != "ASSIGNED") {
				return errors.New("assignment policy activity read-back did not match its destination")
			}
			return nil
		case "IN_PROGRESS":
			if err := waitWithContext(ctx, 5*time.Second); err != nil {
				return errors.New("assignment policy activity wait ended; outcome may be unknown; inspect before retrying")
			}
			response = OrgDeviceActivityResponse{}
			if err := c.assignmentPolicyRequest(ctx, http.MethodGet, "/v1/orgDeviceActivities/"+url.PathEscape(activityID), nil, &response); err != nil {
				return err
			}
		default:
			return errors.New("assignment policy activity failed or has an unknown status; inspect before retrying")
		}
	}
	return errors.New("assignment policy activity did not complete within its bound; inspect before retrying")
}

// GetAssignmentPolicyServer revalidates the exact destination immediately before a planned write.
func (c *Client) GetAssignmentPolicyServer(ctx context.Context, id string) (*AssignmentPolicyServer, error) {
	if !assignmentPolicyID.MatchString(id) {
		return nil, errors.New("assignment policy server identifier is invalid")
	}
	var response MdmServerResponse
	if err := c.assignmentPolicyRequest(ctx, http.MethodGet, "/v1/mdmServers/"+url.PathEscape(id)+"?fields%5BmdmServers%5D=serverName%2CserverType%2Cstatus", nil, &response); err != nil {
		return nil, err
	}
	s := response.Data
	if s.Type != "mdmServers" || s.ID != id || strings.TrimSpace(s.Attributes.ServerName) == "" || s.Attributes.ServerType != "MDM" || s.Attributes.Status == nil || *s.Attributes.Status != MdmServerStatusActive {
		return nil, errors.New("assignment policy destination is not an active MDM server with its expected identity")
	}
	return &AssignmentPolicyServer{ID: s.ID, Name: s.Attributes.ServerName, Type: s.Attributes.ServerType}, nil
}
