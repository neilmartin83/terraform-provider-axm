// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package audit_events_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

	if _, ok := eventsAttr.NestedObject.Attributes["event_data_json"]; !ok {
		t.Error("nested attribute 'event_data_json' not found in events")
	}
}
