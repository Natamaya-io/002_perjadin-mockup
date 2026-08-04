package utils

import (
	"fmt"
	"time"
)

// GetRomanMonths returns a slice of strings representing months in Roman numerals.
func GetRomanMonths() []string {
	return []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
}

// GetIndonesianMonths returns a slice of strings representing month names in Indonesian.
func GetIndonesianMonths() []string {
	return []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
}

// FormatIndonesianDate formats a time.Time object into an Indonesian date string (e.g., "31 Maret 2026").
func FormatIndonesianDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	months := GetIndonesianMonths()
	return fmt.Sprintf("%d %s %d", t.Day(), months[int(t.Month())], t.Year())
}
