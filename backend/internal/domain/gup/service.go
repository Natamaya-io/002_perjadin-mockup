package gup

import (
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// Service menentukan fungsi GUP yang dapat digunakan oleh handler.
type Service interface {
	List(ctx context.Context) ([]Transaction, error)

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*Transaction, error)

	Create(
		ctx context.Context,
		input FormInput,
	) (*Transaction, error)

	Update(
		ctx context.Context,
		id uuid.UUID,
		input FormInput,
	) (*Transaction, error)

	Delete(
		ctx context.Context,
		id uuid.UUID,
	) error

	GetMaster(
		ctx context.Context,
		year int,
	) (*MasterData, error)
}

// service merupakan implementasi Service.
type service struct {
	repo      Repository
	uploadDir string
}

// NewService membuat service GUP baru.
func NewService(repo Repository) Service {
	return &service{
		repo:      repo,
		uploadDir: filepath.Join("uploads", "gup"),
	}
}

// List mengambil semua transaksi GUP.
func (s *service) List(
	ctx context.Context,
) ([]Transaction, error) {
	return s.repo.List(ctx)
}

// GetByID mengambil satu transaksi GUP berdasarkan UUID.
func (s *service) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Transaction, error) {
	if id == uuid.Nil {
		return nil, &ValidationError{
			Message: "ID GUP tidak valid",
		}
	}

	return s.repo.GetByID(ctx, id)
}

// GetMaster mengambil jenis pengadaan dan sumber dana.
func (s *service) GetMaster(
	ctx context.Context,
	year int,
) (*MasterData, error) {
	if year < 2000 || year > 2100 {
		return nil, &ValidationError{
			Message: "Tahun sumber dana tidak valid",
		}
	}

	procurementTypes, err :=
		s.repo.GetProcurementTypes(ctx)

	if err != nil {
		return nil, fmt.Errorf(
			"gagal mengambil jenis pengadaan: %w",
			err,
		)
	}

	fundingSources, err :=
		s.repo.GetFundingSources(ctx, year)

	if err != nil {
		return nil, fmt.Errorf(
			"gagal mengambil sumber dana: %w",
			err,
		)
	}

	return &MasterData{
		ProcurementTypes: procurementTypes,
		FundingSources:   fundingSources,
	}, nil
}

// Create menyimpan transaksi GUP baru.
func (s *service) Create(
	ctx context.Context,
	input FormInput,
) (*Transaction, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	document, err := s.saveDocument(
		input.Document,
	)
	if err != nil {
		return nil, err
	}

	saveInput := convertToSaveInput(
		input,
		document,
	)

	saveInput.BusinessID = createBusinessID()

	result, err := s.repo.Create(
		ctx,
		saveInput,
	)
	if err != nil {
		// Dokumen yang sudah tersimpan harus dihapus
		// jika penyimpanan database gagal.
		if document != nil {
			s.removeDocument(
				document.Filename,
			)
		}

		return nil, fmt.Errorf(
			"gagal menyimpan data GUP: %w",
			err,
		)
	}

	return result, nil
}

// Update memperbarui transaksi GUP.
func (s *service) Update(
	ctx context.Context,
	id uuid.UUID,
	input FormInput,
) (*Transaction, error) {
	if id == uuid.Nil {
		return nil, &ValidationError{
			Message: "ID GUP tidak valid",
		}
	}

	if err := validateInput(input); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return nil, err
	}

	// Secara default gunakan dokumen lama.
	var document *DocumentMetadata

	if existing.DocumentFilename != "" {
		document = &DocumentMetadata{
			Filename: existing.DocumentFilename,
			OriginalName: existing.
				DocumentOriginalName,
			MIME: existing.DocumentMIME,
			Size: existing.DocumentSize,
		}
	}

	hasNewDocument := input.Document != nil

	// Simpan dokumen baru ketika pengguna memilih file.
	if hasNewDocument {
		document, err = s.saveDocument(
			input.Document,
		)
		if err != nil {
			return nil, err
		}
	}

	saveInput := convertToSaveInput(
		input,
		document,
	)

	// ID bisnis tidak boleh berubah saat edit.
	saveInput.BusinessID = existing.BusinessID

	result, err := s.repo.Update(
		ctx,
		id,
		saveInput,
	)
	if err != nil {
		// Dokumen baru dihapus jika database gagal diperbarui.
		if hasNewDocument && document != nil {
			s.removeDocument(
				document.Filename,
			)
		}

		return nil, fmt.Errorf(
			"gagal memperbarui data GUP: %w",
			err,
		)
	}

	// Setelah database berhasil diperbarui,
	// dokumen lama dapat dihapus.
	if hasNewDocument &&
		existing.DocumentFilename != "" &&
		document != nil &&
		existing.DocumentFilename != document.Filename {
		s.removeDocument(
			existing.DocumentFilename,
		)
	}

	return result, nil
}

// Delete menghapus transaksi GUP dan dokumennya.
func (s *service) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	if id == uuid.Nil {
		return &ValidationError{
			Message: "ID GUP tidak valid",
		}
	}

	existing, err := s.repo.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(
		ctx,
		id,
	); err != nil {
		return fmt.Errorf(
			"gagal menghapus data GUP: %w",
			err,
		)
	}

	if existing.DocumentFilename != "" {
		s.removeDocument(
			existing.DocumentFilename,
		)
	}

	return nil
}

// validateInput memeriksa seluruh input wajib.
func validateInput(
	input FormInput,
) error {
	if strings.TrimSpace(
		input.PaymentDescription,
	) == "" {
		return &ValidationError{
			Message: "Pembayaran wajib diisi",
		}
	}

	if input.ProcurementTypeID == uuid.Nil {
		return &ValidationError{
			Message: "Jenis pengadaan wajib dipilih",
		}
	}

	if input.FundingSourceID == nil ||
		*input.FundingSourceID == uuid.Nil {
		return &ValidationError{
			Message: "Sumber dana wajib dipilih",
		}
	}

	if input.ValueAmount < 0 {
		return &ValidationError{
			Message: "Nilai pengajuan tidak boleh negatif",
		}
	}

	if input.PaidAmount < 0 {
		return &ValidationError{
			Message: "Jumlah dibayarkan tidak boleh negatif",
		}
	}

	if input.TaxAmount < 0 {
		return &ValidationError{
			Message: "Pajak tidak boleh negatif",
		}
	}

	if input.ReceiptDate.IsZero() {
		return &ValidationError{
			Message: "Tanggal kwitansi wajib diisi",
		}
	}

	if strings.TrimSpace(
		input.Recipient,
	) == "" {
		return &ValidationError{
			Message: "Penerima wajib diisi",
		}
	}

	if strings.TrimSpace(
		input.PUM,
	) == "" {
		return &ValidationError{
			Message: "PUM wajib diisi",
		}
	}

	return nil
}

// convertToSaveInput mengubah data form menjadi data repository.
func convertToSaveInput(
	input FormInput,
	document *DocumentMetadata,
) SaveInput {
	result := SaveInput{
		PaymentDescription: strings.TrimSpace(
			input.PaymentDescription,
		),
		ProcurementTypeID: input.ProcurementTypeID,
		FundingSourceID:   input.FundingSourceID,
		ValueAmount:       input.ValueAmount,
		PaidAmount:        input.PaidAmount,
		TaxAmount:         input.TaxAmount,
		ReceiptDate:       input.ReceiptDate,
		Recipient: strings.TrimSpace(
			input.Recipient,
		),
		PUM: strings.TrimSpace(
			input.PUM,
		),
	}

	if document != nil {
		result.DocumentFilename =
			document.Filename

		result.DocumentOriginal =
			document.OriginalName

		result.DocumentMIME =
			document.MIME

		result.DocumentSize =
			document.Size
	}

	return result
}

// createBusinessID membuat ID bisnis transaksi GUP.
func createBusinessID() string {
	randomID := strings.ReplaceAll(
		uuid.NewString(),
		"-",
		"",
	)

	return "gup_" + randomID
}

// saveDocument memvalidasi dan menyimpan dokumen GUP.
func (s *service) saveDocument(
	header *multipart.FileHeader,
) (*DocumentMetadata, error) {
	// Dokumen bersifat opsional.
	if header == nil {
		return nil, nil
	}

	if header.Size <= 0 {
		return nil, &ValidationError{
			Message: "Dokumen GUP kosong atau tidak valid",
		}
	}

	if header.Size > MaxDocumentSize {
		return nil, &ValidationError{
			Message: "Ukuran dokumen GUP maksimal 5 MB",
		}
	}

	source, err := header.Open()
	if err != nil {
		return nil, fmt.Errorf(
			"gagal membuka dokumen GUP: %w",
			err,
		)
	}
	defer source.Close()

	// Membaca 512 byte pertama untuk memeriksa MIME asli.
	headerBytes := make([]byte, 512)

	readLength, readErr := source.Read(
		headerBytes,
	)

	if readErr != nil && readErr != io.EOF {
		return nil, fmt.Errorf(
			"gagal membaca dokumen GUP: %w",
			readErr,
		)
	}

	if readLength == 0 {
		return nil, &ValidationError{
			Message: "Dokumen GUP kosong atau tidak valid",
		}
	}

	mimeType := http.DetectContentType(
		headerBytes[:readLength],
	)

	allowedTypes := map[string]string{
		"application/pdf": "pdf",
		"image/jpeg":      "jpg",
	}

	extension, allowed :=
		allowedTypes[mimeType]

	if !allowed {
		return nil, &ValidationError{
			Message: "Dokumen GUP hanya boleh berupa PDF, JPG, atau JPEG",
		}
	}

	// Kembalikan posisi pembacaan ke awal file.
	if _, err := source.Seek(
		0,
		io.SeekStart,
	); err != nil {
		return nil, fmt.Errorf(
			"gagal membaca ulang dokumen GUP: %w",
			err,
		)
	}

	// Memastikan JPEG benar dan tidak rusak.
	if mimeType == "image/jpeg" {
		if _, err := jpeg.DecodeConfig(
			source,
		); err != nil {
			return nil, &ValidationError{
				Message: "File JPG atau JPEG tidak valid atau rusak",
			}
		}

		if _, err := source.Seek(
			0,
			io.SeekStart,
		); err != nil {
			return nil, fmt.Errorf(
				"gagal membaca ulang dokumen GUP: %w",
				err,
			)
		}
	}

	// Membuat folder uploads/gup jika belum tersedia.
	if err := os.MkdirAll(
		s.uploadDir,
		0755,
	); err != nil {
		return nil, fmt.Errorf(
			"gagal membuat folder dokumen GUP: %w",
			err,
		)
	}

	// Nama file di server menggunakan UUID.
	filename := "gup_" +
		uuid.NewString() +
		"." +
		extension

	targetPath := filepath.Join(
		s.uploadDir,
		filename,
	)

	target, err := os.OpenFile(
		targetPath,
		os.O_WRONLY|
			os.O_CREATE|
			os.O_EXCL,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal membuat dokumen GUP: %w",
			err,
		)
	}

	// Membatasi pembacaan untuk mencegah file melebihi batas.
	written, copyErr := io.Copy(
		target,
		io.LimitReader(
			source,
			MaxDocumentSize+1,
		),
	)

	closeErr := target.Close()

	if copyErr != nil {
		_ = os.Remove(targetPath)

		return nil, fmt.Errorf(
			"gagal menyimpan dokumen GUP: %w",
			copyErr,
		)
	}

	if closeErr != nil {
		_ = os.Remove(targetPath)

		return nil, fmt.Errorf(
			"gagal menutup dokumen GUP: %w",
			closeErr,
		)
	}

	if written > MaxDocumentSize {
		_ = os.Remove(targetPath)

		return nil, &ValidationError{
			Message: "Ukuran dokumen GUP maksimal 5 MB",
		}
	}

	originalName := sanitizeOriginalFilename(
		header.Filename,
		extension,
	)

	return &DocumentMetadata{
		Filename:     filename,
		OriginalName: originalName,
		MIME:         mimeType,
		Size:         written,
	}, nil
}

// sanitizeOriginalFilename membersihkan nama file asli.
func sanitizeOriginalFilename(
	filename string,
	fallbackExtension string,
) string {
	normalized := strings.ReplaceAll(
		filename,
		"\\",
		"/",
	)

	originalName := filepath.Base(
		normalized,
	)

	// Menghapus karakter kontrol.
	originalName = strings.Map(
		func(character rune) rune {
			if character < 32 ||
				character == 127 {
				return -1
			}

			return character
		},
		originalName,
	)

	originalName = strings.TrimSpace(
		originalName,
	)

	if originalName == "" ||
		originalName == "." {
		originalName =
			"dokumen." + fallbackExtension
	}

	// Membatasi panjang nama file.
	nameRunes := []rune(originalName)

	if len(nameRunes) > 240 {
		originalName = string(
			nameRunes[:240],
		)
	}

	return originalName
}

// removeDocument menghapus dokumen dari server.
func (s *service) removeDocument(
	filename string,
) {
	safeFilename := filepath.Base(
		filename,
	)

	if safeFilename == "" ||
		safeFilename == "." {
		return
	}

	documentPath := filepath.Join(
		s.uploadDir,
		safeFilename,
	)

	_ = os.Remove(documentPath)
}
