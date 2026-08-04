package dalkot

import (
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Service interface {
	List(ctx context.Context, protocolUserID *uuid.UUID) (*ListResponse, error)
	GetRecap(ctx context.Context, filter RecapFilter) (*RecapResponse, error)
	GetMonthlySPJRecap(ctx context.Context, filter RecapFilter) (*MonthlySPJRecapResponse, error)
	GetMaster(ctx context.Context) (*MasterData, error)
	GetByID(ctx context.Context, id uuid.UUID, protocolUserID *uuid.UUID) (*Record, error)
	Create(ctx context.Context, input FormInput, protocolUserID *uuid.UUID) (*Record, error)
	Update(ctx context.Context, id uuid.UUID, input FormInput, protocolUserID *uuid.UUID) (*Record, error)
	Delete(ctx context.Context, id uuid.UUID) (*Record, error)
}

type service struct {
	repo      Repository
	uploadDir string
}

func NewService(repo Repository) Service {
	return &service{
		repo:      repo,
		uploadDir: filepath.Join("uploads", "dalkot"),
	}
}

func (s *service) List(ctx context.Context, protocolUserID *uuid.UUID) (*ListResponse, error) {
	rows, err := s.repo.List(ctx, protocolUserID)
	if err != nil {
		return nil, err
	}

	summary := Summary{TotalRecords: len(rows)}
	for _, row := range rows {
		summary.TotalOfficers += row.OfficerCount
		summary.TotalSPJAmount += row.TotalSPJAmount
		summary.TotalRiilAmount += row.TotalRiilAmount
	}

	return &ListResponse{Rows: rows, Summary: summary}, nil
}

func (s *service) GetRecap(ctx context.Context, filter RecapFilter) (*RecapResponse, error) {
	if err := validateRecapFilter(filter); err != nil {
		return nil, err
	}

	officerName := ""
	if filter.OfficerID != nil {
		users, err := s.repo.GetUsers(ctx)
		if err != nil {
			return nil, err
		}
		found := false
		for _, user := range users {
			if user.ID == *filter.OfficerID {
				officerName = user.Name
				found = true
				break
			}
		}
		if !found {
			return nil, validationError("Petugas Dalkot tidak ditemukan atau tidak aktif")
		}
	}

	rows, summary, err := s.repo.GetRecap(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &RecapResponse{
		Month:       filter.Month,
		Year:        filter.Year,
		OfficerID:   filter.OfficerID,
		OfficerName: officerName,
		Rows:        rows,
		Summary:     summary,
	}, nil
}

func (s *service) GetMonthlySPJRecap(ctx context.Context, filter RecapFilter) (*MonthlySPJRecapResponse, error) {
	filter.OfficerID = nil
	if err := validateRecapFilter(filter); err != nil {
		return nil, err
	}

	rows, err := s.repo.GetMonthlySPJRecap(ctx, filter)
	if err != nil {
		return nil, err
	}

	summary := MonthlySPJRecapSummary{TotalOfficers: len(rows)}
	for _, row := range rows {
		summary.TotalDays += row.TotalDays
		summary.TotalAmount += row.TotalAmount
	}

	return &MonthlySPJRecapResponse{
		Month:   filter.Month,
		Year:    filter.Year,
		Account: MonthlySPJAccount,
		Title:   MonthlySPJTitle,
		Rows:    rows,
		Summary: summary,
	}, nil
}

func validateRecapFilter(filter RecapFilter) error {
	if filter.Month < 1 || filter.Month > 12 {
		return validationError("Bulan rekap Dalkot tidak valid")
	}
	if filter.Year < 2020 || filter.Year > 2100 {
		return validationError("Tahun rekap Dalkot harus berada antara 2020 dan 2100")
	}
	if filter.OfficerID != nil && *filter.OfficerID == uuid.Nil {
		return validationError("ID Petugas Dalkot tidak valid")
	}
	return nil
}

func (s *service) GetMaster(ctx context.Context) (*MasterData, error) {
	users, err := s.repo.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	return &MasterData{
		Users: users,
		Categories: []MasterOption{
			{Value: CategoryJamKerja, Label: CategoryJamKerja, Amount: JamKerjaAmount},
			{Value: CategoryOvertime, Label: CategoryOvertime, Amount: OvertimeAmount},
			{Value: CategoryHariLibur, Label: CategoryHariLibur, Amount: HariLiburAmount},
		},
		Officials: []MasterOption{
			{Value: "Menteri", Label: "Menteri"},
			{Value: "Wakil Menteri", Label: "Wakil Menteri"},
			{Value: "Ibu Menteri", Label: "Ibu Menteri"},
			{Value: "Ibu Wakil Menteri", Label: "Ibu Wakil Menteri"},
		},
		Locations: []MasterOption{
			{Value: "Jakarta Pusat", Label: "Jakarta Pusat"},
			{Value: "Jakarta Utara", Label: "Jakarta Utara"},
			{Value: "Jakarta Selatan", Label: "Jakarta Selatan"},
			{Value: "Jakarta Timur", Label: "Jakarta Timur"},
			{Value: "Jakarta Barat", Label: "Jakarta Barat"},
		},
		SPJAmount:       SPJAmount,
		JamKerjaAmount:  JamKerjaAmount,
		OvertimeAmount:  OvertimeAmount,
		HariLiburAmount: HariLiburAmount,
	}, nil
}

func (s *service) GetByID(
	ctx context.Context,
	id uuid.UUID,
	protocolUserID *uuid.UUID,
) (*Record, error) {
	if id == uuid.Nil {
		return nil, validationError("ID Dalkot tidak valid")
	}

	record, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if protocolUserID != nil && !recordContainsOfficer(record, *protocolUserID) {
		// Kembalikan not found agar user tidak dapat mengetahui keberadaan
		// pengajuan Dalkot yang tidak mencantumkan dirinya sebagai petugas.
		return nil, ErrNotFound
	}

	return record, nil
}

func (s *service) Create(
	ctx context.Context,
	input FormInput,
	protocolUserID *uuid.UUID,
) (*Record, error) {
	if err := validateInput(input, false); err != nil {
		return nil, err
	}

	if err := validateProtocolAssignment(input, protocolUserID); err != nil {
		return nil, err
	}

	document, err := s.saveDocument(input.Document)
	if err != nil {
		return nil, err
	}

	saveInput, err := s.buildSaveInput(ctx, input, document)
	if err != nil {
		if document != nil {
			s.removeDocument(document.Filename)
		}
		return nil, err
	}

	result, err := s.repo.Create(ctx, saveInput)
	if err != nil {
		if document != nil {
			s.removeDocument(document.Filename)
		}
		return nil, err
	}

	return result, nil
}

func (s *service) Update(
	ctx context.Context,
	id uuid.UUID,
	input FormInput,
	protocolUserID *uuid.UUID,
) (*Record, error) {
	if id == uuid.Nil {
		return nil, validationError("ID Dalkot tidak valid")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if protocolUserID != nil && !recordContainsOfficer(existing, *protocolUserID) {
		return nil, ErrNotFound
	}

	hasExistingDocument := existing.DocumentFilename != ""
	if err := validateInput(input, hasExistingDocument); err != nil {
		return nil, err
	}

	if err := validateProtocolAssignment(input, protocolUserID); err != nil {
		return nil, err
	}

	var document *DocumentMetadata
	if hasExistingDocument {
		document = &DocumentMetadata{
			Filename:     existing.DocumentFilename,
			OriginalName: existing.DocumentOriginalName,
			MIME:         existing.DocumentMIME,
			Size:         existing.DocumentSize,
		}
	}

	hasNewDocument := input.Document != nil
	if hasNewDocument {
		document, err = s.saveDocument(input.Document)
		if err != nil {
			return nil, err
		}
	}

	saveInput, err := s.buildSaveInput(ctx, input, document)
	if err != nil {
		if hasNewDocument && document != nil {
			s.removeDocument(document.Filename)
		}
		return nil, err
	}

	saveInput.BusinessID = existing.BusinessID
	// Status pengajuan tidak diubah oleh form umum. Perubahan status dilakukan
	// melalui alur persetujuan atau pembayaran yang terpisah.
	saveInput.Status = normalizeRecordStatus(existing.Status)

	result, err := s.repo.Update(ctx, id, saveInput)
	if err != nil {
		if hasNewDocument && document != nil {
			s.removeDocument(document.Filename)
		}
		return nil, err
	}

	if hasNewDocument &&
		hasExistingDocument &&
		document != nil &&
		existing.DocumentFilename != document.Filename {
		s.removeDocument(existing.DocumentFilename)
	}

	return result, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) (*Record, error) {
	if id == uuid.Nil {
		return nil, validationError("ID Dalkot tidak valid")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return nil, err
	}
	if existing.DocumentFilename != "" {
		s.removeDocument(existing.DocumentFilename)
	}
	return existing, nil
}

func validateProtocolAssignment(input FormInput, protocolUserID *uuid.UUID) error {
	if protocolUserID == nil {
		return nil
	}

	if containsUUID(input.SPJOfficerIDs, *protocolUserID) ||
		containsUUID(input.RiilOfficerIDs, *protocolUserID) {
		return nil
	}

	return validationError(
		"Nama Anda wajib dipilih sebagai Petugas SPJ atau Petugas Riil",
	)
}

func recordContainsOfficer(record *Record, userID uuid.UUID) bool {
	if record == nil || userID == uuid.Nil {
		return false
	}

	for _, officer := range record.Officers {
		if officer.EmployeeID != nil && *officer.EmployeeID == userID {
			return true
		}
	}

	return false
}

func containsUUID(values []uuid.UUID, target uuid.UUID) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}

func validateInput(input FormInput, hasExistingDocument bool) error {
	if input.ServiceDate.IsZero() {
		return validationError("Tanggal Pelaksanaan Dinas wajib diisi")
	}
	if !isValidCategory(input.Category) {
		return validationError("Kategori Dalkot wajib dipilih")
	}
	if !isValidOfficial(input.AccompaniedOfficial) {
		return validationError("Pejabat yang Didampingi tidak valid")
	}
	if len(input.SPJOfficerIDs) == 0 && len(input.RiilOfficerIDs) == 0 {
		return validationError("Pilih minimal satu Petugas SPJ atau Petugas Rill")
	}
	if err := validateUniqueOfficerIDs(input.SPJOfficerIDs, "SPJ"); err != nil {
		return err
	}
	if err := validateUniqueOfficerIDs(input.RiilOfficerIDs, "Rill"); err != nil {
		return err
	}

	// Tidak ada validasi silang antara SPJ dan RIIL. User Protokol yang sama
	// memang boleh dipilih pada kedua dropdown sekaligus.
	if strings.TrimSpace(input.ActivityName) == "" {
		return validationError("Nama Kegiatan wajib diisi")
	}
	if strings.TrimSpace(input.Location) == "" {
		return validationError("Lokasi wajib dipilih")
	}

	hasSPJ := len(input.SPJOfficerIDs) > 0
	if hasSPJ && strings.TrimSpace(input.ReportContent) == "" {
		return validationError("Isi Laporan wajib diisi apabila terdapat Petugas SPJ")
	}
	if hasSPJ && input.Document == nil && !hasExistingDocument {
		return validationError("Dokumentasi wajib diunggah apabila terdapat Petugas SPJ")
	}

	if !isValidRecordStatus(input.Status) {
		return validationError("Status Pengajuan Dalkot tidak valid")
	}
	return nil
}

func validateUniqueOfficerIDs(ids []uuid.UUID, groupName string) error {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			return validationError("ID Petugas " + groupName + " tidak valid")
		}
		if _, exists := seen[id]; exists {
			return validationError("Petugas " + groupName + " tidak boleh dipilih lebih dari satu kali")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func (s *service) buildSaveInput(
	ctx context.Context,
	input FormInput,
	document *DocumentMetadata,
) (SaveInput, error) {
	spjUsers, err := s.repo.GetUsersByIDs(ctx, input.SPJOfficerIDs)
	if err != nil {
		return SaveInput{}, err
	}
	riilUsers, err := s.repo.GetUsersByIDs(ctx, input.RiilOfficerIDs)
	if err != nil {
		return SaveInput{}, err
	}

	riilAmount := calculateRiilAmount(input.Category)
	if riilAmount <= 0 {
		return SaveInput{}, validationError("Nominal Kategori Dalkot tidak valid")
	}

	officers := make([]OfficerSaveInput, 0, len(spjUsers)+len(riilUsers))
	sortOrder := 1
	for _, user := range spjUsers {
		officers = append(officers, OfficerSaveInput{
			EmployeeID:   user.ID,
			OfficerGroup: OfficerGroupSPJ,
			Name:         strings.TrimSpace(user.Name),
			NIP:          strings.TrimSpace(user.NIP),
			Position:     strings.TrimSpace(user.Position),
			Grade:        strings.TrimSpace(user.Grade),
			Amount:       SPJAmount,
			Status:       StatusPending,
			SortOrder:    sortOrder,
		})
		sortOrder++
	}
	for _, user := range riilUsers {
		officers = append(officers, OfficerSaveInput{
			EmployeeID:   user.ID,
			OfficerGroup: OfficerGroupRiil,
			Name:         strings.TrimSpace(user.Name),
			NIP:          strings.TrimSpace(user.NIP),
			Position:     strings.TrimSpace(user.Position),
			Grade:        strings.TrimSpace(user.Grade),
			Amount:       riilAmount,
			Status:       StatusPending,
			SortOrder:    sortOrder,
		})
		sortOrder++
	}

	dalkotType := TypeKebijakanProtokol
	if len(spjUsers) > 0 {
		dalkotType = TypeSPJRill
	}

	result := SaveInput{
		ServiceDate:            input.ServiceDate,
		Category:               strings.TrimSpace(input.Category),
		AccompaniedOfficial:    strings.TrimSpace(input.AccompaniedOfficial),
		DalkotType:             dalkotType,
		PolicyCategory:         categoryToPolicyCode(input.Category),
		AssignmentLetterNumber: strings.TrimSpace(input.AssignmentLetterNumber),
		AssignmentLetterDate:   input.AssignmentLetterDate,
		ActivityName:           strings.TrimSpace(input.ActivityName),
		Location:               strings.TrimSpace(input.Location),
		ReportContent:          strings.TrimSpace(input.ReportContent),
		Status:                 normalizeRecordStatus(input.Status),
		TotalSPJAmount:         int64(len(spjUsers)) * SPJAmount,
		TotalRiilAmount:        int64(len(riilUsers)) * riilAmount,
		Officers:               officers,
	}

	if document != nil {
		result.DocumentFilename = document.Filename
		result.DocumentOriginalName = document.OriginalName
		result.DocumentMIME = document.MIME
		result.DocumentSize = document.Size
	}
	return result, nil
}

func calculateRiilAmount(category string) int64 {
	switch strings.TrimSpace(category) {
	case CategoryJamKerja:
		return JamKerjaAmount
	case CategoryOvertime:
		return OvertimeAmount
	case CategoryHariLibur:
		return HariLiburAmount
	default:
		return 0
	}
}

func categoryToPolicyCode(category string) string {
	switch strings.TrimSpace(category) {
	case CategoryJamKerja:
		return "JAM_KERJA"
	case CategoryOvertime:
		return "OVERTIME"
	case CategoryHariLibur:
		return "HARI_LIBUR"
	default:
		return ""
	}
}

func isValidCategory(value string) bool {
	switch strings.TrimSpace(value) {
	case CategoryJamKerja, CategoryOvertime, CategoryHariLibur:
		return true
	default:
		return false
	}
}

func isValidOfficial(value string) bool {
	switch strings.TrimSpace(value) {
	case "Menteri", "Wakil Menteri", "Ibu Menteri", "Ibu Wakil Menteri":
		return true
	default:
		return false
	}
}

func isValidRecordStatus(value string) bool {
	switch strings.TrimSpace(value) {
	case "", StatusDraft, StatusPending, StatusSukses:
		return true
	default:
		return false
	}
}

func normalizeRecordStatus(value string) string {
	switch strings.TrimSpace(value) {
	case StatusPending:
		return StatusPending
	case StatusSukses:
		return StatusSukses
	default:
		return StatusDraft
	}
}

func validationError(message string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, message)
}

func (s *service) saveDocument(header *multipart.FileHeader) (*DocumentMetadata, error) {
	if header == nil {
		return nil, nil
	}
	if header.Size <= 0 {
		return nil, validationError("Dokumentasi kosong atau tidak valid")
	}
	if header.Size > MaxDocumentSize {
		return nil, validationError("Ukuran dokumentasi maksimal 5 MB")
	}

	source, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf("gagal membuka dokumentasi Dalkot: %w", err)
	}
	defer source.Close()

	headerBytes := make([]byte, 512)
	readLength, readErr := source.Read(headerBytes)
	if readErr != nil && readErr != io.EOF {
		return nil, fmt.Errorf("gagal membaca dokumentasi Dalkot: %w", readErr)
	}
	if readLength == 0 {
		return nil, validationError("Dokumentasi kosong atau tidak valid")
	}

	mimeType := http.DetectContentType(headerBytes[:readLength])
	allowedTypes := map[string]string{
		"application/pdf": "pdf",
		"image/jpeg":      "jpg",
		"image/png":       "png",
	}
	extension, allowed := allowedTypes[mimeType]
	if !allowed {
		return nil, validationError("Dokumentasi hanya boleh berupa PDF, JPG, JPEG, atau PNG")
	}
	if mimeType == "application/pdf" && !strings.HasPrefix(string(headerBytes[:readLength]), "%PDF-") {
		return nil, validationError("File PDF tidak valid")
	}

	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("gagal membaca ulang dokumentasi Dalkot: %w", err)
	}
	switch mimeType {
	case "image/jpeg":
		if _, err := jpeg.DecodeConfig(source); err != nil {
			return nil, validationError("File JPG atau JPEG tidak valid")
		}
	case "image/png":
		if _, err := png.DecodeConfig(source); err != nil {
			return nil, validationError("File PNG tidak valid")
		}
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("gagal membaca ulang dokumentasi Dalkot: %w", err)
	}

	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("gagal membuat folder dokumentasi Dalkot: %w", err)
	}
	filename := "dalkot_" + uuid.NewString() + "." + extension
	targetPath := filepath.Join(s.uploadDir, filename)
	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat dokumentasi Dalkot: %w", err)
	}

	written, copyErr := io.Copy(target, io.LimitReader(source, MaxDocumentSize+1))
	closeErr := target.Close()
	if copyErr != nil {
		_ = os.Remove(targetPath)
		return nil, fmt.Errorf("gagal menyimpan dokumentasi Dalkot: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(targetPath)
		return nil, fmt.Errorf("gagal menutup dokumentasi Dalkot: %w", closeErr)
	}
	if written > MaxDocumentSize {
		_ = os.Remove(targetPath)
		return nil, validationError("Ukuran dokumentasi maksimal 5 MB")
	}

	return &DocumentMetadata{
		Filename:     filename,
		OriginalName: sanitizeOriginalFilename(header.Filename, extension),
		MIME:         mimeType,
		Size:         written,
	}, nil
}

func sanitizeOriginalFilename(filename, fallbackExtension string) string {
	normalized := strings.ReplaceAll(filename, "\\", "/")
	originalName := filepath.Base(normalized)
	originalName = strings.Map(func(character rune) rune {
		if character < 32 || character == 127 {
			return -1
		}
		return character
	}, originalName)
	originalName = strings.TrimSpace(originalName)
	if originalName == "" || originalName == "." {
		originalName = "dokumentasi." + fallbackExtension
	}
	nameRunes := []rune(originalName)
	if len(nameRunes) > 240 {
		originalName = string(nameRunes[:240])
	}
	return originalName
}

func (s *service) removeDocument(filename string) {
	safeFilename := filepath.Base(filename)
	if safeFilename == "" || safeFilename == "." {
		return
	}
	_ = os.Remove(filepath.Join(s.uploadDir, safeFilename))
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsInvalidInput(err error) bool {
	return errors.Is(err, ErrInvalidInput)
}