package common

import (
	"testing"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
)

func TestConfigureClient_NilProviderData(t *testing.T) {
	c, diags := ConfigureClient(nil, "Data Source")
	if diags.HasError() {
		t.Fatalf("expected no errors, got %v", diags.Errors())
	}
	if c != nil {
		t.Fatalf("expected nil client, got %v", c)
	}
}

func TestConfigureClient_WrongType(t *testing.T) {
	c, diags := ConfigureClient("wrong-type", "Data Source")
	if !diags.HasError() {
		t.Fatal("expected errors, got none")
	}
	if c != nil {
		t.Fatalf("expected nil client, got %v", c)
	}
	found := false
	for _, d := range diags.Errors() {
		if d.Detail() != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected at least one error with detail")
	}
}

func TestConfigureClient_ValidClient(t *testing.T) {
	expected := &client.Client{}
	c, diags := ConfigureClient(expected, "Resource")
	if diags.HasError() {
		t.Fatalf("expected no errors, got %v", diags.Errors())
	}
	if c != expected {
		t.Fatal("expected same client pointer")
	}
}
