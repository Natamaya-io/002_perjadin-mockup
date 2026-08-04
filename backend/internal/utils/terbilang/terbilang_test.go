package terbilang

import "testing"

func TestTerbilang(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		number   int
	}{
		{name: "Zero", number: 0, expected: ""}, // Based on the current logic, 0 returns ""
		{name: "Single digit", number: 5, expected: "Lima"},
		{name: "Eleven", number: 11, expected: "Sebelas"},
		{name: "Teens", number: 15, expected: "Lima Belas"},
		{name: "Tens", number: 20, expected: "Dua Puluh "},
		{name: "Tens with units", number: 45, expected: "Empat Puluh Lima"},
		{name: "One Hundred", number: 100, expected: "Seratus "},
		{name: "Hundreds with tens", number: 150, expected: "Seratus Lima Puluh "},
		{name: "Hundreds with units", number: 505, expected: "Lima Ratus Lima"},
		{name: "One Thousand", number: 1000, expected: "Seribu "},
		{name: "Thousands", number: 2026, expected: "Dua Ribu Dua Puluh Enam"},
		{name: "Millions", number: 1500000, expected: "Satu Juta Lima Ratus  Ribu "}, // Kept double space as this is raw unformatted output
		{name: "Negative", number: -15, expected: "Minus Lima Belas"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := Terbilang(tc.number)
			if result != tc.expected {
				t.Errorf("Terbilang(%d) expected %q, got %q", tc.number, tc.expected, result)
			}
		})
	}
}

func TestFormatTerbilang(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		number   int
	}{
		{name: "Tens trim space", number: 20, expected: "Dua Puluh"},
		{name: "Hundreds trim space", number: 100, expected: "Seratus"},
		{name: "Thousands trim space", number: 1000, expected: "Seribu"},
		{name: "Millions trim space", number: 1500000, expected: "Satu Juta Lima Ratus Ribu"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatTerbilang(tc.number)
			if result != tc.expected {
				t.Errorf("FormatTerbilang(%d) expected %q, got %q", tc.number, tc.expected, result)
			}
		})
	}
}
