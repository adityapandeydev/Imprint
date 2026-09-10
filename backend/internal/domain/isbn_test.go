package domain

import (
	"testing"
)

func TestValidateISBN10(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid ISBN-10 with numeric check digit", "0-306-40615-2", true},
		{"Valid ISBN-10 without hyphens", "0306406152", true},
		{"Valid ISBN-10 with X check digit", "0-8044-2957-X", true},
		{"Valid ISBN-10 lowercase x check digit", "080442957x", true},
		{"Valid ISBN-10 with spaces", "0 306 40615 2", true},
		{"Invalid check digit", "0-306-40615-3", false},
		{"Invalid length too short", "030640615", false},
		{"Invalid length too long", "03064061520", false},
		{"Invalid character in body", "0-306-406A5-2", false},
		{"Empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateISBN10(tc.input)
			if got != tc.expected {
				t.Errorf("ValidateISBN10(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestValidateISBN13(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid ISBN-13 (The Hobbit)", "978-0-261-10221-7", true},
		{"Valid ISBN-13 (Dune)", "978-0-441-01359-3", true},
		{"Valid ISBN-13 without hyphens", "9780261102217", true},
		{"Valid ISBN-13 979 prefix", "979-1-090-63607-1", true},
		{"Invalid check digit", "978-0-261-10221-8", false},
		{"Invalid prefix", "977-0-261-10221-7", false},
		{"Invalid length too short", "978026110221", false},
		{"Invalid length too long", "97802611022177", false},
		{"Invalid non-digit character", "978-0-261-1022A-7", false},
		{"Empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateISBN13(tc.input)
			if got != tc.expected {
				t.Errorf("ValidateISBN13(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestISBN10To13(t *testing.T) {
	tests := []struct {
		name      string
		isbn10    string
		want13    string
		expectErr bool
	}{
		{
			name:      "Valid conversion numeric check digit",
			isbn10:    "0-306-40615-2",
			want13:    "9780306406157",
			expectErr: false,
		},
		{
			name:      "Valid conversion with X check digit",
			isbn10:    "0-8044-2957-X",
			want13:    "9780804429573",
			expectErr: false,
		},
		{
			name:      "Invalid ISBN-10 fails conversion",
			isbn10:    "0-306-40615-9",
			want13:    "",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got13, err := ISBN10To13(tc.isbn10)
			if tc.expectErr && err == nil {
				t.Errorf("ISBN10To13(%q) expected error, got nil", tc.isbn10)
			}
			if !tc.expectErr && err != nil {
				t.Errorf("ISBN10To13(%q) unexpected error: %v", tc.isbn10, err)
			}
			if got13 != tc.want13 {
				t.Errorf("ISBN10To13(%q) = %q, want %q", tc.isbn10, got13, tc.want13)
			}
		})
	}
}

func TestNormalizeISBN(t *testing.T) {
	t.Run("Normalize valid 10 to 13 and 10", func(t *testing.T) {
		isbn13, isbn10, err := NormalizeISBN("0-306-40615-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isbn13 != "9780306406157" {
			t.Errorf("isbn13 = %q, want 9780306406157", isbn13)
		}
		if isbn10 != "0306406152" {
			t.Errorf("isbn10 = %q, want 0306406152", isbn10)
		}
	})

	t.Run("Normalize valid 13 with 978 prefix", func(t *testing.T) {
		isbn13, isbn10, err := NormalizeISBN("978-0-261-10221-7")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if isbn13 != "9780261102217" {
			t.Errorf("isbn13 = %q, want 9780261102217", isbn13)
		}
		if isbn10 != "0261102214" {
			t.Errorf("isbn10 = %q, want 0261102214", isbn10)
		}
	})

	t.Run("Invalid ISBN returns error", func(t *testing.T) {
		_, _, err := NormalizeISBN("invalid-isbn")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestValidateASIN(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid Kindle ASIN starting with B", "B08XYZ1234", true},
		{"Valid 10-char alphanumeric ASIN", "B00B7NPRY8", true},
		{"Invalid length too short", "B08XYZ123", false},
		{"Invalid length too long", "B08XYZ12345", false},
		{"Invalid characters", "B08-XYZ!23", false},
		{"Empty string", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ValidateASIN(tc.input)
			if got != tc.expected {
				t.Errorf("ValidateASIN(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}
