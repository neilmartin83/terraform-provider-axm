// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package audit_events

import (
	"testing"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func TestFlattenAuditEvent(t *testing.T) {
	event := client.AuditEvent{
		ID:   "event-1",
		Type: "auditEvent",
		Attributes: client.AuditEventAttributes{
			EventDateTime:        "2024-06-01T12:00:00Z",
			Type:                 "USER_CREATED",
			Category:             "USER",
			ActorType:            "admin",
			ActorID:              "actor-1",
			ActorName:            "Admin User",
			SubjectType:          "user",
			SubjectID:            "subject-1",
			SubjectName:          "Target User",
			Outcome:              "SUCCESS",
			GroupID:              "group-1",
			EventDataPropertyKey: "userId",
			Additional: map[string]any{
				"userId": "subject-1",
			},
		},
	}

	model := flattenAuditEvent(event)

	if model.ID.ValueString() != "event-1" {
		t.Errorf("expected ID event-1, got %s", model.ID.ValueString())
	}
	if model.Type.ValueString() != "auditEvent" {
		t.Errorf("expected Type auditEvent, got %s", model.Type.ValueString())
	}
	if model.EventDateTime.ValueString() != "2024-06-01T12:00:00Z" {
		t.Errorf("expected EventDateTime 2024-06-01T12:00:00Z, got %s", model.EventDateTime.ValueString())
	}
	if model.EventType.ValueString() != "USER_CREATED" {
		t.Errorf("expected EventType USER_CREATED, got %s", model.EventType.ValueString())
	}
	if model.Category.ValueString() != "USER" {
		t.Errorf("expected Category USER, got %s", model.Category.ValueString())
	}
	if model.ActorType.ValueString() != "admin" {
		t.Errorf("expected ActorType admin, got %s", model.ActorType.ValueString())
	}
	if model.ActorID.ValueString() != "actor-1" {
		t.Errorf("expected ActorID actor-1, got %s", model.ActorID.ValueString())
	}
	if model.ActorName.ValueString() != "Admin User" {
		t.Errorf("expected ActorName Admin User, got %s", model.ActorName.ValueString())
	}
	if model.SubjectType.ValueString() != "user" {
		t.Errorf("expected SubjectType user, got %s", model.SubjectType.ValueString())
	}
	if model.SubjectID.ValueString() != "subject-1" {
		t.Errorf("expected SubjectID subject-1, got %s", model.SubjectID.ValueString())
	}
	if model.SubjectName.ValueString() != "Target User" {
		t.Errorf("expected SubjectName Target User, got %s", model.SubjectName.ValueString())
	}
	if model.Outcome.ValueString() != "SUCCESS" {
		t.Errorf("expected Outcome SUCCESS, got %s", model.Outcome.ValueString())
	}
	if model.GroupID.ValueString() != "group-1" {
		t.Errorf("expected GroupID group-1, got %s", model.GroupID.ValueString())
	}
	if model.EventDataPropertyKey.ValueString() != "userId" {
		t.Errorf("expected EventDataPropertyKey userId, got %s", model.EventDataPropertyKey.ValueString())
	}
	if model.EventDataJSON.ValueString() == "" {
		t.Error("expected non-empty EventDataJSON")
	}
}

func TestFlattenAuditEvent_EmptyOptionalFields(t *testing.T) {
	event := client.AuditEvent{
		ID:         "event-2",
		Type:       "auditEvent",
		Attributes: client.AuditEventAttributes{},
	}

	model := flattenAuditEvent(event)

	if !model.EventDateTime.IsNull() {
		t.Error("expected EventDateTime to be null when empty")
	}
	if !model.EventType.IsNull() {
		t.Error("expected EventType to be null when empty")
	}
	if !model.Category.IsNull() {
		t.Error("expected Category to be null when empty")
	}
	if !model.ActorType.IsNull() {
		t.Error("expected ActorType to be null when empty")
	}
	if !model.ActorID.IsNull() {
		t.Error("expected ActorID to be null when empty")
	}
	if !model.ActorName.IsNull() {
		t.Error("expected ActorName to be null when empty")
	}
	if !model.SubjectType.IsNull() {
		t.Error("expected SubjectType to be null when empty")
	}
	if !model.SubjectID.IsNull() {
		t.Error("expected SubjectID to be null when empty")
	}
	if !model.SubjectName.IsNull() {
		t.Error("expected SubjectName to be null when empty")
	}
	if !model.Outcome.IsNull() {
		t.Error("expected Outcome to be null when empty")
	}
	if !model.GroupID.IsNull() {
		t.Error("expected GroupID to be null when empty")
	}
	if !model.EventDataPropertyKey.IsNull() {
		t.Error("expected EventDataPropertyKey to be null when empty")
	}
	if model.EventDataJSON.ValueString() != "null" {
		t.Errorf("expected EventDataJSON to be 'null' when Additional is nil, got %q", model.EventDataJSON.ValueString())
	}
}
