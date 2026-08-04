package gup

import (
	"context"
	"database/sql"
	"errors"
	"net/url"

	"github.com/google/uuid"
)

type Repository interface {
	List(ctx context.Context) ([]Transaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Transaction, error)
	Create(ctx context.Context, input SaveInput) (*Transaction, error)
	Update(
		ctx context.Context,
		id uuid.UUID,
		input SaveInput,
	) (*Transaction, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetProcurementTypes(
		ctx context.Context,
	) ([]ProcurementType, error)

	GetFundingSources(
		ctx context.Context,
		year int,
	) ([]FundingSource, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const transactionSelect = `
SELECT
	gt.id,
	gt.business_id,
	gt.payment_description,
	gt.procurement_type_id,
	pt.name,
	ac.code,
	ac.mak,
	gt.funding_source_id,
	COALESCE(fs.gup_label, ''),
	gt.value_amount::BIGINT,
	gt.paid_amount::BIGINT,
	gt.tax_amount::BIGINT,
	gt.difference_amount::BIGINT,
	gt.receipt_date,
	COALESCE(gt.recipient, ''),
	COALESCE(gt.pum, ''),
	COALESCE(gt.document_filename, ''),
	COALESCE(gt.document_original_name, ''),
	COALESCE(gt.document_mime, ''),
	COALESCE(gt.document_size, 0)::BIGINT,
	gt.created_at,
	gt.updated_at
FROM gup_transactions gt
JOIN procurement_types pt
	ON pt.id = gt.procurement_type_id
JOIN account_codes ac
	ON ac.id = pt.account_code_id
LEFT JOIN funding_sources fs
	ON fs.id = gt.funding_source_id
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTransaction(scanner rowScanner) (*Transaction, error) {
	var result Transaction
	var fundingSourceID uuid.NullUUID
	var receiptDate sql.NullTime

	err := scanner.Scan(
		&result.ID,
		&result.BusinessID,
		&result.PaymentDescription,
		&result.ProcurementTypeID,
		&result.ProcurementTypeName,
		&result.AccountCode,
		&result.MAK,
		&fundingSourceID,
		&result.FundingSourceLabel,
		&result.ValueAmount,
		&result.PaidAmount,
		&result.TaxAmount,
		&result.DifferenceAmount,
		&receiptDate,
		&result.Recipient,
		&result.PUM,
		&result.DocumentFilename,
		&result.DocumentOriginalName,
		&result.DocumentMIME,
		&result.DocumentSize,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, err
	}

	if fundingSourceID.Valid {
		id := fundingSourceID.UUID
		result.FundingSourceID = &id
	}

	if receiptDate.Valid {
		result.ReceiptDate = receiptDate.Time.Format("2006-01-02")
	}

	if result.DocumentFilename != "" {
		result.DocumentURL = "/uploads/gup/" +
			url.PathEscape(result.DocumentFilename)
	}

	return &result, nil
}

func (r *repository) List(
	ctx context.Context,
) ([]Transaction, error) {
	query := transactionSelect + `
		ORDER BY gt.receipt_date DESC NULLS LAST,
		gt.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]Transaction, 0)

	for rows.Next() {
		transaction, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}

		results = append(results, *transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*Transaction, error) {
	query := transactionSelect + `
		WHERE gt.id = $1
	`

	return scanTransaction(
		r.db.QueryRowContext(ctx, query, id),
	)
}

func (r *repository) Create(
	ctx context.Context,
	input SaveInput,
) (*Transaction, error) {
	query := `
		INSERT INTO gup_transactions (
			business_id,
			payment_description,
			procurement_type_id,
			funding_source_id,
			value_amount,
			paid_amount,
			tax_amount,
			receipt_date,
			recipient,
			pum,
			document_filename,
			document_original_name,
			document_mime,
			document_size
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14
		)
		RETURNING id
	`

	var id uuid.UUID

	err := r.db.QueryRowContext(
		ctx,
		query,
		input.BusinessID,
		input.PaymentDescription,
		input.ProcurementTypeID,
		nullableUUID(input.FundingSourceID),
		input.ValueAmount,
		input.PaidAmount,
		input.TaxAmount,
		input.ReceiptDate,
		nullableString(input.Recipient),
		nullableString(input.PUM),
		nullableString(input.DocumentFilename),
		nullableString(input.DocumentOriginal),
		nullableString(input.DocumentMIME),
		nullableDocumentSize(
			input.DocumentFilename,
			input.DocumentSize,
		),
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *repository) Update(
	ctx context.Context,
	id uuid.UUID,
	input SaveInput,
) (*Transaction, error) {
	query := `
		UPDATE gup_transactions
		SET
			payment_description = $1,
			procurement_type_id = $2,
			funding_source_id = $3,
			value_amount = $4,
			paid_amount = $5,
			tax_amount = $6,
			receipt_date = $7,
			recipient = $8,
			pum = $9,
			document_filename = $10,
			document_original_name = $11,
			document_mime = $12,
			document_size = $13,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $14
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		input.PaymentDescription,
		input.ProcurementTypeID,
		nullableUUID(input.FundingSourceID),
		input.ValueAmount,
		input.PaidAmount,
		input.TaxAmount,
		input.ReceiptDate,
		nullableString(input.Recipient),
		nullableString(input.PUM),
		nullableString(input.DocumentFilename),
		nullableString(input.DocumentOriginal),
		nullableString(input.DocumentMIME),
		nullableDocumentSize(
			input.DocumentFilename,
			input.DocumentSize,
		),
		id,
	)

	if err != nil {
		return nil, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		return nil, ErrNotFound
	}

	return r.GetByID(ctx, id)
}

func (r *repository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM gup_transactions WHERE id = $1`,
		id,
	)

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *repository) GetProcurementTypes(
	ctx context.Context,
) ([]ProcurementType, error) {
	query := `
		SELECT
			pt.id,
			pt.name,
			ac.code,
			ac.mak
		FROM procurement_types pt
		JOIN account_codes ac
			ON ac.id = pt.account_code_id
		WHERE pt.is_active = TRUE
		ORDER BY ac.code, pt.name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]ProcurementType, 0)

	for rows.Next() {
		var item ProcurementType

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.AccountCode,
			&item.MAK,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	return results, rows.Err()
}

func (r *repository) GetFundingSources(
	ctx context.Context,
	year int,
) ([]FundingSource, error) {
	query := `
		SELECT
			id,
			year,
			month_number,
			month_name,
			gup_label
		FROM funding_sources
		WHERE year = $1
		ORDER BY month_number
	`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]FundingSource, 0)

	for rows.Next() {
		var item FundingSource

		if err := rows.Scan(
			&item.ID,
			&item.Year,
			&item.MonthNumber,
			&item.MonthName,
			&item.GUPLabel,
		); err != nil {
			return nil, err
		}

		results = append(results, item)
	}

	return results, rows.Err()
}

func nullableUUID(id *uuid.UUID) any {
	if id == nil {
		return nil
	}

	return *id
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func nullableDocumentSize(
	filename string,
	size int64,
) any {
	if filename == "" {
		return nil
	}

	return size
}