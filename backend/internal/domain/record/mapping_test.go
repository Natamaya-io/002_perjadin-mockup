package record

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHandler_MapTravelToDocument_DatePrioritization(t *testing.T) {
	mockSvc := new(MockService)
	cfg := &config.Config{
		Signatory: config.SignatoryConfig{
			PPKName:        "Arief Hafidiyanto",
			PPKNIP:         "19720827 200312 1 002",
			BendaharaName:  "Bendahara",
			BendaharaNIP:   "12345",
		},
	}
	handler := NewHandler(mockSvc, nil, cfg, nil)

	// Base trip dates
	startDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC) // March (III)
	endDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	// Surat Tugas date (different month)
	stDate := time.Date(2026, 4, 5, 0, 0, 0, 0, time.UTC) // April (IV)

	record := &models.TravelRecord{
		Base: models.Base{
			ID: uuid.New(),
		},
		SPDNumber:        "ID-SPJ-001",
		StartDate:        startDate,
		EndDate:          endDate,
		SuratTugasNumber: "ST/123/2026",
		SuratTugasDate:   stDate,
		Employee: models.User{
			Name: "Test Employee",
			NIP:  "19900101",
		},
	}

	// GetRecords is no longer called in mapTravelToDocument
	vars := handler.mapTravelToDocument(record, 1)

	// Assertions for ST Date prioritization
	// 1. bulan_romawi_st should be "IV" (April) not "III" (March)
	assert.Equal(t, "IV", vars["bulan_romawi_st"], "bulan_romawi_st should follow ST date")
	
	// 2. bulan_romawi should also follow refDate which is now stDate
	assert.Equal(t, "IV", vars["bulan_romawi"], "bulan_romawi should follow refDate (ST date)")

	// 3. tanggal_no_surat should be "5 April 2026"
	assert.Contains(t, vars["tanggal_no_surat"], "5 April 2026")

	// 4. new variables
	assert.Contains(t, vars["tanggal_st"], "5 April 2026")
	assert.Contains(t, vars["tgl_st"], "5 April 2026")
	
	// 5. Trip dates should remain correct
	assert.Contains(t, vars["tanggal_berangkat"], "10 Maret 2026")
	assert.Contains(t, vars["tanggal_selesai"], "15 Maret 2026")
}

func TestHandler_MapTravelToDocument_FallbackToStartDate(t *testing.T) {
	mockSvc := new(MockService)
	handler := NewHandler(mockSvc, nil, &config.Config{}, nil)

	startDate := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	record := &models.TravelRecord{
		Base: models.Base{
			ID: uuid.New(),
		},
		SPDNumber: "ID-SPJ-001",
		StartDate: startDate,
		// No ST Date
	}

	vars := handler.mapTravelToDocument(record, 1)

	// Should fallback to StartDate
	assert.Equal(t, "III", vars["bulan_romawi_st"], "Should fallback to March (III)")
	assert.Equal(t, "III", vars["bulan_romawi"], "Should fallback to March (III)")
}
