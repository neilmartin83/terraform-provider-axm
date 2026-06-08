// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package users

import (
	"testing"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func TestFlattenUser(t *testing.T) {
	user := client.User{
		Type: "users",
		ID:   "user-1",
		Attributes: client.UserAttributes{
			FirstName:           "Jane",
			LastName:            "Doe",
			MiddleName:          "M",
			Status:              "ACTIVE",
			ManagedAppleAccount: "jane.doe@example.com",
			IsExternalUser:      false,
			RoleOuList: []client.UserRoleOuMapping{
				{RoleName: "Teacher", OuID: "ou-1"},
			},
			Email:           "jane@example.com",
			EmployeeNumber:  "12345",
			CostCenter:      "CC-1",
			Division:        "Div-1",
			Department:      "Dept-1",
			JobTitle:        "Teacher",
			StartDateTime:   "2024-01-01T00:00:00Z",
			CreatedDateTime: "2024-01-01T00:00:00Z",
			UpdatedDateTime: "2024-06-01T00:00:00Z",
			PhoneNumbers: []client.UserPhoneNumber{
				{PhoneNumber: "555-0100", Type: "WORK"},
			},
		},
	}

	model := flattenUser(user)

	if model.ID.ValueString() != "user-1" {
		t.Errorf("expected ID user-1, got %s", model.ID.ValueString())
	}
	if model.Type.ValueString() != "users" {
		t.Errorf("expected Type users, got %s", model.Type.ValueString())
	}
	if model.FirstName.ValueString() != "Jane" {
		t.Errorf("expected FirstName Jane, got %s", model.FirstName.ValueString())
	}
	if model.LastName.ValueString() != "Doe" {
		t.Errorf("expected LastName Doe, got %s", model.LastName.ValueString())
	}
	if !model.MiddleName.IsNull() && model.MiddleName.ValueString() != "M" {
		t.Errorf("expected MiddleName M, got %s", model.MiddleName.ValueString())
	}
	if model.Status.ValueString() != "ACTIVE" {
		t.Errorf("expected Status ACTIVE, got %s", model.Status.ValueString())
	}
	if model.ManagedAppleAccount.ValueString() != "jane.doe@example.com" {
		t.Errorf("expected ManagedAppleAccount jane.doe@example.com, got %s", model.ManagedAppleAccount.ValueString())
	}
	if model.IsExternalUser.ValueBool() != false {
		t.Error("expected IsExternalUser false")
	}
	if model.Email.ValueString() != "jane@example.com" {
		t.Errorf("expected Email jane@example.com, got %s", model.Email.ValueString())
	}
	if model.EmployeeNumber.ValueString() != "12345" {
		t.Errorf("expected EmployeeNumber 12345, got %s", model.EmployeeNumber.ValueString())
	}
	if model.CostCenter.ValueString() != "CC-1" {
		t.Errorf("expected CostCenter CC-1, got %s", model.CostCenter.ValueString())
	}
	if model.Division.ValueString() != "Div-1" {
		t.Errorf("expected Division Div-1, got %s", model.Division.ValueString())
	}
	if model.Department.ValueString() != "Dept-1" {
		t.Errorf("expected Department Dept-1, got %s", model.Department.ValueString())
	}
	if model.JobTitle.ValueString() != "Teacher" {
		t.Errorf("expected JobTitle Teacher, got %s", model.JobTitle.ValueString())
	}
	if model.StartDateTime.ValueString() != "2024-01-01T00:00:00Z" {
		t.Errorf("expected StartDateTime 2024-01-01T00:00:00Z, got %s", model.StartDateTime.ValueString())
	}
	if model.CreatedDateTime.ValueString() != "2024-01-01T00:00:00Z" {
		t.Errorf("expected CreatedDateTime 2024-01-01T00:00:00Z, got %s", model.CreatedDateTime.ValueString())
	}
	if model.UpdatedDateTime.ValueString() != "2024-06-01T00:00:00Z" {
		t.Errorf("expected UpdatedDateTime 2024-06-01T00:00:00Z, got %s", model.UpdatedDateTime.ValueString())
	}

	if len(model.RoleOuList) != 1 {
		t.Fatalf("expected 1 role_ou_list entry, got %d", len(model.RoleOuList))
	}
	if model.RoleOuList[0].RoleName.ValueString() != "Teacher" {
		t.Errorf("expected role name Teacher, got %s", model.RoleOuList[0].RoleName.ValueString())
	}
	if model.RoleOuList[0].OuID.ValueString() != "ou-1" {
		t.Errorf("expected ou ID ou-1, got %s", model.RoleOuList[0].OuID.ValueString())
	}

	if len(model.PhoneNumbers) != 1 {
		t.Fatalf("expected 1 phone number, got %d", len(model.PhoneNumbers))
	}
	if model.PhoneNumbers[0].PhoneNumber.ValueString() != "555-0100" {
		t.Errorf("expected phone 555-0100, got %s", model.PhoneNumbers[0].PhoneNumber.ValueString())
	}
	if model.PhoneNumbers[0].Type.ValueString() != "WORK" {
		t.Errorf("expected phone type WORK, got %s", model.PhoneNumbers[0].Type.ValueString())
	}
}

func TestFlattenUser_EmptyOptionalFields(t *testing.T) {
	user := client.User{
		Type: "users",
		ID:   "user-2",
		Attributes: client.UserAttributes{
			FirstName:           "John",
			LastName:            "Smith",
			Status:              "ACTIVE",
			ManagedAppleAccount: "john.smith@example.com",
		},
	}

	model := flattenUser(user)

	if !model.MiddleName.IsNull() {
		t.Error("expected MiddleName to be null when empty")
	}
	if !model.Email.IsNull() {
		t.Error("expected Email to be null when empty")
	}
	if !model.EmployeeNumber.IsNull() {
		t.Error("expected EmployeeNumber to be null when empty")
	}
	if !model.CostCenter.IsNull() {
		t.Error("expected CostCenter to be null when empty")
	}
	if !model.Division.IsNull() {
		t.Error("expected Division to be null when empty")
	}
	if !model.Department.IsNull() {
		t.Error("expected Department to be null when empty")
	}
	if !model.JobTitle.IsNull() {
		t.Error("expected JobTitle to be null when empty")
	}
	if !model.StartDateTime.IsNull() {
		t.Error("expected StartDateTime to be null when empty")
	}
	if !model.CreatedDateTime.IsNull() {
		t.Error("expected CreatedDateTime to be null when empty")
	}
	if !model.UpdatedDateTime.IsNull() {
		t.Error("expected UpdatedDateTime to be null when empty")
	}
	if len(model.RoleOuList) != 0 {
		t.Errorf("expected 0 role_ou_list entries, got %d", len(model.RoleOuList))
	}
	if len(model.PhoneNumbers) != 0 {
		t.Errorf("expected 0 phone numbers, got %d", len(model.PhoneNumbers))
	}
}
