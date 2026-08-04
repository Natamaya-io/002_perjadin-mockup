package utils

import "testing"

func TestFormatRupiah(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		amount   float64
	}{
		{
			name:     "Zero amount",
			amount:   0,
			expected: "-",
		},
		{
			name:     "Hundreds",
			amount:   500,
			expected: "500",
		},
		{
			name:     "Thousands",
			amount:   1500,
			expected: "1.500",
		},
		{
			name:     "Tens of Thousands",
			amount:   50000,
			expected: "50.000",
		},
		{
			name:     "Millions",
			amount:   1250000,
			expected: "1.250.000",
		},
		{
			name:     "Billions",
			amount:   1500000000,
			expected: "1.500.000.000",
		},
		{
			name:     "Decimal truncated",
			amount:   1500.99,
			expected: "1.501", // Notice fmt.Sprintf("%.0f") rounds up .99 to next integer
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatRupiah(tc.amount)
			if result != tc.expected {
				t.Errorf("FormatRupiah(%v) expected %q, got %q", tc.amount, tc.expected, result)
			}
		})
	}
}
