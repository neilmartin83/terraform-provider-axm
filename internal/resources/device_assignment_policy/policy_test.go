// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_assignment_policy

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func testPolicy() Policy {
	return Policy{
		Assignments:            map[string]string{},
		AuthoritativeServerIDs: []string{"dev", "live"},
		FallbackServerID:       "prod",
		ServerNames:            map[string]string{"dev": "Development MDM", "live": "Live MDM", "prod": "Default MDM", "other": "Other MDM"},
		MaxChanges:             100,
	}
}

func testSnapshot() Snapshot {
	return Snapshot{Servers: []Server{
		{ID: "prod", Name: "Default MDM", Type: "MDM"},
		{ID: "dev", Name: "Development MDM", Type: "MDM"},
		{ID: "live", Name: "Live MDM", Type: "MDM"},
		{ID: "other", Name: "Other MDM", Type: "MDM"},
		{ID: "configurator", Name: "Apple Configurator", Type: "APPLE_CONFIGURATOR"},
		{ID: "apple", Name: "Apple Built-in", Type: "APPLE_MDM"},
	}}
}

func testDevice(serial, destination string) Device {
	status, server := "ASSIGNED", destination
	if destination == Unassigned || destination == "RELEASED" {
		status, server = destination, ""
	}
	return Device{ID: "id-" + serial, SerialNumber: serial, Status: status, ServerID: server}
}

func TestBuildPlanTransitionMatrix(t *testing.T) {
	for _, current := range []string{"dev", "live", "prod", "other", "configurator", "apple", Unassigned} {
		for _, desired := range []string{"", "dev", "live", "prod", "other", Unassigned} {
			t.Run(current+"_to_"+desired, func(t *testing.T) {
				policy, snapshot := testPolicy(), testSnapshot()
				snapshot.Devices = []Device{testDevice("serial", current)}
				if desired != "" {
					policy.Assignments["serial"] = desired
				}
				plan, err := BuildPlan(policy, snapshot)
				if err != nil {
					t.Fatal(err)
				}
				want := desired
				if want == "" && (current == "dev" || current == "live") {
					want = "prod"
				}
				if want == "" {
					if len(plan.Observed) != 0 || len(plan.Changes) != 0 {
						t.Fatalf("an out-of-scope device must be untouched: %#v", plan)
					}
					return
				}
				if !reflect.DeepEqual(plan.Observed, map[string]string{"serial": current}) {
					t.Fatalf("observed assignment mismatch: %#v", plan.Observed)
				}
				if current == want {
					if len(plan.Changes) != 0 {
						t.Fatalf("existing destination must be a no-op: %#v", plan.Changes)
					}
					return
				}
				expected := []Change{{DeviceID: "id-serial", SerialNumber: "serial", From: current, To: want}}
				if !reflect.DeepEqual(plan.Changes, expected) {
					t.Fatalf("wrong correction: %#v, want %#v", plan.Changes, expected)
				}
			})
		}
	}
}

func TestBuildPlanCompletePolicyExample(t *testing.T) {
	policy, snapshot := testPolicy(), testSnapshot()
	policy.Assignments = map[string]string{
		"keep-dev": "dev", "restore-from-prod": "dev", "cross-server": "live",
		"explicit-prod": "prod", "explicit-unassigned": Unassigned, "keep-unassigned": Unassigned,
	}
	snapshot.Devices = []Device{
		testDevice("keep-dev", "dev"), testDevice("restore-from-prod", "prod"),
		testDevice("cross-server", "dev"), testDevice("explicit-prod", "live"),
		testDevice("explicit-unassigned", "live"), testDevice("keep-unassigned", Unassigned),
		testDevice("new-extra", "dev"), testDevice("untouched-prod", "prod"),
		testDevice("untouched-unassigned", Unassigned), testDevice("untouched-other", "other"),
		testDevice("untouched-configurator", "configurator"), testDevice("untouched-released", "RELEASED"),
	}
	plan, err := BuildPlan(policy, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	expected := []Change{
		{DeviceID: "id-cross-server", SerialNumber: "cross-server", From: "dev", To: "live"},
		{DeviceID: "id-explicit-prod", SerialNumber: "explicit-prod", From: "live", To: "prod"},
		{DeviceID: "id-explicit-unassigned", SerialNumber: "explicit-unassigned", From: "live", To: Unassigned},
		{DeviceID: "id-new-extra", SerialNumber: "new-extra", From: "dev", To: "prod"},
		{DeviceID: "id-restore-from-prod", SerialNumber: "restore-from-prod", From: "prod", To: "dev"},
	}
	if !reflect.DeepEqual(plan.Changes, expected) {
		t.Fatalf("unexpected changes: %#v", plan.Changes)
	}
	if len(plan.Observed) != len(policy.Assignments)+1 {
		t.Fatalf("observed scope must include explicit devices plus the extra: %#v", plan.Observed)
	}
	for _, device := range snapshot.Devices {
		if strings.HasPrefix(device.SerialNumber, "untouched-") {
			if _, ok := plan.Observed[device.SerialNumber]; ok {
				t.Fatalf("unmanaged device included: %s", device.SerialNumber)
			}
		}
	}
}

func TestBuildPlanDeletionFallsBackOnlyWhileOnAuthoritativeServer(t *testing.T) {
	policy, snapshot := testPolicy(), testSnapshot()
	policy.Assignments["serial"] = "dev"
	snapshot.Devices = []Device{testDevice("serial", "dev")}
	before, err := BuildPlan(policy, snapshot)
	if err != nil || len(before.Changes) != 0 {
		t.Fatalf("expected initial no-op: %#v, %v", before, err)
	}
	delete(policy.Assignments, "serial")
	after, err := BuildPlan(policy, snapshot)
	if err != nil || len(after.Changes) != 1 || after.Changes[0].To != "prod" {
		t.Fatalf("removing desired assignment must fall back: %#v, %v", after, err)
	}
	snapshot.Devices[0] = testDevice("serial", "prod")
	converged, err := BuildPlan(policy, snapshot)
	if err != nil || len(converged.Changes) != 0 || len(converged.Observed) != 0 {
		t.Fatalf("fallback must converge to unmanaged default: %#v, %v", converged, err)
	}
}

func TestBuildPlanEnforcementFlagsDoNotHideChanges(t *testing.T) {
	policy, snapshot := testPolicy(), testSnapshot()
	policy.AdoptionOnly, policy.MaxChanges = true, 0
	snapshot.Devices = []Device{testDevice("new-extra", "dev")}
	plan, err := BuildPlan(policy, snapshot)
	if err != nil || len(plan.Changes) != 1 || plan.Changes[0].To != "prod" {
		t.Fatalf("discovery must expose changes even when writes are disabled: %#v, %v", plan, err)
	}
}

func TestBuildPlanNoAuthoritativeServers(t *testing.T) {
	policy, snapshot := testPolicy(), testSnapshot()
	policy.AuthoritativeServerIDs = nil
	policy.Assignments["explicit"] = "prod"
	snapshot.Devices = []Device{testDevice("explicit", "dev"), testDevice("unlisted", "dev")}
	plan, err := BuildPlan(policy, snapshot)
	if err != nil || len(plan.Observed) != 1 || len(plan.Changes) != 1 || plan.Changes[0].SerialNumber != "explicit" {
		t.Fatalf("only explicit entries should be managed: %#v, %v", plan, err)
	}
}

func TestBuildPlanDoesNotMutateInputs(t *testing.T) {
	policy, snapshot := testPolicy(), testSnapshot()
	policy.Assignments["one"] = "live"
	snapshot.Devices = []Device{testDevice("two", "dev"), testDevice("one", "prod")}
	beforePolicy, _ := json.Marshal(policy)
	beforeSnapshot, _ := json.Marshal(snapshot)
	plan, err := BuildPlan(policy, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	plan.Observed["one"] = "changed-output"
	plan.Changes[0].To = "changed-output"
	afterPolicy, _ := json.Marshal(policy)
	afterSnapshot, _ := json.Marshal(snapshot)
	if string(beforePolicy) != string(afterPolicy) || string(beforeSnapshot) != string(afterSnapshot) {
		t.Fatal("plan construction or output mutation changed an input")
	}
}

func TestBuildPlanDeterministicAcrossInventoryOrder(t *testing.T) {
	policy, snapshot := testPolicy(), testSnapshot()
	policy.Assignments = map[string]string{"z": Unassigned, "a": "dev"}
	snapshot.Devices = []Device{testDevice("z", "live"), testDevice("extra", "dev"), testDevice("a", "prod")}
	original, err := BuildPlan(policy, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 20 {
		snapshot.Devices[0], snapshot.Devices[2] = snapshot.Devices[2], snapshot.Devices[0]
		snapshot.Servers[0], snapshot.Servers[1] = snapshot.Servers[1], snapshot.Servers[0]
		policy.AuthoritativeServerIDs[0], policy.AuthoritativeServerIDs[1] = policy.AuthoritativeServerIDs[1], policy.AuthoritativeServerIDs[0]
		got, err := BuildPlan(policy, snapshot)
		if err != nil || !reflect.DeepEqual(got, original) {
			t.Fatalf("iteration %d changed plan: %#v, %v", i, got, err)
		}
	}
}

func TestValidatePolicyRejectsInvalidConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Policy)
	}{
		{"negative_max_changes", func(p *Policy) { p.MaxChanges = -1 }},
		{"missing_fallback", func(p *Policy) { p.FallbackServerID = "" }},
		{"unassigned_fallback", func(p *Policy) { p.FallbackServerID = Unassigned }},
		{"unpinned_fallback", func(p *Policy) { delete(p.ServerNames, "prod") }},
		{"fallback_is_authoritative", func(p *Policy) { p.AuthoritativeServerIDs = append(p.AuthoritativeServerIDs, "prod") }},
		{"repeated_authoritative", func(p *Policy) { p.AuthoritativeServerIDs = append(p.AuthoritativeServerIDs, "dev") }},
		{"empty_authoritative", func(p *Policy) { p.AuthoritativeServerIDs = append(p.AuthoritativeServerIDs, "") }},
		{"unassigned_authoritative", func(p *Policy) { p.AuthoritativeServerIDs = append(p.AuthoritativeServerIDs, Unassigned) }},
		{"unpinned_authoritative", func(p *Policy) { delete(p.ServerNames, "dev") }},
		{"blank_pin_id", func(p *Policy) { p.ServerNames[""] = "valid" }},
		{"invalid_pin_id", func(p *Policy) { p.ServerNames["private\nidentifier"] = "valid" }},
		{"unassigned_pin_id", func(p *Policy) { p.ServerNames[Unassigned] = "valid" }},
		{"blank_pin_name", func(p *Policy) { p.ServerNames["dev"] = "  " }},
		{"control_pin_name", func(p *Policy) { p.ServerNames["dev"] = "private\nname" }},
		{"invalid_utf8_pin_name", func(p *Policy) { p.ServerNames["dev"] = string([]byte{0xff}) }},
		{"empty_explicit_serial", func(p *Policy) { p.Assignments[""] = "dev" }},
		{"whitespace_explicit_serial", func(p *Policy) { p.Assignments["private serial"] = "dev" }},
		{"invalid_utf8_serial", func(p *Policy) { p.Assignments[string([]byte{0xff})] = "dev" }},
		{"empty_explicit_destination", func(p *Policy) { p.Assignments["private-serial"] = "" }},
		{"unpinned_explicit_destination", func(p *Policy) { p.Assignments["private-serial"] = "private-server" }},
		{"malformed_explicit_destination", func(p *Policy) { p.Assignments["private-serial"] = "dev " }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy := testPolicy()
			test.mutate(&policy)
			err := ValidatePolicy(policy)
			if err == nil {
				t.Fatal("invalid configuration was accepted")
			}
			assertPrivateError(t, err)
			plan, err := BuildPlan(policy, testSnapshot())
			assertRejectedPlan(t, plan, err)
		})
	}
}

func TestBuildPlanRejectsInvalidInventory(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Policy, *Snapshot)
	}{
		{"missing_pinned_server", func(_ *Policy, s *Snapshot) { s.Servers = s.Servers[1:] }},
		{"duplicate_server", func(_ *Policy, s *Snapshot) { s.Servers = append(s.Servers, s.Servers[0]) }},
		{"blank_server_id", func(_ *Policy, s *Snapshot) { s.Servers[0].ID = "" }},
		{"unassigned_server_id", func(_ *Policy, s *Snapshot) { s.Servers[0].ID = Unassigned }},
		{"blank_server_name", func(_ *Policy, s *Snapshot) { s.Servers[0].Name = "\t" }},
		{"renamed_server", func(_ *Policy, s *Snapshot) { s.Servers[0].Name = "private-renamed-server" }},
		{"changed_type", func(_ *Policy, s *Snapshot) { s.Servers[0].Type = "APPLE_MDM" }},
		{"unknown_type_even_if_unmanaged", func(_ *Policy, s *Snapshot) { s.Servers[5].Type = "private-unknown-type" }},
		{"explicit_configurator", func(p *Policy, _ *Snapshot) {
			p.ServerNames["configurator"] = "Apple Configurator"
			p.Assignments["private-serial"] = "configurator"
		}},
		{"explicit_apple", func(p *Policy, _ *Snapshot) {
			p.ServerNames["apple"] = "Apple Built-in"
			p.Assignments["private-serial"] = "apple"
		}},
		{"missing_explicit_device", func(p *Policy, _ *Snapshot) { p.Assignments["private-missing-serial"] = "dev" }},
		{"released_explicit_device", func(p *Policy, s *Snapshot) {
			p.Assignments["private-serial"] = "dev"
			s.Devices[0] = testDevice("private-serial", "RELEASED")
		}},
		{"released_explicit_unassigned", func(p *Policy, s *Snapshot) {
			p.Assignments["private-serial"] = Unassigned
			s.Devices[0] = testDevice("private-serial", "RELEASED")
		}},
		{"duplicate_serial", func(_ *Policy, s *Snapshot) {
			d := s.Devices[0]
			d.ID = "different-private-id"
			s.Devices = append(s.Devices, d)
		}},
		{"duplicate_device_id", func(_ *Policy, s *Snapshot) {
			d := s.Devices[0]
			d.SerialNumber = "different-private-serial"
			s.Devices = append(s.Devices, d)
		}},
		{"blank_device_id", func(_ *Policy, s *Snapshot) { s.Devices[0].ID = "" }},
		{"blank_serial", func(_ *Policy, s *Snapshot) { s.Devices[0].SerialNumber = "" }},
		{"whitespace_serial", func(_ *Policy, s *Snapshot) { s.Devices[0].SerialNumber = "private serial" }},
		{"unknown_device_status", func(_ *Policy, s *Snapshot) { s.Devices[0].Status = "private-status" }},
		{"assigned_missing_server", func(_ *Policy, s *Snapshot) { s.Devices[0].ServerID = "" }},
		{"assigned_unknown_server", func(_ *Policy, s *Snapshot) { s.Devices[0].ServerID = "private-unknown-server" }},
		{"assigned_unassigned_sentinel", func(_ *Policy, s *Snapshot) { s.Devices[0].ServerID = Unassigned }},
		{"unassigned_with_server", func(_ *Policy, s *Snapshot) { s.Devices[0].Status = Unassigned }},
		{"released_with_server", func(_ *Policy, s *Snapshot) { s.Devices[0].Status = "RELEASED" }},
		{"malformed_unmanaged_device", func(_ *Policy, s *Snapshot) {
			d := testDevice("private-unmanaged-serial", "prod")
			d.Status = "private-status"
			s.Devices = append(s.Devices, d)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			policy, snapshot := testPolicy(), testSnapshot()
			snapshot.Devices = []Device{testDevice("private-serial", "dev")}
			test.mutate(&policy, &snapshot)
			plan, err := BuildPlan(policy, snapshot)
			assertRejectedPlan(t, plan, err)
		})
	}
}

func TestBuildPlanEmptyInventoryDoesNotInventDevices(t *testing.T) {
	plan, err := BuildPlan(testPolicy(), testSnapshot())
	if err != nil || plan.Observed == nil || plan.Changes == nil || len(plan.Observed) != 0 || len(plan.Changes) != 0 {
		t.Fatalf("expected a deterministic empty plan: %#v, %v", plan, err)
	}
}

func assertRejectedPlan(t *testing.T, plan Plan, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("invalid input was accepted")
	}
	if !reflect.DeepEqual(plan, Plan{}) {
		t.Fatalf("rejected input returned a partial plan: %#v", plan)
	}
	assertPrivateError(t, err)
}

func assertPrivateError(t *testing.T, err error) {
	t.Helper()
	if strings.Contains(err.Error(), "private") || strings.Contains(err.Error(), "Development MDM") || strings.Contains(err.Error(), "id-") {
		t.Fatalf("error exposed input data: %q", err.Error())
	}
}
