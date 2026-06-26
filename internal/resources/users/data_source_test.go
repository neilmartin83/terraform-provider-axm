// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package users_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
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

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"axm": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}
	for _, envVar := range []string{"AXM_CLIENT_ID", "AXM_KEY_ID", "AXM_PRIVATE_KEY", "AXM_SCOPE"} {
		if os.Getenv(envVar) == "" {
			t.Skipf("%s must be set for acceptance tests", envVar)
		}
	}
}

func TestAccUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "axm_users" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.axm_users.all", "id"),
				),
			},
		},
	})
}
