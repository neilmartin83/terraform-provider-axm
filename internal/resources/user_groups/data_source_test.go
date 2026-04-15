// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package user_groups_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/user_groups"
)

func TestUserGroupsDataSourceMetadata(t *testing.T) {
	ds := user_groups.NewUserGroupsDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_user_groups" {
		t.Errorf("expected TypeName %q, got %q", "axm_user_groups", resp.TypeName)
	}
}

func TestUserGroupsDataSourceSchema(t *testing.T) {
	ds := user_groups.NewUserGroupsDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	groupsAttr, ok := resp.Schema.Attributes["groups"]
	if !ok {
		t.Fatal("attribute 'groups' not found")
	}
	listNested, ok := groupsAttr.(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'groups' to be a ListNestedAttribute")
	}

	expectedNested := []string{
		"id", "type", "ou_id", "name", "group_type", "total_member_count",
		"status", "created_date_time", "updated_date_time",
	}
	nestedAttrs := listNested.NestedObject.Attributes
	for _, name := range expectedNested {
		if _, ok := nestedAttrs[name]; !ok {
			t.Errorf("nested attribute %q not found in groups", name)
		}
	}
}
