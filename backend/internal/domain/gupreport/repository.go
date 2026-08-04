package gupreport

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type Repository interface {
	GetRows(ctx context.Context, year int) ([]ReportRow, error)
	SaveBudget(ctx context.Context, year int, procurementTypeID uuid.UUID, amount int64) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetRows(ctx context.Context, year int) ([]ReportRow, error) {
	const query = `
		SELECT
			pt.id,
			pt.name,
			ac.code,
			ac.mak,
			COUNT(gt.id)::BIGINT AS transaction_count,
			COALESCE(SUM(gt.paid_amount), 0)::BIGINT AS realization,
			COALESCE(SUM(gt.value_amount), 0)::BIGINT AS submission_value,
			COALESCE(MAX(b.amount), 0)::BIGINT AS budget
		FROM procurement_types pt
		JOIN account_codes ac
			ON ac.id = pt.account_code_id
		LEFT JOIN gup_transactions gt
			ON gt.procurement_type_id = pt.id
			AND COALESCE(gt.receipt_date, gt.created_at::date) >= make_date($1, 1, 1)
			AND COALESCE(gt.receipt_date, gt.created_at::date) < make_date($1 + 1, 1, 1)
		LEFT JOIN budgets b
			ON b.procurement_type_id = pt.id
			AND b.year = $1
		WHERE pt.is_active = TRUE
		GROUP BY pt.id, pt.name, ac.code, ac.mak
		ORDER BY ac.code, pt.name
	`

	rows, err := r.db.QueryContext(ctx, query, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]ReportRow, 0)
	for rows.Next() {
		var row ReportRow
		if err := rows.Scan(
			&row.ProcurementTypeID,
			&row.ProcurementTypeName,
			&row.AccountCode,
			&row.MAK,
			&row.TransactionCount,
			&row.Realization,
			&row.SubmissionValue,
			&row.Budget,
		); err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *repository) SaveBudget(
	ctx context.Context,
	year int,
	procurementTypeID uuid.UUID,
	amount int64,
) error {
	const query = `
		INSERT INTO budgets (
			year,
			procurement_type_id,
			amount,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (year, procurement_type_id)
		DO UPDATE SET
			amount = EXCLUDED.amount,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, year, procurementTypeID, amount)
	return err
}
