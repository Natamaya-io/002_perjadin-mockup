package record

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/db"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/sqlc-dev/pqtype"
)

type CursorData struct {
	SPD    string `json:"spd"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

func encodeCursor(data CursorData) string {
	b, _ := json.Marshal(data)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(c string) CursorData {
	var data CursorData
	if c == "" {
		return data
	}
	// Backward compatibility with raw offset string
	if offset, err := strconv.Atoi(c); err == nil {
		data.Offset = offset
		return data
	}

	b, err := base64.RawURLEncoding.DecodeString(c)
	if err == nil {
		json.Unmarshal(b, &data)
	}
	return data
}

type Repository interface {
	NextSpdNumber(ctx context.Context) (int64, error)
	CreateTravelRecord(ctx context.Context, record *models.TravelRecord) error
	// CreateTravelRecordsBulk inserts all records in a single database transaction.
	// If any insert fails the entire batch is rolled back — no partial state.
	CreateTravelRecordsBulk(ctx context.Context, records []*models.TravelRecord) error
	GetTravelRecords(ctx context.Context, filters map[string]interface{}) ([]models.TravelRecord, error)
	GetTravelRecordByID(ctx context.Context, id uuid.UUID) (*models.TravelRecord, error)
	GetOverlappingRecords(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]models.TravelRecord, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdateTravelRecord(ctx context.Context, record *models.TravelRecord) error
	// SyncReportBySpd atomically pushes report updates to all other records sharing the same SPD.
	// This avoids pulling the entire database into application memory.
	SyncReportBySpd(ctx context.Context, spd string, sourceRecord *models.TravelRecord) error
	DeleteTravelRecord(ctx context.Context, id uuid.UUID) error
	DeleteTravelRecordsBySpd(ctx context.Context, spd string) error
	SyncSpdSequence(ctx context.Context) error
	GetDashboardSummary(ctx context.Context, role string, userIDStr string) (*models.DashboardSummary, error)
	GetPaginatedRecords(ctx context.Context, params models.PaginatedParams) (*models.PaginatedResponse, error)
	GetFactualSequenceNumber(ctx context.Context, id uuid.UUID, createdAt time.Time) (int, error)
	RecalculateSequenceNumbers(ctx context.Context) ([]uuid.UUID, error)
}

type repository struct {
	q *db.Queries
	d *sql.DB
}

func NewRepository(sqlDB *sql.DB) Repository {
	return &repository{
		q: db.New(sqlDB),
		d: sqlDB,
	}
}

// Helpers
func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func fromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func toNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func fromNullTime(nt sql.NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

func toNullFloat(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: true}
}

func fromNullFloat(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func toNullInt32(i int) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

func fromNullInt32(ni sql.NullInt32) int {
	if ni.Valid {
		return int(ni.Int32)
	}
	return 0
}

func toJsonb(d json.RawMessage) pqtype.NullRawMessage {
	if len(d) == 0 {
		return pqtype.NullRawMessage{RawMessage: json.RawMessage("{}"), Valid: true}
	}
	return pqtype.NullRawMessage{RawMessage: json.RawMessage(d), Valid: true}
}

func fromJsonb(j pqtype.NullRawMessage) json.RawMessage {
	if !j.Valid || len(j.RawMessage) == 0 {
		return nil
	}
	return json.RawMessage(j.RawMessage)
}

func mapDBUser(dbu db.User) models.User {
	return models.User{
		Base: models.Base{
			ID:        dbu.ID,
			CreatedAt: fromNullTime(dbu.CreatedAt),
			UpdatedAt: fromNullTime(dbu.UpdatedAt),
		},
		Email:        dbu.Email,
		Password:     dbu.Password,
		Name:         dbu.Name,
		Role:         dbu.Role,
		NIP:          fromNullString(dbu.Nip),
		NomorHP:      fromNullString(dbu.NomorHp),
		Pangkat:      fromNullString(dbu.Pangkat),
		Golongan:     fromNullString(dbu.Golongan),
		Jabatan:      fromNullString(dbu.Jabatan),
		TingkatBiaya: fromNullString(dbu.TingkatBiaya),
		SessionID:    fromNullString(dbu.SessionID),
		DemoPassword: fromNullString(dbu.DemoPassword),
	}
}

func mapDBRecord(dbr db.TravelRecord) models.TravelRecord {
	return models.TravelRecord{
		Base: models.Base{
			ID:        dbr.ID,
			CreatedAt: fromNullTime(dbr.CreatedAt),
			UpdatedAt: fromNullTime(dbr.UpdatedAt),
		},
		SPDNumber:        fromNullString(dbr.SpdNumber),
		EmployeeID:       dbr.EmployeeID,
		CreatorID:        dbr.CreatorID,
		StartDate:        fromNullTime(dbr.StartDate),
		EndDate:          fromNullTime(dbr.EndDate),
		Location:         fromNullString(dbr.Location),
		Province:         fromNullString(dbr.Province),
		Type:             fromNullString(dbr.Type),
		Purpose:          fromNullString(dbr.Purpose),
		Stakeholder:      fromNullString(dbr.Stakeholder),
		Agenda:           fromNullString(dbr.Agenda),
		Status:           fromNullString(dbr.Status),
		IsViewed:         dbr.IsViewed.Bool,
		ReportStatus:     fromNullString(dbr.ReportStatus),
		PaymentStatus:    fromNullString(dbr.PaymentStatus),
		TotalCost:        fromNullFloat(dbr.TotalCost),
		SuratTugasPath:   fromNullString(dbr.SuratTugasPath),
		SuratTugasNumber: fromNullString(dbr.SuratTugasNumber),
		SuratTugasDate:   fromNullTime(dbr.SuratTugasDate),
		SequenceNumber:   fromNullInt32(dbr.SequenceNumber),
	}
}

func mapDBLocation(dbl db.TravelLocation) models.TravelLocation {
	return models.TravelLocation{
		Base: models.Base{
			ID:        dbl.ID,
			CreatedAt: dbl.CreatedAt.Time,
			UpdatedAt: dbl.UpdatedAt.Time,
		},
		TravelRecordID: dbl.TravelRecordID,
		Location:       dbl.Location,
		Province:       dbl.Province,
		StartDate:      dbl.StartDate,
		EndDate:        dbl.EndDate,
	}
}

func mapDBCost(dbc db.TravelCost) models.TravelCost {
	return models.TravelCost{
		TravelRecordID:     dbc.TravelRecordID,
		TicketGo:           fromNullFloat(dbc.TicketGo),
		TicketBack:         fromNullFloat(dbc.TicketBack),
		DailyAllowanceDays: fromNullInt32(dbc.DailyAllowanceDays),
		DailyAllowanceRate: fromNullFloat(dbc.DailyAllowanceRate),
		HotelDays:          fromNullInt32(dbc.HotelDays),
		HotelRate:          fromNullFloat(dbc.HotelRate),
		LocalTransport:     fromNullFloat(dbc.LocalTransport),
		RegionalTransport:  fromNullFloat(dbc.RegionalTransport),
		TransportMode:      fromNullString(dbc.TransportMode),
		TransportAmount:    fromNullFloat(dbc.TransportAmount),
		OtherCost:          fromNullFloat(dbc.OtherCost),
		OtherCostDesc:      fromNullString(dbc.OtherCostDesc),
		ReceiptFiles:       fromJsonb(dbc.ReceiptFiles),
		TicketGoFile:       fromJsonb(dbc.TicketGoFile),
		TicketBackFile:     fromJsonb(dbc.TicketBackFile),
		BoardingPassFile:   fromJsonb(dbc.BoardingPassFile),
		HotelFile:          fromJsonb(dbc.HotelFile),
		TransportFile:      fromJsonb(dbc.TransportFile),
		AdditionalCosts:    fromJsonb(dbc.AdditionalCosts),
		Details:            fromJsonb(dbc.Details),
	}
}

func mapDBReport(dbrep db.TravelReport) models.TravelReport {

	return models.TravelReport{
		TravelRecordID: dbrep.TravelRecordID,
		Text:           fromNullString(dbrep.Text),
		SubmittedAt:    fromNullTime(dbrep.SubmittedAt),
		Files:          fromJsonb(dbrep.Files),
		SppdFile:       fromJsonb(dbrep.SppdFile),
		SuratTugasFile: fromJsonb(dbrep.SuratTugasFile),
		PPKName:        fromNullString(dbrep.PpkName),
		PPKNIP:         fromNullString(dbrep.PpkNip),
		BendaharaName:  fromNullString(dbrep.BendaharaName),
		BendaharaNIP:   fromNullString(dbrep.BendaharaNip),
		TanggalMerah:   fromJsonb(dbrep.TanggalMerah),
	}
}

func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	dbu, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u := mapDBUser(dbu)
	return &u, nil
}

// NextSpdNumber atomically fetches the next sequence value from PostgreSQL.
// Each call is guaranteed to return a unique, incrementing int64 regardless
// of concurrent callers — no application-level locking required.
func (r *repository) NextSpdNumber(ctx context.Context) (int64, error) {
	return r.q.NextSpdNumber(ctx)
}

// CreateTravelRecordsBulk inserts every record in [records] inside a single
// PostgreSQL transaction. Any failure triggers a full rollback of all preceding
// inserts — the database is never left in a partial state.
func (r *repository) CreateTravelRecordsBulk(ctx context.Context, records []*models.TravelRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.d.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("CreateTravelRecordsBulk: begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback() //nolint:errcheck
	}() // no-op after Commit; guard against any early return

	qtx := r.q.WithTx(tx)

	for _, record := range records {
		if record.ID == uuid.Nil {
			record.ID = uuid.New()
		}

		dbr, err := qtx.CreateTravelRecord(ctx, db.CreateTravelRecordParams{
			ID:               record.ID,
			SpdNumber:        toNullString(record.SPDNumber),
			EmployeeID:       record.EmployeeID,
			CreatorID:        record.CreatorID,
			StartDate:        toNullTime(record.StartDate),
			EndDate:          toNullTime(record.EndDate),
			Location:         toNullString(record.Location),
			Province:         toNullString(record.Province),
			Type:             toNullString(record.Type),
			Purpose:          toNullString(record.Purpose),
			Stakeholder:      toNullString(record.Stakeholder),
			Agenda:           toNullString(record.Agenda),
			Status:           toNullString(record.Status),
			IsViewed:         sql.NullBool{Bool: record.IsViewed, Valid: true},
			ReportStatus:     toNullString(record.ReportStatus),
			PaymentStatus:    toNullString(record.PaymentStatus),
			TotalCost:        toNullFloat(record.TotalCost),
			SuratTugasPath:   toNullString(record.SuratTugasPath),
			SuratTugasNumber: toNullString(record.SuratTugasNumber),
			SuratTugasDate:   toNullTime(record.SuratTugasDate),
		})
		if err != nil {
			return fmt.Errorf("CreateTravelRecordsBulk: insert record for employee %s: %w", record.EmployeeID, err)
		}

		// Sync the DB-assigned timestamps back onto the in-memory struct.
		originalLocations := record.Locations
		*record = mapDBRecord(dbr)
		record.Locations = originalLocations

		for _, loc := range record.Locations {
			_, err := qtx.CreateTravelLocation(ctx, db.CreateTravelLocationParams{
				ID:             uuid.New(),
				TravelRecordID: record.ID,
				Location:       loc.Location,
				Province:       loc.Province,
				StartDate:      loc.StartDate,
				EndDate:        loc.EndDate,
			})
			if err != nil {
				return fmt.Errorf("CreateTravelRecordsBulk: insert location for record %s: %w", record.ID, err)
			}
		}

		if record.Cost != nil {
			_, err := qtx.CreateTravelCost(ctx, db.CreateTravelCostParams{
				TravelRecordID:     record.ID,
				TicketGo:           toNullFloat(record.Cost.TicketGo),
				TicketBack:         toNullFloat(record.Cost.TicketBack),
				DailyAllowanceDays: toNullInt32(record.Cost.DailyAllowanceDays),
				DailyAllowanceRate: toNullFloat(record.Cost.DailyAllowanceRate),
				HotelDays:          toNullInt32(record.Cost.HotelDays),
				HotelRate:          toNullFloat(record.Cost.HotelRate),
				LocalTransport:     toNullFloat(record.Cost.LocalTransport),
				RegionalTransport:  toNullFloat(record.Cost.RegionalTransport),
				TransportMode:      toNullString(record.Cost.TransportMode),
				TransportAmount:    toNullFloat(record.Cost.TransportAmount),
				OtherCost:          toNullFloat(record.Cost.OtherCost),
				OtherCostDesc:      toNullString(record.Cost.OtherCostDesc),
				ReceiptFiles:       toJsonb(record.Cost.ReceiptFiles),
				TicketGoFile:       toJsonb(record.Cost.TicketGoFile),
				TicketBackFile:     toJsonb(record.Cost.TicketBackFile),
				BoardingPassFile:   toJsonb(record.Cost.BoardingPassFile),
				HotelFile:          toJsonb(record.Cost.HotelFile),
				TransportFile:      toJsonb(record.Cost.TransportFile),
				AdditionalCosts:    toJsonb(record.Cost.AdditionalCosts),
				Details:            toJsonb(record.Cost.Details),
			})
			if err != nil {
				return fmt.Errorf("CreateTravelRecordsBulk: insert cost for record %s: %w", record.ID, err)
			}
		}

		if record.Report != nil {
			_, err := qtx.CreateTravelReport(ctx, db.CreateTravelReportParams{
				TravelRecordID: record.ID,
				Text:           toNullString(record.Report.Text),
				SubmittedAt:    toNullTime(record.Report.SubmittedAt),
				Files:          toJsonb(record.Report.Files),
				SppdFile:       toJsonb(record.Report.SppdFile),
				SuratTugasFile: toJsonb(record.Report.SuratTugasFile),
				PpkName:        toNullString(record.Report.PPKName),
				PpkNip:         toNullString(record.Report.PPKNIP),
				BendaharaName:  toNullString(record.Report.BendaharaName),
				BendaharaNip:   toNullString(record.Report.BendaharaNIP),
				TanggalMerah:   toJsonb(record.Report.TanggalMerah),
			})
			if err != nil {
				return fmt.Errorf("CreateTravelRecordsBulk: insert report for record %s: %w", record.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateTravelRecordsBulk: commit: %w", err)
	}
	return nil
}

func (r *repository) CreateTravelRecord(ctx context.Context, record *models.TravelRecord) error {
	tx, err := r.d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback() //nolint:errcheck
	}()

	qtx := r.q.WithTx(tx)

	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}

	dbr, err := qtx.CreateTravelRecord(ctx, db.CreateTravelRecordParams{
		ID:               record.ID,
		SpdNumber:        toNullString(record.SPDNumber),
		EmployeeID:       record.EmployeeID,
		CreatorID:        record.CreatorID,
		StartDate:        toNullTime(record.StartDate),
		EndDate:          toNullTime(record.EndDate),
		Location:         toNullString(record.Location),
		Province:         toNullString(record.Province),
		Type:             toNullString(record.Type),
		Purpose:          toNullString(record.Purpose),
		Stakeholder:      toNullString(record.Stakeholder),
		Agenda:           toNullString(record.Agenda),
		Status:           toNullString(record.Status),
		IsViewed:         sql.NullBool{Bool: record.IsViewed, Valid: true},
		ReportStatus:     toNullString(record.ReportStatus),
		PaymentStatus:    toNullString(record.PaymentStatus),
		TotalCost:        toNullFloat(record.TotalCost),
		SuratTugasPath:   toNullString(record.SuratTugasPath),
		SuratTugasNumber: toNullString(record.SuratTugasNumber),
		SuratTugasDate:   toNullTime(record.SuratTugasDate),
	})
	if err != nil {
		return err
	}

	// Preserve locations before mapping (mapDBRecord might clear them as they aren't in the DB TravelRecord struct)
	originalLocations := record.Locations
	*record = mapDBRecord(dbr)
	record.Locations = originalLocations

	// Create locations - only if they don't already exist or as a clean batch
	for _, loc := range record.Locations {
		locID := uuid.New()
		_, err := qtx.CreateTravelLocation(ctx, db.CreateTravelLocationParams{
			ID:             locID,
			TravelRecordID: record.ID,
			Location:       loc.Location,
			Province:       loc.Province,
			StartDate:      loc.StartDate,
			EndDate:        loc.EndDate,
		})
		if err != nil {
			return err
		}
	}

	if record.Cost != nil {
		_, err := qtx.CreateTravelCost(ctx, db.CreateTravelCostParams{
			TravelRecordID:     record.ID,
			TicketGo:           toNullFloat(record.Cost.TicketGo),
			TicketBack:         toNullFloat(record.Cost.TicketBack),
			DailyAllowanceDays: toNullInt32(record.Cost.DailyAllowanceDays),
			DailyAllowanceRate: toNullFloat(record.Cost.DailyAllowanceRate),
			HotelDays:          toNullInt32(record.Cost.HotelDays),
			HotelRate:          toNullFloat(record.Cost.HotelRate),
			LocalTransport:     toNullFloat(record.Cost.LocalTransport),
			RegionalTransport:  toNullFloat(record.Cost.RegionalTransport),
			TransportMode:      toNullString(record.Cost.TransportMode),
			TransportAmount:    toNullFloat(record.Cost.TransportAmount),
			OtherCost:          toNullFloat(record.Cost.OtherCost),
			OtherCostDesc:      toNullString(record.Cost.OtherCostDesc),
			ReceiptFiles:       toJsonb(record.Cost.ReceiptFiles),
			TicketGoFile:       toJsonb(record.Cost.TicketGoFile),
			TicketBackFile:     toJsonb(record.Cost.TicketBackFile),
			BoardingPassFile:   toJsonb(record.Cost.BoardingPassFile),
			HotelFile:          toJsonb(record.Cost.HotelFile),
			TransportFile:      toJsonb(record.Cost.TransportFile),
			AdditionalCosts:    toJsonb(record.Cost.AdditionalCosts),
			Details:            toJsonb(record.Cost.Details),
		})
		if err != nil {
			return err
		}
	}

	if record.Report != nil {
		_, err := qtx.CreateTravelReport(ctx, db.CreateTravelReportParams{
			TravelRecordID: record.ID,
			Text:           toNullString(record.Report.Text),
			SubmittedAt:    toNullTime(record.Report.SubmittedAt),
			Files:          toJsonb(record.Report.Files),
			SppdFile:       toJsonb(record.Report.SppdFile),
			SuratTugasFile: toJsonb(record.Report.SuratTugasFile),
			PpkName:        toNullString(record.Report.PPKName),
			PpkNip:         toNullString(record.Report.PPKNIP),
			BendaharaName:  toNullString(record.Report.BendaharaName),
			BendaharaNip:   toNullString(record.Report.BendaharaNIP),
			TanggalMerah:   toJsonb(record.Report.TanggalMerah),
		})
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// SyncReportBySpd executes an optimized O(1) bulk update via SQL directly.
// It patches the travel_records and UPSERTs travel_reports for all employees
// under the same SPD number, fully averting O(N) memory blowout.
func (r *repository) SyncReportBySpd(ctx context.Context, spd string, src *models.TravelRecord) error {
	if src == nil || src.Report == nil || spd == "" {
		return nil
	}

	tx, err := r.d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback() //nolint:errcheck
	}()

	// 1. Update travel_records fields (ST Number, Date, Report Status)
	_, err = tx.ExecContext(ctx, `
		UPDATE travel_records SET
			surat_tugas_number = COALESCE(NULLIF($2, ''), surat_tugas_number),
			surat_tugas_date = CASE WHEN $3::timestamp IS NOT NULL THEN $3 ELSE surat_tugas_date END,
			report_status = COALESCE(NULLIF($4, ''), report_status),
			updated_at = CURRENT_TIMESTAMP
		WHERE spd_number = $1 AND deleted_at IS NULL
	`, spd, src.SuratTugasNumber, toNullTime(src.SuratTugasDate), src.ReportStatus)
	if err != nil {
		return fmt.Errorf("sync travel_records: %w", err)
	}

	// 2. Upsert travel_reports for all matching records
	_, err = tx.ExecContext(ctx, `
		INSERT INTO travel_reports (
			travel_record_id, text, submitted_at, files, sppd_file, surat_tugas_file, 
			ppk_name, ppk_nip, bendahara_name, bendahara_nip, tanggal_merah
		)
		SELECT 
			id, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		FROM travel_records 
		WHERE spd_number = $1 AND deleted_at IS NULL
		ON CONFLICT (travel_record_id) DO UPDATE SET
			text = EXCLUDED.text,
			submitted_at = EXCLUDED.submitted_at,
			files = EXCLUDED.files,
			sppd_file = EXCLUDED.sppd_file,
			surat_tugas_file = EXCLUDED.surat_tugas_file,
			ppk_name = EXCLUDED.ppk_name,
			ppk_nip = EXCLUDED.ppk_nip,
			bendahara_name = EXCLUDED.bendahara_name,
			bendahara_nip = EXCLUDED.bendahara_nip,
			tanggal_merah = EXCLUDED.tanggal_merah
	`,
		spd,
		toNullString(src.Report.Text),
		toNullTime(src.Report.SubmittedAt),
		toJsonb(src.Report.Files),
		toJsonb(src.Report.SppdFile),
		toJsonb(src.Report.SuratTugasFile),
		toNullString(src.Report.PPKName),
		toNullString(src.Report.PPKNIP),
		toNullString(src.Report.BendaharaName),
		toNullString(src.Report.BendaharaNIP),
		toJsonb(src.Report.TanggalMerah),
	)
	if err != nil {
		return fmt.Errorf("sync travel_reports upsert: %w", err)
	}

	return tx.Commit()
}

func (r *repository) GetTravelRecords(ctx context.Context, filters map[string]interface{}) ([]models.TravelRecord, error) {
	var statusFilter string
	if status, ok := filters["status"]; ok && status != "" {
		statusFilter, _ = status.(string) //nolint:errcheck
	}
	var spdFilter string
	if spd, ok := filters["spd"]; ok && spd != "" {
		spdFilter, _ = spd.(string) //nolint:errcheck
	}

	limit := 1000
	if l, ok := filters["limit"]; ok && l != "" {
		lStr, _ := l.(string) //nolint:errcheck
		if parsedLimit, err := strconv.Atoi(lStr); err == nil {
			limit = parsedLimit
		}
	}

	var dbrs []db.TravelRecord
	var err error

	if spdFilter != "" {
		// Use the optimized query that filters by SPD at the database level
		dbrs, err = r.q.GetRecordsBySPDs(ctx, []string{spdFilter})
		if err != nil {
			return nil, err
		}

		// Apply status filter in-memory if it exists
		if statusFilter != "" {
			var filtered []db.TravelRecord
			for _, dbr := range dbrs {
				if fromNullString(dbr.Status) == statusFilter {
					filtered = append(filtered, dbr)
				}
			}
			dbrs = filtered
		}
	} else {
		// Fallback to the paginated/limited query if no specific SPD is requested
		// Need a custom query if limit is dynamic, but since we rely on sqlc let's see.
		// For now we just return standard limited records.
		dbrs, err = r.q.GetTravelRecords(ctx, statusFilter)
		if err != nil {
			return nil, err
		}
		if len(dbrs) > limit {
			dbrs = dbrs[:limit]
		}
	}

	if len(dbrs) == 0 {
		return []models.TravelRecord{}, nil
	}

	var recordIDs []uuid.UUID
	var userIDs []uuid.UUID
	userIDMap := make(map[uuid.UUID]bool)

	for _, rec := range dbrs {
		recordIDs = append(recordIDs, rec.ID)

		if !userIDMap[rec.EmployeeID] {
			userIDs = append(userIDs, rec.EmployeeID)
			userIDMap[rec.EmployeeID] = true
		}
		if !userIDMap[rec.CreatorID] {
			userIDs = append(userIDs, rec.CreatorID)
			userIDMap[rec.CreatorID] = true
		}
	}

	// Fetch dynamic sequence numbers in bulk
	var seqNumbers []db.GetSequenceNumbersRow
	if len(recordIDs) > 0 {
		var errSeq error
		seqNumbers, errSeq = r.q.GetSequenceNumbers(ctx, recordIDs)
		if errSeq != nil {
			fmt.Printf("GetSequenceNumbers err: %v\n", errSeq)
		}
	}
	seqMap := make(map[uuid.UUID]int)
	for _, sq := range seqNumbers {
		seqMap[sq.ID] = int(sq.SequenceNumber)
	}

	// Fetch Users in Bulk
	var users []db.User
	if len(userIDs) > 0 {
		var errU error
		users, errU = r.q.GetUsersByIDs(ctx, userIDs)
		if errU != nil {
			fmt.Printf("GetUsersByIDs err: %v\n", errU)
		}
	}
	userMap := make(map[uuid.UUID]models.User)
	for _, u := range users {
		userMap[u.ID] = mapDBUser(u)
	}

	// Fetch Costs in Bulk
	var costs []db.TravelCost
	if len(recordIDs) > 0 {
		var errC error
		costs, errC = r.q.GetTravelCostsByRecordIDs(ctx, recordIDs)
		if errC != nil {
			fmt.Printf("GetTravelCosts err: %v\n", errC)
		}
	}
	costMap := make(map[uuid.UUID]*models.TravelCost)
	for _, c := range costs {
		mc := mapDBCost(c)

		// SECURITY & PERFORMANCE: DO NOT return huge Base64 PDF/Image files in the bulk list response!
		mc.TransportFile = nil
		mc.HotelFile = nil
		mc.BoardingPassFile = nil
		mc.TicketBackFile = nil
		mc.TicketGoFile = nil
		mc.ReceiptFiles = nil


		costMap[c.TravelRecordID] = &mc
	}

	// Fetch Reports in Bulk
	var reports []db.TravelReport
	if len(recordIDs) > 0 {
		var errR error
		reports, errR = r.q.GetTravelReportsByRecordIDs(ctx, recordIDs)
		if errR != nil {
			fmt.Printf("GetTravelReports err: %v\n", errR)
		}
	}
	reportMap := make(map[uuid.UUID]*models.TravelReport)
	for _, rep := range reports {
		mr := mapDBReport(rep)
		
		// SECURITY & PERFORMANCE: Strip huge Base64 files from Report bulk response
		mr.Files = nil
		mr.SppdFile = nil
		mr.SuratTugasFile = nil
		
		reportMap[rep.TravelRecordID] = &mr
	}

	// Fetch Locations in Bulk
	var locs []db.TravelLocation
	if len(recordIDs) > 0 {
		var errL error
		locs, errL = r.q.GetTravelLocationsByRecordIDs(ctx, recordIDs)
		if errL != nil {
			fmt.Printf("GetTravelLocations err: %v\n", errL)
		}
	}
	locMap := make(map[uuid.UUID][]models.TravelLocation)
	for _, l := range locs {
		locMap[l.TravelRecordID] = append(locMap[l.TravelRecordID], mapDBLocation(l))
	}

	var travelRecords []models.TravelRecord
	for _, dbr := range dbrs {
		tr := mapDBRecord(dbr)

		if u, ok := userMap[dbr.EmployeeID]; ok {
			tr.Employee = u
		}
		if u, ok := userMap[dbr.CreatorID]; ok {
			tr.Creator = u
		}
		if c, ok := costMap[dbr.ID]; ok {
			tr.Cost = c
		}
		if rep, ok := reportMap[dbr.ID]; ok {
			tr.Report = rep
		}
		if l, ok := locMap[dbr.ID]; ok {
			tr.Locations = l
		}

		if seq, ok := seqMap[dbr.ID]; ok {
			tr.SequenceNumber = seq
		}

		travelRecords = append(travelRecords, tr)
	}

	return travelRecords, nil
}

func (r *repository) RecalculateSequenceNumbers(ctx context.Context) ([]uuid.UUID, error) {
	query := `
		WITH group_min_time AS (
			SELECT spd_number, MIN(created_at) as min_time
			FROM travel_records
			WHERE deleted_at IS NULL
			GROUP BY spd_number
		),
		ranked_records AS (
			SELECT 
				t.id, 
				ROW_NUMBER() OVER (
					ORDER BY g.min_time ASC, t.spd_number ASC, t.created_at ASC, t.id ASC
				) as dynamic_sequence
			FROM travel_records t
			JOIN group_min_time g ON t.spd_number = g.spd_number
			WHERE t.deleted_at IS NULL
		)
		UPDATE travel_records tr
		SET sequence_number = r.dynamic_sequence
		FROM ranked_records r
		WHERE tr.id = r.id AND (tr.sequence_number IS NULL OR tr.sequence_number != r.dynamic_sequence)
		RETURNING tr.id
	`
	rows, err := r.d.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updatedIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err == nil {
			updatedIDs = append(updatedIDs, id)
		}
	}
	return updatedIDs, nil
}

func (r *repository) GetTravelRecordByID(ctx context.Context, id uuid.UUID) (*models.TravelRecord, error) {
	dbr, err := r.q.GetTravelRecordByID(ctx, id)
	if err != nil {
		return nil, err
	}

	rec := mapDBRecord(dbr)
	recPtr := &rec

	emp, errEmp := r.GetUserByID(ctx, recPtr.EmployeeID)
	if errEmp != nil {
		fmt.Printf("GetUserByID err: %v\n", errEmp)
	}
	if emp != nil {
		recPtr.Employee = *emp
	}

	creator, errCr := r.GetUserByID(ctx, recPtr.CreatorID)
	if errCr != nil {
		fmt.Printf("GetUserByID err: %v\n", errCr)
	}
	if creator != nil {
		recPtr.Creator = *creator
	}

	dbc, err := r.q.GetTravelCostByRecordID(ctx, recPtr.ID)
	if err == nil {
		cost := mapDBCost(dbc)
		recPtr.Cost = &cost
	}

	dbrep, err := r.q.GetTravelReportByRecordID(ctx, recPtr.ID)
	if err == nil {
		rep := mapDBReport(dbrep)
		recPtr.Report = &rep
	}

	dbls, err := r.q.GetTravelLocationsByRecordID(ctx, recPtr.ID)
	if err == nil {
		locs := make([]models.TravelLocation, len(dbls))
		for j, dbl := range dbls {
			locs[j] = mapDBLocation(dbl)
		}
		recPtr.Locations = locs
	}

	return recPtr, nil
}

func (r *repository) GetDashboardSummary(ctx context.Context, role string, userIDStr string) (*models.DashboardSummary, error) {
	summary := &models.DashboardSummary{}

	var nullUserID uuid.NullUUID
	if role == "protokol" && userIDStr != "" {
		if uid, err := uuid.Parse(userIDStr); err == nil {
			nullUserID = uuid.NullUUID{UUID: uid, Valid: true}
		}
	}

	var wg sync.WaitGroup
	var errs []error
	var errMu sync.Mutex

	// 1. Total Trips
	wg.Add(1)
	go func() {
		defer wg.Done()
		total, err := r.q.GetTotalTripsCount(ctx, nullUserID)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("GetTotalTripsCount: %w", err))
			errMu.Unlock()
		} else {
			summary.TotalTrips = total
		}
	}()

	// 2. Active Trips
	wg.Add(1)
	go func() {
		defer wg.Done()
		active, err := r.q.GetActiveTripsCount(ctx, nullUserID)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("GetActiveTripsCount: %w", err))
			errMu.Unlock()
		} else {
			summary.ActiveTrips = active
		}
	}()

	// 3. Status Counts
	wg.Add(1)
	go func() {
		defer wg.Done()
		counts, err := r.q.GetDashboardStatusCounts(ctx, nullUserID)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("GetDashboardStatusCounts: %w", err))
			errMu.Unlock()
		} else {
			for _, row := range counts {
				status := fromNullString(row.Status)
				paymentStatus := fromNullString(row.PaymentStatus)
				reportStatus := fromNullString(row.ReportStatus)

				switch {
				case paymentStatus == "Paid":
					summary.StatusCompleted += row.Count
				case reportStatus == "Completed":
					summary.StatusInProgress += row.Count
				case status == "Submitted" || status == "Approved":
					summary.StatusInProgress += row.Count
				case status == "Draft" || status == "Assigned":
					summary.StatusAssigned += row.Count
				case status == "Rejected":
					summary.StatusRejected += row.Count
				}
			}
		}
	}()

	// 4. Report Counts
	wg.Add(1)
	go func() {
		defer wg.Done()
		counts, err := r.q.GetDashboardReportCounts(ctx, nullUserID)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("GetDashboardReportCounts: %w", err))
			errMu.Unlock()
		} else {
			for _, row := range counts {
				reportStatus := fromNullString(row.ReportStatus)
				if reportStatus == "Completed" {
					summary.ReportCompleted += row.Count
				} else {
					summary.ReportPending += row.Count
				}
			}
		}
	}()

	// 5. Recent Records
	wg.Add(1)
	go func() {
		defer wg.Done()
		dbRecs, err := r.q.GetRecentRecords(ctx, nullUserID)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("GetRecentRecords: %w", err))
			errMu.Unlock()
		} else {
			recs := make([]models.TravelRecord, len(dbRecs))
			var userIDs []uuid.UUID
			userSet := make(map[uuid.UUID]bool)
			for _, dbRec := range dbRecs {
				if !userSet[dbRec.EmployeeID] {
					userIDs = append(userIDs, dbRec.EmployeeID)
					userSet[dbRec.EmployeeID] = true
				}
			}

			userMap := make(map[uuid.UUID]models.User)
			if len(userIDs) > 0 {
				if users, err := r.q.GetUsersByIDs(ctx, userIDs); err == nil {
					for _, u := range users {
						userMap[u.ID] = mapDBUser(u)
					}
				}
			}

			for i, dbRec := range dbRecs {
				rec := mapDBRecord(dbRec)
				if emp, ok := userMap[rec.EmployeeID]; ok {
					rec.Employee = emp
				}
				recs[i] = rec
			}
			summary.RecentRecords = recs
		}
	}()

	// 6. Budgets
	wg.Add(1)
	go func() {
		defer wg.Done()
		dbBudgets, err := r.q.GetDashboardBudgets(ctx, nullUserID)
		if err != nil {
			errMu.Lock()
			errs = append(errs, fmt.Errorf("GetDashboardBudgets: %w", err))
			errMu.Unlock()
		} else {
			budgets := make([]models.DashboardBudget, len(dbBudgets))
			for i, dbb := range dbBudgets {
				budgets[i] = models.DashboardBudget{
					Year:  int(dbb.Year),
					Month: int(dbb.Month),
					Total: dbb.Total,
				}
			}
			summary.Budgets = budgets
		}
	}()

	wg.Wait()

	if len(errs) > 0 {
		fmt.Printf("[Dashboard Summary] Encountered %d errors during aggregation. Partial data will be returned. First error: %v\n", len(errs), errs[0])
	}

	return summary, nil
}

func (r *repository) GetOverlappingRecords(ctx context.Context, employeeID uuid.UUID, startDate, endDate time.Time) ([]models.TravelRecord, error) {
	dbrs, err := r.q.GetOverlappingRecords(ctx, db.GetOverlappingRecordsParams{
		EmployeeID:   employeeID,
		NewEndDate:   toNullTime(endDate),
		NewStartDate: toNullTime(startDate),
	})
	if err != nil {
		return nil, err
	}

	var records []models.TravelRecord
	for _, dbr := range dbrs {
		records = append(records, mapDBRecord(dbr))
	}
	return records, nil
}

func (r *repository) UpdateTravelRecord(ctx context.Context, record *models.TravelRecord) error {
	tx, err := r.d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback() //nolint:errcheck
	}()

	qtx := r.q.WithTx(tx)

	_, err = qtx.UpdateTravelRecord(ctx, db.UpdateTravelRecordParams{
		ID:               record.ID,
		SpdNumber:        toNullString(record.SPDNumber),
		EmployeeID:       record.EmployeeID,
		CreatorID:        record.CreatorID,
		StartDate:        toNullTime(record.StartDate),
		EndDate:          toNullTime(record.EndDate),
		Location:         toNullString(record.Location),
		Province:         toNullString(record.Province),
		Type:             toNullString(record.Type),
		Purpose:          toNullString(record.Purpose),
		Stakeholder:      toNullString(record.Stakeholder),
		Agenda:           toNullString(record.Agenda),
		Status:           toNullString(record.Status),
		IsViewed:         sql.NullBool{Bool: record.IsViewed, Valid: true},
		ReportStatus:     toNullString(record.ReportStatus),
		PaymentStatus:    toNullString(record.PaymentStatus),
		TotalCost:        toNullFloat(record.TotalCost),
		SuratTugasPath:   toNullString(record.SuratTugasPath),
		SuratTugasNumber: toNullString(record.SuratTugasNumber),
		SuratTugasDate:   toNullTime(record.SuratTugasDate),
	})
	if err != nil {
		return fmt.Errorf("UpdateTravelRecord: %w", err)
	}

	// Update locations (simpler to delete and recreate)
	if err := qtx.DeleteTravelLocationsByRecordID(ctx, record.ID); err != nil {
		return fmt.Errorf("DeleteTravelLocationsByRecordID: %w", err)
	}
	for i, loc := range record.Locations {
		// Always generate a new ID to prevent primary key conflict with the soft-deleted row
		loc.ID = uuid.New()
		record.Locations[i].ID = loc.ID

		_, err := qtx.CreateTravelLocation(ctx, db.CreateTravelLocationParams{
			ID:             loc.ID,
			TravelRecordID: record.ID,
			Location:       loc.Location,
			Province:       loc.Province,
			StartDate:      loc.StartDate,
			EndDate:        loc.EndDate,
		})
		if err != nil {
			return fmt.Errorf("CreateTravelLocation: %w", err)
		}
	}

	if record.Cost != nil {
		_, err := qtx.UpdateTravelCost(ctx, db.UpdateTravelCostParams{
			TravelRecordID:     record.ID,
			TicketGo:           toNullFloat(record.Cost.TicketGo),
			TicketBack:         toNullFloat(record.Cost.TicketBack),
			DailyAllowanceDays: toNullInt32(record.Cost.DailyAllowanceDays),
			DailyAllowanceRate: toNullFloat(record.Cost.DailyAllowanceRate),
			HotelDays:          toNullInt32(record.Cost.HotelDays),
			HotelRate:          toNullFloat(record.Cost.HotelRate),
			LocalTransport:     toNullFloat(record.Cost.LocalTransport),
			RegionalTransport:  toNullFloat(record.Cost.RegionalTransport),
			TransportMode:      toNullString(record.Cost.TransportMode),
			TransportAmount:    toNullFloat(record.Cost.TransportAmount),
			OtherCost:          toNullFloat(record.Cost.OtherCost),
			OtherCostDesc:      toNullString(record.Cost.OtherCostDesc),
			ReceiptFiles:       toJsonb(record.Cost.ReceiptFiles),
			TicketGoFile:       toJsonb(record.Cost.TicketGoFile),
			TicketBackFile:     toJsonb(record.Cost.TicketBackFile),
			BoardingPassFile:   toJsonb(record.Cost.BoardingPassFile),
			HotelFile:          toJsonb(record.Cost.HotelFile),
			TransportFile:      toJsonb(record.Cost.TransportFile),
			AdditionalCosts:    toJsonb(record.Cost.AdditionalCosts),
			Details:            toJsonb(record.Cost.Details),
		})
		if err != nil {
			_, err = qtx.CreateTravelCost(ctx, db.CreateTravelCostParams{
				TravelRecordID:     record.ID,
				TicketGo:           toNullFloat(record.Cost.TicketGo),
				TicketBack:         toNullFloat(record.Cost.TicketBack),
				DailyAllowanceDays: toNullInt32(record.Cost.DailyAllowanceDays),
				DailyAllowanceRate: toNullFloat(record.Cost.DailyAllowanceRate),
				HotelDays:          toNullInt32(record.Cost.HotelDays),
				HotelRate:          toNullFloat(record.Cost.HotelRate),
				LocalTransport:     toNullFloat(record.Cost.LocalTransport),
				RegionalTransport:  toNullFloat(record.Cost.RegionalTransport),
				TransportMode:      toNullString(record.Cost.TransportMode),
				TransportAmount:    toNullFloat(record.Cost.TransportAmount),
				OtherCost:          toNullFloat(record.Cost.OtherCost),
				OtherCostDesc:      toNullString(record.Cost.OtherCostDesc),
				ReceiptFiles:       toJsonb(record.Cost.ReceiptFiles),
				TicketGoFile:       toJsonb(record.Cost.TicketGoFile),
				TicketBackFile:     toJsonb(record.Cost.TicketBackFile),
				BoardingPassFile:   toJsonb(record.Cost.BoardingPassFile),
				HotelFile:          toJsonb(record.Cost.HotelFile),
				TransportFile:      toJsonb(record.Cost.TransportFile),
				AdditionalCosts:    toJsonb(record.Cost.AdditionalCosts),
				Details:            toJsonb(record.Cost.Details),
			})
			if err != nil {
				return fmt.Errorf("CreateTravelCost fallback: %w", err)
			}
		}
	}

	if record.Report != nil {
		_, err := qtx.UpdateTravelReport(ctx, db.UpdateTravelReportParams{
			TravelRecordID: record.ID,
			Text:           toNullString(record.Report.Text),
			SubmittedAt:    toNullTime(record.Report.SubmittedAt),
			Files:          toJsonb(record.Report.Files),
			SppdFile:       toJsonb(record.Report.SppdFile),
			SuratTugasFile: toJsonb(record.Report.SuratTugasFile),
			PpkName:        toNullString(record.Report.PPKName),
			PpkNip:         toNullString(record.Report.PPKNIP),
			BendaharaName:  toNullString(record.Report.BendaharaName),
			BendaharaNip:   toNullString(record.Report.BendaharaNIP),
			TanggalMerah:   toJsonb(record.Report.TanggalMerah),
		})
		if err != nil {
			_, err = qtx.CreateTravelReport(ctx, db.CreateTravelReportParams{
				TravelRecordID: record.ID,
				Text:           toNullString(record.Report.Text),
				SubmittedAt:    toNullTime(record.Report.SubmittedAt),
				Files:          toJsonb(record.Report.Files),
				SppdFile:       toJsonb(record.Report.SppdFile),
				SuratTugasFile: toJsonb(record.Report.SuratTugasFile),
				PpkName:        toNullString(record.Report.PPKName),
				PpkNip:         toNullString(record.Report.PPKNIP),
				BendaharaName:  toNullString(record.Report.BendaharaName),
				BendaharaNip:   toNullString(record.Report.BendaharaNIP),
				TanggalMerah:   toJsonb(record.Report.TanggalMerah),
			})
			if err != nil {
				return fmt.Errorf("CreateTravelReport fallback: %w", err)
			}
		}
	}

	return tx.Commit()
}

func (r *repository) DeleteTravelRecord(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteTravelRecord(ctx, id)
}

func (r *repository) DeleteTravelRecordsBySpd(ctx context.Context, spd string) error {
	tx, err := r.d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback() //nolint:errcheck
	}()

	qtx := r.q.WithTx(tx)
	spdStr := toNullString(spd)

	if err := qtx.DeleteTravelLocationsBySpd(ctx, spdStr); err != nil {
		return err
	}

	if err := qtx.DeleteTravelRecordBySpd(ctx, spdStr); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *repository) SyncSpdSequence(ctx context.Context) error {
	return r.q.SyncSpdSequence(ctx)
}

func (r *repository) GetPaginatedRecords(ctx context.Context, params models.PaginatedParams) (*models.PaginatedResponse, error) {
	if params.Limit < 1 {
		params.Limit = 10
	}

	cursorData := decodeCursor(params.Cursor)
	offset := cursorData.Offset

	// 1. Build common WHERE clause and arguments
	whereClause := "WHERE travel_records.deleted_at IS NULL"
	var args []interface{}
	argId := 1

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND travel_records.status = $%d", argId)
		args = append(args, params.Status)
		argId++
	}
	if params.ReportStatus != "" {
		whereClause += fmt.Sprintf(" AND travel_records.report_status = $%d", argId)
		args = append(args, params.ReportStatus)
		argId++
	}
	if params.PaymentStatus != "" {
		whereClause += fmt.Sprintf(" AND travel_records.payment_status = $%d", argId)
		args = append(args, params.PaymentStatus)
		argId++
	}

	joinClause := ""
	if params.Search != "" {
		joinClause = "LEFT JOIN users ON travel_records.employee_id = users.id"
		whereClause += fmt.Sprintf(` AND (
			travel_records.spd_number ILIKE '%%' || $%d || '%%'
			OR travel_records.location ILIKE '%%' || $%d || '%%'
			OR users.name ILIKE '%%' || $%d || '%%'
		)`, argId, argId, argId)
		args = append(args, params.Search)
		argId++
	}

	if params.StartDate != nil {
		whereClause += fmt.Sprintf(" AND travel_records.start_date >= $%d", argId)
		args = append(args, params.StartDate)
		argId++
	}
	if params.EndDate != nil {
		whereClause += fmt.Sprintf(" AND travel_records.start_date <= $%d", argId)
		args = append(args, params.EndDate)
		argId++
	}
	if params.UserID != nil {
		whereClause += fmt.Sprintf(" AND (travel_records.employee_id = $%d OR travel_records.creator_id = $%d)", argId, argId)
		args = append(args, params.UserID)
		argId++
	}

	// 2. Prepare parallel queries
	var wg sync.WaitGroup
	var spdStrings []string
	var totalItems int64
	var totalRecords int64
	var errMain, errCountSPDs, errCountRecords error

	wg.Add(3)

	// A. Main Query (Get Paginated SPDs)
	go func() {
		defer wg.Done()

		mainWhereClause := whereClause
		mainArgs := append([]interface{}{}, args...)
		mainArgId := argId

		if cursorData.SPD != "" && (params.SortBy == "spj-desc" || params.SortBy == "") {
			mainWhereClause += fmt.Sprintf(" AND CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) < CAST(SUBSTRING($%d FROM 'ID-SPJ-([0-9]+)') AS NUMERIC)", mainArgId)
			mainArgs = append(mainArgs, cursorData.SPD)
			mainArgId++
			offset = 0 // Keyset Pagination: No offset needed
		} else if cursorData.SPD != "" && params.SortBy == "spj-asc" {
			mainWhereClause += fmt.Sprintf(" AND CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) > CAST(SUBSTRING($%d FROM 'ID-SPJ-([0-9]+)') AS NUMERIC)", mainArgId)
			mainArgs = append(mainArgs, cursorData.SPD)
			mainArgId++
			offset = 0 // Keyset Pagination: No offset needed
		}

		mainQuery := fmt.Sprintf(`
			SELECT travel_records.spd_number
			FROM travel_records
			%s
			%s
			GROUP BY travel_records.spd_number
		`, joinClause, mainWhereClause)

		// Sorting
		switch params.SortBy {
		case "spj-asc":
			mainQuery += " ORDER BY CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) ASC"
		case "spj-desc":
			mainQuery += " ORDER BY CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) DESC"
		case "date-asc":
			mainQuery += " ORDER BY MAX(travel_records.start_date) ASC, CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) ASC"
		case "date-desc":
			mainQuery += " ORDER BY MAX(travel_records.start_date) DESC, CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) DESC"
		case "cost-asc":
			mainQuery += " ORDER BY SUM(travel_records.total_cost) ASC, CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) ASC"
		case "cost-desc":
			mainQuery += " ORDER BY SUM(travel_records.total_cost) DESC, CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) DESC"
		default:
			mainQuery += " ORDER BY CAST(SUBSTRING(travel_records.spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) DESC"
		}

		mainQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", mainArgId, mainArgId+1)
		mainArgs = append(mainArgs, params.Limit, offset)

		rows, err := r.d.QueryContext(ctx, mainQuery, mainArgs...)
		if err != nil {
			errMain = fmt.Errorf("dynamic GetPaginatedSPDs: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var spd sql.NullString
			if err := rows.Scan(&spd); err != nil {
				errMain = err
				return
			}
			if spd.Valid {
				spdStrings = append(spdStrings, spd.String)
			}
		}
		errMain = rows.Err()
	}()

	// B. Count Total SPDs
	go func() {
		defer wg.Done()
		countSPDsQuery := fmt.Sprintf(`
			SELECT COUNT(DISTINCT travel_records.spd_number)
			FROM travel_records
			%s
			%s
		`, joinClause, whereClause)

		errCountSPDs = r.d.QueryRowContext(ctx, countSPDsQuery, args...).Scan(&totalItems)
	}()

	// C. Count Total Records (Employees)
	go func() {
		defer wg.Done()
		countRecordsQuery := fmt.Sprintf(`
			SELECT COUNT(travel_records.id)
			FROM travel_records
			%s
			%s
		`, joinClause, whereClause)

		errCountRecords = r.d.QueryRowContext(ctx, countRecordsQuery, args...).Scan(&totalRecords)
	}()

	wg.Wait()

	// Check for any errors from parallel queries
	if errMain != nil {
		return nil, errMain
	}
	if errCountSPDs != nil {
		return nil, fmt.Errorf("count total items: %w", errCountSPDs)
	}
	if errCountRecords != nil {
		return nil, fmt.Errorf("count total records: %w", errCountRecords)
	}

	// 3. If no SPDs found, return empty early
	if len(spdStrings) == 0 {
		return &models.PaginatedResponse{
			Data:         []models.TravelRecord{},
			TotalItems:   totalItems,
			TotalRecords: totalRecords,
			NextCursor:   "",
			Limit:        params.Limit,
		}, nil
	}

	// Calculate NextCursor based on offset & last SPD
	nextCursor := ""
	if len(spdStrings) == params.Limit {
		nextCursorData := CursorData{
			SPD:    spdStrings[len(spdStrings)-1],
			Offset: cursorData.Offset + len(spdStrings),
			Sort:   params.SortBy,
		}
		nextCursor = encodeCursor(nextCursorData)
	}

	// 4. Fetch the actual records for those SPDs
	records, err := r.q.GetRecordsBySPDs(ctx, spdStrings)
	if err != nil {
		return nil, fmt.Errorf("GetRecordsBySPDs: %w", err)
	}

	// 5. Bulk Fetch Relationships to prevent N+1 Queries
	var recordIDs []uuid.UUID
	var userIDs []uuid.UUID
	userIDMap := make(map[uuid.UUID]bool)

	for _, rec := range records {
		recordIDs = append(recordIDs, rec.ID)

		if !userIDMap[rec.EmployeeID] {
			userIDs = append(userIDs, rec.EmployeeID)
			userIDMap[rec.EmployeeID] = true
		}
		if !userIDMap[rec.CreatorID] {
			userIDs = append(userIDs, rec.CreatorID)
			userIDMap[rec.CreatorID] = true
		}
	}

	// Fetch dynamic sequence numbers in bulk
	var seqNumbers []db.GetSequenceNumbersRow
	if len(recordIDs) > 0 {
		var errSeq error
		seqNumbers, errSeq = r.q.GetSequenceNumbers(ctx, recordIDs)
		if errSeq != nil {
			fmt.Printf("GetSequenceNumbers err: %v\n", errSeq)
		}
	}
	seqMap := make(map[uuid.UUID]int)
	for _, sq := range seqNumbers {
		seqMap[sq.ID] = int(sq.SequenceNumber)
	}

	// Fetch Users in Bulk
	var users []db.User
	if len(userIDs) > 0 {
		var errU error
		users, errU = r.q.GetUsersByIDs(ctx, userIDs)
		if errU != nil {
			fmt.Printf("GetUsersByIDs err: %v\n", errU)
		}
	}
	userMap := make(map[uuid.UUID]models.User)
	for _, u := range users {
		userMap[u.ID] = mapDBUser(u)
	}

	// Fetch Costs in Bulk
	var costs []db.TravelCost
	if len(recordIDs) > 0 {
		var errC error
		costs, errC = r.q.GetTravelCostsByRecordIDs(ctx, recordIDs)
		if errC != nil {
			fmt.Printf("GetTravelCosts err: %v\n", errC)
		}
	}
	costMap := make(map[uuid.UUID]*models.TravelCost)
	for _, c := range costs {
		mc := mapDBCost(c)

		// SECURITY & PERFORMANCE: DO NOT return huge Base64 PDF/Image files in the bulk list response!
		mc.TransportFile = nil
		mc.HotelFile = nil
		mc.BoardingPassFile = nil
		mc.TicketBackFile = nil
		mc.TicketGoFile = nil
		mc.ReceiptFiles = nil
		mc.AdditionalCosts = nil // Also contains base64 files!
		mc.Details = nil // Also contains base64 files!

		costMap[c.TravelRecordID] = &mc
	}

	// Fetch Reports in Bulk
	var reports []db.TravelReport
	if len(recordIDs) > 0 {
		var errR error
		reports, errR = r.q.GetTravelReportsByRecordIDs(ctx, recordIDs)
		if errR != nil {
			fmt.Printf("GetTravelReports err: %v\n", errR)
		}
	}
	reportMap := make(map[uuid.UUID]*models.TravelReport)
	for _, rep := range reports {
		mr := mapDBReport(rep)
		// SECURITY & PERFORMANCE: DO NOT return huge Base64 PDF files in the paginated table response!
		// This prevents 200MB+ memory spikes and OOM Killer crashes.
		mr.Files = nil
		mr.SppdFile = nil
		mr.SuratTugasFile = nil
		mr.TanggalMerah = nil

		reportMap[rep.TravelRecordID] = &mr
	}

	// Fetch Locations in Bulk
	var locs []db.TravelLocation
	if len(recordIDs) > 0 {
		var errL error
		locs, errL = r.q.GetTravelLocationsByRecordIDs(ctx, recordIDs)
		if errL != nil {
			fmt.Printf("GetTravelLocations err: %v\n", errL)
		}
	}
	locMap := make(map[uuid.UUID][]models.TravelLocation)
	for _, l := range locs {
		locMap[l.TravelRecordID] = append(locMap[l.TravelRecordID], mapDBLocation(l))
	}

	// 6. Map to Domain Models in O(1) time
	var travelRecords []models.TravelRecord

	// Create a map to group records by SPD first
	recordsBySpd := make(map[string][]models.TravelRecord)
	for _, rec := range records {
		spdStr := fromNullString(rec.SpdNumber)
		tr := mapDBRecord(rec)

		if u, ok := userMap[rec.EmployeeID]; ok {
			tr.Employee = u
		}
		if u, ok := userMap[rec.CreatorID]; ok {
			tr.Creator = u
		}
		if c, ok := costMap[rec.ID]; ok {
			tr.Cost = c
		}
		if rep, ok := reportMap[rec.ID]; ok {
			tr.Report = rep
		}
		if l, ok := locMap[rec.ID]; ok {
			tr.Locations = l
		}

		if seq, ok := seqMap[rec.ID]; ok {
			tr.SequenceNumber = seq
		}

		recordsBySpd[spdStr] = append(recordsBySpd[spdStr], tr)
	}

	// Now append them to travelRecords strictly in the order of spdStrings
	for _, spd := range spdStrings {
		travelRecords = append(travelRecords, recordsBySpd[spd]...)
	}

	return &models.PaginatedResponse{
		Data:         travelRecords,
		TotalItems:   totalItems,
		TotalRecords: totalRecords,
		NextCursor:   nextCursor,
		Limit:        params.Limit,
	}, nil
}

func (r *repository) GetFactualSequenceNumber(ctx context.Context, id uuid.UUID, createdAt time.Time) (int, error) {
	query := `
		WITH ordered AS (
			SELECT id, ROW_NUMBER() OVER (
				ORDER BY 
					CAST(SUBSTRING(spd_number FROM 'ID-SPJ-([0-9]+)') AS NUMERIC) ASC, 
					sequence_number ASC,
					created_at ASC
			) as actual_rank 
			FROM travel_records 
			WHERE deleted_at IS NULL AND spd_number IS NOT NULL AND spd_number <> ''
		)
		SELECT actual_rank FROM ordered WHERE id = $1
	`
	var seq int
	err := r.d.QueryRowContext(ctx, query, id).Scan(&seq)
	if err != nil {
		// Fallback in case of parsing errors or missing
		return 0, err
	}
	return seq, nil
}
