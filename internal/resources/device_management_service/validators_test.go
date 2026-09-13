// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func TestNormalizeDeadline(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "rfc3339_w_z", input: "2026-03-15T17:00:00Z", want: "2026-03-15T17:00:00.000Z"},
		{name: "rfc3339_with_offset", input: "2026-03-15T10:00:00-07:00", want: "2026-03-15T17:00:00.000Z"},
		{name: "already_ms", input: "2026-03-15T17:00:00.123Z", want: "2026-03-15T17:00:00.123Z"},
		{name: "invalid", input: "not-a-date", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDeadline(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestPartitionMigrations(t *testing.T) {
	serials := []string{"S1", "S2", "S3", "S4"}
	migrated := map[string]string{
		"S1": "2026-03-15T17:00:00.000Z",
		"S2": "2026-03-15T17:00:00.000Z",
		"S3": "2026-04-01T00:00:00.000Z",
	}

	plain, byDeadline := partitionMigrations(serials, migrated)

	if len(plain) != 1 || plain[0] != "S4" {
		t.Errorf("expected plain=[S4], got %v", plain)
	}
	if len(byDeadline) != 2 {
		t.Fatalf("expected 2 deadline groups, got %d", len(byDeadline))
	}

	byDeadlineMap := make(map[string][]string)
	for _, g := range byDeadline {
		byDeadlineMap[g.deadline] = g.devices
	}
	if got := byDeadlineMap["2026-03-15T17:00:00.000Z"]; len(got) != 2 || !containsAll(got, "S1", "S2") {
		t.Errorf("expected deadline group to contain S1,S2, got %v", got)
	}
	if got := byDeadlineMap["2026-04-01T00:00:00.000Z"]; len(got) != 1 || got[0] != "S3" {
		t.Errorf("expected deadline group to contain S3, got %v", got)
	}
}

func TestPartitionMigrations_NoMigrations(t *testing.T) {
	serials := []string{"S1", "S2"}
	migrated := map[string]string{}

	plain, byDeadline := partitionMigrations(serials, migrated)
	if len(plain) != 2 {
		t.Errorf("expected all devices plain, got %v", plain)
	}
	if len(byDeadline) != 0 {
		t.Errorf("expected no deadline groups, got %v", byDeadline)
	}
}

func TestMigrationMap(t *testing.T) {
	devices := []MdmDeviceMigrationModel{
		{ID: types.StringValue("S1"), MigrationDeadline: types.StringValue("2026-03-15T17:00:00Z")},
		{ID: types.StringValue("S2"), MigrationDeadline: types.StringValue("2026-04-01T00:00:00.000Z")},
	}

	got, err := migrationMap(devices)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["S1"] != "2026-03-15T17:00:00.000Z" {
		t.Errorf("expected normalized S1 deadline, got %q", got["S1"])
	}
	if got["S2"] != "2026-04-01T00:00:00.000Z" {
		t.Errorf("expected normalized S2 deadline, got %q", got["S2"])
	}
}

func TestMigrationMap_InvalidDeadline(t *testing.T) {
	devices := []MdmDeviceMigrationModel{
		{ID: types.StringValue("S1"), MigrationDeadline: types.StringValue("nope")},
	}
	if _, err := migrationMap(devices); err == nil {
		t.Fatal("expected error for invalid deadline, got nil")
	}
}

func TestMigrationMap_MissingDeadline(t *testing.T) {
	devices := []MdmDeviceMigrationModel{
		{ID: types.StringValue("S1"), MigrationDeadline: types.StringNull()},
	}
	_, err := migrationMap(devices)
	if err == nil {
		t.Fatal("expected error for missing deadline, got nil")
	}
	if !strings.Contains(err.Error(), "S1") {
		t.Errorf("expected error to reference device S1, got %q", err.Error())
	}
}

func TestVerifyMigrationsCapable(t *testing.T) {
	oldInterval := migrationEligibilityCheckInterval
	migrationEligibilityCheckInterval = 10 * time.Millisecond
	t.Cleanup(func() { migrationEligibilityCheckInterval = oldInterval })

	serials := []string{"S1", "S2"}

	capable := func(ctx context.Context, id string) (*client.OrgDevice, error) {
		return &client.OrgDevice{
			ID:         id,
			Attributes: client.DeviceAttribute{IsMdmMigrationCapable: true, Status: "ASSIGNED"},
		}, nil
	}
	if err := verifyMigrationsCapable(context.Background(), serials, capable, time.Second); err != nil {
		t.Errorf("expected no error for capable devices, got %v", err)
	}

	// A device that is unassigned can never be migrated and must fail fast.
	unassignedFetcher := func(ctx context.Context, id string) (*client.OrgDevice, error) {
		return &client.OrgDevice{
			ID:         id,
			Attributes: client.DeviceAttribute{IsMdmMigrationCapable: false, Status: "UNASSIGNED"},
		}, nil
	}
	start := time.Now()
	err := verifyMigrationsCapable(context.Background(), serials, unassignedFetcher, time.Minute)
	if err == nil {
		t.Fatal("expected error for unassigned non-capable device")
	}
	if time.Since(start) > 2*time.Second {
		t.Error("expected fast failure for unassigned device")
	}
	if !strings.Contains(err.Error(), "S1") {
		t.Errorf("expected error to reference S1, got %q", err.Error())
	}

	// An assigned device that stays non-capable fails once the timeout elapses.
	laggyFetcher := func(ctx context.Context, id string) (*client.OrgDevice, error) {
		return &client.OrgDevice{
			ID:         id,
			Attributes: client.DeviceAttribute{IsMdmMigrationCapable: false, Status: "ASSIGNED"},
		}, nil
	}
	err = verifyMigrationsCapable(context.Background(), serials, laggyFetcher, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected error after timeout for laggy device")
	}

	// A device that becomes capable on a later poll succeeds.
	attempts := 0
	eventuallyCapable := func(ctx context.Context, id string) (*client.OrgDevice, error) {
		attempts++
		if attempts < 3 {
			return &client.OrgDevice{
				ID:         id,
				Attributes: client.DeviceAttribute{IsMdmMigrationCapable: false, Status: "ASSIGNED"},
			}, nil
		}
		return &client.OrgDevice{
			ID:         id,
			Attributes: client.DeviceAttribute{IsMdmMigrationCapable: true, Status: "ASSIGNED"},
		}, nil
	}
	if err := verifyMigrationsCapable(context.Background(), serials, eventuallyCapable, time.Second); err != nil {
		t.Errorf("expected eventual success, got %v", err)
	}

	// A persistent lookup failure fails fast via the timeout.
	fetchError := errors.New("boom")
	errorFetcher := func(ctx context.Context, id string) (*client.OrgDevice, error) {
		return nil, fetchError
	}
	start = time.Now()
	err = verifyMigrationsCapable(context.Background(), serials, errorFetcher, 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected error when fetch fails")
	}
	if time.Since(start) > 5*time.Second {
		t.Error("expected timeout-bound failure for persistent lookup errors")
	}
	if !strings.Contains(err.Error(), "S1") {
		t.Errorf("expected error to reference S1, got %q", err.Error())
	}

	// A transient lookup error should retry and recover once fetches succeed.
	attempts = 0
	flakyFetcher := func(ctx context.Context, id string) (*client.OrgDevice, error) {
		attempts++
		if attempts < 2 {
			return nil, errors.New("transient")
		}
		return &client.OrgDevice{
			ID:         id,
			Attributes: client.DeviceAttribute{IsMdmMigrationCapable: true, Status: "ASSIGNED"},
		}, nil
	}
	if err := verifyMigrationsCapable(context.Background(), serials, flakyFetcher, 2*time.Second); err != nil {
		t.Errorf("expected success after transient lookup errors, got %v", err)
	}
}

func TestCheckMigrationReadiness(t *testing.T) {
	fetch := func(ctx context.Context, id string) (*client.OrgDevice, error) {
		switch id {
		case "CAPABLE":
			return &client.OrgDevice{ID: id,
				Attributes: client.DeviceAttribute{IsMdmMigrationCapable: true, Status: "ASSIGNED"}}, nil
		case "NONCAP":
			return &client.OrgDevice{ID: id,
				Attributes: client.DeviceAttribute{IsMdmMigrationCapable: false, Status: "ASSIGNED"}}, nil
		case "UNASSIGNED":
			return &client.OrgDevice{ID: id,
				Attributes: client.DeviceAttribute{IsMdmMigrationCapable: false, Status: "UNASSIGNED"}}, nil
		default:
			return nil, errors.New("boom")
		}
	}

	migrated := []MdmDeviceMigrationModel{
		{ID: types.StringValue("CAPABLE"), MigrationDeadline: types.StringValue("2026-03-15T17:00:00.000Z")},
		{ID: types.StringValue("NONCAP"), MigrationDeadline: types.StringValue("2026-03-15T17:00:00.000Z")},
		{ID: types.StringValue("UNASSIGNED"), MigrationDeadline: types.StringValue("2026-03-15T17:00:00.000Z")},
		{ID: types.StringValue("MISSING"), MigrationDeadline: types.StringValue("2026-03-15T17:00:00.000Z")},
	}

	readiness := checkMigrationReadiness(context.Background(), migrated, fetch)
	if len(readiness) != 4 {
		t.Fatalf("expected 4 readiness entries, got %d", len(readiness))
	}

	byID := make(map[string]migrationReadiness, len(readiness))
	for _, r := range readiness {
		byID[r.DeviceID] = r
	}

	if r := byID["CAPABLE"]; !r.Capable || r.LookupErr != nil {
		t.Errorf("CAPABLE: expected capable, got %+v", r)
	}
	for _, nonCapable := range []string{"NONCAP", "UNASSIGNED"} {
		if r := byID[nonCapable]; r.Capable || r.LookupErr != nil {
			t.Errorf("%s: expected non-capable, got %+v", nonCapable, r)
		}
	}
	if r := byID["MISSING"]; r.LookupErr == nil {
		t.Errorf("MISSING: expected lookup error, got %+v", r)
	}
}

func TestPlanMigrationDiagnostics(t *testing.T) {
	readiness := []migrationReadiness{
		{DeviceID: "OK", Capable: true},
		{DeviceID: "NONCAP", Capable: false},
		{DeviceID: "MISSING", LookupErr: errors.New("boom")},
	}

	var diags diag.Diagnostics
	planMigrationDiagnostics(readiness, &diags)

	if len(diags) != 2 {
		t.Fatalf("expected 2 warning diagnostics, got %d: %v", len(diags), diags.Errors())
	}

	var warnedNonCap, warnedLookup bool
	for _, d := range diags {
		if d.Severity() != diag.SeverityWarning {
			t.Errorf("expected all warnings, got severity %s", d.Severity())
		}
		if strings.Contains(d.Detail(), "NONCAP") {
			warnedNonCap = true
		}
		if strings.Contains(d.Detail(), "MISSING") {
			warnedLookup = true
		}
	}
	if !warnedNonCap {
		t.Error("expected a warning for the non-capable device")
	}
	if !warnedLookup {
		t.Error("expected a warning for the lookup failure")
	}
}

func TestJoinQuoted(t *testing.T) {
	got := joinQuoted([]string{"a", "b"})
	want := `"a", "b"`
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
	if got := joinQuoted(nil); got != "" {
		t.Errorf("expected empty string for nil, got %q", got)
	}
}

func TestMigratedOutsideDeviceIDs(t *testing.T) {
	migrated := []MdmDeviceMigrationModel{
		{ID: types.StringValue("S1"), MigrationDeadline: types.StringValue("2026-03-15T17:00:00.000Z")},
		{ID: types.StringValue("S9"), MigrationDeadline: types.StringValue("2026-04-01T00:00:00.000Z")},
	}
	outside := migratedOutsideDeviceIDs([]string{"S1", "S2"}, migrated)
	if len(outside) != 1 || outside[0] != "S9" {
		t.Errorf("expected [S9], got %v", outside)
	}
	if got := migratedOutsideDeviceIDs([]string{"S1", "S9"}, migrated); len(got) != 0 {
		t.Errorf("expected none outside, got %v", got)
	}
	if got := migratedOutsideDeviceIDs(nil, nil); len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func containsAll(items []string, wanted ...string) bool {
	itemSet := make(map[string]bool, len(items))
	for _, i := range items {
		itemSet[i] = true
	}
	for _, w := range wanted {
		if !itemSet[w] {
			return false
		}
	}
	return true
}
