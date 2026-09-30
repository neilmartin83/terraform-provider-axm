// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func policyPage(data string) string { return `{"data":` + data + `,"meta":{"paging":{}},"links":{}}` }
func policyDevice(id, serial, status string) string {
	return `{"type":"orgDevices","id":"` + id + `","attributes":{"serialNumber":"` + serial + `","status":"` + status + `","releasedFromOrgDateTime":null}}`
}
func policyServer() string {
	return `{"type":"mdmServers","id":"server-1","attributes":{"serverName":"Test MDM","serverType":"MDM","status":"ACTIVE"}}`
}
func policyInventory(t *testing.T, overrides map[string]string) (*Client, *int) {
	t.Helper()
	routes := map[string]string{
		"/v1/mdmServers": policyPage("[" + policyServer() + "]"),
		"/v1/orgDevices": policyPage("[" + policyDevice("device-1", "serial-1", "ASSIGNED") + "," + policyDevice("device-2", "serial-2", "UNASSIGNED") + "," + strings.ReplaceAll(policyDevice("device-3", "serial-3", "UNASSIGNED"), `"releasedFromOrgDateTime":null`, `"releasedFromOrgDateTime":"2026-01-01T00:00:00Z"`) + "]"),
		"/v1/mdmServers/server-1/relationships/devices": policyPage(`[{"type":"orgDevices","id":"device-1"}]`),
	}
	for path, response := range overrides {
		routes[path] = response
	}
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet {
			t.Error("unexpected write")
		}
		if response, ok := routes[r.URL.Path]; ok {
			_, _ = w.Write([]byte(response))
		} else {
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return newTestClient(t, server), &calls
}

func TestAssignmentPolicySnapshot(t *testing.T) {
	c, _ := policyInventory(t, nil)
	snapshot, err := c.GetAssignmentPolicySnapshot(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Devices) != 3 || len(snapshot.Servers) != 1 {
		t.Fatal("incorrect inventory size")
	}
	if d := snapshot.Devices[0]; d.ID != "device-1" || d.SerialNumber != "serial-1" || d.ServerID != "server-1" {
		t.Fatal("incorrect identity mapping")
	}
	if snapshot.Devices[1].Status != "UNASSIGNED" || snapshot.Devices[1].ServerID != "" {
		t.Fatal("unassigned status lost")
	}
	if snapshot.Devices[2].Status != "RELEASED" {
		t.Fatal("released status lost")
	}
}

func TestAssignmentPolicySnapshotRejectsInconsistentInventory(t *testing.T) {
	cases := map[string]map[string]string{
		"missing status":         {"/v1/orgDevices": policyPage("[" + policyDevice("device-1", "serial-1", "") + "]")},
		"duplicate ID":           {"/v1/orgDevices": policyPage("[" + policyDevice("device-1", "serial-1", "ASSIGNED") + "," + policyDevice("device-1", "serial-2", "ASSIGNED") + "]")},
		"duplicate serial":       {"/v1/orgDevices": policyPage("[" + policyDevice("device-1", "serial-1", "ASSIGNED") + "," + policyDevice("device-2", "serial-1", "UNASSIGNED") + "]")},
		"unknown linkage":        {"/v1/mdmServers/server-1/relationships/devices": policyPage(`[{"type":"orgDevices","id":"device-9"}]`)},
		"unassigned linkage":     {"/v1/mdmServers/server-1/relationships/devices": policyPage(`[{"type":"orgDevices","id":"device-2"}]`)},
		"missing membership":     {"/v1/mdmServers/server-1/relationships/devices": policyPage(`[]`)},
		"duplicate membership":   {"/v1/mdmServers/server-1/relationships/devices": policyPage(`[{"type":"orgDevices","id":"device-1"},{"type":"orgDevices","id":"device-1"}]`)},
		"missing data":           {"/v1/orgDevices": `{"meta":{"paging":{}}}`},
		"missing paging":         {"/v1/orgDevices": `{"data":[]}`},
		"null data":              {"/v1/orgDevices": policyPage("null")},
		"incomplete total":       {"/v1/orgDevices": `{"data":[],"meta":{"paging":{"total":10}}}`},
		"unsafe id":              {"/v1/orgDevices": policyPage("[" + policyDevice("..", "serial-1", "ASSIGNED") + "]")},
		"unknown server type":    {"/v1/mdmServers": policyPage(`[{"type":"mdmServers","id":"server-1","attributes":{"serverName":"Test","serverType":"UNKNOWN"}}]`)},
		"deleted server":         {"/v1/mdmServers": policyPage(strings.ReplaceAll("["+policyServer()+"]", "ACTIVE", "DELETED"))},
		"external next":          {"/v1/orgDevices": `{"data":[],"meta":{"paging":{"nextCursor":"abc"}},"links":{"next":"https://example.invalid/v1/orgDevices?cursor=abc"}}`},
		"protocol relative next": {"/v1/orgDevices": `{"data":[],"meta":{"paging":{"nextCursor":"abc"}},"links":{"next":"//example.invalid/v1/orgDevices?cursor=abc"}}`},
	}
	for name, routes := range cases {
		t.Run(name, func(t *testing.T) {
			c, _ := policyInventory(t, routes)
			_, err := c.GetAssignmentPolicySnapshot(context.Background())
			if err == nil {
				t.Fatal("invalid inventory accepted")
			}
			if strings.Contains(err.Error(), "serial-") || strings.Contains(err.Error(), "device-") {
				t.Fatal("identifier leaked")
			}
		})
	}
}

func TestAssignmentPolicyPagination(t *testing.T) {
	for _, cycle := range []bool{false, true} {
		t.Run(map[bool]string{false: "multiple pages", true: "cycle"}[cycle], func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Query().Get("limit") != "1000" {
					t.Error("missing bound")
				}
				if calls == 1 || cycle {
					_, _ = w.Write([]byte(`{"data":[{}],"meta":{"paging":{"nextCursor":"next"}},"links":{"next":"/v1/orgDevices?cursor=next"}}`))
				} else {
					_, _ = w.Write([]byte(policyPage(`[]`)))
				}
			}))
			defer server.Close()
			c := newTestClient(t, server)
			count := 0
			err := c.assignmentPolicyPages(context.Background(), "/v1/orgDevices", func(json.RawMessage) error { count++; return nil })
			if cycle && err == nil {
				t.Fatal("cycle accepted")
			}
			if !cycle && (err != nil || count != 1 || calls != 2) {
				t.Fatal("pagination failed")
			}
			if calls > 2 {
				t.Fatal("cycle not bounded")
			}
		})
	}
}

func TestAssignmentPolicyDeviceExplicitRelationship(t *testing.T) {
	for _, test := range []struct {
		name, status, relationship string
		wantErr                    bool
	}{
		{"unassigned", "UNASSIGNED", `{"data":null}`, false},
		{"assigned", "ASSIGNED", `{"data":{"type":"mdmServers","id":"server-1"}}`, false},
		{"omitted", "ASSIGNED", `{}`, true},
		{"empty object", "ASSIGNED", `{"data":{}}`, true},
		{"contradictory", "ASSIGNED", `{"data":null}`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "assignedServer") {
					_, _ = w.Write([]byte(test.relationship))
				} else {
					_, _ = w.Write([]byte(`{"data":` + policyDevice("device-1", "serial-1", test.status) + `}`))
				}
			}))
			defer server.Close()
			_, err := newTestClient(t, server).GetAssignmentPolicyDevice(context.Background(), "device-1")
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func TestAssignPolicyDeviceRequestAndReadback(t *testing.T) {
	for _, destination := range []string{"server-1", ""} {
		t.Run(map[bool]string{true: "assign", false: "unassign"}[destination != ""], func(t *testing.T) {
			posts := 0
			gets := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					posts++
					var body OrgDeviceActivityCreateRequest
					if json.NewDecoder(r.Body).Decode(&body) != nil {
						t.Fatal("bad request")
					}
					expected := string(OrgDeviceActivityAssignDevices)
					if destination == "" {
						expected = string(OrgDeviceActivityUnassignDevices)
					}
					if body.Data.Attributes.ActivityType != expected || body.Data.Attributes.ActivityTypeMetadata != nil || len(body.Data.Relationships.Devices.Data) != 1 || body.Data.Relationships.Devices.Data[0].ID != "device-1" {
						t.Error("wrong activity")
					}
					if destination == "" && body.Data.Relationships.MdmServer != nil {
						t.Error("unassign included server")
					}
					if destination != "" && (body.Data.Relationships.MdmServer == nil || body.Data.Relationships.MdmServer.Data.ID != destination) {
						t.Error("wrong server")
					}
					_, _ = w.Write([]byte(`{"data":{"type":"orgDeviceActivities","id":"activity-1","attributes":{"status":"COMPLETED","subStatus":"COMPLETED_WITH_SUCCESS"}}}`))
					return
				}
				gets++
				if strings.HasSuffix(r.URL.Path, "assignedServer") {
					if destination == "" {
						_, _ = w.Write([]byte(`{"data":null}`))
					} else {
						_, _ = w.Write([]byte(`{"data":{"type":"mdmServers","id":"server-1"}}`))
					}
					return
				}
				status := "ASSIGNED"
				if destination == "" {
					status = "UNASSIGNED"
				}
				_, _ = w.Write([]byte(`{"data":` + policyDevice("device-1", "serial-1", status) + `}`))
			}))
			defer server.Close()
			if err := newTestClient(t, server).AssignPolicyDevice(context.Background(), "device-1", destination); err != nil {
				t.Fatal(err)
			}
			expectedReads := 2
			if destination == "" {
				expectedReads = 1
			}
			if posts != 1 || gets != expectedReads {
				t.Fatal("wrong request count")
			}
		})
	}
}

func TestAssignPolicyDeviceFailuresNeverRetry(t *testing.T) {
	for _, test := range []struct {
		name string
		code int
		body string
	}{
		{"rate limit", 429, `secret-serial`}, {"gateway", 502, `secret-serial`}, {"redirect", 307, `secret-serial`},
		{"partial", 200, `{"data":{"type":"orgDeviceActivities","id":"activity-1","attributes":{"status":"COMPLETED","subStatus":"COMPLETED_WITH_ERRORS"}}}`},
		{"unknown status", 200, `{"data":{"type":"orgDeviceActivities","id":"activity-1","attributes":{"status":"secret-serial"}}}`},
		{"malformed", 200, `secret-serial`},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("Location", "/v1/replayed")
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(test.code)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()
			err := newTestClient(t, server).AssignPolicyDevice(context.Background(), "device-1", "server-1")
			if err == nil || strings.Contains(err.Error(), "secret-serial") {
				t.Fatal("unsafe error")
			}
			if calls != 1 {
				t.Fatal("write replayed")
			}
		})
	}
}

type assignmentPolicyTransport func(*http.Request) (*http.Response, error)

func (f assignmentPolicyTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestAssignmentPolicyNetworkErrorIsRedacted(t *testing.T) {
	calls := 0
	c := &Client{baseURL: "https://example.invalid", httpClient: &http.Client{Transport: assignmentPolicyTransport(func(r *http.Request) (*http.Response, error) { calls++; return nil, errors.New("secret-serial") })}}
	err := c.AssignPolicyDevice(context.Background(), "device-1", "server-1")
	if err == nil || strings.Contains(err.Error(), "secret-serial") || calls != 1 {
		t.Fatal("unsafe transport behavior")
	}
}

func TestAssignmentPolicyWriteReadbackMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"data":{"type":"orgDeviceActivities","id":"activity-1","attributes":{"status":"COMPLETED","subStatus":"COMPLETED_WITH_SUCCESS"}}}`))
		case strings.HasSuffix(r.URL.Path, "assignedServer"):
			_, _ = w.Write([]byte(`{"data":{"type":"mdmServers","id":"server-other"}}`))
		default:
			_, _ = w.Write([]byte(`{"data":` + policyDevice("device-1", "serial-1", "ASSIGNED") + `}`))
		}
	}))
	defer server.Close()
	if err := newTestClient(t, server).AssignPolicyDevice(context.Background(), "device-1", "server-1"); err == nil {
		t.Fatal("mismatched outcome accepted")
	}
}

type assignmentPolicyForbiddenLogger struct{ t *testing.T }

func (l assignmentPolicyForbiddenLogger) LogRequest(context.Context, string, string, []byte) {
	l.t.Error("request logged")
}
func (l assignmentPolicyForbiddenLogger) LogResponse(context.Context, int, http.Header, []byte) {
	l.t.Error("response logged")
}
func (l assignmentPolicyForbiddenLogger) LogAuth(context.Context, string, map[string]any) {
	l.t.Error("auth event logged")
}

func TestAssignmentPolicyDoesNotLogOrMakeReplayableRequests(t *testing.T) {
	calls := 0
	c := &Client{baseURL: "https://example.invalid", logger: assignmentPolicyForbiddenLogger{t}, httpClient: &http.Client{Transport: assignmentPolicyTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.ContentLength <= 0 {
			t.Error("request length is not explicit")
		}
		if r.GetBody != nil {
			t.Error("request body is replayable")
		}
		if r.Header.Get("Idempotency-Key") != "" {
			t.Error("unexpected idempotency marker")
		}
		return nil, errors.New("secret-device")
	})}}
	if err := c.AssignPolicyDevice(context.Background(), "device-1", "server-1"); err == nil || calls != 1 {
		t.Fatal("wrong request behavior")
	}
}

func TestAssignmentPolicyServerValidation(t *testing.T) {
	for _, test := range []struct {
		name, body string
		wantErr    bool
	}{
		{"valid", `{"data":` + policyServer() + `}`, false},
		{"different ID", `{"data":` + strings.ReplaceAll(policyServer(), "server-1", "server-2") + `}`, true},
		{"nonMDM", `{"data":` + strings.ReplaceAll(policyServer(), `"MDM"`, `"APPLE_CONFIGURATOR"`) + `}`, true},
		{"inactive", `{"data":` + strings.ReplaceAll(policyServer(), "ACTIVE", "INACTIVE") + `}`, true},
		{"missing status", `{"data":` + strings.ReplaceAll(policyServer(), `,"status":"ACTIVE"`, "") + `}`, true},
		{"null status", `{"data":` + strings.ReplaceAll(policyServer(), `"status":"ACTIVE"`, `"status":null`) + `}`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(test.body)) }))
			defer server.Close()
			_, err := newTestClient(t, server).GetAssignmentPolicyServer(context.Background(), "server-1")
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected result: %v", err)
			}
		})
	}
}

func TestAssignPolicyDevicePollCannotReuseMissingIdentity(t *testing.T) {
	posts := 0
	polls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts++
			_, _ = w.Write([]byte(`{"data":{"id":"activity-1","type":"orgDeviceActivities","attributes":{"status":"IN_PROGRESS"}}}`))
		} else {
			polls++
			_, _ = w.Write([]byte(`{"data":{"attributes":{"status":"COMPLETED","subStatus":"COMPLETED_WITH_SUCCESS"}}}`))
		}
	}))
	defer server.Close()
	err := newTestClient(t, server).AssignPolicyDevice(context.Background(), "device-1", "server-1")
	if err == nil || posts != 1 || polls != 1 {
		t.Fatal("poll reused stale identity")
	}
}

func TestAssignPolicyDeviceCancellationNeverRepeatsWrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts++
		_, _ = w.Write([]byte(`{"data":{"id":"activity-1","type":"orgDeviceActivities","attributes":{"status":"IN_PROGRESS"}}}`))
		cancel()
	}))
	defer server.Close()
	if err := newTestClient(t, server).AssignPolicyDevice(ctx, "device-1", "server-1"); err == nil || posts != 1 {
		t.Fatal("cancelled operation was accepted or retried")
	}
}

func TestAssignmentPolicyReadRetriesRateLimitWithoutLogging(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet {
			t.Error("read sent a write")
		}
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("secret-device"))
			return
		}
		_, _ = w.Write([]byte(`{"data":` + policyServer() + `}`))
	}))
	defer server.Close()
	c := newTestClient(t, server)
	c.logger = assignmentPolicyForbiddenLogger{t}
	result, err := c.GetAssignmentPolicyServer(context.Background(), "server-1")
	if err != nil || result == nil || result.ID != "server-1" || calls != 2 {
		t.Fatalf("bounded read retry failed: %v", err)
	}
}

func TestAssignmentPolicyReadRetryIsBoundedAndRedacted(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("secret-device"))
	}))
	defer server.Close()
	c := newTestClient(t, server)
	c.logger = assignmentPolicyForbiddenLogger{t}
	_, err := c.GetAssignmentPolicyServer(context.Background(), "server-1")
	if err == nil || strings.Contains(err.Error(), "secret-device") || calls != maxRetries {
		t.Fatalf("read retry was unbounded or leaked data: %v", err)
	}
}

func TestAssignmentPolicyReadDoesNotFollowRedirect(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Location", "/v1/leaked")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	defer server.Close()
	_, err := newTestClient(t, server).GetAssignmentPolicyServer(context.Background(), "server-1")
	if err == nil || calls != 1 {
		t.Fatal("read followed a redirect")
	}
}

func TestAssignmentPolicyExplicitUnassignedSkipsMissingRelationship(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if strings.HasSuffix(r.URL.Path, "assignedServer") {
			t.Error("unassigned device requested nonexistent relationship")
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":` + policyDevice("device-1", "serial-1", "UNASSIGNED") + `}`))
	}))
	defer server.Close()
	device, err := newTestClient(t, server).GetAssignmentPolicyDevice(context.Background(), "device-1")
	if err != nil || device == nil || device.Status != "UNASSIGNED" || device.ServerID != "" || calls != 1 {
		t.Fatalf("explicit unassigned status not preserved: %v", err)
	}
}

func TestAssignmentPolicyNotFoundNeverMeansUnassigned(t *testing.T) {
	for _, missingRelationship := range []bool{false, true} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !missingRelationship || strings.HasSuffix(r.URL.Path, "assignedServer") {
				http.NotFound(w, r)
				return
			}
			_, _ = w.Write([]byte(`{"data":` + policyDevice("device-1", "serial-1", "ASSIGNED") + `}`))
		}))
		device, err := newTestClient(t, server).GetAssignmentPolicyDevice(context.Background(), "device-1")
		server.Close()
		if err == nil || device != nil {
			t.Fatal("HTTP404 became an unassigned device")
		}
	}
}

func TestAssignmentPolicyMigrationStates(t *testing.T) {
	for _, test := range []struct {
		status    string
		wantError bool
	}{
		{"", false}, {"SUCCESS", false}, {"FAILED", false}, {"REQUESTED", true}, {"STARTED", true}, {"NONE", true}, {"UNKNOWN", true},
	} {
		t.Run(test.status, func(t *testing.T) {
			raw := strings.ReplaceAll(policyDevice("device-1", "serial-1", "ASSIGNED"), `"status":"ASSIGNED"`, `"status":"ASSIGNED","mdmMigrationStatus":"`+test.status+`"`)
			_, err := assignmentPolicyDecodeDevice(json.RawMessage(raw))
			if (err != nil) != test.wantError {
				t.Fatalf("migration-state parsing mismatch: %v", err)
			}
		})
	}
}

func TestAssignmentPolicyReleaseTimestampIsAuthoritative(t *testing.T) {
	for _, status := range []string{"ASSIGNED", "UNASSIGNED"} {
		raw := strings.ReplaceAll(policyDevice("device-1", "serial-1", status), `"releasedFromOrgDateTime":null`, `"releasedFromOrgDateTime":"2026-01-01T00:00:00Z"`)
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; _, _ = w.Write([]byte(`{"data":` + raw + `}`)) }))
		device, err := newTestClient(t, server).GetAssignmentPolicyDevice(context.Background(), "device-1")
		server.Close()
		if err != nil || device == nil || device.Status != "RELEASED" || calls != 1 {
			t.Fatalf("release timestamp not preserved: %v", err)
		}
	}
	raw := strings.ReplaceAll(policyDevice("device-1", "serial-1", "UNASSIGNED"), `"releasedFromOrgDateTime":null`, `"releasedFromOrgDateTime":"invalid"`)
	if _, err := assignmentPolicyDecodeDevice(json.RawMessage(raw)); err == nil {
		t.Fatal("invalid release timestamp accepted")
	}
	if _, err := assignmentPolicyDecodeDevice(json.RawMessage(policyDevice("device-1", "serial-1", "RELEASED"))); err == nil {
		t.Fatal("undocumented API status accepted")
	}
}

func TestAssignmentPolicySingleDeviceRequiresExplicitReleaseEvidence(t *testing.T) {
	for _, test := range []struct {
		name, replacement string
		wantError         bool
	}{
		{"null", `,"releasedFromOrgDateTime":null`, false},
		{"missing", "", true},
		{"empty", `,"releasedFromOrgDateTime":""`, true},
		{"invalid", `,"releasedFromOrgDateTime":"invalid"`, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			raw := strings.ReplaceAll(policyDevice("device-1", "serial-1", "UNASSIGNED"), `,"releasedFromOrgDateTime":null`, test.replacement)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Has("fields[orgDevices]") {
					t.Error("sparse selector may exclude migration state")
				}
				_, _ = w.Write([]byte(`{"data":` + raw + `}`))
			}))
			defer server.Close()
			_, err := newTestClient(t, server).GetAssignmentPolicyDevice(context.Background(), "device-1")
			if (err != nil) != test.wantError {
				t.Fatalf("release evidence handling mismatch: %v", err)
			}
		})
	}
}

func TestAssignmentPolicySuccessfulActivityRejectsReleasedReadback(t *testing.T) {
	posts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts++
			_, _ = w.Write([]byte(`{"data":{"type":"orgDeviceActivities","id":"activity-1","attributes":{"status":"COMPLETED","subStatus":"COMPLETED_WITH_SUCCESS"}}}`))
			return
		}
		raw := strings.ReplaceAll(policyDevice("device-1", "serial-1", "UNASSIGNED"), `"releasedFromOrgDateTime":null`, `"releasedFromOrgDateTime":"2026-01-01T00:00:00Z"`)
		_, _ = w.Write([]byte(`{"data":` + raw + `}`))
	}))
	defer server.Close()
	if err := newTestClient(t, server).AssignPolicyDevice(context.Background(), "device-1", ""); err == nil || posts != 1 {
		t.Fatal("released device accepted as successful unassignment")
	}
}
