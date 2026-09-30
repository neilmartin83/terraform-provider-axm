// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package device_assignment_policy

import (
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Unassigned is an explicit desired state, distinct from an omitted device.
const Unassigned = "UNASSIGNED"

// Policy declares explicit device destinations and the servers whose unlisted
// members should be moved to the fallback server.
type Policy struct {
	Assignments            map[string]string
	AuthoritativeServerIDs []string
	FallbackServerID       string
	ServerNames            map[string]string
	AdoptionOnly           bool
	MaxChanges             int64
}

// Server is the identity needed to validate a destination against live inventory.
type Server struct {
	ID   string
	Name string
	Type string
}

// Device is the assignment information needed to reconcile one device.
// Unassigned and released devices must have an empty ServerID.
type Device struct {
	ID           string
	SerialNumber string
	Status       string
	ServerID     string
}

// Snapshot is a complete inventory, assembled without modifying remote devices.
type Snapshot struct {
	Servers []Server
	Devices []Device
}

// Change describes the single final destination for one device.
type Change struct {
	DeviceID     string
	SerialNumber string
	From         string
	To           string
}

// Plan includes observed assignments in scope and changes in serial-number order.
type Plan struct {
	Observed map[string]string
	Changes  []Change
}

// ValidatePolicy checks the configuration without relying on live inventory.
// Errors intentionally exclude all configured values and device identifiers.
func ValidatePolicy(policy Policy) error {
	if policy.MaxChanges < 0 {
		return errors.New("maximum changes must not be negative")
	}
	if !validIdentifier(policy.FallbackServerID) || policy.FallbackServerID == Unassigned {
		return errors.New("a valid fallback MDM server is required")
	}
	for id, name := range policy.ServerNames {
		if !validIdentifier(id) || id == Unassigned || !validName(name) {
			return errors.New("server identity pins must contain valid IDs and names")
		}
	}
	if _, ok := policy.ServerNames[policy.FallbackServerID]; !ok {
		return errors.New("the fallback server must have an identity pin")
	}
	authoritative := make(map[string]struct{}, len(policy.AuthoritativeServerIDs))
	for _, id := range policy.AuthoritativeServerIDs {
		if !validIdentifier(id) || id == Unassigned {
			return errors.New("authoritative servers must contain valid MDM server IDs")
		}
		if id == policy.FallbackServerID {
			return errors.New("the fallback server cannot also be authoritative")
		}
		if _, ok := authoritative[id]; ok {
			return errors.New("authoritative server IDs must not be repeated")
		}
		if _, ok := policy.ServerNames[id]; !ok {
			return errors.New("every authoritative server must have an identity pin")
		}
		authoritative[id] = struct{}{}
	}
	for serial, destination := range policy.Assignments {
		if !validIdentifier(serial) || !validIdentifier(destination) {
			return errors.New("explicit assignments must contain valid device and destination identifiers")
		}
		if destination != Unassigned {
			if _, ok := policy.ServerNames[destination]; !ok {
				return errors.New("every explicit MDM destination must have an identity pin")
			}
		}
	}
	return nil
}

// BuildPlan reconciles policy against one complete snapshot without side effects.
// AdoptionOnly and MaxChanges govern writes, not discovery of intended changes;
// callers must enforce both again at the apply boundary.
func BuildPlan(policy Policy, snapshot Snapshot) (Plan, error) {
	if err := ValidatePolicy(policy); err != nil {
		return Plan{}, err
	}
	servers := make(map[string]Server, len(snapshot.Servers))
	for _, server := range snapshot.Servers {
		if !validIdentifier(server.ID) || server.ID == Unassigned || !validName(server.Name) {
			return Plan{}, errors.New("server inventory contains a malformed identity")
		}
		switch server.Type {
		case "MDM", "APPLE_CONFIGURATOR", "APPLE_MDM":
		default:
			return Plan{}, errors.New("server inventory contains an unsupported server type")
		}
		if _, ok := servers[server.ID]; ok {
			return Plan{}, errors.New("server inventory contains a repeated server ID")
		}
		servers[server.ID] = server
	}
	for id, name := range policy.ServerNames {
		server, ok := servers[id]
		if !ok {
			return Plan{}, errors.New("a pinned MDM server is missing from inventory")
		}
		if server.Type != "MDM" {
			return Plan{}, errors.New("assignment destinations must be external MDM servers; Apple-managed and Configurator destinations are unsupported")
		}
		if server.Name != name {
			return Plan{}, errors.New("a pinned MDM server name does not match inventory")
		}
	}
	devices := make(map[string]Device, len(snapshot.Devices))
	deviceIDs := make(map[string]struct{}, len(snapshot.Devices))
	for _, device := range snapshot.Devices {
		if !validIdentifier(device.ID) || !validIdentifier(device.SerialNumber) {
			return Plan{}, errors.New("device inventory contains a malformed identity")
		}
		if _, ok := devices[device.SerialNumber]; ok {
			return Plan{}, errors.New("device inventory contains a repeated serial number")
		}
		if _, ok := deviceIDs[device.ID]; ok {
			return Plan{}, errors.New("device inventory contains a repeated device ID")
		}
		switch device.Status {
		case "ASSIGNED":
			if _, ok := servers[device.ServerID]; !ok {
				return Plan{}, errors.New("an assigned device has an unknown or missing server")
			}
		case Unassigned, "RELEASED":
			if device.ServerID != "" {
				return Plan{}, errors.New("device assignment status is inconsistent with its server")
			}
		default:
			return Plan{}, errors.New("device inventory contains an unsupported assignment status")
		}
		devices[device.SerialNumber] = device
		deviceIDs[device.ID] = struct{}{}
	}
	for serial := range policy.Assignments {
		device, ok := devices[serial]
		if !ok {
			return Plan{}, errors.New("an explicitly managed device is missing from inventory")
		}
		if device.Status == "RELEASED" {
			return Plan{}, errors.New("an explicitly managed device has been released from the organization")
		}
	}
	authoritative := make(map[string]struct{}, len(policy.AuthoritativeServerIDs))
	for _, id := range policy.AuthoritativeServerIDs {
		authoritative[id] = struct{}{}
	}
	plan := Plan{Observed: map[string]string{}, Changes: []Change{}}
	for serial, device := range devices {
		destination, explicit := policy.Assignments[serial]
		_, inAuthoritativeServer := authoritative[device.ServerID]
		if !explicit && !inAuthoritativeServer {
			continue
		}
		if !explicit {
			destination = policy.FallbackServerID
		}
		current := device.ServerID
		if device.Status == Unassigned {
			current = Unassigned
		}
		plan.Observed[serial] = current
		if destination != current {
			plan.Changes = append(plan.Changes, Change{
				DeviceID: device.ID, SerialNumber: serial, From: current, To: destination,
			})
		}
	}
	sort.Slice(plan.Changes, func(i, j int) bool {
		return plan.Changes[i].SerialNumber < plan.Changes[j].SerialNumber
	})
	return plan, nil
}

func validIdentifier(value string) bool {
	return value != "" && utf8.ValidString(value) && strings.IndexFunc(value, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) == -1
}

func validName(value string) bool {
	return strings.TrimSpace(value) != "" && utf8.ValidString(value) && strings.IndexFunc(value, unicode.IsControl) == -1
}
