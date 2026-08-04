package terbilang

import "strings"

var (
	huruf = []string{"", "Satu", "Dua", "Tiga", "Empat", "Lima", "Enam", "Tujuh", "Delapan", "Sembilan", "Sepuluh", "Sebelas"}
)

func Terbilang(n int) string {
	switch {
	case n < 0:
		return "Minus " + Terbilang(-n)
	case n < 12:
		return huruf[n]
	case n < 20:
		return Terbilang(n-10) + " Belas"
	case n < 100:
		return Terbilang(n/10) + " Puluh " + Terbilang(n%10)
	case n < 200:
		return "Seratus " + Terbilang(n-100)
	case n < 1000:
		return Terbilang(n/100) + " Ratus " + Terbilang(n%100)
	case n < 2000:
		return "Seribu " + Terbilang(n-1000)
	case n < 1000000:
		return Terbilang(n/1000) + " Ribu " + Terbilang(n%1000)
	case n < 1000000000:
		return Terbilang(n/1000000) + " Juta " + Terbilang(n%1000000)
	case n < 1000000000000:
		return Terbilang(n/1000000000) + " Miliar " + Terbilang(n%1000000000)
	default:
		return ""
	}
}

func FormatTerbilang(n int) string {
	// Clean up any double spaces that might occur from the recursive string concatenation
	res := Terbilang(n)
	res = strings.ReplaceAll(res, "  ", " ")
	return strings.TrimSpace(res)
}
