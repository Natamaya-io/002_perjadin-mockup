package dalkot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	List(ctx context.Context) ([]Record, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Record, error)
	GetOfficerMasters(ctx context.Context) ([]OfficerMaster, error)
	Create(ctx context.Context, input SaveInput) (*Record, error)
	Update(ctx context.Context, id uuid.UUID, input SaveInput) (*Record, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) List(ctx context.Context) ([]Record, error) {
	const query = `
		SELECT
			dr.id,
			dr.business_id,
			dr.active_email,
			dr.service_date,
			dr.category,
			dr.accompanied_official,
			dr.dalkot_type,
			COALESCE(dr.policy_category, ''),
			dr.activity_name,
			dr.location,
			COALESCE(dr.report_content, ''),
			dr.status,
			dr.total_amount,
			COALESCE(dr.document_filename, ''),
			COALESCE(dr.document_original_name, ''),
			COALESCE(dr.document_mime, ''),
			dr.document_size,
			dr.created_at,
			dr.updated_at,
			COUNT(dro.id) AS officer_count
		FROM dalkot_records dr
		LEFT JOIN dalkot_record_officers dro
			ON dro.dalkot_id = dr.id
		GROUP BY dr.id
		ORDER BY dr.created_at DESC, dr.service_date DESC, dr.business_id DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar Dalkot: %w", err)
	}
	defer rows.Close()

	records := make([]Record, 0)

	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}

		officers, err := r.listOfficers(ctx, record.ID)
		if err != nil {
			return nil, err
		}

		record.Officers = officers
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri daftar Dalkot: %w", err)
	}

	return records, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Record, error) {
	const query = `
		SELECT
			dr.id,
			dr.business_id,
			dr.active_email,
			dr.service_date,
			dr.category,
			dr.accompanied_official,
			dr.dalkot_type,
			COALESCE(dr.policy_category, ''),
			dr.activity_name,
			dr.location,
			COALESCE(dr.report_content, ''),
			dr.status,
			dr.total_amount,
			COALESCE(dr.document_filename, ''),
			COALESCE(dr.document_original_name, ''),
			COALESCE(dr.document_mime, ''),
			dr.document_size,
			dr.created_at,
			dr.updated_at,
			COUNT(dro.id) AS officer_count
		FROM dalkot_records dr
		LEFT JOIN dalkot_record_officers dro
			ON dro.dalkot_id = dr.id
		WHERE dr.id = $1
		GROUP BY dr.id
	`

	row := r.db.QueryRowContext(ctx, query, id)
	record, err := scanRecord(row)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	officers, err := r.listOfficers(ctx, record.ID)
	if err != nil {
		return nil, err
	}

	record.Officers = officers
	return &record, nil
}

func (r *repository) GetOfficerMasters(
	ctx context.Context,
) ([]OfficerMaster, error) {
	const query = `
		SELECT
			id,
			name,
			COALESCE(nip, ''),
			COALESCE(position, ''),
			sort_order
		FROM dalkot_officer_master
		WHERE is_active = TRUE
		ORDER BY sort_order ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil master petugas Dalkot: %w", err)
	}
	defer rows.Close()

	items := make([]OfficerMaster, 0)

	for rows.Next() {
		var item OfficerMaster

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.NIP,
			&item.Position,
			&item.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("gagal membaca master petugas Dalkot: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri master petugas Dalkot: %w", err)
	}

	return items, nil
}

func (r *repository) Create(
	ctx context.Context,
	input SaveInput,
) (*Record, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi Dalkot: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	const insertRecord = `
		INSERT INTO dalkot_records (
			business_id,
			active_email,
			service_date,
			category,
			accompanied_official,
			dalkot_type,
			policy_category,
			activity_name,
			location,
			report_content,
			status,
			total_amount,
			document_filename,
			document_original_name,
			document_mime,
			document_size
		)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			NULLIF($7, ''),
			$8, $9,
			NULLIF($10, ''),
			$11, $12,
			NULLIF($13, ''),
			NULLIF($14, ''),
			NULLIF($15, ''),
			$16
		)
		RETURNING id
	`

	var id uuid.UUID

	if err := tx.QueryRowContext(
		ctx,
		insertRecord,
		input.BusinessID,
		input.ActiveEmail,
		input.ServiceDate,
		input.Category,
		input.AccompaniedOfficial,
		input.DalkotType,
		input.PolicyCategory,
		input.ActivityName,
		input.Location,
		input.ReportContent,
		input.Status,
		input.TotalAmount,
		input.DocumentFilename,
		input.DocumentOriginalName,
		input.DocumentMIME,
		input.DocumentSize,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("gagal menyimpan pengajuan Dalkot: %w", err)
	}

	if err := insertOfficers(ctx, tx, id, input.Officers); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyelesaikan transaksi Dalkot: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *repository) Update(
	ctx context.Context,
	id uuid.UUID,
	input SaveInput,
) (*Record, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi perubahan Dalkot: %w", err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	const updateRecord = `
		UPDATE dalkot_records
		SET
			active_email = $2,
			service_date = $3,
			category = $4,
			accompanied_official = $5,
			dalkot_type = $6,
			policy_category = NULLIF($7, ''),
			activity_name = $8,
			location = $9,
			report_content = NULLIF($10, ''),
			status = $11,
			total_amount = $12,
			document_filename = NULLIF($13, ''),
			document_original_name = NULLIF($14, ''),
			document_mime = NULLIF($15, ''),
			document_size = $16,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := tx.ExecContext(
		ctx,
		updateRecord,
		id,
		input.ActiveEmail,
		input.ServiceDate,
		input.Category,
		input.AccompaniedOfficial,
		input.DalkotType,
		input.PolicyCategory,
		input.ActivityName,
		input.Location,
		input.ReportContent,
		input.Status,
		input.TotalAmount,
		input.DocumentFilename,
		input.DocumentOriginalName,
		input.DocumentMIME,
		input.DocumentSize,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui pengajuan Dalkot: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("gagal memastikan perubahan Dalkot: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM dalkot_record_officers WHERE dalkot_id = $1`,
		id,
	); err != nil {
		return nil, fmt.Errorf("gagal mengganti petugas Dalkot: %w", err)
	}

	if err := insertOfficers(ctx, tx, id, input.Officers); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyelesaikan perubahan Dalkot: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM dalkot_records WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("gagal menghapus data Dalkot: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("gagal memastikan penghapusan Dalkot: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRecord(scanner rowScanner) (Record, error) {
	var record Record
	var serviceDate time.Time

	err := scanner.Scan(
		&record.ID,
		&record.BusinessID,
		&record.ActiveEmail,
		&serviceDate,
		&record.Category,
		&record.AccompaniedOfficial,
		&record.DalkotType,
		&record.PolicyCategory,
		&record.ActivityName,
		&record.Location,
		&record.ReportContent,
		&record.Status,
		&record.TotalAmount,
		&record.DocumentFilename,
		&record.DocumentOriginalName,
		&record.DocumentMIME,
		&record.DocumentSize,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.OfficerCount,
	)
	if err != nil {
		return Record{}, err
	}

	record.ServiceDate = serviceDate.Format("2006-01-02")

	if record.DocumentFilename != "" {
		record.DocumentURL = "/uploads/dalkot/" + record.DocumentFilename
	}

	return record, nil
}

func (r *repository) listOfficers(
	ctx context.Context,
	dalkotID uuid.UUID,
) ([]Officer, error) {
	const query = `
		SELECT
			id,
			dalkot_id,
			employee_id,
			name,
			COALESCE(nip, ''),
			COALESCE(position, ''),
			amount,
			sort_order,
			created_at
		FROM dalkot_record_officers
		WHERE dalkot_id = $1
		ORDER BY sort_order ASC, created_at ASC, name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, dalkotID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil rincian petugas Dalkot: %w", err)
	}
	defer rows.Close()

	officers := make([]Officer, 0)

	for rows.Next() {
		var officer Officer
		var employeeID sql.NullString

		if err := rows.Scan(
			&officer.ID,
			&officer.DalkotID,
			&employeeID,
			&officer.Name,
			&officer.NIP,
			&officer.Position,
			&officer.Amount,
			&officer.SortOrder,
			&officer.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal membaca petugas Dalkot: %w", err)
		}

		if employeeID.Valid {
			parsedID, err := uuid.Parse(employeeID.String)
			if err == nil {
				officer.EmployeeID = &parsedID
			}
		}

		officers = append(officers, officer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri petugas Dalkot: %w", err)
	}

	return officers, nil
}

func insertOfficers(
	ctx context.Context,
	tx *sql.Tx,
	dalkotID uuid.UUID,
	officers []OfficerSaveInput,
) error {
	const query = `
		INSERT INTO dalkot_record_officers (
			dalkot_id,
			employee_id,
			name,
			nip,
			position,
			amount,
			sort_order
		)
		VALUES (
			$1, $2, $3,
			NULLIF($4, ''),
			NULLIF($5, ''),
			$6, $7
		)
	`

	for _, officer := range officers {
		if _, err := tx.ExecContext(
			ctx,
			query,
			dalkotID,
			officer.EmployeeID,
			officer.Name,
			officer.NIP,
			officer.Position,
			officer.Amount,
			officer.SortOrder,
		); err != nil {
			return fmt.Errorf("gagal menyimpan petugas Dalkot: %w", err)
		}
	}

	return nil
}
