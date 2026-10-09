package chatbot

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var ErrForbidden = errors.New("akses analisis tidak diizinkan")

type Service interface {
	GetGUPAnalysis(ctx context.Context, filter PeriodFilter) (*GUPAnalysis, error)
	GetOutsideCityAnalysis(ctx context.Context, filter PeriodFilter, role, userID string) (*OutsideCityAnalysis, error)
	GetInsideCityAnalysis(ctx context.Context, filter PeriodFilter, role, userID string) (*InsideCityAnalysis, error)
	HandleMessage(ctx context.Context, message, role, userID string) (*ChatMessageResponse, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func validatePeriod(filter PeriodFilter) error {
	if filter.Year < 2000 || filter.Year > 2100 {
		return fmt.Errorf("tahun analisis tidak valid")
	}
	if filter.Month < 0 || filter.Month > 12 {
		return fmt.Errorf("bulan analisis tidak valid")
	}
	return nil
}

func periodInfo(filter PeriodFilter) PeriodInfo {
	label := fmt.Sprintf("Tahun %d", filter.Year)
	if filter.Month > 0 {
		months := []string{
			"", "Januari", "Februari", "Maret", "April", "Mei", "Juni",
			"Juli", "Agustus", "September", "Oktober", "November", "Desember",
		}
		label = fmt.Sprintf("%s %d", months[filter.Month], filter.Year)
	}
	return PeriodInfo{Year: filter.Year, Month: filter.Month, Label: label}
}

func scopeForRole(role, userID string) (*uuid.UUID, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role == "super_admin" || role == "kasubag" {
		return nil, nil
	}
	if role != "protokol" {
		return nil, ErrForbidden
	}
	id, err := uuid.Parse(strings.TrimSpace(userID))
	if err != nil || id == uuid.Nil {
		return nil, fmt.Errorf("ID pengguna tidak valid")
	}
	return &id, nil
}

func canAccessGUP(role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	return role == "super_admin" || role == "kasubag"
}

func (s *service) GetGUPAnalysis(ctx context.Context, filter PeriodFilter) (*GUPAnalysis, error) {
	if err := validatePeriod(filter); err != nil {
		return nil, err
	}
	result, err := s.repo.GetGUPAnalysis(ctx, filter)
	if err != nil {
		return nil, err
	}
	result.Period = periodInfo(filter)
	return result, nil
}

func (s *service) GetOutsideCityAnalysis(ctx context.Context, filter PeriodFilter, role, userID string) (*OutsideCityAnalysis, error) {
	if err := validatePeriod(filter); err != nil {
		return nil, err
	}
	scope, err := scopeForRole(role, userID)
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetOutsideCityAnalysis(ctx, filter, scope)
	if err != nil {
		return nil, err
	}
	result.Period = periodInfo(filter)
	return result, nil
}

func (s *service) GetInsideCityAnalysis(ctx context.Context, filter PeriodFilter, role, userID string) (*InsideCityAnalysis, error) {
	if err := validatePeriod(filter); err != nil {
		return nil, err
	}
	scope, err := scopeForRole(role, userID)
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetInsideCityAnalysis(ctx, filter, scope)
	if err != nil {
		return nil, err
	}
	result.Period = periodInfo(filter)
	return result, nil
}

var yearPattern = regexp.MustCompile(`\b(20\d{2})\b`)

func normalizeCommand(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = strings.ReplaceAll(value, "_", " ")
	value = strings.ReplaceAll(value, "-", " ")
	value = strings.Join(strings.Fields(value), " ")
	return value
}

func periodFromMessage(message string) PeriodFilter {
	filter := PeriodFilter{Year: time.Now().Year(), Month: 0}
	normalized := normalizeCommand(message)

	if match := yearPattern.FindStringSubmatch(normalized); len(match) == 2 {
		var parsed int
		_, _ = fmt.Sscanf(match[1], "%d", &parsed)
		if parsed > 0 {
			filter.Year = parsed
		}
	}

	months := []struct {
		name  string
		value int
	}{
		{"januari", 1}, {"februari", 2}, {"maret", 3}, {"april", 4},
		{"mei", 5}, {"juni", 6}, {"juli", 7}, {"agustus", 8},
		{"september", 9}, {"oktober", 10}, {"november", 11}, {"desember", 12},
	}
	for _, month := range months {
		if strings.Contains(normalized, month.name) {
			filter.Month = month.value
			break
		}
	}
	return filter
}

func containsAny(value string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(value, word) {
			return true
		}
	}
	return false
}

func gupSuggestions() []string {
	return []string{
		"Analisa Keuangan GUP",
		"Berikan 3 ranking pengeluaran GUP terbesar dan pengeluaran terkecil",
		"Laporan LS yang sudah digunakan rekap setiap bulan",
		"Dari total LS yang tersedia, pengeluaran terbesar dari akun apa",
		"Tampilkan chart batang pemasukan, realisasi, dan balance",
	}
}

func outsideSuggestions() []string {
	return []string{
		"Analisa Keuangan Perjalanan Dinas Luar Kota",
		"Jumlahkan masing masing user Protokol untuk perjalanan Luar Kota dan berikan jumlah statusnya",
	}
}

func insideSuggestions() []string {
	return []string{
		"Analisa Keuangan Perjalanan Dinas Dalam Kota",
		"Jumlahkan masing masing user Protokol untuk perjalanan Dalam Kota dan berikan jumlah statusnya",
	}
}

// HandleMessage is intentionally deterministic. The command is interpreted here,
// then all financial calculations are executed by the analytics repository.
func (s *service) HandleMessage(ctx context.Context, message, role, userID string) (*ChatMessageResponse, error) {
	normalized := normalizeCommand(message)
	if normalized == "" {
		return &ChatMessageResponse{
			Intent:  "empty",
			Message: "Silakan ketik perintah analisis terlebih dahulu.",
		}, nil
	}

	// Menu commands only return available commands. No financial query is executed.
	if normalized == "#menu gup" || normalized == "menu gup" {
		if !canAccessGUP(role) {
			return nil, ErrForbidden
		}
		return &ChatMessageResponse{
			Intent:      "menu_gup",
			Message:     "Menu GUP tersedia. Pilih atau ketik salah satu perintah berikut.",
			Suggestions: gupSuggestions(),
		}, nil
	}
	if normalized == "#menu luar kota" || normalized == "menu luar kota" {
		if _, err := scopeForRole(role, userID); err != nil {
			return nil, err
		}
		return &ChatMessageResponse{
			Intent:      "menu_luar_kota",
			Message:     "Menu Luar Kota tersedia. Pilih atau ketik salah satu perintah berikut.",
			Suggestions: outsideSuggestions(),
		}, nil
	}
	if normalized == "#menu dalam kota" || normalized == "menu dalam kota" || normalized == "#menu dalkot" {
		if _, err := scopeForRole(role, userID); err != nil {
			return nil, err
		}
		return &ChatMessageResponse{
			Intent:      "menu_dalam_kota",
			Message:     "Menu Dalam Kota tersedia. Pilih atau ketik salah satu perintah berikut.",
			Suggestions: insideSuggestions(),
		}, nil
	}

	filter := periodFromMessage(message)

	isOutside := containsAny(normalized, "luar kota", "luarkota")
	isInside := containsAny(normalized, "dalam kota", "dalkot")
	isGUP := containsAny(normalized, "gup", "laporan ls", "total ls", "pemasukan", "akun apa") && !isOutside && !isInside

	if isGUP {
		if !canAccessGUP(role) {
			return nil, ErrForbidden
		}
		data, err := s.GetGUPAnalysis(ctx, filter)
		if err != nil {
			return nil, err
		}

		intent := "gup_analysis"
		messageText := fmt.Sprintf("Analisis keuangan GUP untuk %s berhasil dibuat.", data.Period.Label)
		switch {
		case containsAny(normalized, "chart", "grafik", "batang"):
			intent = "gup_chart"
			messageText = fmt.Sprintf("Berikut chart pemasukan, realisasi, dan balance GUP untuk %s.", data.Period.Label)
		case containsAny(normalized, "akun") && containsAny(normalized, "terbesar", "tertinggi", "paling besar"):
			intent = "gup_account"
			messageText = fmt.Sprintf("Berikut rekap akun dengan realisasi GUP terbesar untuk %s.", data.Period.Label)
		case containsAny(normalized, "laporan ls", "rekap ls", "ls yang sudah", "ls digunakan", "ls yang digunakan"):
			intent = "gup_ls"
			messageText = fmt.Sprintf("Berikut rekap penggunaan LS untuk %s.", data.Period.Label)
		case containsAny(normalized, "rangking", "ranking", "peringkat"):
			intent = "gup_ranking"
			messageText = fmt.Sprintf("Berikut ranking pengeluaran GUP terbesar dan terkecil untuk %s.", data.Period.Label)
		}
		return &ChatMessageResponse{Intent: intent, Message: messageText, Data: data}, nil
	}

	if isOutside {
		data, err := s.GetOutsideCityAnalysis(ctx, filter, role, userID)
		if err != nil {
			return nil, err
		}
		return &ChatMessageResponse{
			Intent:  "luar_kota_analysis",
			Message: fmt.Sprintf("Analisis perjalanan dinas Luar Kota untuk %s berhasil dibuat.", data.Period.Label),
			Data:    data,
		}, nil
	}

	if isInside {
		data, err := s.GetInsideCityAnalysis(ctx, filter, role, userID)
		if err != nil {
			return nil, err
		}
		return &ChatMessageResponse{
			Intent:  "dalam_kota_analysis",
			Message: fmt.Sprintf("Analisis perjalanan dinas Dalam Kota untuk %s berhasil dibuat.", data.Period.Label),
			Data:    data,
		}, nil
	}

	return &ChatMessageResponse{
		Intent:  "unknown",
		Message: "Perintah belum dikenali. Ketik #Menu-GUP, #Menu-Luar_Kota, atau #Menu-Dalam_Kota untuk melihat perintah yang tersedia.",
		Suggestions: []string{
			"#Menu-GUP",
			"#Menu-Luar_Kota",
			"#Menu-Dalam_Kota",
		},
	}, nil
}
