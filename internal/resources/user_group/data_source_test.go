// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package user_group_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/neilmartin83/terraform-provider-axm/internal/resources/user_group"
)

func TestUserGroupDataSourceMetadata(t *testing.T) {
	ds := user_group.NewUserGroupDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_user_group" {
		t.Errorf("expected TypeName %q, got %q", "axm_user_group", resp.TypeName)
	}
}

func TestUserGroupDataSourceSchema(t *testing.T) {
	ds := user_group.NewUserGroupDataSource()
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

	userIDsAttr, ok := resp.Schema.Attributes["user_ids"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'user_ids' to be a ListAttribute")
	}
	if userIDsAttr.ElementType != types.StringType {
		t.Errorf("expected 'user_ids' ElementType to be StringType")
	}
}
