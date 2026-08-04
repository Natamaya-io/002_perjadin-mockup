package utils

import (
	"fmt"
	"strings"
)

// FormatRupiah memformat float64 menjadi string mata uang Rupiah tanpa simbol Rp
func FormatRupiah(amount float64) string {
	if amount == 0 {
		return "-"
	}
	s := fmt.Sprintf("%.0f", amount)
	if len(s) <= 3 {
		return s
	}

	var res []string
	for len(s) > 3 {
		res = append([]string{s[len(s)-3:]}, res...)
		s = s[:len(s)-3]
	}
	res = append([]string{s}, res...)
	return strings.Join(res, ".")
}
