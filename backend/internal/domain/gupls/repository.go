package gupls

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

var ErrFundingSourceNotFound = errors.New("sumber dana GUP tidak ditemukan")

type Repository interface {
	GetRows(ctx context.Context, year int) ([]Row, error)
	SaveAmount(ctx context.Context, year int, fundingSourceID uuid.UUID, amount int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetRows(ctx context.Context, year int) ([]Row, error) {
	const query = `
		SELECT
			fs.id,
			fs.year,
			fs.month_number,
			fs.month_name,
			fs.gup_label,
			COALESCE(ml.amount, 0)::BIGINT AS ls_amount,
			COUNT(gt.id)::BIGINT AS transaction_count,
			COALESCE(SUM(gt.paid_amount), 0)::BIGINT AS realization
		FROM funding_sources fs
		LEFT JOIN monthly_ls ml
			ON ml.funding_source_id = fs.id
		LEFT JOIN gup_transactions gt
			ON gt.funding_source_id = fs.id
		WHERE fs.year = $1
		GROUP BY
			fs.id,
			fs.year,
			fs.month_number,
			fs.month_name,
			fs.gup_label,
			ml.amount
		ORDER BY fs.month_number
	`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Row, 0, 12)
	for rows.Next() {
		var row Row
		if err := rows.Scan(
			&row.FundingSourceID,
			&row.Year,
			&row.MonthNumber,
			&row.MonthName,
			&row.GUPLabel,
			&row.Amount,
			&row.TransactionCount,
			&row.Realization,
		); err != nil {
			return nil, err
		}

		row.AccountCode = LSAccountCode
		row.Remaining = row.Amount - row.Realization
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *repository) SaveAmount(
	ctx context.Context,
	year int,
	fundingSourceID uuid.UUID,
	amount int64,
) error {
	const query = `
		INSERT INTO monthly_ls (
			funding_source_id,
			amount,
			created_at,
			updated_at
		)
		SELECT
			fs.id,
			$3,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		FROM funding_sources fs
		WHERE fs.id = $1
			AND fs.year = $2
		ON CONFLICT (funding_source_id)
		DO UPDATE SET
			amount = EXCLUDED.amount,
			updated_at = CURRENT_TIMESTAMP
		RETURNING funding_source_id
	`

	var savedID uuid.UUID
	err := r.db.QueryRowContext(
		ctx,
		query,
		fundingSourceID,
		year,
		amount,
	).Scan(&savedID)

	if errors.Is(err, sql.ErrNoRows) {
		return ErrFundingSourceNotFound
	}

	return err
}
