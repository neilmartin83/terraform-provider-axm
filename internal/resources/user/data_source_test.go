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

	listAttrs := []string{"role_ou_list", "phone_numbers"}
	for _, name := range listAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q not found", name)
		}
		if _, ok := attr.(dsschema.ListNestedAttribute); !ok {
			t.Errorf("expected attribute %q to be a ListNestedAttribute", name)
		}
	}
}
