package utils

import (
	"testing"
	"time"
)

func TestGetRomanMonths(t *testing.T) {
	months := GetRomanMonths()
	if len(months) != 13 {
		t.Errorf("Expected length 13, got %d", len(months))
	}
	if months[1] != "I" {
		t.Errorf("Expected 'I' at index 1, got %s", months[1])
	}
	if months[12] != "XII" {
		t.Errorf("Expected 'XII' at index 12, got %s", months[12])
	}
}

func TestGetIndonesianMonths(t *testing.T) {
	months := GetIndonesianMonths()
	if len(months) != 13 {
		t.Errorf("Expected length 13, got %d", len(months))
	}
	if months[1] != "Januari" {
		t.Errorf("Expected 'Januari' at index 1, got %s", months[1])
	}
	if months[8] != "Agustus" {
		t.Errorf("Expected 'Agustus' at index 8, got %s", months[8])
	}
}

func TestFormatIndonesianDate(t *testing.T) {
	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "Zero Time",
			date:     time.Time{},
			expected: "-",
		},
		{
			name:     "Independence Day",
			date:     time.Date(1945, time.August, 17, 10, 0, 0, 0, time.UTC),
			expected: "17 Agustus 1945",
		},
		{
			name:     "New Year",
			date:     time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			expected: "1 Januari 2025",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatIndonesianDate(tc.date)
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}
