package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/gofpdi"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils/terbilang"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GenerateSpdOverlay creates a PDF by importing the official SPD template and overlaying data.
func GenerateSpdOverlay(record *models.TravelRecord) ([]byte, error) {
	templateName := "Berkas Luar Kota - SPD.pdf"
	templatePath := templateName // Try current directory first (Docker)
	
	// Check if template exists
	if _, err := os.Stat(templatePath); err != nil {
		// Try parent directory (Local development from backend dir)
		templatePath = "../" + templateName
		if _, err := os.Stat(templatePath); err != nil {
			// Try root directory from backend
			templatePath = filepath.Join("..", "..", templateName)
			if _, err := os.Stat(templatePath); err != nil {
				return nil, fmt.Errorf("template file not found in current, parent, or root directory")
			}
		}
	}

	// Custom size: F4 / Legal (21.59 cm x 33.02 cm)
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr:    "mm",
		SizeStr:    "F4",
		Size:       gofpdf.SizeType{Wd: 215.9, Ht: 330.2},
		OrientationStr: "P",
	})

	// Import the official PDF using the contrib package
	tpl := gofpdi.ImportPage(pdf, templatePath, 1, "/MediaBox")
	
	pdf.AddPage()
	
	// Draw the official background first
	gofpdi.UseImportedTemplate(pdf, tpl, 0, 0, 215.9, 330.2)

	// Now overlay the data with Arial font (cleaner for this form)
	pdf.SetFont("Arial", "", 10)

	// Helper to draw text at absolute X, Y (in mm)
	drawText := func(x, y float64, text string) {
		pdf.SetXY(x, y)
		pdf.CellFormat(0, 5, text, "", 0, "L", false, 0, "")
	}
	
	// Helper to draw bold text
	drawTextBold := func(x, y float64, text string) {
		pdf.SetFont("Arial", "B", 10)
		pdf.SetXY(x, y)
		pdf.CellFormat(0, 5, text, "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10) // reset
	}

	// Helper to draw multi-line text (for purpose/destinations)
	drawMultiText := func(x, y, w float64, text string) {
		pdf.SetXY(x, y)
		pdf.MultiCell(w, 4.5, text, "", "L", false)
	}

	titleCaser := cases.Title(language.Indonesian)

	// ==========================================
	// ABSOLUTE COORDINATES (Refined for Template)
	// ==========================================

	// Header - Nomor SPD (Right top side)
	drawText(135, 33.5, record.SPDNumber)

	// 1. Pejabat berwenang
	drawText(100, 58, "KPA Biro Umum Sekretariat Jenderal Kemnaker")

	// 2. Nama / NIP
	drawTextBold(100, 72.5, record.Employee.Name)
	drawText(160, 72.5, record.Employee.NIP)

	// 3. Pangkat, Jabatan, Tingkat Biaya
	pangkatGol := fmt.Sprintf("%s (%s)", record.Employee.Pangkat, record.Employee.Golongan)
	if record.Employee.Pangkat == "" || record.Employee.Pangkat == "-" {
		pangkatGol = record.Employee.Jabatan
	}
	drawText(105, 87.5, pangkatGol)
	
	jabatan := record.Employee.Jabatan
	if jabatan == "" {
		jabatan = "Staf Protokol"
	}
	drawText(105, 95.5, jabatan)
	
	tingkatBiaya := record.Employee.TingkatBiaya
	if tingkatBiaya == "" {
		tingkatBiaya = "C"
	}
	drawText(105, 103, tingkatBiaya)

	// 4. Maksud Perjalanan Dinas
	var locStrings []string
	if len(record.Locations) > 0 {
		for _, l := range record.Locations {
			locStrings = append(locStrings, titleCaser.String(l.Location))
		}
	} else {
		locStrings = append(locStrings, titleCaser.String(record.Location))
	}
	destinations := strings.Join(locStrings, " & ")

	prov := titleCaser.String(record.Province)
	if len(record.Locations) > 0 && record.Locations[0].Province != "" {
		prov = titleCaser.String(record.Locations[0].Province)
	}

	// Filling the blanks in the pre-printed sentence:
	// "Biaya Perjalanan Dinas dalam rangka _______ kunjungan kerja _________ di _______ , Provinsi __________;"
	drawMultiText(150, 116, 45, record.Purpose)      // Blank 1: Purpose
	drawText(95, 123.5, record.Employee.Name)        // Blank 2: Name
	drawText(140, 123.5, destinations)               // Blank 3: Destinations
	drawText(185, 123.5, prov)                       // Blank 4: Province

	// 5. Alat Angkutan
	transportMode := "-"
	if record.Cost != nil && record.Cost.TransportMode != "" {
		transportMode = record.Cost.TransportMode
	}
	drawText(100, 142.5, transportMode)

	// 6. Tempat berangkat & Tempat tujuan
	// Assuming "Jakarta" is often already in template but if not, align it
	drawText(105, 151, "Jakarta")
	drawText(105, 158.5, destinations)

	// 7. Lamanya, Tanggal Berangkat, Tanggal Kembali
	days := int(record.EndDate.Sub(record.StartDate).Hours()/24) + 1
	daysStr := fmt.Sprintf("%d (%s) Hari", days, terbilang.FormatTerbilang(days))
	drawText(105, 172.5, daysStr)

	formatDate := func(t time.Time) string {
		months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
		if t.IsZero() {
			return "________________"
		}
		return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()], t.Year())
	}

	drawText(105, 180, formatDate(record.StartDate))
	drawText(105, 187, formatDate(record.EndDate))

	// Footer / Tanda Tangan
	drawText(155, 271, "Jakarta")
	
	spdDate := record.SuratTugasDate
	if spdDate.IsZero() {
		spdDate = record.StartDate
	}
	if spdDate.IsZero() {
		spdDate = time.Now()
	}
	drawText(155, 277, formatDate(spdDate)) 
	
	// PPK Details
	drawTextBold(150, 305, "Arief Hafidiyanto")
	drawText(150, 310, "NIP. 19720827 200312 1 002")
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}
