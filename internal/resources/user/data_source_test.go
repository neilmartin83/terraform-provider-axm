// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package user_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/user"
)

func TestUserDataSourceMetadata(t *testing.T) {
	ds := user.NewUserDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_user" {
		t.Errorf("expected TypeName %q, got %q", "axm_user", resp.TypeName)
	}
}

func TestUserDataSourceSchema(t *testing.T) {
	ds := user.NewUserDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	idAttr, ok := resp.Schema.Attributes["id"]
	if !ok {
		t.Fatal("attribute 'id' not found")
	}
	if !idAttr.IsRequired() {
		t.Error("expected 'id' to be Required")
	}

	computedAttrs := []string{
		"type", "first_name", "last_name", "middle_name", "status",
		"managed_apple_account", "email", "employee_number", "cost_center",
		"division", "department", "job_title", "start_date_time",
		"created_date_time", "updated_date_time",
	}
	for _, name := range computedAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("attribute %q not found", name)
			continue
		}
		if !attr.IsComputed() {
			t.Errorf("expected attribute %q to be Computed", name)
		}
	}

	isExternalUserAttr, ok := resp.Schema.Attributes["is_external_user"]
	if !ok {
		t.Fatal("attribute 'is_external_user' not found")
	}
	if !isExternalUserAttr.IsComputed() {
		t.Error("expected 'is_external_user' to be Computed")
	}
	if _, ok := isExternalUserAttr.(dsschema.BoolAttribute); !ok {
		t.Error("expected 'is_external_user' to be a BoolAttribute")
	}

	listAttrs := []string{"role_ou_list", "phone_numbers"}
	for _, name := range listAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q not found", name)
		}
		listNested, ok := attr.(dsschema.ListNestedAttribute)
		if !ok {
			t.Errorf("expected attribute %q to be a ListNestedAttribute", name)
			continue
		}
		if !listNested.IsComputed() {
			t.Errorf("expected attribute %q to be Computed", name)
		}
	}

	roleAttr, ok := resp.Schema.Attributes["role_ou_list"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'role_ou_list' to be a ListNestedAttribute")
	}
	for _, name := range []string{"role_name", "ou_id"} {
		if _, ok := roleAttr.NestedObject.Attributes[name]; !ok {
			t.Errorf("nested attribute %q not found in role_ou_list", name)
		}
	}

	phoneAttr, ok := resp.Schema.Attributes["phone_numbers"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'phone_numbers' to be a ListNestedAttribute")
	}
	for _, name := range []string{"phone_number", "type"} {
		if _, ok := phoneAttr.NestedObject.Attributes[name]; !ok {
			t.Errorf("nested attribute %q not found in phone_numbers", name)
		}
	}
}
