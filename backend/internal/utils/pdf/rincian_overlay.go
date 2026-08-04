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
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/kemnaker/perjadin-backend/internal/utils/terbilang"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// GenerateRincianOverlay creates a PDF by importing the official Rincian template and overlaying data.
func GenerateRincianOverlay(record *models.TravelRecord, cfg *config.Config) ([]byte, error) {
	templateName := "Berkas Luar Kota - rincian pembayaran.pdf"
	// Try multiple paths to find the template
	paths := []string{
		templateName,
		filepath.Join("backend", templateName),
		filepath.Join("templates", templateName),
		filepath.Join("backend", "templates", templateName),
	}
	
	templatePath := ""
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			templatePath = p
			break
		}
	}

	if templatePath == "" {
		return nil, fmt.Errorf("rincian template file not found: %s", templateName)
	}

	// Legal / F4 size (215.9 mm x 330.2 mm)
	pdf := gofpdf.NewCustom(&gofpdf.InitType{
		UnitStr:    "mm",
		SizeStr:    "F4",
		Size:       gofpdf.SizeType{Wd: 215.9, Ht: 330.2},
		OrientationStr: "P",
	})

	tpl := gofpdi.ImportPage(pdf, templatePath, 1, "/MediaBox")
	pdf.AddPage()
	gofpdi.UseImportedTemplate(pdf, tpl, 0, 0, 215.9, 330.2)

	pdf.SetFont("Arial", "", 10)
	titleCaser := cases.Title(language.Indonesian)

	drawText := func(x, y float64, text string) {
		pdf.SetXY(x, y)
		pdf.CellFormat(0, 5, text, "", 0, "L", false, 0, "")
	}

	drawTextRight := func(x, y, w float64, text string) {
		pdf.SetXY(x, y)
		pdf.CellFormat(w, 5, text, "", 0, "R", false, 0, "")
	}

	formatDate := func(t time.Time) string {
		months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
		if t.IsZero() {
			return "-"
		}
		return fmt.Sprintf("%d %s %d", t.Day(), months[int(t.Month())], t.Year())
	}

	romanMonths := []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	
	// 1. Header Info
	// Format: 1/ID-SPJ-001/UM.06.00/Prot/IV/2026
	now := time.Now()
	romanMonth := romanMonths[int(now.Month())]
	fullSpdNumber := fmt.Sprintf("1/%s/UM.06.00/Prot/%s/%d", record.SPDNumber, romanMonth, now.Year())
	drawText(75, 43, fullSpdNumber) // Lampiran SPPD Nomor
	
	months := []string{"", "Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	payMonth := fmt.Sprintf("%s %d", months[int(time.Now().Month())], time.Now().Year())
	drawText(75, 48, payMonth) // Tanggal (Bulan Pembayaran)

	// 2. Keterangan Section (Middle Right)
	// Biaya Perjalanan Dinas dalam rangka <Purpose> kunjungan kerja <Name> ke <Destinations>, Provinsi <Province> selama <Days> Hari pada tanggal <Date>
	drawText(125, 68, record.Purpose)
	drawText(125, 78, record.Employee.Name)
	
	var locStrings []string
	if len(record.Locations) > 0 {
		for _, l := range record.Locations {
			locStrings = append(locStrings, titleCaser.String(l.Location))
		}
	} else {
		locStrings = append(locStrings, titleCaser.String(record.Location))
	}
	destinations := strings.Join(locStrings, " & ")
	drawText(125, 83, destinations)
	
	prov := titleCaser.String(record.Province)
	if len(record.Locations) > 0 && record.Locations[0].Province != "" {
		prov = titleCaser.String(record.Locations[0].Province)
	}
	drawText(125, 87.5, prov)
	
	days := int(record.EndDate.Sub(record.StartDate).Hours()/24) + 1
	drawText(118, 92.5, fmt.Sprintf("%d (%s)", days, terbilang.FormatTerbilang(days)))
	drawText(160, 92.5, fmt.Sprintf("%s s.d %s", record.StartDate.Format("02/01/2006"), record.EndDate.Format("02/01/2006")))

	// 3. Table Financial Details
	if record.Cost != nil {
		// A. Transport PP
		drawTextRight(155, 116, 30, utils.FormatRupiah(record.Cost.TicketGo+record.Cost.TicketBack))
		drawTextRight(155, 121, 30, utils.FormatRupiah(record.Cost.LocalTransport))
		drawTextRight(155, 126, 30, utils.FormatRupiah(record.Cost.RegionalTransport))

		// 1. Uang Harian
		harianText := fmt.Sprintf("%d hari x Rp %s", record.Cost.DailyAllowanceDays, utils.FormatRupiah(record.Cost.DailyAllowanceRate))
		drawText(40, 150, harianText)
		harianTotal := record.Cost.DailyAllowanceRate * float64(record.Cost.DailyAllowanceDays)
		drawTextRight(155, 150, 30, utils.FormatRupiah(harianTotal))

		// 2. Penginapan
		penginapanText := fmt.Sprintf("%d hari x Rp %s", record.Cost.HotelDays, utils.FormatRupiah(record.Cost.HotelRate))
		drawText(40, 170, penginapanText)
		hotelTotal := record.Cost.HotelRate * float64(record.Cost.HotelDays)
		drawTextRight(155, 170, 30, utils.FormatRupiah(hotelTotal))
	}

	// Total
	total := record.TotalCost
	drawTextRight(155, 222, 30, utils.FormatRupiah(total))
	
	terbilangText := titleCaser.String(terbilang.FormatTerbilang(int(total)))
	drawText(50, 230, terbilangText)

	// Signatures
	drawText(140, 245, "Jakarta, " + formatDate(time.Now()))
	
	// Left: Bendahara
	bendaharaName := cfg.Signatory.BendaharaName
	bendaharaNip := cfg.Signatory.BendaharaNIP
	drawText(25, 300, bendaharaName)
	drawText(25, 305, "NIP. " + bendaharaNip)

	// Middle Right: Penerima
	drawText(140, 300, record.Employee.Name)
	drawText(140, 305, "NIP. " + record.Employee.NIP)

	// Bottom Right: PPK
	ppkName := cfg.Signatory.PPKName
	ppkNip := cfg.Signatory.PPKNIP
	drawText(140, 355, ppkName) // Note: Adjusting y for F4 bottom area
	drawText(140, 360, "NIP. " + ppkNip)

	// Since F4 is 330.2mm, y=355 is actually off-page! 
	// Let's re-adjust for the footer area within 330mm.
	
	// Corrected Footer Positions
	ySigs := 275.0
	drawText(25, ySigs, bendaharaName)
	drawText(25, ySigs+5, "NIP. " + bendaharaNip)
	
	drawText(140, ySigs, record.Employee.Name)
	drawText(140, ySigs+5, "NIP. " + record.Employee.NIP)
	
	yPpk := 315.0
	drawText(140, yPpk, ppkName)
	drawText(140, yPpk+5, "NIP. " + ppkNip)

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}
