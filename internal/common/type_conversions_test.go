package common

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringsToTypesStrings(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  int
	}{
		{name: "empty_slice", input: []string{}, want: 0},
		{name: "nil_slice", input: nil, want: 0},
		{name: "single_element", input: []string{"hello"}, want: 1},
		{name: "multiple_elements", input: []string{"a", "b", "c", "d"}, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringsToTypesStrings(tt.input)
			if len(result) != tt.want {
				t.Fatalf("expected length %d, got %d", tt.want, len(result))
			}
			for i, v := range result {
				if v.ValueString() != tt.input[i] {
					t.Errorf("element[%d]: expected %q, got %q", i, tt.input[i], v.ValueString())
				}
			}
		})
	}
}

func TestStringPointerOrNil(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
	}{
		{name: "empty_string", input: "", wantNil: true},
		{name: "non_empty_string", input: "hello", wantNil: false},
		{name: "whitespace_only", input: "  ", wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StringPointerOrNil(tt.input)
			if tt.wantNil {
				if result != nil {
					t.Fatalf("expected nil, got %q", *result)
				}
				return
			}
			if result == nil {
				t.Fatal("expected non-nil pointer")
			}
			if *result != tt.input {
				t.Fatalf("expected %q, got %q", tt.input, *result)
			}
		})
	}

	_ = types.StringType
}
