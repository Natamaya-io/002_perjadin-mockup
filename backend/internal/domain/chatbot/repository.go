package chatbot

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetGUPAnalysis(ctx context.Context, filter PeriodFilter) (*GUPAnalysis, error)
	GetOutsideCityAnalysis(ctx context.Context, filter PeriodFilter, scopeUserID *uuid.UUID) (*OutsideCityAnalysis, error)
	GetInsideCityAnalysis(ctx context.Context, filter PeriodFilter, scopeUserID *uuid.UUID) (*InsideCityAnalysis, error)
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func periodBounds(filter PeriodFilter) (time.Time, time.Time) {
	month := time.January
	if filter.Month > 0 {
		month = time.Month(filter.Month)
	}
	start := time.Date(filter.Year, month, 1, 0, 0, 0, 0, time.UTC)
	if filter.Month > 0 {
		return start, start.AddDate(0, 1, 0)
	}
	return start, start.AddDate(1, 0, 0)
}

func (r *repository) GetGUPAnalysis(ctx context.Context, filter PeriodFilter) (*GUPAnalysis, error) {
	result := &GUPAnalysis{
		Monthly:        make([]GUPMonthlyRow, 0),
		TopExpenses:    make([]GUPExpenseRank, 0),
		LowestExpenses: make([]GUPExpenseRank, 0),
		AccountRanking: make([]GUPAccountRank, 0),
	}

	monthlyQuery := `
		SELECT
			fs.month_number,
			fs.month_name,
			fs.gup_label,
			COALESCE(ml.amount, 0)::BIGINT AS income,
			COALESCE(SUM(gt.paid_amount), 0)::BIGINT AS realization,
			COUNT(gt.id)::BIGINT AS transaction_count
		FROM funding_sources fs
		LEFT JOIN monthly_ls ml ON ml.funding_source_id = fs.id
		LEFT JOIN gup_transactions gt
			ON gt.funding_source_id = fs.id
		WHERE fs.year = $1
			AND ($2::int = 0 OR fs.month_number = $2)
		GROUP BY fs.id, fs.month_number, fs.month_name, fs.gup_label, ml.amount
		ORDER BY fs.month_number ASC
	`

	rows, err := r.db.QueryContext(ctx, monthlyQuery, filter.Year, filter.Month)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil rekap bulanan GUP: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row GUPMonthlyRow
		if err := rows.Scan(&row.MonthNumber, &row.MonthName, &row.GUPLabel, &row.Income, &row.Realization, &row.TransactionCount); err != nil {
			return nil, fmt.Errorf("gagal membaca rekap bulanan GUP: %w", err)
		}
		row.Balance = row.Income - row.Realization
		row.Status = gupUsageStatus(row.Income, row.Realization)
		result.Monthly = append(result.Monthly, row)
		result.Summary.TotalLS += row.Income
		result.Summary.TotalRealization += row.Realization
		result.Summary.TransactionCount += row.TransactionCount
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri rekap bulanan GUP: %w", err)
	}

	result.Summary.Balance = result.Summary.TotalLS - result.Summary.TotalRealization
	if result.Summary.TotalLS > 0 {
		result.Summary.RealizationPercentage = math.Round((float64(result.Summary.TotalRealization)/float64(result.Summary.TotalLS))*10000) / 100
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(gt.tax_amount), 0)::BIGINT
		FROM gup_transactions gt
		JOIN funding_sources fs ON fs.id = gt.funding_source_id
		WHERE fs.year = $1 AND ($2::int = 0 OR fs.month_number = $2)
	`, filter.Year, filter.Month).Scan(&result.Summary.TotalTax); err != nil {
		return nil, fmt.Errorf("gagal menghitung total pajak GUP: %w", err)
	}

	top, err := r.getGUPExpenseRanks(ctx, filter, false)
	if err != nil {
		return nil, err
	}
	result.TopExpenses = top

	lowest, err := r.getGUPExpenseRanks(ctx, filter, true)
	if err != nil {
		return nil, err
	}
	result.LowestExpenses = lowest

	accountRows, err := r.db.QueryContext(ctx, `
		SELECT
			ac.code,
			ac.mak,
			COALESCE(SUM(gt.paid_amount), 0)::BIGINT AS paid_amount,
			COUNT(gt.id)::BIGINT AS transaction_count
		FROM gup_transactions gt
		JOIN procurement_types pt ON pt.id = gt.procurement_type_id
		JOIN account_codes ac ON ac.id = pt.account_code_id
		JOIN funding_sources fs ON fs.id = gt.funding_source_id
		WHERE fs.year = $1 AND ($2::int = 0 OR fs.month_number = $2)
		GROUP BY ac.id, ac.code, ac.mak
		HAVING COALESCE(SUM(gt.paid_amount), 0) > 0
		ORDER BY paid_amount DESC, ac.code ASC
	`, filter.Year, filter.Month)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ranking akun GUP: %w", err)
	}
	defer accountRows.Close()

	for accountRows.Next() {
		var row GUPAccountRank
		if err := accountRows.Scan(&row.AccountCode, &row.MAK, &row.PaidAmount, &row.TransactionCount); err != nil {
			return nil, fmt.Errorf("gagal membaca ranking akun GUP: %w", err)
		}
		row.Rank = len(result.AccountRanking) + 1
		if result.Summary.TotalRealization > 0 {
			row.Percentage = math.Round((float64(row.PaidAmount)/float64(result.Summary.TotalRealization))*10000) / 100
		}
		result.AccountRanking = append(result.AccountRanking, row)
	}
	if err := accountRows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri ranking akun GUP: %w", err)
	}
	if len(result.AccountRanking) > 0 {
		largest := result.AccountRanking[0]
		result.LargestAccount = &largest
	}

	return result, nil
}

func (r *repository) getGUPExpenseRanks(ctx context.Context, filter PeriodFilter, ascending bool) ([]GUPExpenseRank, error) {
	order := "DESC"
	if ascending {
		order = "ASC"
	}
	query := fmt.Sprintf(`
		SELECT
			gt.business_id,
			gt.payment_description,
			pt.name,
			ac.code,
			ac.mak,
			COALESCE(fs.gup_label, ''),
			COALESCE(gt.recipient, ''),
			gt.receipt_date,
			gt.paid_amount::BIGINT
		FROM gup_transactions gt
		JOIN procurement_types pt ON pt.id = gt.procurement_type_id
		JOIN account_codes ac ON ac.id = pt.account_code_id
		LEFT JOIN funding_sources fs ON fs.id = gt.funding_source_id
		WHERE fs.year = $1
			AND ($2::int = 0 OR fs.month_number = $2)
			AND gt.paid_amount > 0
		ORDER BY gt.paid_amount %s, gt.receipt_date DESC, gt.business_id ASC
		LIMIT 3
	`, order)

	rows, err := r.db.QueryContext(ctx, query, filter.Year, filter.Month)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ranking transaksi GUP: %w", err)
	}
	defer rows.Close()

	result := make([]GUPExpenseRank, 0, 3)
	for rows.Next() {
		var row GUPExpenseRank
		var receiptDate time.Time
		if err := rows.Scan(
			&row.BusinessID,
			&row.PaymentDescription,
			&row.ProcurementTypeName,
			&row.AccountCode,
			&row.MAK,
			&row.FundingSourceLabel,
			&row.Recipient,
			&receiptDate,
			&row.PaidAmount,
		); err != nil {
			return nil, fmt.Errorf("gagal membaca ranking transaksi GUP: %w", err)
		}
		row.Rank = len(result) + 1
		row.ReceiptDate = receiptDate.Format("2006-01-02")
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri ranking transaksi GUP: %w", err)
	}
	return result, nil
}

func gupUsageStatus(income, realization int64) string {
	switch {
	case income == 0:
		return "Belum Tersedia"
	case realization == 0:
		return "Belum Digunakan"
	case realization < income:
		return "Digunakan Sebagian"
	case realization == income:
		return "Habis"
	default:
		return "Melebihi LS"
	}
}

func (r *repository) GetOutsideCityAnalysis(ctx context.Context, filter PeriodFilter, scopeUserID *uuid.UUID) (*OutsideCityAnalysis, error) {
	start, end := periodBounds(filter)
	var scope any
	if scopeUserID != nil {
		scope = *scopeUserID
	}

	const query = `
		SELECT
			u.id,
			u.name,
			u.email,
			COALESCE(u.nip, ''),
			COALESCE(u.jabatan, ''),
			COUNT(tr.id)::BIGINT AS total_trips,
			COUNT(tr.id) FILTER (
				WHERE tr.end_date::date <= CURRENT_DATE
				  AND COALESCE(tr.status, '') <> 'Rejected'
			)::BIGINT AS executed_trips,
			COALESCE(SUM(tr.total_cost) FILTER (
				WHERE tr.end_date::date <= CURRENT_DATE
				  AND COALESCE(tr.status, '') <> 'Rejected'
			), 0)::BIGINT AS executed_cost,
			COUNT(tr.id) FILTER (
				WHERE COALESCE(tr.payment_status, '') = 'Paid'
			)::BIGINT AS completed,
			COUNT(tr.id) FILTER (
				WHERE COALESCE(tr.payment_status, '') <> 'Paid'
				  AND (
					COALESCE(tr.report_status, '') = 'Completed'
					OR COALESCE(tr.status, '') IN ('Submitted', 'Approved')
				  )
			)::BIGINT AS in_progress,
			COUNT(tr.id) FILTER (
				WHERE COALESCE(tr.payment_status, '') <> 'Paid'
				  AND COALESCE(tr.report_status, '') <> 'Completed'
				  AND COALESCE(tr.status, '') IN ('Draft', 'Assigned')
			)::BIGINT AS assigned,
			COUNT(tr.id) FILTER (
				WHERE COALESCE(tr.payment_status, '') <> 'Paid'
				  AND COALESCE(tr.report_status, '') <> 'Completed'
				  AND COALESCE(tr.status, '') = 'Rejected'
			)::BIGINT AS rejected
		FROM users u
		LEFT JOIN travel_records tr
			ON tr.employee_id = u.id
			AND tr.deleted_at IS NULL
			AND LOWER(COALESCE(tr.type, '')) = 'luar_kota'
			AND tr.start_date >= $1
			AND tr.start_date < $2
		WHERE u.deleted_at IS NULL
			AND LOWER(TRIM(u.role)) = 'protokol'
			AND ($3::uuid IS NULL OR u.id = $3::uuid)
		GROUP BY u.id, u.name, u.email, u.nip, u.jabatan
		ORDER BY executed_cost DESC, LOWER(u.name) ASC
	`

	rows, err := r.db.QueryContext(ctx, query, start, end, scope)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil analisis perjalanan luar kota: %w", err)
	}
	defer rows.Close()

	result := &OutsideCityAnalysis{Users: make([]OutsideCityUserRow, 0)}
	for rows.Next() {
		var row OutsideCityUserRow
		if err := rows.Scan(
			&row.UserID,
			&row.Name,
			&row.Email,
			&row.NIP,
			&row.Position,
			&row.TotalTrips,
			&row.ExecutedTrips,
			&row.ExecutedCost,
			&row.Statuses.Completed,
			&row.Statuses.InProgress,
			&row.Statuses.Assigned,
			&row.Statuses.Rejected,
		); err != nil {
			return nil, fmt.Errorf("gagal membaca analisis perjalanan luar kota: %w", err)
		}
		result.Users = append(result.Users, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri analisis perjalanan luar kota: %w", err)
	}

	for i := range result.Users {
		result.Users[i].Rank = i + 1
		result.Summary.ProtocolUsers++
		result.Summary.TotalTrips += result.Users[i].TotalTrips
		result.Summary.ExecutedTrips += result.Users[i].ExecutedTrips
		result.Summary.ExecutedCost += result.Users[i].ExecutedCost
		result.Summary.Statuses.Assigned += result.Users[i].Statuses.Assigned
		result.Summary.Statuses.InProgress += result.Users[i].Statuses.InProgress
		result.Summary.Statuses.Completed += result.Users[i].Statuses.Completed
		result.Summary.Statuses.Rejected += result.Users[i].Statuses.Rejected
	}

	return result, nil
}

func (r *repository) GetInsideCityAnalysis(ctx context.Context, filter PeriodFilter, scopeUserID *uuid.UUID) (*InsideCityAnalysis, error) {
	start, end := periodBounds(filter)
	var scope any
	if scopeUserID != nil {
		scope = *scopeUserID
	}

	const query = `
		SELECT
			u.id,
			u.name,
			u.email,
			COALESCE(u.nip, ''),
			COALESCE(u.jabatan, ''),
			COUNT(DISTINCT dr.id)::BIGINT AS total_activities,
			COUNT(DISTINCT dr.id) FILTER (
				WHERE dr.service_date <= CURRENT_DATE
				  AND dr.status IN ('Pending', 'Sukses')
			)::BIGINT AS executed_activities,
			COALESCE(SUM(dro.amount) FILTER (
				WHERE dro.officer_group = 'SPJ'
				  AND dr.service_date <= CURRENT_DATE
				  AND dr.status IN ('Pending', 'Sukses')
			), 0)::BIGINT AS total_spj,
			COALESCE(SUM(dro.amount) FILTER (
				WHERE dro.officer_group = 'RIIL'
				  AND dr.service_date <= CURRENT_DATE
				  AND dr.status IN ('Pending', 'Sukses')
			), 0)::BIGINT AS total_riil,
			COUNT(DISTINCT dr.id) FILTER (WHERE dr.status = 'Draft')::BIGINT AS draft_count,
			COUNT(DISTINCT dr.id) FILTER (WHERE dr.status = 'Pending')::BIGINT AS pending_count,
			COUNT(DISTINCT dr.id) FILTER (WHERE dr.status = 'Sukses')::BIGINT AS success_count
		FROM users u
		LEFT JOIN dalkot_record_officers dro ON dro.employee_id = u.id
		LEFT JOIN dalkot_records dr
			ON dr.id = dro.dalkot_id
			AND dr.service_date >= $1
			AND dr.service_date < $2
		WHERE u.deleted_at IS NULL
			AND LOWER(TRIM(u.role)) = 'protokol'
			AND ($3::uuid IS NULL OR u.id = $3::uuid)
		GROUP BY u.id, u.name, u.email, u.nip, u.jabatan
		ORDER BY total_riil DESC, total_spj DESC, LOWER(u.name) ASC
	`

	rows, err := r.db.QueryContext(ctx, query, start, end, scope)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil analisis perjalanan dalam kota: %w", err)
	}
	defer rows.Close()

	result := &InsideCityAnalysis{Users: make([]InsideCityUserRow, 0)}
	for rows.Next() {
		var row InsideCityUserRow
		if err := rows.Scan(
			&row.UserID,
			&row.Name,
			&row.Email,
			&row.NIP,
			&row.Position,
			&row.TotalActivities,
			&row.ExecutedActivities,
			&row.TotalSPJ,
			&row.TotalRIIL,
			&row.Statuses.Draft,
			&row.Statuses.Pending,
			&row.Statuses.Success,
		); err != nil {
			return nil, fmt.Errorf("gagal membaca analisis perjalanan dalam kota: %w", err)
		}
		row.Balance = row.TotalSPJ - row.TotalRIIL
		result.Users = append(result.Users, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri analisis perjalanan dalam kota: %w", err)
	}

	// Ranking is based on RIIL because RIIL represents actual expenditure.
	sort.SliceStable(result.Users, func(i, j int) bool {
		if result.Users[i].TotalRIIL == result.Users[j].TotalRIIL {
			return result.Users[i].Name < result.Users[j].Name
		}
		return result.Users[i].TotalRIIL > result.Users[j].TotalRIIL
	})
	for i := range result.Users {
		result.Users[i].Rank = i + 1
		result.Summary.ProtocolUsers++
		result.Summary.TotalSPJ += result.Users[i].TotalSPJ
		result.Summary.TotalRIIL += result.Users[i].TotalRIIL
	}
	result.Summary.Balance = result.Summary.TotalSPJ - result.Summary.TotalRIIL

	const summaryQuery = `
		SELECT
			COUNT(DISTINCT dr.id)::BIGINT,
			COUNT(DISTINCT dr.id) FILTER (
				WHERE dr.service_date <= CURRENT_DATE
				  AND dr.status IN ('Pending', 'Sukses')
			)::BIGINT,
			COUNT(DISTINCT dr.id) FILTER (WHERE dr.status = 'Draft')::BIGINT,
			COUNT(DISTINCT dr.id) FILTER (WHERE dr.status = 'Pending')::BIGINT,
			COUNT(DISTINCT dr.id) FILTER (WHERE dr.status = 'Sukses')::BIGINT
		FROM dalkot_records dr
		WHERE dr.service_date >= $1
		  AND dr.service_date < $2
		  AND (
			$3::uuid IS NULL OR EXISTS (
				SELECT 1 FROM dalkot_record_officers dro
				WHERE dro.dalkot_id = dr.id AND dro.employee_id = $3::uuid
			)
		  )
	`
	if err := r.db.QueryRowContext(ctx, summaryQuery, start, end, scope).Scan(
		&result.Summary.TotalActivities,
		&result.Summary.ExecutedActivities,
		&result.Summary.Statuses.Draft,
		&result.Summary.Statuses.Pending,
		&result.Summary.Statuses.Success,
	); err != nil {
		return nil, fmt.Errorf("gagal menghitung ringkasan perjalanan dalam kota: %w", err)
	}

	return result, nil
}
