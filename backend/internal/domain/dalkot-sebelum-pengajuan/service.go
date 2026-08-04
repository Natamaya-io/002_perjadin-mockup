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
	"time"

	"github.com/google/uuid"
)

type Service interface {
	List(ctx context.Context) (*ListResponse, error)
	GetMaster(ctx context.Context) (*MasterData, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Record, error)
	Create(ctx context.Context, input FormInput) (*Record, error)
	Update(ctx context.Context, id uuid.UUID, input FormInput) (*Record, error)
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

func (s *service) List(ctx context.Context) (*ListResponse, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	var summary Summary
	summary.TotalRecords = len(rows)

	for _, row := range rows {
		summary.TotalOfficers += row.OfficerCount
		summary.TotalAmount += row.TotalAmount

		if row.Status == StatusDraft || row.Status == StatusDiajukan {
			summary.PendingRecords++
		}
	}

	return &ListResponse{
		Rows:    rows,
		Summary: summary,
	}, nil
}

func (s *service) GetMaster(ctx context.Context) (*MasterData, error) {
	officers, err := s.repo.GetOfficerMasters(ctx)
	if err != nil {
		return nil, err
	}

	return &MasterData{
		Officers: officers,
		Categories: []MasterOption{
			{Value: "Pendampingan Pimpinan", Label: "Pendampingan Pimpinan"},
			{Value: "Kegiatan Kedinasan", Label: "Kegiatan Kedinasan"},
			{Value: "Koordinasi Protokol", Label: "Koordinasi Protokol"},
			{Value: "Lainnya", Label: "Lainnya"},
		},
		Officials: []MasterOption{
			{Value: "Menteri Ketenagakerjaan", Label: "Menteri Ketenagakerjaan"},
			{Value: "Wakil Menteri Ketenagakerjaan", Label: "Wakil Menteri Ketenagakerjaan"},
			{Value: "Sekretaris Jenderal", Label: "Sekretaris Jenderal"},
			{Value: "Pejabat Lainnya", Label: "Pejabat Lainnya"},
		},
		Locations: []MasterOption{
			{Value: "Jakarta Pusat", Label: "Jakarta Pusat"},
			{Value: "Jakarta Selatan", Label: "Jakarta Selatan"},
			{Value: "Jakarta Barat", Label: "Jakarta Barat"},
			{Value: "Jakarta Timur", Label: "Jakarta Timur"},
			{Value: "Jakarta Utara", Label: "Jakarta Utara"},
			{Value: "Lainnya", Label: "Lainnya"},
		},
		PolicyCategories: []MasterOption{
			{Value: PolicyJamKerja, Label: "Jam Kerja"},
			{Value: PolicyOvertime, Label: "Overtime"},
			{Value: PolicyHariLibur, Label: "Hari Libur"},
		},
		SPJRillAmount:   SPJRillAmount,
		JamKerjaAmount:  JamKerjaAmount,
		OvertimeAmount:  OvertimeAmount,
		HariLiburAmount: HariLiburAmount,
	}, nil
}

func (s *service) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Record, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("ID Dalkot tidak valid")
	}

	return s.repo.GetByID(ctx, id)
}

func (s *service) Create(
	ctx context.Context,
	input FormInput,
) (*Record, error) {
	if err := validateInput(input, true); err != nil {
		return nil, err
	}

	document, err := s.saveDocument(input.Document)
	if err != nil {
		return nil, err
	}

	saveInput := buildSaveInput(input, document)
	saveInput.BusinessID = createBusinessID()

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
) (*Record, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("ID Dalkot tidak valid")
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	requireDocument := existing.DocumentFilename == ""
	if err := validateInput(input, requireDocument); err != nil {
		return nil, err
	}

	var document *DocumentMetadata

	if existing.DocumentFilename != "" {
		document = &DocumentMetadata{
			Filename:     existing.DocumentFilename,
			OriginalName: existing.DocumentOriginalName,
			MIME:         existing.DocumentMIME,
			Size:         existing.DocumentSize,
		}
	}

	hasNewDocument := input.Document != nil
	removeExistingDocument := false

	if input.DalkotType == TypeKebijakanProtokol {
		document = nil
		hasNewDocument = false
		removeExistingDocument = existing.DocumentFilename != ""
	} else if hasNewDocument {
		document, err = s.saveDocument(input.Document)
		if err != nil {
			return nil, err
		}
	}

	saveInput := buildSaveInput(input, document)
	saveInput.BusinessID = existing.BusinessID

	result, err := s.repo.Update(ctx, id, saveInput)
	if err != nil {
		if hasNewDocument && document != nil {
			s.removeDocument(document.Filename)
		}

		return nil, err
	}

	if removeExistingDocument {
		s.removeDocument(existing.DocumentFilename)
	} else if hasNewDocument &&
		existing.DocumentFilename != "" &&
		document != nil &&
		existing.DocumentFilename != document.Filename {
		s.removeDocument(existing.DocumentFilename)
	}

	return result, nil
}

func (s *service) Delete(
	ctx context.Context,
	id uuid.UUID,
) (*Record, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("ID Dalkot tidak valid")
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

func validateInput(input FormInput, requireDocument bool) error {
	if strings.TrimSpace(input.ActiveEmail) == "" {
		return fmt.Errorf("Email Aktif wajib diisi")
	}
	if !strings.Contains(input.ActiveEmail, "@") {
		return fmt.Errorf("Email Aktif tidak valid")
	}
	if input.ServiceDate.IsZero() {
		return fmt.Errorf("Tanggal Pelaksanaan Dinas wajib diisi")
	}
	if strings.TrimSpace(input.Category) == "" {
		return fmt.Errorf("Kategori Dalkot wajib dipilih")
	}
	if strings.TrimSpace(input.AccompaniedOfficial) == "" {
		return fmt.Errorf("Pejabat yang Didampingi wajib dipilih")
	}
	if input.DalkotType != TypeSPJRill &&
		input.DalkotType != TypeKebijakanProtokol {
		return fmt.Errorf("Jenis Dalkot tidak valid")
	}
	if len(input.Officers) == 0 {
		return fmt.Errorf("Pilih minimal satu petugas")
	}
	if strings.TrimSpace(input.ActivityName) == "" {
		return fmt.Errorf("Nama Kegiatan wajib diisi")
	}
	if strings.TrimSpace(input.Location) == "" {
		return fmt.Errorf("Lokasi wajib dipilih")
	}
	if input.DalkotType == TypeSPJRill {
		if strings.TrimSpace(input.ReportContent) == "" {
			return fmt.Errorf("Isi Laporan wajib diisi untuk SPJ RILL")
		}
		if requireDocument && input.Document == nil {
			return fmt.Errorf("Dokumentasi wajib diunggah untuk SPJ RILL")
		}
	}
	if input.DalkotType == TypeKebijakanProtokol {
		switch input.PolicyCategory {
		case PolicyJamKerja, PolicyOvertime, PolicyHariLibur:
		default:
			return fmt.Errorf("Kategori Kebijakan Protokol wajib dipilih")
		}
	}

	for _, officer := range input.Officers {
		if strings.TrimSpace(officer.Name) == "" {
			return fmt.Errorf("Nama petugas tidak boleh kosong")
		}
	}

	return nil
}

func buildSaveInput(
	input FormInput,
	document *DocumentMetadata,
) SaveInput {
	amountPerOfficer := calculateAmountPerOfficer(
		input.DalkotType,
		input.PolicyCategory,
	)

	officers := make([]OfficerSaveInput, 0, len(input.Officers))
	var totalAmount int64

	for index, officer := range input.Officers {
		officers = append(officers, OfficerSaveInput{
			EmployeeID: officer.EmployeeID,
			Name:       strings.TrimSpace(officer.Name),
			NIP:        strings.TrimSpace(officer.NIP),
			Position:   strings.TrimSpace(officer.Position),
			Amount:     amountPerOfficer,
			SortOrder:  index + 1,
		})

		totalAmount += amountPerOfficer
	}

	status := strings.TrimSpace(input.Status)
	if status == "" {
		status = StatusDiajukan
	}

	result := SaveInput{
		ActiveEmail:         strings.TrimSpace(input.ActiveEmail),
		ServiceDate:         input.ServiceDate,
		Category:            strings.TrimSpace(input.Category),
		AccompaniedOfficial: strings.TrimSpace(input.AccompaniedOfficial),
		DalkotType:          input.DalkotType,
		PolicyCategory:      input.PolicyCategory,
		ActivityName:        strings.TrimSpace(input.ActivityName),
		Location:            strings.TrimSpace(input.Location),
		ReportContent:       strings.TrimSpace(input.ReportContent),
		Status:              status,
		TotalAmount:         totalAmount,
		Officers:            officers,
	}

	if input.DalkotType == TypeSPJRill {
		result.PolicyCategory = ""
	}

	if document != nil {
		result.DocumentFilename = document.Filename
		result.DocumentOriginalName = document.OriginalName
		result.DocumentMIME = document.MIME
		result.DocumentSize = document.Size
	}

	return result
}

func calculateAmountPerOfficer(
	dalkotType string,
	policyCategory string,
) int64 {
	if dalkotType == TypeSPJRill {
		return SPJRillAmount
	}

	switch policyCategory {
	case PolicyJamKerja:
		return JamKerjaAmount
	case PolicyOvertime:
		return OvertimeAmount
	case PolicyHariLibur:
		return HariLiburAmount
	default:
		return 0
	}
}

func createBusinessID() string {
	return "DLK-" + time.Now().Format("20060102") + "-" +
		strings.ToUpper(uuid.NewString()[:8])
}

func (s *service) saveDocument(
	header *multipart.FileHeader,
) (*DocumentMetadata, error) {
	if header == nil {
		return nil, nil
	}
	if header.Size <= 0 {
		return nil, fmt.Errorf("Dokumentasi kosong atau tidak valid")
	}
	if header.Size > MaxDocumentSize {
		return nil, fmt.Errorf("Ukuran dokumentasi maksimal 5 MB")
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
		return nil, fmt.Errorf("Dokumentasi kosong atau tidak valid")
	}

	mimeType := http.DetectContentType(headerBytes[:readLength])

	allowedTypes := map[string]string{
		"application/pdf": "pdf",
		"image/jpeg":      "jpg",
		"image/png":       "png",
	}

	extension, allowed := allowedTypes[mimeType]
	if !allowed {
		return nil, fmt.Errorf(
			"Dokumentasi hanya boleh berupa PDF, JPG, JPEG, atau PNG",
		)
	}

	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("gagal membaca ulang dokumentasi Dalkot: %w", err)
	}

	if mimeType == "image/jpeg" {
		if _, err := jpeg.DecodeConfig(source); err != nil {
			return nil, fmt.Errorf("File JPG atau JPEG tidak valid")
		}
	}
	if mimeType == "image/png" {
		if _, err := png.DecodeConfig(source); err != nil {
			return nil, fmt.Errorf("File PNG tidak valid")
		}
	}

	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("gagal membaca ulang dokumentasi Dalkot: %w", err)
	}

	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("gagal membuat folder dokumentasi Dalkot: %w", err)
	}

	filename := "dalkot_" + uuid.NewString() + "." + extension
	targetPath := filepath.Join(s.uploadDir, filename)

	target, err := os.OpenFile(
		targetPath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat dokumentasi Dalkot: %w", err)
	}

	written, copyErr := io.Copy(
		target,
		io.LimitReader(source, MaxDocumentSize+1),
	)
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
		return nil, fmt.Errorf("Ukuran dokumentasi maksimal 5 MB")
	}

	originalName := sanitizeOriginalFilename(header.Filename, extension)

	return &DocumentMetadata{
		Filename:     filename,
		OriginalName: originalName,
		MIME:         mimeType,
		Size:         written,
	}, nil
}

func sanitizeOriginalFilename(
	filename string,
	fallbackExtension string,
) string {
	normalized := strings.ReplaceAll(filename, "\\", "/")
	originalName := filepath.Base(normalized)

	originalName = strings.Map(
		func(character rune) rune {
			if character < 32 || character == 127 {
				return -1
			}
			return character
		},
		originalName,
	)

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
