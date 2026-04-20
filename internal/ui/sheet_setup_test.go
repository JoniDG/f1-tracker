package ui

import "testing"

func TestParseSpreadsheetID(t *testing.T) {
	validID := "1AbCdEfGhIjKlMnOpQrStUvWxYz0123456789AbCdEf"

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"valid raw id", validID, validID},
		{"valid google url", "https://docs.google.com/spreadsheets/d/" + validID + "/edit#gid=0", validID},
		{"valid google url without trailing slash", "https://docs.google.com/spreadsheets/d/" + validID, validID},
		{"valid google url with query", "https://docs.google.com/spreadsheets/d/" + validID + "?foo=bar", validID},
		{"non-google url", "https://example.com/algo", ""},
		{"non-google url looking like sheets", "https://algo-que-no-es-google.com/spreadsheets/d/" + validID, ""},
		{"short raw id", "short", ""},
		{"raw id with spaces", "has space in middle", ""},
		{"raw id with special chars", "abc$def!ghi", ""},
		{"google url with bogus short id", "https://docs.google.com/spreadsheets/d/short/edit", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSpreadsheetID(tt.input)
			if got != tt.want {
				t.Errorf("parseSpreadsheetID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
