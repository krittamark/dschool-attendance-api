package parser

import (
	"testing"
)

func TestParseInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"3,534", 3534},
		{"-", 0},
		{"", 0},
		{"  45  ", 45},
	}

	for _, tt := range tests {
		got := ParseInt(tt.input)
		if got != tt.expected {
			t.Errorf("ParseInt(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestClassToEdLevel(t *testing.T) {
	tests := []struct {
		input       string
		wantEdLevel string
		wantClass   string
	}{
		{"101", "2", "101"},
		{"ม.101", "2", "101"},
		{"ม. 205", "2", "205"},
		{"315", "2", "315"},
		{"401", "3", "401"},
		{"ม.405", "3", "405"},
		{"501", "3", "501"},
		{"617", "3", "617"},
	}

	for _, tt := range tests {
		ed, c := ClassToEdLevel(tt.input)
		if ed != tt.wantEdLevel || c != tt.wantClass {
			t.Errorf("ClassToEdLevel(%q) = (%q, %q), want (%q, %q)", tt.input, ed, c, tt.wantEdLevel, tt.wantClass)
		}
	}
}

func TestFormatDateParam(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-09-08", "69/09/08"},
		{"2026-01-15", "69/01/15"},
		{"15/01/2026", "69/01/15"},
		{"69/09/08", "69/09/08"},
	}

	for _, tt := range tests {
		got := FormatDateParam(tt.input)
		if got != tt.expected {
			t.Errorf("FormatDateParam(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
