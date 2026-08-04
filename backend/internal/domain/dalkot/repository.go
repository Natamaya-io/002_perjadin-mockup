package dalkot

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	List(ctx context.Context, officerID *uuid.UUID) ([]Record, error)
	GetRecap(ctx context.Context, filter RecapFilter) ([]RecapRow, RecapSummary, error)
	GetMonthlySPJRecap(ctx context.Context, filter RecapFilter) ([]MonthlySPJRecapRow, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Record, error)
	GetUsers(ctx context.Context) ([]UserOption, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]UserOption, error)
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

func (r *repository) GetUsers(ctx context.Context) ([]UserOption, error) {
	// to_jsonb digunakan agar kode tetap kompatibel dengan nama kolom profil
	// User yang mungkin berbeda antarversi project, misalnya jabatan/position.
	const query = `
		SELECT to_jsonb(u)
		FROM users AS u
		WHERE LOWER(TRIM(COALESCE(to_jsonb(u)->>'role', ''))) = 'protokol'
		  AND COALESCE(TRIM(to_jsonb(u)->>'deleted_at'), '') = ''
		  AND COALESCE(NULLIF(LOWER(TRIM(to_jsonb(u)->>'is_active')), ''), 'true')
		      NOT IN ('false', '0', 'no', 'tidak', 'nonaktif')
		  AND COALESCE(NULLIF(LOWER(TRIM(to_jsonb(u)->>'active')), ''), 'true')
		      NOT IN ('false', '0', 'no', 'tidak', 'nonaktif')
		ORDER BY LOWER(COALESCE(
			to_jsonb(u)->>'name',
			to_jsonb(u)->>'full_name',
			to_jsonb(u)->>'username',
			to_jsonb(u)->>'email',
			''
		)) ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil User Protokol: %w", err)
	}
	defer rows.Close()

	users := make([]UserOption, 0)

	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("gagal membaca User Protokol: %w", err)
		}

		item, err := decodeUserOption(raw)
		if err != nil {
			return nil, err
		}
		users = append(users, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri User Protokol: %w", err)
	}

	return users, nil
}

func decodeUserOption(raw json.RawMessage) (UserOption, error) {
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return UserOption{}, fmt.Errorf("format data User Protokol tidak valid: %w", err)
	}

	idText := readText(data, "id")
	id, err := uuid.Parse(idText)
	if err != nil || id == uuid.Nil {
		return UserOption{}, fmt.Errorf("ID User Protokol tidak valid: %q", idText)
	}

	name := readText(data, "name", "full_name", "fullName", "username")
	if name == "" {
		name = readText(data, "email")
	}
	if name == "" {
		return UserOption{}, fmt.Errorf("nama User Protokol dengan ID %s kosong", id.String())
	}

	return UserOption{
		ID:       id,
		Name:     name,
		Email:    readText(data, "email"),
		NIP:      readText(data, "nip", "employee_number", "employeeNumber"),
		Position: readText(data, "jabatan", "position", "job_title", "jobTitle"),
		Grade:    readText(data, "golongan", "grade", "pangkat"),
		Role:     strings.ToLower(readText(data, "role")),
		IsActive: true,
	}, nil
}

func readText(data map[string]any, keys ...string) string {
	for _, key := range keys {
		value, exists := data[key]
		if !exists || value == nil {
			continue
		}

		switch typed := value.(type) {
		case string:
			if text := strings.TrimSpace(typed); text != "" {
				return text
			}
		case float64:
			return strings.TrimSpace(fmt.Sprintf("%.0f", typed))
		case bool:
			if typed {
				return "true"
			}
			return "false"
		case map[string]any:
			for _, nestedKey := range []string{"name", "label", "value", "code"} {
				if nested, ok := typed[nestedKey].(string); ok && strings.TrimSpace(nested) != "" {
					return strings.TrimSpace(nested)
				}
			}
		}
	}
	return ""
}

func (r *repository) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]UserOption, error) {
	if len(ids) == 0 {
		return []UserOption{}, nil
	}

	allUsers, err := r.GetUsers(ctx)
	if err != nil {
		return nil, err
	}

	userByID := make(map[uuid.UUID]UserOption, len(allUsers))
	for _, item := range allUsers {
		userByID[item.ID] = item
	}

	result := make([]UserOption, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			return nil, fmt.Errorf("ID User Protokol tidak valid")
		}

		item, exists := userByID[id]
		if !exists {
			return nil, fmt.Errorf("User Protokol dengan ID %s tidak ditemukan atau tidak aktif", id.String())
		}
		result = append(result, item)
	}

	return result, nil
}

func (r *repository) List(ctx context.Context, officerID *uuid.UUID,) ([]Record, error) {
	const query = `
		SELECT
			dr.id,
			dr.business_id,
			COALESCE(dr.active_email, ''),
			dr.service_date,
			dr.category,
			dr.accompanied_official,
			dr.dalkot_type,
			COALESCE(dr.assignment_letter_number, ''),
			dr.assignment_letter_date,
			dr.activity_name,
			dr.location,
			COALESCE(dr.report_content, ''),
			dr.status,
			COALESCE(dr.total_spj_amount, 0),
			COALESCE(dr.total_riil_amount, 0),
			COALESCE(dr.document_filename, ''),
			COALESCE(dr.document_original_name, ''),
			COALESCE(dr.document_mime, ''),
			COALESCE(dr.document_size, 0),
			dr.created_at,
			dr.updated_at
		FROM dalkot_records AS dr
		WHERE (
			$1::uuid IS NULL
			OR EXISTS (
				SELECT 1
				FROM dalkot_record_officers AS dro
				WHERE dro.dalkot_id = dr.id
				  AND dro.employee_id = $1::uuid
			)
		)
		ORDER BY
			dr.created_at DESC,
			dr.service_date DESC,
			dr.business_id DESC
	`

	var officerFilter any
	if officerID != nil {
		officerFilter = *officerID
	}

	rows, err := r.db.QueryContext(
		ctx,
		query,
		officerFilter,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"gagal mengambil daftar Dalkot: %w",
			err,
		)
	}
	defer rows.Close()

	records := make([]Record, 0)

	for rows.Next() {
		item, err := scanRecord(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"gagal membaca daftar Dalkot: %w",
				err,
			)
		}

		item.Officers, err = r.listOfficers(
			ctx,
			item.ID,
		)
		if err != nil {
			return nil, err
		}

		item.OfficerCount = countUniqueOfficers(
			item.Officers,
		)

		records = append(records, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"gagal menelusuri daftar Dalkot: %w",
			err,
		)
	}

	return records, nil
}

func (r *repository) GetRecap(ctx context.Context, filter RecapFilter) ([]RecapRow, RecapSummary, error) {
	periodStart := time.Date(filter.Year, time.Month(filter.Month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	var officerID any
	if filter.OfficerID != nil {
		officerID = *filter.OfficerID
	}

	const rowsQuery = `
		SELECT
			dr.business_id,
			dr.service_date,
			dr.accompanied_official,
			dr.activity_name,
			dro.name,
			dr.category,
			dr.dalkot_type,
			COALESCE(dro.amount, 0)
		FROM dalkot_record_officers AS dro
		JOIN dalkot_records AS dr ON dr.id = dro.dalkot_id
		WHERE dro.officer_group = 'RIIL'
		  AND dr.service_date >= $1
		  AND dr.service_date < $2
		  AND ($3::uuid IS NULL OR dro.employee_id = $3::uuid)
		ORDER BY dr.service_date ASC, dr.created_at ASC, dro.sort_order ASC, LOWER(dro.name) ASC
	`

	rows, err := r.db.QueryContext(ctx, rowsQuery, periodStart, periodEnd, officerID)
	if err != nil {
		return nil, RecapSummary{}, fmt.Errorf("gagal mengambil rincian rekap Dalkot: %w", err)
	}
	defer rows.Close()

	items := make([]RecapRow, 0)
	for rows.Next() {
		var item RecapRow
		var serviceDate time.Time
		if err := rows.Scan(
			&item.BusinessID,
			&serviceDate,
			&item.AccompaniedOfficial,
			&item.ActivityName,
			&item.OfficerName,
			&item.Category,
			&item.DalkotType,
			&item.Nominal,
		); err != nil {
			return nil, RecapSummary{}, fmt.Errorf("gagal membaca rincian rekap Dalkot: %w", err)
		}
		item.ServiceDate = serviceDate.Format("2006-01-02")
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, RecapSummary{}, fmt.Errorf("gagal menelusuri rincian rekap Dalkot: %w", err)
	}

	const summaryQuery = `
		SELECT
			COALESCE(SUM(CASE WHEN dro.officer_group = 'SPJ' THEN dro.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN dro.officer_group = 'RIIL' THEN dro.amount ELSE 0 END), 0)
		FROM dalkot_record_officers AS dro
		JOIN dalkot_records AS dr ON dr.id = dro.dalkot_id
		WHERE dr.service_date >= $1
		  AND dr.service_date < $2
		  AND ($3::uuid IS NULL OR dro.employee_id = $3::uuid)
	`

	var summary RecapSummary
	if err := r.db.QueryRowContext(ctx, summaryQuery, periodStart, periodEnd, officerID).Scan(
		&summary.TotalSPJAmount,
		&summary.TotalRiilAmount,
	); err != nil {
		return nil, RecapSummary{}, fmt.Errorf("gagal menghitung ringkasan rekap Dalkot: %w", err)
	}

	summary.TotalRows = len(items)
	for _, item := range items {
		summary.TotalNominal += item.Nominal
	}
	summary.Balance = summary.TotalSPJAmount - summary.TotalRiilAmount
	summary.BalanceMinus = summary.Balance < 0

	return items, summary, nil
}

type monthlySPJAccumulator struct {
	officerID   *uuid.UUID
	name        string
	grade       string
	dates       map[string]struct{}
	letters     map[string]struct{}
	letterDates map[string]struct{}
}

func (r *repository) GetMonthlySPJRecap(ctx context.Context, filter RecapFilter) ([]MonthlySPJRecapRow, error) {
	periodStart := time.Date(filter.Year, time.Month(filter.Month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	const query = `
		SELECT
			dro.employee_id,
			dro.name,
			COALESCE(NULLIF(TRIM(dro.grade), ''), '-'),
			dr.service_date,
			COALESCE(dr.assignment_letter_number, ''),
			dr.assignment_letter_date
		FROM dalkot_record_officers AS dro
		JOIN dalkot_records AS dr ON dr.id = dro.dalkot_id
		WHERE dro.officer_group = 'SPJ'
		  AND dro.amount > 0
		  AND dr.dalkot_type = 'SPJ_RILL'
		  AND dr.service_date >= $1
		  AND dr.service_date < $2
		ORDER BY LOWER(dro.name) ASC, dr.service_date ASC, dr.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil rekap bulanan SPJ Dalkot: %w", err)
	}
	defer rows.Close()

	grouped := make(map[string]*monthlySPJAccumulator)
	order := make([]string, 0)

	for rows.Next() {
		var employeeID uuid.NullUUID
		var name, grade, letterNumber string
		var serviceDate time.Time
		var letterDate sql.NullTime
		if err := rows.Scan(&employeeID, &name, &grade, &serviceDate, &letterNumber, &letterDate); err != nil {
			return nil, fmt.Errorf("gagal membaca rekap bulanan SPJ Dalkot: %w", err)
		}

		key := "name:" + strings.ToLower(strings.TrimSpace(name))
		var officerID *uuid.UUID
		if employeeID.Valid {
			value := employeeID.UUID
			officerID = &value
			key = "id:" + value.String()
		}

		item, exists := grouped[key]
		if !exists {
			item = &monthlySPJAccumulator{
				officerID:   officerID,
				name:        strings.TrimSpace(name),
				grade:       strings.TrimSpace(grade),
				dates:       make(map[string]struct{}),
				letters:     make(map[string]struct{}),
				letterDates: make(map[string]struct{}),
			}
			if item.grade == "" {
				item.grade = "-"
			}
			grouped[key] = item
			order = append(order, key)
		}

		item.dates[serviceDate.Format("2006-01-02")] = struct{}{}
		if value := strings.TrimSpace(letterNumber); value != "" {
			item.letters[value] = struct{}{}
		}
		if letterDate.Valid {
			item.letterDates[letterDate.Time.Format("2006-01-02")] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri rekap bulanan SPJ Dalkot: %w", err)
	}

	sort.SliceStable(order, func(i, j int) bool {
		return strings.ToLower(grouped[order[i]].name) < strings.ToLower(grouped[order[j]].name)
	})

	result := make([]MonthlySPJRecapRow, 0, len(order))
	for _, key := range order {
		item := grouped[key]
		dates := sortedStringSet(item.dates)
		letters := sortedStringSet(item.letters)
		letterDates := sortedStringSet(item.letterDates)
		totalDays := len(dates)
		result = append(result, MonthlySPJRecapRow{
			OfficerID:               item.officerID,
			Name:                    item.name,
			Grade:                   item.grade,
			Dates:                   dates,
			Rate:                    SPJAmount,
			TotalDays:               totalDays,
			TotalAmount:             int64(totalDays) * SPJAmount,
			AssignmentLetterNumbers: letters,
			AssignmentLetterDates:   letterDates,
		})
	}

	return result, nil
}

func sortedStringSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Record, error) {
	const query = `
		SELECT
			id,
			business_id,
			COALESCE(active_email, ''),
			service_date,
			category,
			accompanied_official,
			dalkot_type,
			COALESCE(assignment_letter_number, ''),
			assignment_letter_date,
			activity_name,
			location,
			COALESCE(report_content, ''),
			status,
			COALESCE(total_spj_amount, 0),
			COALESCE(total_riil_amount, 0),
			COALESCE(document_filename, ''),
			COALESCE(document_original_name, ''),
			COALESCE(document_mime, ''),
			COALESCE(document_size, 0),
			created_at,
			updated_at
		FROM dalkot_records
		WHERE id = $1
	`

	item, err := scanRecord(r.db.QueryRowContext(ctx, query, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail Dalkot: %w", err)
	}

	item.Officers, err = r.listOfficers(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	item.OfficerCount = countUniqueOfficers(item.Officers)

	return &item, nil
}

func nextDalkotBusinessID(ctx context.Context, tx *sql.Tx) (string, error) {
	// Kunci transaksional mencegah dua pengajuan memperoleh nomor yang sama.
	// Nomor baru hanya menjadi permanen jika transaksi insert berhasil di-commit.
	if _, err := tx.ExecContext(
		ctx,
		`SELECT pg_advisory_xact_lock(hashtext('dalkot_business_id'))`,
	); err != nil {
		return "", fmt.Errorf("gagal mengunci nomor ID Dalkot: %w", err)
	}

	const query = `
		SELECT COALESCE(MAX(
			CASE
				WHEN business_id ~ '^DLK-[0-9]+$'
				THEN SUBSTRING(business_id FROM 5)::BIGINT
				ELSE 0
			END
		), 0) + 1
		FROM dalkot_records
	`

	var sequenceNumber int64
	if err := tx.QueryRowContext(ctx, query).Scan(&sequenceNumber); err != nil {
		return "", fmt.Errorf("gagal membuat ID Dalkot: %w", err)
	}

	return fmt.Sprintf("DLK-%03d", sequenceNumber), nil
}

func (r *repository) Create(ctx context.Context, input SaveInput) (*Record, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi Dalkot: %w", err)
	}
	defer tx.Rollback()

	businessID, err := nextDalkotBusinessID(ctx, tx)
	if err != nil {
		return nil, err
	}
	input.BusinessID = businessID

	const query = `
		INSERT INTO dalkot_records (
			business_id,
			active_email,
			service_date,
			category,
			accompanied_official,
			dalkot_type,
			policy_category,
			assignment_letter_number,
			assignment_letter_date,
			activity_name,
			location,
			report_content,
			status,
			total_spj_amount,
			total_riil_amount,
			total_amount,
			document_filename,
			document_original_name,
			document_mime,
			document_size
		)
		VALUES (
			$1, '', $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8,
			$9, $10, NULLIF($11, ''), $12, $13, $14, 0,
			NULLIF($15, ''), NULLIF($16, ''), NULLIF($17, ''), $18
		)
		RETURNING id
	`

	var id uuid.UUID
	err = tx.QueryRowContext(
		ctx,
		query,
		input.BusinessID,
		input.ServiceDate,
		input.Category,
		input.AccompaniedOfficial,
		input.DalkotType,
		input.PolicyCategory,
		input.AssignmentLetterNumber,
		input.AssignmentLetterDate,
		input.ActivityName,
		input.Location,
		input.ReportContent,
		input.Status,
		input.TotalSPJAmount,
		input.TotalRiilAmount,
		input.DocumentFilename,
		input.DocumentOriginalName,
		input.DocumentMIME,
		input.DocumentSize,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan Pengajuan Dalkot: %w", err)
	}

	if err := insertOfficers(ctx, tx, id, input.Officers, nil); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyelesaikan transaksi Dalkot: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, input SaveInput) (*Record, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("gagal memulai transaksi perubahan Dalkot: %w", err)
	}
	defer tx.Rollback()

	statusMap, err := readOfficerStatusMap(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	const query = `
		UPDATE dalkot_records
		SET
			service_date = $2,
			category = $3,
			accompanied_official = $4,
			dalkot_type = $5,
			policy_category = NULLIF($6, ''),
			assignment_letter_number = NULLIF($7, ''),
			assignment_letter_date = $8,
			activity_name = $9,
			location = $10,
			report_content = NULLIF($11, ''),
			status = $12,
			total_spj_amount = $13,
			total_riil_amount = $14,
			total_amount = 0,
			document_filename = NULLIF($15, ''),
			document_original_name = NULLIF($16, ''),
			document_mime = NULLIF($17, ''),
			document_size = $18,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := tx.ExecContext(
		ctx,
		query,
		id,
		input.ServiceDate,
		input.Category,
		input.AccompaniedOfficial,
		input.DalkotType,
		input.PolicyCategory,
		input.AssignmentLetterNumber,
		input.AssignmentLetterDate,
		input.ActivityName,
		input.Location,
		input.ReportContent,
		input.Status,
		input.TotalSPJAmount,
		input.TotalRiilAmount,
		input.DocumentFilename,
		input.DocumentOriginalName,
		input.DocumentMIME,
		input.DocumentSize,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui Pengajuan Dalkot: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("gagal memastikan perubahan Dalkot: %w", err)
	}
	if affected == 0 {
		return nil, ErrNotFound
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM dalkot_record_officers WHERE dalkot_id = $1`, id); err != nil {
		return nil, fmt.Errorf("gagal mengganti petugas Dalkot: %w", err)
	}

	if err := insertOfficers(ctx, tx, id, input.Officers, statusMap); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("gagal menyelesaikan perubahan Dalkot: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM dalkot_records WHERE id = $1`, id)
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
	var item Record
	var serviceDate time.Time
	var assignmentLetterDate sql.NullTime

	err := scanner.Scan(
		&item.ID,
		&item.BusinessID,
		&item.ActiveEmail,
		&serviceDate,
		&item.Category,
		&item.AccompaniedOfficial,
		&item.DalkotType,
		&item.AssignmentLetterNumber,
		&assignmentLetterDate,
		&item.ActivityName,
		&item.Location,
		&item.ReportContent,
		&item.Status,
		&item.TotalSPJAmount,
		&item.TotalRiilAmount,
		&item.DocumentFilename,
		&item.DocumentOriginalName,
		&item.DocumentMIME,
		&item.DocumentSize,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return Record{}, err
	}

	item.ServiceDate = serviceDate.Format("2006-01-02")
	if assignmentLetterDate.Valid {
		value := assignmentLetterDate.Time.Format("2006-01-02")
		item.AssignmentLetterDate = &value
	}
	item.Officers = make([]Officer, 0)
	if item.DocumentFilename != "" {
		item.DocumentURL = "/uploads/dalkot/" + item.DocumentFilename
	}
	return item, nil
}

func (r *repository) listOfficers(ctx context.Context, dalkotID uuid.UUID) ([]Officer, error) {
	const query = `
		SELECT
			id,
			dalkot_id,
			employee_id,
			officer_group,
			name,
			COALESCE(nip, ''),
			COALESCE(position, ''),
			COALESCE(grade, ''),
			amount,
			status,
			sort_order,
			created_at
		FROM dalkot_record_officers
		WHERE dalkot_id = $1
		ORDER BY
			CASE officer_group WHEN 'SPJ' THEN 1 WHEN 'RIIL' THEN 2 ELSE 3 END,
			sort_order ASC,
			LOWER(name) ASC
	`

	rows, err := r.db.QueryContext(ctx, query, dalkotID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil petugas Dalkot: %w", err)
	}
	defer rows.Close()

	items := make([]Officer, 0)
	for rows.Next() {
		var item Officer
		var employeeID uuid.NullUUID
		if err := rows.Scan(
			&item.ID,
			&item.DalkotID,
			&employeeID,
			&item.OfficerGroup,
			&item.Name,
			&item.NIP,
			&item.Position,
			&item.Grade,
			&item.Amount,
			&item.Status,
			&item.SortOrder,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("gagal membaca petugas Dalkot: %w", err)
		}

		if employeeID.Valid {
			value := employeeID.UUID
			item.EmployeeID = &value
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri petugas Dalkot: %w", err)
	}
	return items, nil
}

func readOfficerStatusMap(ctx context.Context, tx *sql.Tx, dalkotID uuid.UUID) (map[string]string, error) {
	rows, err := tx.QueryContext(
		ctx,
		`SELECT employee_id, officer_group, status FROM dalkot_record_officers WHERE dalkot_id = $1`,
		dalkotID,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca status petugas lama: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var employeeID uuid.NullUUID
		var group, status string
		if err := rows.Scan(&employeeID, &group, &status); err != nil {
			return nil, fmt.Errorf("gagal membaca status petugas lama: %w", err)
		}
		if employeeID.Valid {
			result[officerStatusKey(employeeID.UUID, group)] = status
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("gagal menelusuri status petugas lama: %w", err)
	}
	return result, nil
}

func insertOfficers(
	ctx context.Context,
	tx *sql.Tx,
	dalkotID uuid.UUID,
	officers []OfficerSaveInput,
	statusMap map[string]string,
) error {
	if len(officers) == 0 {
		return fmt.Errorf("Pengajuan Dalkot harus memiliki minimal satu petugas")
	}

	const query = `
		INSERT INTO dalkot_record_officers (
			dalkot_id,
			employee_id,
			officer_group,
			name,
			nip,
			position,
			grade,
			amount,
			status,
			sort_order
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10)
	`

	statement, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("gagal menyiapkan penyimpanan petugas Dalkot: %w", err)
	}
	defer statement.Close()

	seen := make(map[string]struct{})
	for _, item := range officers {
		if item.EmployeeID == uuid.Nil {
			return fmt.Errorf("ID petugas Dalkot tidak valid")
		}
		if item.OfficerGroup != OfficerGroupSPJ && item.OfficerGroup != OfficerGroupRiil {
			return fmt.Errorf("kelompok petugas Dalkot tidak valid")
		}

		key := officerStatusKey(item.EmployeeID, item.OfficerGroup)
		if _, exists := seen[key]; exists {
			return fmt.Errorf("petugas yang sama tidak boleh diduplikasi dalam kelompok %s", item.OfficerGroup)
		}
		seen[key] = struct{}{}

		status := normalizeOfficerStatus(item.Status)
		if previous, exists := statusMap[key]; exists {
			status = normalizeOfficerStatus(previous)
		}

		if _, err := statement.ExecContext(
			ctx,
			dalkotID,
			item.EmployeeID,
			item.OfficerGroup,
			strings.TrimSpace(item.Name),
			strings.TrimSpace(item.NIP),
			strings.TrimSpace(item.Position),
			strings.TrimSpace(item.Grade),
			item.Amount,
			status,
			item.SortOrder,
		); err != nil {
			return fmt.Errorf("gagal menyimpan petugas Dalkot: %w", err)
		}
	}
	return nil
}

func officerStatusKey(employeeID uuid.UUID, group string) string {
	return employeeID.String() + "|" + strings.ToUpper(strings.TrimSpace(group))
}

func normalizeOfficerStatus(value string) string {
	switch strings.TrimSpace(value) {
	case StatusDraft:
		return StatusDraft
	case StatusSukses:
		return StatusSukses
	default:
		return StatusPending
	}
}

func countUniqueOfficers(officers []Officer) int {
	unique := make(map[string]struct{})
	for _, officer := range officers {
		if officer.EmployeeID != nil {
			unique[officer.EmployeeID.String()] = struct{}{}
			continue
		}
		name := strings.ToLower(strings.TrimSpace(officer.Name))
		if name != "" {
			unique["name:"+name] = struct{}{}
		}
	}
	return len(unique)
}
