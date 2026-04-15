// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package users_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/users"
)

func TestUsersDataSourceMetadata(t *testing.T) {
	ds := users.NewUsersDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_users" {
		t.Errorf("expected TypeName %q, got %q", "axm_users", resp.TypeName)
	}
}

func TestUsersDataSourceSchema(t *testing.T) {
	ds := users.NewUsersDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("attribute 'id' not found")
	}
	if !idAttr.IsComputed() {
		t.Error("expected 'id' to be Computed")
	}

	usersAttr, ok := resp.Schema.Attributes["users"]
	if !ok {
		t.Fatal("attribute 'users' not found")
	}
	listNested, ok := usersAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'users' to be a ListNestedAttribute")
	}
	if !usersAttr.IsComputed() {
		t.Error("expected 'users' to be Computed")
	}

	nestedAttrs := listNested.NestedObject.Attributes
	expectedNested := []string{
		"id", "type", "first_name", "last_name", "middle_name", "status",
		"managed_apple_account", "is_external_user", "role_ou_list", "email",
		"employee_number", "cost_center", "division", "department", "job_title",
		"start_date_time", "created_date_time", "updated_date_time", "phone_numbers",
	}
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in users", name)
		}
	}

	roleAttr, ok := nestedAttrs["role_ou_list"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'role_ou_list' to be a ListNestedAttribute")
	}
	for _, name := range []string{"role_name", "ou_id"} {
		if _, ok := roleAttr.NestedObject.Attributes[name]; !ok {
			t.Errorf("nested attribute %q not found in role_ou_list", name)
		}
	}

	phoneAttr, ok := nestedAttrs["phone_numbers"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'phone_numbers' to be a ListNestedAttribute")
	}
	for _, name := range []string{"phone_number", "type"} {
		if _, ok := phoneAttr.NestedObject.Attributes[name]; !ok {
			t.Errorf("nested attribute %q not found in phone_numbers", name)
		}
	}
}
