package fundsource

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrNotFound          = errors.New("sumber dana tidak ditemukan")
	ErrDuplicate         = errors.New("kode sumber dana sudah digunakan pada tahun yang sama")
	ErrSourceInUse       = errors.New("sumber dana masih memiliki alokasi GUP/LS")
	ErrAllocationInUse   = errors.New("alokasi masih digunakan oleh perjalanan dinas")
	ErrInsufficientFunds = errors.New("saldo sumber dana tidak mencukupi")
)

type Repository interface {
	ListSources(ctx context.Context, year int) ([]Source, error)
	ListAllocations(ctx context.Context, year int) ([]Allocation, error)
	ListSourceOptions(ctx context.Context, year int, budgetType string, spdNumber string) ([]SourceOption, error)
	GetSource(ctx context.Context, id uuid.UUID) (*Source, error)
	CreateSource(ctx context.Context, input SaveSourceRequest) (*Source, error)
	UpdateSource(ctx context.Context, id uuid.UUID, input SaveSourceRequest) (*Source, error)
	DeleteSource(ctx context.Context, id uuid.UUID) error
	CreateAllocation(ctx context.Context, input SaveAllocationRequest) (*Allocation, error)
	DeleteAllocation(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

const sourceSelect = `
WITH allocation_summary AS (
    SELECT
        finance_source_id,
        COALESCE(SUM(amount) FILTER (WHERE allocation_type = 'GUP'), 0) AS allocated_gup,
        COALESCE(SUM(amount) FILTER (WHERE allocation_type = 'LS'), 0) AS allocated_ls,
        COALESCE(SUM(amount), 0) AS allocated_total
    FROM finance_allocations
    GROUP BY finance_source_id
),
usage_summary AS (
    SELECT
        finance_source_id,
        COALESCE(SUM(amount) FILTER (WHERE budget_type = 'GUP'), 0) AS used_gup,
        COALESCE(SUM(amount) FILTER (WHERE budget_type = 'LS'), 0) AS used_ls,
        COALESCE(SUM(amount), 0) AS used_total
    FROM finance_usages
    GROUP BY finance_source_id
)
SELECT
    fs.id,
    fs.year,
    fs.code,
    fs.name,
    fs.amount::BIGINT,
    COALESCE(a.allocated_gup, 0)::BIGINT,
    COALESCE(a.allocated_ls, 0)::BIGINT,
    COALESCE(a.allocated_total, 0)::BIGINT,
    COALESCE(u.used_gup, 0)::BIGINT,
    COALESCE(u.used_ls, 0)::BIGINT,
    COALESCE(u.used_total, 0)::BIGINT,
    (COALESCE(a.allocated_gup, 0) - COALESCE(u.used_gup, 0))::BIGINT,
    (COALESCE(a.allocated_ls, 0) - COALESCE(u.used_ls, 0))::BIGINT,
    (fs.amount - COALESCE(u.used_total, 0))::BIGINT,
    (fs.amount - COALESCE(a.allocated_total, 0))::BIGINT,
    COALESCE(fs.notes, ''),
    fs.is_active,
    fs.created_at,
    fs.updated_at
FROM finance_sources fs
LEFT JOIN allocation_summary a ON a.finance_source_id = fs.id
LEFT JOIN usage_summary u ON u.finance_source_id = fs.id
`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSource(scanner rowScanner) (*Source, error) {
	var result Source
	if err := scanner.Scan(
		&result.ID,
		&result.Year,
		&result.Code,
		&result.Name,
		&result.Amount,
		&result.AllocatedGUP,
		&result.AllocatedLS,
		&result.AllocatedTotal,
		&result.UsedGUP,
		&result.UsedLS,
		&result.UsedTotal,
		&result.RemainingGUP,
		&result.RemainingLS,
		&result.RemainingTotal,
		&result.Available,
		&result.Notes,
		&result.IsActive,
		&result.CreatedAt,
		&result.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &result, nil
}

func (r *repository) ListSources(ctx context.Context, year int) ([]Source, error) {
	query := sourceSelect + `
WHERE fs.year = $1
ORDER BY fs.code, fs.name
`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Source, 0)
	for rows.Next() {
		source, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *source)
	}
	return result, rows.Err()
}

func (r *repository) GetSource(ctx context.Context, id uuid.UUID) (*Source, error) {
	query := sourceSelect + `
WHERE fs.id = $1
`
	return scanSource(r.db.QueryRowContext(ctx, query, id))
}

func (r *repository) ListSourceOptions(ctx context.Context, year int, budgetType string, spdNumber string) ([]SourceOption, error) {
	const query = `
WITH allocation_data AS (
    SELECT
        finance_source_id,
        COALESCE(SUM(amount), 0)::BIGINT AS allocated
    FROM finance_allocations
    WHERE allocation_type = $2
    GROUP BY finance_source_id
),
usage_data AS (
    SELECT
        finance_source_id,
        COALESCE(SUM(amount), 0)::BIGINT AS used_other
    FROM finance_usages
    WHERE budget_type = $2
      AND ($3 = '' OR spd_number <> $3)
    GROUP BY finance_source_id
)
SELECT
    fs.id,
    fs.year,
    fs.code,
    fs.name,
    $2::text,
    COALESCE(a.allocated, 0)::BIGINT,
    COALESCE(u.used_other, 0)::BIGINT,
    (COALESCE(a.allocated, 0) - COALESCE(u.used_other, 0))::BIGINT AS available
FROM finance_sources fs
LEFT JOIN allocation_data a ON a.finance_source_id = fs.id
LEFT JOIN usage_data u ON u.finance_source_id = fs.id
WHERE fs.year = $1
  AND fs.is_active = TRUE
  AND COALESCE(a.allocated, 0) > 0
  AND (COALESCE(a.allocated, 0) - COALESCE(u.used_other, 0)) > 0
ORDER BY fs.code, fs.name
`

	rows, err := r.db.QueryContext(ctx, query, year, budgetType, spdNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]SourceOption, 0)
	for rows.Next() {
		var item SourceOption
		if err := rows.Scan(
			&item.ID,
			&item.Year,
			&item.Code,
			&item.Name,
			&item.BudgetType,
			&item.Allocated,
			&item.UsedOther,
			&item.Available,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *repository) ListAllocations(ctx context.Context, year int) ([]Allocation, error) {
	const query = `
SELECT
    fa.id,
    fa.finance_source_id,
    fs.code,
    fs.name,
    fa.allocation_type,
    fa.amount::BIGINT,
    COALESCE(fa.notes, ''),
    fa.created_at,
    fa.updated_at
FROM finance_allocations fa
JOIN finance_sources fs ON fs.id = fa.finance_source_id
WHERE fs.year = $1
ORDER BY fa.created_at DESC
`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Allocation, 0)
	for rows.Next() {
		var item Allocation
		if err := rows.Scan(
			&item.ID,
			&item.FinanceSourceID,
			&item.SourceCode,
			&item.SourceName,
			&item.AllocationType,
			&item.Amount,
			&item.Notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *repository) CreateSource(ctx context.Context, input SaveSourceRequest) (*Source, error) {
	const query = `
INSERT INTO finance_sources (year, code, name, amount, notes, is_active)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)
RETURNING id
`
	var id uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		input.Year,
		input.Code,
		input.Name,
		input.Amount,
		input.Notes,
		input.IsActive,
	).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	return r.GetSource(ctx, id)
}

func (r *repository) UpdateSource(ctx context.Context, id uuid.UUID, input SaveSourceRequest) (*Source, error) {
	const query = `
UPDATE finance_sources
SET
    year = $1,
    code = $2,
    name = $3,
    amount = $4,
    notes = NULLIF($5, ''),
    is_active = $6,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $7
`
	result, err := r.db.ExecContext(
		ctx,
		query,
		input.Year,
		input.Code,
		input.Name,
		input.Amount,
		input.Notes,
		input.IsActive,
		id,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicate
		}
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, ErrNotFound
	}
	return r.GetSource(ctx, id)
}

func (r *repository) DeleteSource(ctx context.Context, id uuid.UUID) error {
	var allocationCount int64
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM finance_allocations WHERE finance_source_id = $1`,
		id,
	).Scan(&allocationCount); err != nil {
		return err
	}
	if allocationCount > 0 {
		return ErrSourceInUse
	}

	var usageCount int64
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM finance_usages WHERE finance_source_id = $1`,
		id,
	).Scan(&usageCount); err != nil {
		return err
	}
	if usageCount > 0 {
		return ErrSourceInUse
	}

	result, err := r.db.ExecContext(ctx, `DELETE FROM finance_sources WHERE id = $1`, id)
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

func (r *repository) CreateAllocation(ctx context.Context, input SaveAllocationRequest) (*Allocation, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var totalAmount int64
	var isActive bool
	err = tx.QueryRowContext(
		ctx,
		`SELECT amount::BIGINT, is_active FROM finance_sources WHERE id = $1 FOR UPDATE`,
		input.FinanceSourceID,
	).Scan(&totalAmount, &isActive)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !isActive {
		return nil, ErrInsufficientFunds
	}

	var allocated int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(amount), 0)::BIGINT FROM finance_allocations WHERE finance_source_id = $1`,
		input.FinanceSourceID,
	).Scan(&allocated); err != nil {
		return nil, err
	}

	if input.Amount > totalAmount-allocated {
		return nil, ErrInsufficientFunds
	}

	const insertQuery = `
INSERT INTO finance_allocations (finance_source_id, allocation_type, amount, notes)
VALUES ($1, $2, $3, NULLIF($4, ''))
RETURNING id
`
	var id uuid.UUID
	if err := tx.QueryRowContext(
		ctx,
		insertQuery,
		input.FinanceSourceID,
		input.AllocationType,
		input.Amount,
		input.Notes,
	).Scan(&id); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.getAllocation(ctx, id)
}

func (r *repository) getAllocation(ctx context.Context, id uuid.UUID) (*Allocation, error) {
	const query = `
SELECT
    fa.id,
    fa.finance_source_id,
    fs.code,
    fs.name,
    fa.allocation_type,
    fa.amount::BIGINT,
    COALESCE(fa.notes, ''),
    fa.created_at,
    fa.updated_at
FROM finance_allocations fa
JOIN finance_sources fs ON fs.id = fa.finance_source_id
WHERE fa.id = $1
`
	var item Allocation
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID,
		&item.FinanceSourceID,
		&item.SourceCode,
		&item.SourceName,
		&item.AllocationType,
		&item.Amount,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *repository) DeleteAllocation(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var sourceID uuid.UUID
	var allocationType string
	var targetAmount int64

	err = tx.QueryRowContext(
		ctx,
		`SELECT finance_source_id, allocation_type, amount::BIGINT
         FROM finance_allocations
         WHERE id = $1
         FOR UPDATE`,
		id,
	).Scan(&sourceID, &allocationType, &targetAmount)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	var totalAllocated int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(amount), 0)::BIGINT
         FROM finance_allocations
         WHERE finance_source_id = $1 AND allocation_type = $2`,
		sourceID,
		allocationType,
	).Scan(&totalAllocated); err != nil {
		return err
	}

	var used int64
	if err := tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(SUM(amount), 0)::BIGINT
         FROM finance_usages
         WHERE finance_source_id = $1 AND budget_type = $2`,
		sourceID,
		allocationType,
	).Scan(&used); err != nil {
		return err
	}

	if totalAllocated-targetAmount < used {
		return ErrAllocationInUse
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM finance_allocations WHERE id = $1`, id)
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

	return tx.Commit()
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}
