package record

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"
)

// ImportResult holds the summary of an import operation
type ImportResult struct {
	Details   []ImportDetail `json:"details"`
	TotalRows int            `json:"totalRows"`
	Imported  int            `json:"imported"`
	Updated   int            `json:"updated"`
	Skipped   int            `json:"skipped"`
	Failed    int            `json:"failed"`
}

type ImportDetail struct {
	SPJID   string `json:"spjId"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Row     int    `json:"row"`
}

// ImportExcel handles bulk import of travel records from an Excel file
func (h *Handler) ImportExcel(c echo.Context) error {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "File tidak ditemukan. Silakan upload file .xlsx")
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".xlsx") {
		return echo.NewHTTPError(http.StatusBadRequest, "Format file harus .xlsx")
	}

	src, err := file.Open()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal membuka file")
	}
	defer src.Close()

	// Parse Excel
	f, err := excelize.OpenReader(src)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Gagal membaca file Excel: %v", err))
	}
	defer f.Close()

	// Get the first sheet
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "File Excel kosong")
	}

	// Use RawCellValue to bypass local formatting like "Rp", "." and "," for thousands/decimals
	rows, err := f.GetRows(sheetName, excelize.Options{RawCellValue: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Gagal membaca data: %v", err))
	}

	if len(rows) < 3 {
		return echo.NewHTTPError(http.StatusBadRequest, "File Excel tidak memiliki data (minimal 3 baris: header + data)")
	}

	// Get creator ID from JWT (the logged-in superadmin)
	creatorID := uuid.Nil
	if creatorIDStr, ok := c.Get("user_id").(string /*nolint:errcheck*/); ok && creatorIDStr != "" {
		if cid, parseErr := uuid.Parse(creatorIDStr); parseErr == nil {
			creatorID = cid
		}
	}
	if creatorID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User tidak terautentikasi")
	}

	// Load all users and build name->UUID map (case-insensitive)
	allUsers, err := h.userRepo.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal memuat data pegawai")
	}
	userMap := make(map[string]uuid.UUID)
	for _, u := range allUsers {
		normalizedName := strings.TrimSpace(strings.ToLower(u.Name))
		userMap[normalizedName] = u.ID
	}

	// Load existing records to check duplicates (by SPD number + employee name)
	existingRecords, err := h.svc.GetRecords(c.Request().Context(), map[string]interface{}{})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal memuat data record existing")
	}
	// Map key: "spd|name" → existing TravelRecord (for update if costs are 0)
	existingMap := make(map[string]models.TravelRecord)
	for _, r := range existingRecords {
		key := fmt.Sprintf("%s|%s", strings.ToLower(r.SPDNumber), strings.ToLower(r.Employee.Name))
		existingMap[key] = r
	}

	result := ImportResult{
		Details: []ImportDetail{},
	}

	// Process each data row (skip row 1=empty, row 2=headers, start from row 3 = index 2)
	for rowIdx := 2; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		result.TotalRows++

		// Helper to safely get cell value
		getCell := func(colIdx int) string {
			if colIdx < len(row) {
				return strings.TrimSpace(row[colIdx])
			}
			return ""
		}

		// Column mapping (0-indexed):
		// A=0: ID SPJ, B=1: No, C=2: Nama, D=3: Daerah
		// F=5: Tiket, G=6: Jumlah Hari, I=8: Uang Harian
		// K=10: Jml Hari Menginap, L=11: Hotel
		// N=13: Transport Lokal, O=14: Transport Daerah, P=15: Jumlah
		// R=17: KETERANGAN, S=18: Provinsi
		// V=21: Alat Angkut, W=22: Mulai, X=23: Selesai
		// AB=27: Status

		idSPJ := getCell(0)
		if idSPJ == "" {
			continue // Skip empty rows
		}

		name := getCell(2)
		daerah := getCell(3)
		statusCol := getCell(27)
		province := getCell(18)
		purpose := getCell(17)
		transportMode := getCell(21)

		rowNum := rowIdx + 1 // 1-indexed for user display

		isTHR := strings.ToUpper(strings.TrimSpace(statusCol)) == "THR"

		// Check if user exists
		normalizedName := strings.TrimSpace(strings.ToLower(name))
		employeeID, userFound := userMap[normalizedName]
		if !userFound {
			// Auto create alumni user
			username := strings.ReplaceAll(normalizedName, " ", "") + "_alumni"
			
			// Cek apakah email/username ini sudah ada di database
			if existingUser, err := h.userRepo.GetUserByEmail(username); err == nil && existingUser != nil {
				employeeID = existingUser.ID
				userMap[normalizedName] = employeeID
			} else {
				// Just ensure it's a valid string for the Email/Username field
				newUser := models.User{
					Base:  models.Base{ID: uuid.New()},
					Name:  name,
					Email: username, // The system uses Email field as Username
					Role:  "alumni_staff",
				}
				
				if err := h.userRepo.CreateUser(&newUser); err == nil {
					employeeID = newUser.ID
					userMap[normalizedName] = employeeID
				} else {
					// Fallback mutlak: Jika masih bentrok (mungkin karena race condition atau keunikan tersembunyi)
					newUser.Email = username + "_" + strings.ReplaceAll(uuid.New().String()[:8], "-", "")
					if errFallback := h.userRepo.CreateUser(&newUser); errFallback == nil {
						employeeID = newUser.ID
						userMap[normalizedName] = employeeID
					} else {
						result.Failed++
						result.Details = append(result.Details, ImportDetail{
							Row: rowNum, SPJID: idSPJ, Name: name,
							Status: "failed", Message: "Gagal membuat user alumni otomatis: " + errFallback.Error(),
						})
						continue
					}
				}
			}
		}

		// Check duplicates — if already exists, UPDATE costs instead of skipping
		dupKey := fmt.Sprintf("%s|%s", strings.ToLower(idSPJ), normalizedName)
		if existingRec, isDuplicate := existingMap[dupKey]; isDuplicate {
			// Parse cost fields for the update
			ticketCostUpd := safeParseFloat(getCell(5))
			dailyRateUpd := safeParseFloat(getCell(8))
			dailyDaysUpd := safeParseInt(getCell(6))
			hotelNightsUpd := safeParseInt(getCell(10))
			hotelRateUpd := safeParseFloat(getCell(11))
			localTransportUpd := safeParseFloat(getCell(13))
			regionalTransportUpd := safeParseFloat(getCell(14))
			totalCostUpd := safeParseFloat(getCell(15))
			transportModeUpd := getCell(21)

			var additionalCostsUpd []map[string]interface{}
			if localTransportUpd > 0 {
				additionalCostsUpd = append(additionalCostsUpd, map[string]interface{}{
					"name":   "Transport Lokal",
					"amount": localTransportUpd,
					"file":   nil,
				})
			} else {
				additionalCostsUpd = make([]map[string]interface{}, 0)
			}

			detailsArrUpd := []map[string]interface{}{
				{
					"transportMode":   transportModeUpd,
					"ticketGo":        ticketCostUpd,
					"ticketBack":      0,
					"hotelDays":       hotelNightsUpd,
					"hotelRate":       hotelRateUpd,
					"transportAmount": regionalTransportUpd,
					"additionalCosts": additionalCostsUpd,
				},
			}
			detailsBytesUpd, errM := json.Marshal(detailsArrUpd)
			if errM != nil {
				detailsBytesUpd = []byte("[]")
			}

			existingRec.TotalCost = totalCostUpd
			existingRec.Cost = &models.TravelCost{
				TicketGo:           ticketCostUpd,
				DailyAllowanceDays: dailyDaysUpd,
				DailyAllowanceRate: dailyRateUpd,
				HotelDays:          hotelNightsUpd,
				HotelRate:          hotelRateUpd,
				LocalTransport:     localTransportUpd,
				RegionalTransport:  regionalTransportUpd,
				TransportAmount:    regionalTransportUpd,
				TransportMode:      transportModeUpd,
				Details:            detailsBytesUpd,
			}

			if err := h.svc.UpdateRecord(c.Request().Context(), &existingRec); err != nil {
				result.Failed++
				result.Details = append(result.Details, ImportDetail{
					Row: rowNum, SPJID: idSPJ, Name: name,
					Status: "failed", Message: "Gagal update biaya: " + err.Error(),
				})
			} else {
				result.Updated++
				result.Details = append(result.Details, ImportDetail{
					Row: rowNum, SPJID: idSPJ, Name: name,
					Status: "updated", Message: "Biaya berhasil diperbarui dari Excel",
				})
				// Update the map to prevent double-update in same batch
				existingMap[dupKey] = existingRec
			}
			continue
		}

		// Parse dates - since we use RawCellValue, we can actually read the raw unformatted serial or standard text
		colW, errW := excelize.ColumnNumberToName(23)
		if errW != nil {
			colW = "W"
		}
		colX, errX := excelize.ColumnNumberToName(24)
		if errX != nil {
			colX = "X"
		}
		startDate := readExcelDate(f, sheetName, fmt.Sprintf("%s%d", colW, rowNum))
		endDate := readExcelDate(f, sheetName, fmt.Sprintf("%s%d", colX, rowNum))

		if startDate.IsZero() || endDate.IsZero() {
			result.Failed++
			result.Details = append(result.Details, ImportDetail{
				Row: rowNum, SPJID: idSPJ, Name: name,
				Status: "failed", Message: "Tanggal mulai atau selesai tidak valid",
			})
			continue
		}

		// Parse cost fields
		ticketCost := safeParseFloat(getCell(5))         // F: Tiket
		dailyRate := safeParseFloat(getCell(8))          // I: Uang Harian
		dailyDays := safeParseInt(getCell(6))            // G: Jumlah Hari
		hotelNights := safeParseInt(getCell(10))         // K: Jml Hari Menginap
		hotelRate := safeParseFloat(getCell(11))         // L: Hotel
		localTransport := safeParseFloat(getCell(13))    // N: Transport Lokal
		regionalTransport := safeParseFloat(getCell(14)) // O: Transport Daerah
		totalCost := safeParseFloat(getCell(15))         // P: Jumlah

		// Check if cell has a background fill color (e.g. red highlight = incomplete)
		cellA := fmt.Sprintf("A%d", rowNum)
		styleID, err := f.GetCellStyle(sheetName, cellA)

		recordStatus := "Approved"
		reportStatus := "Completed"
		paymentStatus := "Paid"
		recordPurpose := purpose

		if isTHR {
			if recordPurpose == "" {
				recordPurpose = "Pembayaran THR"
			} else {
				recordPurpose += " (THR)"
			}
		} else {
			if err == nil {
				style, getStyleErr := f.GetStyle(styleID)
				// Pattern 1 is solid fill
				if getStyleErr == nil && style != nil && style.Fill.Pattern == 1 && len(style.Fill.Color) > 0 {
					color := strings.ToUpper(style.Fill.Color[0])
					// Ignore white/black
					if !strings.HasSuffix(color, "FFFFFF") && !strings.HasSuffix(color, "000000") {
						recordStatus = "Assigned"
						reportStatus = "Pending"
						paymentStatus = "Pending"
					}
				}
			}
		}

		// Create the travel record (bypass service.CreateRecord to skip overlap check & WA notification)
		record := &models.TravelRecord{
			SPDNumber:     idSPJ,
			EmployeeID:    employeeID,
			CreatorID:     creatorID,
			StartDate:     startDate,
			EndDate:       endDate,
			Location:      daerah,
			Province:      province,
			Type:          "luar_kota",
			Purpose:       recordPurpose,
			Stakeholder:   "",
			Agenda:        "",
			Status:        recordStatus,
			ReportStatus:  reportStatus,
			PaymentStatus: paymentStatus,
			TotalCost:     totalCost,
			Locations: []models.TravelLocation{
				{
					Location:  daerah,
					Province:  province,
					StartDate: startDate,
					EndDate:   endDate,
				},
			},
		}

		// Sync with SvelteKit array structures (1:1 UI data match)
		var additionalCosts []map[string]interface{}
		if localTransport > 0 {
			additionalCosts = append(additionalCosts, map[string]interface{}{
				"name":   "Transport Lokal",
				"amount": localTransport,
				"file":   nil,
			})
		} else {
			// SvelteKit expects an empty array rather than null
			additionalCosts = make([]map[string]interface{}, 0)
		}

		detailsArr := []map[string]interface{}{
			{
				"transportMode":   transportMode,
				"ticketGo":        ticketCost,
				"ticketBack":      0,
				"hotelDays":       hotelNights,
				"hotelRate":       hotelRate,
				"transportAmount": regionalTransport,
				"additionalCosts": additionalCosts,
			},
		}

		detailsBytes, errM := json.Marshal(detailsArr)
		if errM != nil {
			detailsBytes = []byte("[]")
		}

		record.Cost = &models.TravelCost{
			TicketGo:           ticketCost,
			DailyAllowanceDays: dailyDays,
			DailyAllowanceRate: dailyRate,
			HotelDays:          hotelNights,
			HotelRate:          hotelRate,
			LocalTransport:     localTransport,
			RegionalTransport:  regionalTransport,
			TransportAmount:    regionalTransport, // Make backend standard variables match too
			TransportMode:      transportMode,
			Details:            detailsBytes,
		}

		// Use CreateRecordDirect to bypass overlap check and WA notification
		if err := h.svc.CreateRecordDirect(c.Request().Context(), record); err != nil {
			result.Failed++
			result.Details = append(result.Details, ImportDetail{
				Row: rowNum, SPJID: idSPJ, Name: name,
				Status: "failed", Message: fmt.Sprintf("Gagal menyimpan: %v", err),
			})
			continue
		}

		// Mark as existing to prevent duplicate within same import batch
		existingMap[dupKey] = *record

		result.Imported++
		result.Details = append(result.Details, ImportDetail{
			Row: rowNum, SPJID: idSPJ, Name: name,
			Status: "imported", Message: "Berhasil diimport",
		})
	}

	// Invalidate cache once at the end
	h.svc.InvalidateAllCache(c.Request().Context())

	return c.JSON(http.StatusOK, result)
}

// readExcelDate reads a date value from an Excel cell reference (e.g. "W3")
func readExcelDate(f *excelize.File, sheet, cellRef string) time.Time {
	rawVal, err := f.GetCellValue(sheet, cellRef, excelize.Options{RawCellValue: true})
	if err != nil || rawVal == "" {
		return time.Time{}
	}

	// Try to parse as Excel serial date number
	serial, err := strconv.ParseFloat(rawVal, 64)
	if err == nil && serial > 0 {
		t, err := excelize.ExcelDateToTime(serial, false)
		if err == nil {
			return t
		}
	}

	formattedVal, valErr := f.GetCellValue(sheet, cellRef)
	if valErr != nil || formattedVal == "" {
		return time.Time{}
	}

	formats := []string{
		"02-01-2006",
		"02/01/2006",
		"02-01-06",
		"02/01/06",
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"02-Jan-2006",
		"01-02-06",
		"1/2/2006",
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, formattedVal); err == nil {
			return t
		}
	}

	return time.Time{}
}

// safeParseFloat safely parses a string to float64.
// Handles both standard format ("386000") and Indonesian/Excel format:
// - Dots as thousands separators: "386.000" → 386000
// - Comma as decimal separator: "1.234,56" → 1234.56
// - Currency prefix: "Rp430.000" → 430000
func safeParseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0
	}

	// Remove currency prefix/suffix (Rp, IDR, etc.)
	s = strings.TrimPrefix(s, "Rp")
	s = strings.TrimPrefix(s, "IDR")
	s = strings.TrimSpace(s)

	// If no comma: dots are thousands separators (Indonesian format: "386.000", "1.421.370")
	// If has comma: dot=thousands, comma=decimal ("1.234,56")
	if strings.Contains(s, ",") {
		// Remove dots (thousands) then replace comma with dot (decimal)
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else {
		// Count dots: if more than one, or if dot is followed by exactly 3 digits at the end → thousands sep
		dotCount := strings.Count(s, ".")
		if dotCount > 1 {
			// Multiple dots → all are thousands separators
			s = strings.ReplaceAll(s, ".", "")
		} else if dotCount == 1 {
			// Single dot: check if it's a thousands separator (followed by exactly 3 digits)
			parts := strings.Split(s, ".")
			if len(parts[1]) == 3 {
				// e.g. "386.000" → thousands separator
				s = strings.ReplaceAll(s, ".", "")
			}
			// else: "386.5" → keep as decimal point
		}
	}

	// Remove any remaining non-numeric characters except dot and minus
	var cleaned strings.Builder
	for i, ch := range s {
		if ch >= '0' && ch <= '9' {
			cleaned.WriteRune(ch)
		} else if ch == '.' || (ch == '-' && i == 0) {
			cleaned.WriteRune(ch)
		}
	}
	s = cleaned.String()

	if s == "" {
		return 0
	}

	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return val
}

// safeParseInt safely parses a string to int.
// Handles Indonesian number format (dots as thousands separators).
func safeParseInt(s string) int {
	// Use safeParseFloat to handle all formats, then truncate
	f := safeParseFloat(s)
	return int(f)
}
