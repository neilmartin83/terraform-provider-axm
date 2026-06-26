// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package audit_events_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/audit_events"
)

func TestAuditEventsDataSourceMetadata(t *testing.T) {
	ds := audit_events.NewAuditEventsDataSource()
	resp := datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_audit_events" {
		t.Errorf("expected TypeName %q, got %q", "axm_audit_events", resp.TypeName)
	}
}

func TestAuditEventsDataSourceSchema(t *testing.T) {
	ds := audit_events.NewAuditEventsDataSource()
	resp := datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	requiredAttrs := []string{"start_timestamp", "end_timestamp"}
	for _, name := range requiredAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q not found", name)
		}
		if !attr.IsRequired() {
			t.Errorf("expected attribute %q to be Required", name)
		}
	}

	optionalAttrs := []string{"actor_id", "subject_id", "event_type", "limit", "fields", "cursor"}
	for _, name := range optionalAttrs {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("attribute %q not found", name)
		}
		if !attr.IsOptional() {
			t.Errorf("expected attribute %q to be Optional", name)
		}
	}

	fieldsAttr, ok := resp.Schema.Attributes["fields"].(dsschema.ListAttribute)
	if !ok {
		t.Fatal("expected 'fields' to be a ListAttribute")
	}
	if fieldsAttr.ElementType != types.StringType {
		t.Errorf("expected 'fields' ElementType to be StringType")
	}

	eventsAttr, ok := resp.Schema.Attributes["events"].(dsschema.ListNestedAttribute)
	if !ok {
		t.Fatal("expected 'events' to be a ListNestedAttribute")
	}
	if !eventsAttr.IsComputed() {
		t.Error("expected 'events' to be Computed")
	}

	allNested := []string{
		"id", "type", "event_date_time", "event_type", "category",
		"actor_type", "actor_id", "actor_name", "subject_type",
		"subject_id", "subject_name", "outcome", "group_id",
		"event_data_property_key", "event_data_json",
	}
	for _, name := range allNested {
		if _, ok := eventsAttr.NestedObject.Attributes[name]; !ok {
			t.Errorf("nested attribute %q not found in events", name)
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

func TestAccAuditEventsDataSource(t *testing.T) {
	end := time.Now().UTC().Truncate(24 * time.Hour)
	start := end.Add(-7 * 24 * time.Hour)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					data "axm_audit_events" "test" {
						start_timestamp = %q
						end_timestamp   = %q
						limit           = 1
					}
				`, start.Format(time.RFC3339), end.Format(time.RFC3339)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.axm_audit_events.test", "id"),
				),
			},
		},
	})
}
