package record

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/domain/master"
	"github.com/kemnaker/perjadin-backend/internal/domain/user"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/kemnaker/perjadin-backend/internal/utils/document"
	"github.com/kemnaker/perjadin-backend/internal/utils/terbilang"
	"github.com/labstack/echo/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Handler struct {
	svc       Service
	userRepo  user.Repository
	masterSvc master.Service
	cfg       *config.Config
	docGen    *document.Generator
}

func NewHandler(s Service, userRepo user.Repository, cfg *config.Config, masterSvc master.Service) *Handler {
	gotenbergURL := os.Getenv("GOTENBERG_URL")
	if gotenbergURL == "" {
		gotenbergURL = "http://gotenberg:3000"
	}

	return &Handler{
		svc:       s,
		userRepo:  userRepo,
		masterSvc: masterSvc,
		cfg:       cfg,
		docGen:    document.NewGenerator(gotenbergURL, "templates"),
	}
}

func (h *Handler) notifyEmployee(record *models.TravelRecord) {
	if h.cfg.Fonnte.Token == "" {
		return
	}

	u, err := h.userRepo.GetUserByID(record.EmployeeID)
	if err != nil || u == nil || u.NomorHP == "" {
		return
	}

	locationsStr := ""
	if len(record.Locations) > 0 {
		for _, loc := range record.Locations {
			locationsStr += fmt.Sprintf("\n📍 *Tujuan:* %s, %s\n📅 *Tanggal:* %s s/d %s",
				loc.Location, loc.Province, loc.StartDate.Format("02 Jan 2006"), loc.EndDate.Format("02 Jan 2006"))
		}
	} else {
		locationsStr = fmt.Sprintf("\n📍 *Tujuan:* %s, %s\n📅 *Tanggal:* %s s/d %s",
			record.Location, record.Province, record.StartDate.Format("02 Jan 2006"), record.EndDate.Format("02 Jan 2006"))
	}

	message := fmt.Sprintf("Halo *%s*,\n\nAnda telah ditugaskan untuk melaksanakan perjalanan dinas dengan rincian sebagai berikut:\n%s\n\n🎯 *Kegiatan:* %s\n\nHarap persiapkan diri Anda dan cek aplikasi untuk detail lebih lanjut.\n\n_Pesan ini dikirim otomatis oleh Sistem Perjadin Protokol Kemnaker RI_",
		u.Name, locationsStr, record.Purpose)

	err = utils.SendWhatsAppMessage(h.cfg, u.NomorHP, message)
	if err != nil {
		fmt.Printf("Failed to send WhatsApp message to %s: %v\n", u.NomorHP, err)
	} else {
		fmt.Printf("Successfully sent WhatsApp notification to %s\n", u.NomorHP)
	}
}

func addWorkingDays(t time.Time, days int, holidays []time.Time) time.Time {
	isHoliday := func(d time.Time) bool {
		for _, h := range holidays {
			if d.Year() == h.Year() && d.Month() == h.Month() && d.Day() == h.Day() {
				return true
			}
		}
		return false
	}

	for i := 0; i < days; i++ {
		t = t.AddDate(0, 0, 1)
		for t.Weekday() == time.Saturday || t.Weekday() == time.Sunday || isHoliday(t) {
			t = t.AddDate(0, 0, 1)
		}
	}
	return t
}

func (h *Handler) mapTravelToDocument(record *models.TravelRecord, localIndex int) map[string]interface{} {
	// spdSubNumber is the sequence number of the employee within the specific ID-SPJ group.
	// It is formatted as "001", "002", etc.
	titleCaser := cases.Title(language.Indonesian)
	days := 0
	if !record.StartDate.IsZero() && !record.EndDate.IsZero() {
		days = int(record.EndDate.Sub(record.StartDate).Hours()/24) + 1
	}

	formatRupiahWithRp := func(amount float64) string {
		if amount == 0 {
			return "-"
		}
		return fmt.Sprintf("Rp %s", utils.FormatRupiah(amount))
	}
	formatRupiahNoRp := func(amount float64) string {
		if amount == 0 {
			return "-"
		}
		return utils.FormatRupiah(amount)
	}
	formatNumber := func(n int) string {
		if n == 0 {
			return "-"
		}
		return fmt.Sprintf("%d", n)
	}

	dest := record.Location
	prov := record.Province
	if len(record.Locations) > 0 {
		var locs []string
		var provs []string
		seenProvs := make(map[string]bool)
		for _, l := range record.Locations {
			locs = append(locs, titleCaser.String(l.Location))
			pName := titleCaser.String(l.Province)
			if pName != "" && !seenProvs[pName] {
				provs = append(provs, pName)
				seenProvs[pName] = true
			}
		}
		dest = strings.Join(locs, " & ")
		prov = strings.Join(provs, " & ")
	}
	transportMode := ""
	if record.Cost != nil && record.Cost.TransportMode != "" {
		transportMode = record.Cost.TransportMode
	}
	spdSubNumber := fmt.Sprintf("%03d", localIndex)

	noSurat := strings.TrimSpace(record.SuratTugasNumber)
	tglSurat := record.SuratTugasDate
	numberGap := "\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0"
	dateGap := "\u00A0\u00A0\u00A0\u00A0\u00A0"

	extractNumericID := func(id string) string {
		if id == "" {
			return spdSubNumber
		}
		parts := strings.Split(id, "-")
		if len(parts) > 0 {
			suffix := parts[len(parts)-1]
			if len(suffix) < 3 {
				return fmt.Sprintf("%03s", suffix)
			}
			return suffix
		}
		return id
	}

	tanggalPerjalananDoc := ""
	if !record.StartDate.IsZero() && !record.EndDate.IsZero() {
		if record.StartDate.Year() == record.EndDate.Year() && record.StartDate.Month() == record.EndDate.Month() && record.StartDate.Day() == record.EndDate.Day() {
			tanggalPerjalananDoc = fmt.Sprintf("Tanggal %d %s %d", record.StartDate.Day(), utils.GetIndonesianMonths()[int(record.StartDate.Month())], record.StartDate.Year())
		} else {
			tanggalPerjalananDoc = fmt.Sprintf("Tanggal %d \u2013 %d %s %d", record.StartDate.Day(), record.EndDate.Day(), utils.GetIndonesianMonths()[int(record.StartDate.Month())], record.StartDate.Year())
			if record.StartDate.Month() != record.EndDate.Month() || record.StartDate.Year() != record.EndDate.Year() {
				tanggalPerjalananDoc = fmt.Sprintf("Tanggal %s \u2013 %s", utils.FormatIndonesianDate(record.StartDate), utils.FormatIndonesianDate(record.EndDate))
			}
		}
	}
	if noSurat == "" {
		noSurat = numberGap // Ruang kosong untuk diisi manual
	} else {
		groupID := extractNumericID(record.SPDNumber)
		if groupID != "" && groupID != spdSubNumber {
			noSurat = strings.Replace(noSurat, groupID, spdSubNumber, 1)
		}
	}
	noSpd := record.SPDNumber
	cleanMaksud := regexp.MustCompile(`\s+`).ReplaceAllString(record.Purpose, " ")
	reKunjungan := regexp.MustCompile(`(?i)\s*kunjungan kerja\s*`)
	cleanMaksud = strings.TrimSpace(reKunjungan.ReplaceAllString(cleanMaksud, " "))
	cleanStakeholder := regexp.MustCompile(`\s+`).ReplaceAllString(record.Stakeholder, " ")
	namaPpk := h.cfg.Signatory.PPKName
	nipPpk := h.cfg.Signatory.PPKNIP
	namaBendahara := h.cfg.Signatory.BendaharaName
	nipBendahara := h.cfg.Signatory.BendaharaNIP
	jabPpk := "Pejabat Pembuat Komitmen"
	jabBendahara := "Bendahara Pengeluaran Pembantu"
	if h.masterSvc != nil {
		gs, errGs := h.masterSvc.GetSettings(context.Background())
		if errGs == nil {
			if gs["ppk_name"] != "" {
			namaPpk = gs["ppk_name"]
		}
		if gs["ppk_nip"] != "" {
			nipPpk = gs["ppk_nip"]
		}
		if gs["bendahara_name"] != "" {
			namaBendahara = gs["bendahara_name"]
		}
		if gs["bendahara_nip"] != "" {
			nipBendahara = gs["bendahara_nip"]
		}
		}
	}
	if record.Report != nil {
		if record.Report.PPKName != "" {
			namaPpk = record.Report.PPKName
		}
		if record.Report.PPKNIP != "" {
			nipPpk = record.Report.PPKNIP
		}
		if record.Report.BendaharaName != "" {
			namaBendahara = record.Report.BendaharaName
		}
		if record.Report.BendaharaNIP != "" {
			nipBendahara = record.Report.BendaharaNIP
		}
	}
	if record.Cost == nil {
		record.Cost = &models.TravelCost{DailyAllowanceRate: 0, DailyAllowanceDays: days}
	}
	tglCetak := tglSurat
	if tglCetak.IsZero() {
		tglCetak = time.Now()
	}
	tanggalMulaiStr := "-"
	if !record.StartDate.IsZero() && !record.EndDate.IsZero() {
		if record.StartDate.Month() == record.EndDate.Month() && record.StartDate.Year() == record.EndDate.Year() {
			if record.StartDate.Day() == record.EndDate.Day() {
				tanggalMulaiStr = ""
			} else {
				tanggalMulaiStr = fmt.Sprintf("%d", record.StartDate.Day())
			}
		} else {
			tanggalMulaiStr = utils.FormatIndonesianDate(record.StartDate)
		}
	}
	isiLaporan := "-"
	if record.Report != nil && record.Report.Text != "" {
		isiLaporan = record.Report.Text
	}

	terbilangHari := "-"
	if days > 0 {
		terbilangHari = strings.ToLower(terbilang.FormatTerbilang(days))
	}

	tingkatBiaya := record.Employee.TingkatBiaya
	if tingkatBiaya == "" {
		tingkatBiaya = "C"
	}

	pejabatBerwenang := "KPA Biro Umum Sekretariat Jenderal Kemnaker"
	instansi := "Kementerian Ketenagakerjaan RI"
	akunAnggaran := "-" // Can be updated if database field is added

	pangkat := record.Employee.Pangkat
	if pangkat == "" {
		pangkat = "-"
	}
	golongan := record.Employee.Golongan
	if golongan == "" {
		golongan = "-"
	}

	now := time.Now()
	refDate := now
	tanggalLaporan := now
	tanggalRincian := now

	var holidays []time.Time
	if record.Report != nil && len(record.Report.TanggalMerah) > 0 {
		var dateStrings []string
		if err := json.Unmarshal(record.Report.TanggalMerah, &dateStrings); err == nil {
			for _, ds := range dateStrings {
				if t, err := time.Parse("2006-01-02", ds); err == nil {
					holidays = append(holidays, t)
				}
			}
		}
	}

	if !record.StartDate.IsZero() {
		refDate = record.StartDate
	}
	if !tglSurat.IsZero() {
		refDate = tglSurat
	}
	if !record.EndDate.IsZero() {
		tanggalLaporan = addWorkingDays(record.EndDate, 1, holidays)
		tanggalRincian = addWorkingDays(record.EndDate, 3, holidays)
	}

	tanggalNoSuratStr := fmt.Sprintf("%s %s %d", dateGap, utils.GetIndonesianMonths()[int(refDate.Month())], refDate.Year())
	bulanRomawiST := utils.GetRomanMonths()[int(refDate.Month())]
	tahunST := refDate.Year()
	if !tglSurat.IsZero() {
		tanggalNoSuratStr = utils.FormatIndonesianDate(tglSurat)
		bulanRomawiST = utils.GetRomanMonths()[int(tglSurat.Month())]
		tahunST = tglSurat.Year()
	}

	vars := map[string]interface{}{
		"bulan_romawi_st": bulanRomawiST,
		"tahun_st":        tahunST,
		"no_spd":          spdSubNumber, "id_spj": noSpd, "no_surat": noSurat,
		"pejabat_berwenang":   pejabatBerwenang,
		"tingkat_biaya":       tingkatBiaya,
		"instansi":            instansi,
		"akun_anggaran":       akunAnggaran,
		"tempat_berangkat":    "Jakarta",
		"tempat_tujuan":       dest,
		"kota_atau_kabupaten": dest,
		"keterangan":          "-",
		"lama_perjalanan":     formatNumber(days),
		"nama_stakeholder":    cleanStakeholder,
		"bulan_no_surat":      utils.GetRomanMonths()[int(refDate.Month())], "tahun_no_surat": refDate.Year(),
		"bulan_romawi":     utils.GetRomanMonths()[int(refDate.Month())],
		"tanggal_no_surat": tanggalNoSuratStr, "bulan_pembayaran": utils.GetIndonesianMonths()[int(refDate.Month())],
		"tahun_pembayaran": refDate.Year(), "tahun_saat_ini": fmt.Sprintf("%d", refDate.Year()),
		"nama": record.Employee.Name, "nama_petugas": record.Employee.Name, "nip": record.Employee.NIP, "nip_petugas": record.Employee.NIP,
		"pangkat": pangkat, "golongan": golongan,
		"pangkat_gol": fmt.Sprintf("%s (%s)", pangkat, golongan), "jabatan": record.Employee.Jabatan,
		"maksud_perjalanan": cleanMaksud, "tujuan_perjalanan": cleanMaksud, "stakeholder": cleanStakeholder,
		"tujuan": dest, "kota": dest, "provinsi": prov, "transportasi": transportMode, "mode_transportasi": transportMode,
		"tanggal_berangkat": utils.FormatIndonesianDate(record.StartDate), "tanggal_mulai": tanggalMulaiStr,
		"tanggal_selesai": utils.FormatIndonesianDate(record.EndDate), "lama_hari": formatNumber(days),
		"tanggal_perjalanan_doc": tanggalPerjalananDoc,
		"terbilang":              terbilangHari,
		"bulan":                  utils.GetIndonesianMonths()[int(record.StartDate.Month())], "tahun": fmt.Sprintf("%d", record.StartDate.Year()),
		"tgl_cetak": utils.FormatIndonesianDate(tglCetak), "nama_ppk": namaPpk, "nip_ppk": nipPpk,
		"nama_bendahara": namaBendahara, "nip_bendahara": nipBendahara, "jabatan_ppk": jabPpk, "jabatan_bendahara": jabBendahara,
		"isi_laporan": isiLaporan, "tanggal_dikeluarkan": utils.FormatIndonesianDate(tglCetak),
		"tanggal_laporan": utils.FormatIndonesianDate(tanggalLaporan), "tanggal_rincian": utils.FormatIndonesianDate(tanggalRincian),
		"tanggal_spd": utils.FormatIndonesianDate(tglCetak), "tgl_spd": utils.FormatIndonesianDate(tglCetak),
		"tanggal_st": utils.FormatIndonesianDate(tglSurat), "tgl_st": utils.FormatIndonesianDate(tglSurat),
	}
	type Detail struct {
		TransportMode   string `json:"transportMode"`
		AdditionalCosts []struct {
			Name       string  `json:"name"`
			Amount     float64 `json:"amount"`
			TicketGo   float64 `json:"ticketGo"`
			TicketBack float64 `json:"ticketBack"`
			HotelDays  int     `json:"hotelDays"`
			HotelRate  float64 `json:"hotelRate"`
		} `json:"additionalCosts"`
		TicketGo        float64 `json:"ticketGo"`
		TicketBack      float64 `json:"ticketBack"`
		HotelDays       int     `json:"hotelDays"`
		HotelRate       float64 `json:"hotelRate"`
		TransportAmount float64 `json:"transportAmount"`
	}
	var costDetails []Detail
	if record.Cost != nil && len(record.Cost.Details) > 0 {
		if errUm := json.Unmarshal(record.Cost.Details, &costDetails); errUm != nil {
			fmt.Printf("Unmarshal Cost.Details err: %v\n", errUm)
		}
	}
	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}
	for i, ordinal := range ordinals {
		if i < len(record.Locations) {
			loc := record.Locations[i]
			locDays := int(loc.EndDate.Sub(loc.StartDate).Hours()/24) + 1
			vars[fmt.Sprintf("tujuan_%s", ordinal)] = titleCaser.String(loc.Location)
			vars[fmt.Sprintf("hari_%s", ordinal)] = formatNumber(locDays)
			var sbmRate float64
			if record.Cost != nil {
				sbmRate = record.Cost.DailyAllowanceRate
			}
			vars[fmt.Sprintf("sbm_rate_%s", ordinal)] = formatRupiahNoRp(sbmRate)
			if i < len(costDetails) {
				cd := costDetails[i]
				vars[fmt.Sprintf("hotel_rate_%s", ordinal)] = formatRupiahNoRp(cd.HotelRate)
				vars[fmt.Sprintf("tiket_%s", ordinal)] = formatRupiahNoRp(cd.TicketGo + cd.TicketBack)
			}
		}
	}
	if record.Cost != nil {
		var aggTicket, aggLokal, aggHotel, aggTambahan float64
		var totalDays, hotelDays int
		for _, cd := range costDetails {
			aggTicket += cd.TicketGo + cd.TicketBack
			aggLokal += cd.TransportAmount
			aggHotel += float64(cd.HotelDays) * cd.HotelRate
			hotelDays += cd.HotelDays
			for _, ac := range cd.AdditionalCosts {
				switch ac.Name {
				case "Extend Tiket":
					aggTicket += ac.TicketGo + ac.TicketBack
					aggTambahan += ac.Amount
				case "Extend Penginapan":
					aggHotel += float64(ac.HotelDays) * ac.HotelRate
					hotelDays += ac.HotelDays
				default:
					aggTambahan += ac.Amount
				}
			}
		}
		for _, loc := range record.Locations {
			d := int(loc.EndDate.Sub(loc.StartDate).Hours()/24) + 1
			totalDays += d
		}
		if aggTicket == 0 {
			aggTicket = record.Cost.TicketGo + record.Cost.TicketBack
		}
		if aggLokal == 0 {
			aggLokal = record.Cost.LocalTransport
		}
		if aggHotel == 0 {
			aggHotel = record.Cost.HotelRate * float64(record.Cost.HotelDays)
		}
		if hotelDays == 0 {
			hotelDays = record.Cost.HotelDays
		}
		aggSbm := record.Cost.DailyAllowanceRate * float64(totalDays)
		totalAgg := aggTicket + aggLokal + aggSbm + aggHotel + aggTambahan

		penginapanRate := record.Cost.HotelRate
		if hotelDays > 0 {
			penginapanRate = aggHotel / float64(hotelDays)
		}

		vars["total_biaya"] = formatRupiahWithRp(totalAgg)
		vars["biaya_hotel"] = formatRupiahWithRp(aggHotel)
		vars["biaya_pesawat"] = formatRupiahWithRp(aggTicket)
		if totalAgg > 0 {
			vars["terbilang_biaya"] = fmt.Sprintf("%s RUPIAH", strings.ToUpper(terbilang.FormatTerbilang(int(totalAgg))))
		} else {
			vars["terbilang_biaya"] = "-"
		}

		// Rincian specific variables
		vars["tiket_pesawat"] = formatRupiahNoRp(aggTicket)
		vars["transport_lokal"] = formatRupiahNoRp(aggTambahan)
		vars["transport_daerah"] = formatRupiahNoRp(aggLokal)
		vars["sbm"] = formatRupiahNoRp(record.Cost.DailyAllowanceRate)
		vars["total_sbm"] = formatRupiahNoRp(aggSbm)

		if hotelDays > 0 {
			vars["durasi_hotel"] = hotelDays
		} else {
			vars["durasi_hotel"] = "-"
		}

		vars["p"] = formatRupiahNoRp(penginapanRate)
		vars["total_penginapan"] = formatRupiahNoRp(aggHotel)
		vars["total"] = formatRupiahNoRp(totalAgg)

		if totalAgg > 0 {
			vars["total_terbilang"] = strings.ToUpper(terbilang.FormatTerbilang(int(totalAgg))) + " RUPIAH"
		} else {
			vars["total_terbilang"] = "-"
		}
	}
	return vars
}

func (h *Handler) ExportSpdPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - SPD.docx", "SPD")
}

func (h *Handler) ExportSpdDocx(c echo.Context) error {
	return h.exportDocumentDocx(c, "Berkas Luar Kota - SPD.docx", "SPD")
}

func (h *Handler) ExportLaporanPDF(c echo.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(c.Request().Context(), id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	spdGroupRecords := h.getGroupRecordsSorted(record.SPDNumber)

	seqIndex := 0
	if noSpdStr := c.QueryParam("no_spd"); noSpdStr != "" {
		fmt.Sscanf(noSpdStr, "%d", &seqIndex)
	}
	if seqIndex == 0 {
		seqIndex, _ = h.svc.GetFactualSequenceNumber(c.Request().Context(), record.ID, record.CreatedAt)
	}
	vars := h.mapTravelToDocument(record, seqIndex)
	vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
	vars["tgl_cetak"] = vars["tanggal_laporan"]
	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}
	for i, ordinal := range ordinals {
		if i < len(spdGroupRecords) {
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = spdGroupRecords[i].Employee.Name
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = spdGroupRecords[i].Employee.NIP
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = fmt.Sprintf("%d.", i+1)
		} else {
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = "__REMOVE_ROW__"
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = ""
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = ""
		}
	}

	fotoOrdinals := []string{"satu", "dua", "tiga", "empat"}

	images := make(map[string]document.ImageData)
	var pdfAttachments [][]byte
	lampiranIndex := 1
	fotoIndex := 0

	addLamp := func(data, path, mime string, isFoto bool) {
		if data == "" && path != "" {
			// [ANTI-PATH MISMATCH]: Bersihkan string path dari awalan / atau uploads/ ganda dari Database
			cleanPath := strings.TrimPrefix(path, "/")
			cleanPath = strings.TrimPrefix(cleanPath, "uploads/")
			cleanPath = strings.TrimPrefix(cleanPath, "/")

			content, err := os.ReadFile(filepath.Join("uploads", cleanPath))
			if err == nil {
				data = base64.StdEncoding.EncodeToString(content)
			}
		}
		if data == "" {
			return
		}
		if strings.HasPrefix(mime, "image/") {
			if isFoto {
				if fotoIndex < len(fotoOrdinals) {
					images[fmt.Sprintf("foto_dokumentasi_%s", fotoOrdinals[fotoIndex])] = document.ImageData{Data: data, MimeType: mime}
					fotoIndex++
				}
			} else {
				if lampiranIndex <= 100 {
					images[fmt.Sprintf("lampiran_%d", lampiranIndex)] = document.ImageData{Data: data, MimeType: mime}
					lampiranIndex++
				}
			}
		} else if mime == "application/pdf" {
			raw := data
			if i := strings.Index(raw, ","); i != -1 {
				raw = raw[i+1:]
			}
			b, decErr := base64.StdEncoding.DecodeString(raw)
			if decErr == nil {
				pdfAttachments = append(pdfAttachments, b)
			}
		}
	}

	type FileObj struct {
		Data string `json:"data"`
		Type string `json:"type"`
		Path string `json:"path"`
	}
	processedFoto := false
	for _, r := range spdGroupRecords {
		if !processedFoto && r.Report != nil && len(r.Report.Files) > 0 {
			var files []FileObj
			if umErr := json.Unmarshal(r.Report.Files, &files); umErr == nil {
				for _, f := range files {
					addLamp(f.Data, f.Path, f.Type, true)
				}
				processedFoto = true
			}
		}
		if r.Cost != nil && len(r.Cost.Details) > 0 {
			var details []struct {
				TicketGoFile      *FileObj  `json:"ticketGoFile"`
				TicketBackFile    *FileObj  `json:"ticketBackFile"`
				HotelFile         *FileObj  `json:"hotelFile"`
				TransportFile     *FileObj  `json:"transportFile"`
				BoardingPassFiles []FileObj `json:"boardingPassFiles"`
				AdditionalCosts   []struct {
					File              *FileObj  `json:"file"`
					TicketGoFile      *FileObj  `json:"ticketGoFile"`
					TicketBackFile    *FileObj  `json:"ticketBackFile"`
					BoardingPassFiles []FileObj `json:"boardingPassFiles"`
				} `json:"additionalCosts"`
			}
			if umErr := json.Unmarshal(r.Cost.Details, &details); umErr == nil {
				for _, d := range details {
					if d.TicketGoFile != nil {
						addLamp(d.TicketGoFile.Data, d.TicketGoFile.Path, d.TicketGoFile.Type, false)
					}
					if d.TicketBackFile != nil {
						addLamp(d.TicketBackFile.Data, d.TicketBackFile.Path, d.TicketBackFile.Type, false)
					}
					for _, bp := range d.BoardingPassFiles {
						addLamp(bp.Data, bp.Path, bp.Type, false)
					}
					if d.HotelFile != nil {
						addLamp(d.HotelFile.Data, d.HotelFile.Path, d.HotelFile.Type, false)
					}
					if d.TransportFile != nil {
						addLamp(d.TransportFile.Data, d.TransportFile.Path, d.TransportFile.Type, false)
					}
					for _, ac := range d.AdditionalCosts {
						if ac.File != nil {
							addLamp(ac.File.Data, ac.File.Path, ac.File.Type, false)
						}
						if ac.TicketGoFile != nil {
							addLamp(ac.TicketGoFile.Data, ac.TicketGoFile.Path, ac.TicketGoFile.Type, false)
						}
						if ac.TicketBackFile != nil {
							addLamp(ac.TicketBackFile.Data, ac.TicketBackFile.Path, ac.TicketBackFile.Type, false)
						}
						for _, abp := range ac.BoardingPassFiles {
							addLamp(abp.Data, abp.Path, abp.Type, false)
						}
					}
				}
			}
		}
	}

	for _, ordinal := range fotoOrdinals {
		key := fmt.Sprintf("foto_dokumentasi_%s", ordinal)
		if _, ok := images[key]; !ok {
			vars[key] = "__REMOVE_ROW__"
		}
	}

	for i := 1; i <= 100; i++ {
		key := fmt.Sprintf("lampiran_%d", i)
		if _, ok := images[key]; !ok {
			vars[key] = "__REMOVE_P__"
		}
	}

	payload := document.DocumentRequest{TemplateName: "Berkas Luar Kota - Laporan.docx", Variables: vars}
	pdfBytes, err := h.docGen.GenerateWithImages(c.Request().Context(), payload, images)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if len(pdfAttachments) > 0 {
		merged, err := h.docGen.MergePDFs(c.Request().Context(), append([][]byte{pdfBytes}, pdfAttachments...))
		if err == nil {
			pdfBytes = merged
		}
	}

	filename := fmt.Sprintf("Laporan_%s_%s.pdf", record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) ExportLaporanDocx(c echo.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(c.Request().Context(), id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	spdGroupRecords := h.getGroupRecordsSorted(record.SPDNumber)

	seqIndex := 0
	if noSpdStr := c.QueryParam("no_spd"); noSpdStr != "" {
		fmt.Sscanf(noSpdStr, "%d", &seqIndex)
	}
	if seqIndex == 0 {
		seqIndex, _ = h.svc.GetFactualSequenceNumber(c.Request().Context(), record.ID, record.CreatedAt)
	}
	vars := h.mapTravelToDocument(record, seqIndex)
	vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
	vars["tgl_cetak"] = vars["tanggal_laporan"]

	ordinals := []string{"satu", "dua", "tiga", "empat", "lima", "enam", "tujuh", "delapan", "sembilan", "sepuluh"}
	for i, ordinal := range ordinals {
		if i < len(spdGroupRecords) {
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = spdGroupRecords[i].Employee.Name
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = spdGroupRecords[i].Employee.NIP
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = fmt.Sprintf("%d.", i+1)
		} else {
			vars[fmt.Sprintf("nama_petugas_%s", ordinal)] = "__REMOVE_ROW__"
			vars[fmt.Sprintf("nip_petugas_%s", ordinal)] = ""
			vars[fmt.Sprintf("no_urut_%s", ordinal)] = ""
		}
	}

	payload := document.DocumentRequest{TemplateName: "Berkas Luar Kota - Laporan.docx", Variables: vars}
	docxBytes, errD := h.docGen.GenerateDocx(c.Request().Context(), payload)
	if errD != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errD.Error())
	}
	filename := fmt.Sprintf("Laporan_%s_%s.docx", record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

func (h *Handler) ExportRincianPDF(c echo.Context) error {
	return h.exportDocument(c, "Berkas Luar Kota - rincian pembayaran.docx", "Rincian")
}

func (h *Handler) ExportRincianDocx(c echo.Context) error {
	return h.exportDocumentDocx(c, "Berkas Luar Kota - rincian pembayaran.docx", "Rincian")
}

func (h *Handler) exportDocument(c echo.Context, templateName, prefix string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(c.Request().Context(), id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	seqIndex := 0
	if noSpdStr := c.QueryParam("no_spd"); noSpdStr != "" {
		fmt.Sscanf(noSpdStr, "%d", &seqIndex)
	}
	if seqIndex == 0 {
		seqIndex, _ = h.svc.GetFactualSequenceNumber(c.Request().Context(), record.ID, record.CreatedAt)
	}
	vars := h.mapTravelToDocument(record, seqIndex)
	switch prefix {
	case "Laporan":
		vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
		vars["tgl_cetak"] = vars["tanggal_laporan"]
		vars["tanggal_no_surat"] = vars["tanggal_laporan"]
	case "Rincian":
		vars["tanggal_dikeluarkan"] = vars["tanggal_rincian"]
		vars["tgl_cetak"] = vars["tanggal_rincian"]
		vars["tanggal_no_surat"] = vars["tanggal_rincian"]
		vars["bulan"] = ""
		vars["tahun"] = ""
	case "SPD":
		if record.SuratTugasDate.IsZero() {
			vars["tanggal_dikeluarkan"] = "\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0"
		} else {
			vars["tanggal_dikeluarkan"] = utils.FormatIndonesianDate(record.SuratTugasDate)
		}
	}
	payload := document.DocumentRequest{TemplateName: templateName, Variables: vars}
	pdfBytes, err := h.docGen.Generate(c.Request().Context(), payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	filename := fmt.Sprintf("%s_%s_%s.pdf", prefix, record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func (h *Handler) exportDocumentDocx(c echo.Context, templateName string, prefix string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("panic: %v", r))
		}
	}()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid record ID")
	}
	record, err := h.svc.GetRecordByID(c.Request().Context(), id)
	if err != nil || record == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Record not found")
	}

	seqIndex := 0
	if noSpdStr := c.QueryParam("no_spd"); noSpdStr != "" {
		fmt.Sscanf(noSpdStr, "%d", &seqIndex)
	}
	if seqIndex == 0 {
		seqIndex, _ = h.svc.GetFactualSequenceNumber(c.Request().Context(), record.ID, record.CreatedAt)
	}
	vars := h.mapTravelToDocument(record, seqIndex)
	switch prefix {
	case "Laporan":
		vars["tanggal_dikeluarkan"] = vars["tanggal_laporan"]
		vars["tgl_cetak"] = vars["tanggal_laporan"]
		vars["tanggal_no_surat"] = vars["tanggal_laporan"]
	case "Rincian":
		vars["tanggal_dikeluarkan"] = vars["tanggal_rincian"]
		vars["tgl_cetak"] = vars["tanggal_rincian"]
		vars["tanggal_no_surat"] = vars["tanggal_rincian"]
		vars["bulan"] = ""
		vars["tahun"] = ""
	case "SPD":
		if record.SuratTugasDate.IsZero() {
			vars["tanggal_dikeluarkan"] = "\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0\u00A0"
		} else {
			vars["tanggal_dikeluarkan"] = utils.FormatIndonesianDate(record.SuratTugasDate)
		}
	}
	payload := document.DocumentRequest{TemplateName: templateName, Variables: vars}
	docxBytes, errD := h.docGen.GenerateDocx(c.Request().Context(), payload)
	if errD != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, errD.Error())
	}
	filename := fmt.Sprintf("%s_%s_%s.docx", prefix, record.Employee.Name, record.SPDNumber)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=\"%s\"", filename))
	return c.Blob(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}

func (h *Handler) UploadFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No file provided")
	}

	// SECURITY: Validasi ekstensi yang diizinkan untuk mencegah Upload Vulnerability
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".pdf": true, ".docx": true, ".xlsx": true, ".xls": true, ".doc": true,
	}
	if !allowed[ext] {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid file extension: '%s' (Filename: %s)", ext, file.Filename))
	}

	src, errS := file.Open()
	if errS != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal membuka file")
	}
	defer src.Close()
	filename := uuid.New().String() + "_" + filepath.Base(file.Filename)
	dst, errD := os.Create(filepath.Join("uploads", filename))
	if errD != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal membuat file")
	}
	defer dst.Close()

	// BACKEND COMPRESSION: If it's an image, decode and re-encode it with high compression
	if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
		img, _, errImg := image.Decode(src)
		if errImg == nil {
			// Save it explicitly as a highly compressed JPEG regardless of original
			errEncode := jpeg.Encode(dst, img, &jpeg.Options{Quality: 60}) // 60 is a good balance for documents/proofs
			if errEncode == nil {
				return c.JSON(http.StatusOK, map[string]string{"path": filename})
			}
			// If encoding fails, fallback to direct copy below
		}
		// Reset the reader pointer for fallback
		if seeker, ok := src.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
	}

	if _, errC := io.Copy(dst, src); errC != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal menyimpan file")
	}
	return c.JSON(http.StatusOK, map[string]string{"path": filename})
}

func (h *Handler) CreateRecord(c echo.Context) error {
	var r models.TravelRecord
	if err := c.Bind(&r); err != nil {
		fmt.Printf("Error binding record: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// SECURITY: Sanitasi XSS dasar untuk input string
	r.Purpose = strings.ReplaceAll(strings.ReplaceAll(r.Purpose, "<", ""), ">", "")
	r.Stakeholder = strings.ReplaceAll(strings.ReplaceAll(r.Stakeholder, "<", ""), ">", "")
	r.Location = strings.ReplaceAll(strings.ReplaceAll(r.Location, "<", ""), ">", "")

	if creatorIDStr, ok := c.Get("user_id").(string /*nolint:errcheck*/); ok && creatorIDStr != "" {
		if cid, err := uuid.Parse(creatorIDStr); err == nil {
			r.CreatorID = cid
		}
	}

	// Bulk path: all employees share the same SPD and are inserted atomically.
	// A failure for any one employee rolls back the entire batch.
	if len(r.EmployeeIDs) > 0 {
		batch := make([]*models.TravelRecord, 0, len(r.EmployeeIDs))
		for _, eid := range r.EmployeeIDs {
			nr := r
			nr.ID = uuid.Nil
			nr.EmployeeID = eid
			batch = append(batch, &nr)
		}
		if err := h.svc.CreateRecordsBulk(c.Request().Context(), batch); err != nil {
			fmt.Printf("Error creating bulk records: %v\n", err)
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusCreated, batch)
	}

	// Single employee path — unchanged behaviour.
	if r.EmployeeID == uuid.Nil {
		r.EmployeeID = r.CreatorID
	}

	if err := h.svc.CreateRecord(c.Request().Context(), &r); err != nil {
		fmt.Printf("Error creating record: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	go h.notifyEmployee(&r)
	return c.JSON(http.StatusCreated, r)
}

func (h *Handler) GetRecords(c echo.Context) error {
	role := c.Get("role").(string) //nolint:errcheck
	userIDStr := c.Get("user_id").(string) //nolint:errcheck

	filters := map[string]interface{}{}
	if spd := c.QueryParam("spd"); spd != "" {
		filters["spd"] = spd
	}
	if limit := c.QueryParam("limit"); limit != "" {
		filters["limit"] = limit
	}

	recs, err := h.svc.GetRecords(c.Request().Context(), filters)
	if err != nil {
		fmt.Printf("[ERROR] GetRecords service failed: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if role == "protokol" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			fmt.Printf("[ERROR] Failed to parse userID from context: %v\n", err)
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid User ID")
		}

		// Optimization: If a specific SPD is requested, check if the user has access to it
		if spd := c.QueryParam("spd"); spd != "" {
			hasAccess := false
			for _, r := range recs {
				if r.EmployeeID == uid || r.CreatorID == uid {
					hasAccess = true
					break
				}
			}
			if !hasAccess {
				return c.JSON(http.StatusOK, []models.TravelRecord{})
			}
			// If has access to the SPD, they can see all records in it
			return c.JSON(http.StatusOK, recs)
		}

		// Otherwise, filter normally (e.g. for a global list)
		spds := make(map[string]bool)
		for _, r := range recs {
			if r.EmployeeID == uid || r.CreatorID == uid {
				spds[r.SPDNumber] = true
			}
		}

		res := make([]models.TravelRecord, 0)
		for _, r := range recs {
			if spds[r.SPDNumber] {
				res = append(res, r)
			}
		}
		return c.JSON(http.StatusOK, res)
	}

	return c.JSON(http.StatusOK, recs)
}

func (h *Handler) GetPaginatedRecords(c echo.Context) error {
	var params models.PaginatedParams
	if err := c.Bind(&params); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid pagination parameters")
	}

	role := c.Get("role").(string) //nolint:errcheck
	userIDStr := c.Get("user_id").(string) //nolint:errcheck

	// If the user is a protokol (employee), they can only see their own records.
	// If the user is super_admin or kasubag, they can see all.
	if role == "protokol" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid User ID")
		}
		params.UserID = &uid
	}

	res, err := h.svc.GetPaginatedRecords(c.Request().Context(), params)
	if err != nil {
		fmt.Printf("[ERROR] GetPaginatedRecords failed: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) GetRecordByID(c echo.Context) error {
	id, errI := uuid.Parse(c.Param("id"))
	if errI != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}
	r, errR := h.svc.GetRecordByID(c.Request().Context(), id)
	if errR != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}
	return c.JSON(http.StatusOK, r)
}

func (h *Handler) UpdateRecord(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "ID tidak valid")
	}

	// Read body once for debugging if needed, but Bind is preferred
	r, err := h.svc.GetRecordByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}

	// SECURITY: IDOR Protection, verifikasi pemilik atau petugas yang ditugaskan
	userIDStr := c.Get("user_id").(string) //nolint:errcheck
	userRole := c.Get("role").(string) //nolint:errcheck

	hasAccess := false
	switch {
	case userRole == "super_admin" || userRole == "kasubag":
		hasAccess = true
	case r.CreatorID.String() == userIDStr:
		hasAccess = true
	case r.EmployeeID.String() == userIDStr:
		hasAccess = true
	case userRole == "protokol":
		// Protokol can edit records of other employees IF they are in the same SPD group
		allRecords, err := h.svc.GetRecords(c.Request().Context(), nil)
		if err == nil {
			for _, rec := range allRecords {
				if rec.SPDNumber == r.SPDNumber && rec.EmployeeID.String() == userIDStr {
					hasAccess = true
					break
				}
			}
		}
	}

	if !hasAccess {
		return echo.NewHTTPError(http.StatusForbidden, "Anda tidak memiliki akses untuk mengubah data ini")
	}

	// 1. Baca raw JSON ke map untuk memastikan kita tidak kehilangan nested fields seperti json.RawMessage
	bodyBytes, errB := io.ReadAll(c.Request().Body)
	if errB != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal membaca body")
	}
	fmt.Printf("PUT PAYLOAD: %s\n", string(bodyBytes))
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var partialMap struct {
		ReportData struct {
			TanggalMerah json.RawMessage `json:"tanggalMerah"`
		} `json:"reportData"`
	}
	var explicitTanggalMerah []byte
	var hasExplicitTanggalMerah bool

	if errUm := json.Unmarshal(bodyBytes, &partialMap); errUm == nil {
		if partialMap.ReportData.TanggalMerah != nil {
			hasExplicitTanggalMerah = true
			if string(partialMap.ReportData.TanggalMerah) == "null" {
				explicitTanggalMerah = []byte("[]")
			} else {
				explicitTanggalMerah = partialMap.ReportData.TanggalMerah
			}
		}
	}

	// 2. Lakukan Bind bawaan
	if err := c.Bind(r); err != nil {
		fmt.Printf("UpdateRecord Bind Error: %v\n", err)
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Gagal memproses data: %v", err))
	}

	// 3. Kembalikan secara paksa nilai TanggalMerah jika Frontend mengirimnya
	if hasExplicitTanggalMerah && r.Report != nil {
		r.Report.TanggalMerah = explicitTanggalMerah
	}

	// SECURITY: Sanitasi XSS
	r.Purpose = strings.ReplaceAll(strings.ReplaceAll(r.Purpose, "<", ""), ">", "")
	r.Stakeholder = strings.ReplaceAll(strings.ReplaceAll(r.Stakeholder, "<", ""), ">", "")
	r.Location = strings.ReplaceAll(strings.ReplaceAll(r.Location, "<", ""), ">", "")

	r.ID = id // Ensure ID stays correct

	if err := h.svc.UpdateRecord(c.Request().Context(), r); err != nil {
		fmt.Printf("UpdateRecord Service Error: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal menyimpan ke database: %v", err))
	}

	// PERFORMANCE: Jangan serialize dan kirim ulang payload base64 yang masif (hingga belasan MB)
	r.Report = nil
	r.Cost = nil

	return c.JSON(http.StatusOK, r)
}

func (h *Handler) DeleteRecord(c echo.Context) error {
	id, errI := uuid.Parse(c.Param("id"))
	if errI != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	// SECURITY: IDOR Protection
	r, err := h.svc.GetRecordByID(c.Request().Context(), id)
	if err != nil || r == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Data tidak ditemukan")
	}
	userIDStr := c.Get("user_id").(string) //nolint:errcheck
	userRole := c.Get("role").(string) //nolint:errcheck
	if userRole != "super_admin" && userRole != "kasubag" && r.CreatorID.String() != userIDStr {
		return echo.NewHTTPError(http.StatusForbidden, "Anda tidak memiliki izin menghapus data ini")
	}

	if err := h.svc.DeleteRecord(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal menghapus data: %v", err))
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteRecordsBySpd(c echo.Context) error {
	userRole := c.Get("role").(string) //nolint:errcheck
	if userRole != "super_admin" && userRole != "kasubag" {
		return echo.NewHTTPError(http.StatusForbidden, "Hanya admin yang dapat menghapus SPD batch")
	}
	if err := h.svc.DeleteRecordsBySpd(c.Request().Context(), c.Param("spd")); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Gagal menghapus SPD: %v", err))
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *Handler) GetDashboardSummary(c echo.Context) error {
	role := c.Get("role").(string) //nolint:errcheck
	userIDStr := c.Get("user_id").(string) //nolint:errcheck

	summary, err := h.svc.GetDashboardSummary(c.Request().Context(), role, userIDStr)
	if err != nil {
		fmt.Printf("GetDashboardSummary Error: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Gagal mengambil data dashboard")
	}
	return c.JSON(http.StatusOK, summary)
}
