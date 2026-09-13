// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_management_service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

// rfc3339StringValidator validates that a string value is a valid RFC 3339
// timestamp.
type rfc3339StringValidator struct{}

// Description describes the validation in plain text formatting.
func (v rfc3339StringValidator) Description(_ context.Context) string {
	return "string must be a valid RFC 3339 timestamp"
}

// MarkdownDescription describes the validation in Markdown format.
func (v rfc3339StringValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v rfc3339StringValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := time.Parse(time.RFC3339, req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid RFC 3339 timestamp",
			fmt.Sprintf("%q is not a valid RFC 3339 timestamp: %v", req.ConfigValue.ValueString(), err),
		)
	}
}

// validRFC3339 returns a string validator that requires RFC 3339 timestamps.
func validRFC3339() validator.String {
	return rfc3339StringValidator{}
}

// orgDeviceFetcher retrieves an organization device by serial number.
type orgDeviceFetcher func(ctx context.Context, id string) (*client.OrgDevice, error)

// migrationReadiness describes a single device's MDM migration eligibility as
// reported by the API at a point in time.
type migrationReadiness struct {
	DeviceID  string
	Capable   bool
	LookupErr error
}

// checkMigrationReadiness fetches each migrated device and reports its current
// migration eligibility.
func checkMigrationReadiness(ctx context.Context, migrated []MdmDeviceMigrationModel, fetch orgDeviceFetcher) []migrationReadiness {
	readiness := make([]migrationReadiness, 0, len(migrated))
	for _, device := range migrated {
		if device.ID.IsNull() || device.ID.IsUnknown() {
			continue
		}
		serial := device.ID.ValueString()
		r := migrationReadiness{DeviceID: serial}
		orgDevice, err := fetch(ctx, serial)
		if err != nil {
			r.LookupErr = err
			readiness = append(readiness, r)
			continue
		}
		r.Capable = orgDevice.Attributes.IsMdmMigrationCapable
		readiness = append(readiness, r)
	}
	return readiness
}

// planMigrationDiagnostics emits plan-time warnings for migrated devices whose
// eligibility cannot be confirmed right now. A device may legitimately be
// reported as non-eligible while a migration to this server is in progress, so
// eligibility must not block planning; it is definitively enforced when the
// resource is applied (see verifyMigrationsCapable).
func planMigrationDiagnostics(readiness []migrationReadiness, diags *diag.Diagnostics) {
	for _, r := range readiness {
		switch {
		case r.LookupErr != nil:
			diags.Append(diag.NewAttributeWarningDiagnostic(
				path.Root("migrated_devices"),
				"Could not verify MDM migration eligibility",
				fmt.Sprintf("could not verify migration eligibility for device %q during planning: %v. Eligibility is re-checked when the resource is applied.", r.DeviceID, r.LookupErr),
			))
		case !r.Capable:
			diags.Append(diag.NewAttributeWarningDiagnostic(
				path.Root("migrated_devices"),
				"Device not currently eligible for MDM migration",
				fmt.Sprintf("device %q is not currently reported as eligible for MDM migration. The device must be assigned to another MDM server before it can be migrated; eligibility is enforced when the resource is applied.", r.DeviceID),
			))
		}
	}
}

var (
	migrationEligibilityCheckInterval = 5 * time.Second
	migrationEligibilityTimeout       = 90 * time.Second
)

// verifyMigrationsCapable re-checks migration eligibility for the given serial
// numbers at apply time. Apple may take a moment to reflect a device as eligible
// for migration, so the check is retried until it succeeds or the timeout
// elapses. Devices that are not assigned to any server can never be migrated and
// are reported immediately.
func verifyMigrationsCapable(ctx context.Context, serials []string, fetch orgDeviceFetcher, timeout time.Duration) error {
	start := time.Now()
	var nonCapable, lookupFailed []string

	for {
		var pending []string
		nonCapable = nil
		lookupFailed = nil
		allReady := true

		for _, serial := range serials {
			orgDevice, err := fetch(ctx, serial)
			if err != nil {
				lookupFailed = append(lookupFailed, fmt.Sprintf("%s (%v)", serial, err))
				allReady = false
				continue
			}
			if orgDevice.Attributes.IsMdmMigrationCapable {
				continue
			}
			allReady = false
			if orgDevice.Attributes.Status == "UNASSIGNED" {
				nonCapable = append(nonCapable, serial)
			} else {
				pending = append(pending, serial)
			}
		}

		if allReady {
			return nil
		}
		if time.Since(start) >= timeout {
			nonCapable = append(nonCapable, pending...)
			break
		}
		if len(nonCapable) > 0 {
			break
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(migrationEligibilityCheckInterval):
		}
	}

	if len(nonCapable) > 0 {
		return fmt.Errorf("the following devices are not eligible for MDM migration: %s", joinQuoted(nonCapable))
	}
	if len(lookupFailed) > 0 {
		return fmt.Errorf("could not verify migration eligibility for the following devices: %s", joinQuoted(lookupFailed))
	}
	return nil
}

// joinQuoted joins a slice of strings with commas, quoting each item.
func joinQuoted(items []string) string {
	var s strings.Builder
	for i, item := range items {
		if i > 0 {
			s.WriteString(", ")
		}
		fmt.Fprintf(&s, "%q", item)
	}
	return s.String()
}

// normalizeDeadline converts an RFC 3339 timestamp into the millisecond
// precision ISO 8601 UTC format expected by the Apple Business Manager API.
func normalizeDeadline(value string) (string, error) {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return "", fmt.Errorf("invalid migration_deadline %q: %w", value, err)
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z"), nil
}

// deadlineAssignment captures a deadline group of devices for a migration activity.
type deadlineAssignment struct {
	deadline string
	devices  []string
}

// partitionMigrations splits assignable device serial numbers into those that
// carry an MDM migration deadline and those that are plain assignments.
// Devices in the migrated map are grouped by their normalized deadline.
// Devices without a deadline (or the empty deadline) go to plain.
func partitionMigrations(deviceSerials []string, migrated map[string]string) (plain []string, byDeadline []deadlineAssignment) {
	groups := make(map[string][]string)
	for _, serial := range deviceSerials {
		deadline, ok := migrated[serial]
		if !ok || deadline == "" {
			plain = append(plain, serial)
			continue
		}
		groups[deadline] = append(groups[deadline], serial)
	}

	for deadline, devices := range groups {
		byDeadline = append(byDeadline, deadlineAssignment{deadline: deadline, devices: devices})
	}
	return plain, byDeadline
}

// migrationMap builds a serial-number to deadline map from the migrated devices
// model, normalizing each deadline to the Apple API format.
func migrationMap(migrated []MdmDeviceMigrationModel) (map[string]string, error) {
	result := make(map[string]string, len(migrated))
	for _, device := range migrated {
		if device.ID.IsNull() || device.ID.IsUnknown() {
			continue
		}
		if device.MigrationDeadline.IsNull() || device.MigrationDeadline.IsUnknown() {
			return nil, fmt.Errorf("migration_deadline is required for migrated device %q", device.ID.ValueString())
		}
		normalized, err := normalizeDeadline(device.MigrationDeadline.ValueString())
		if err != nil {
			return nil, err
		}
		result[device.ID.ValueString()] = normalized
	}
	return result, nil
}

// migratedOutsideDeviceIDs returns the migrated device serial numbers that are
// not present in the given device_ids.
func migratedOutsideDeviceIDs(deviceIDs []string, migrated []MdmDeviceMigrationModel) []string {
	deviceIDSet := make(map[string]bool, len(deviceIDs))
	for _, id := range deviceIDs {
		deviceIDSet[id] = true
	}

	var outside []string
	for _, migratedDevice := range migrated {
		if migratedDevice.ID.IsNull() || migratedDevice.ID.IsUnknown() {
			continue
		}
		if !deviceIDSet[migratedDevice.ID.ValueString()] {
			outside = append(outside, migratedDevice.ID.ValueString())
		}
	}
	return outside
}
