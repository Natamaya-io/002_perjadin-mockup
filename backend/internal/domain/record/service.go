package record

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/models"
	"github.com/kemnaker/perjadin-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type Service interface {
	GenerateSpdNumber(ctx context.Context) (string, error)
	CreateRecord(ctx context.Context, record *models.TravelRecord) error
	// CreateRecordsBulk validates and inserts a group of records atomically.
	// All overlap checks run before any DB write — if any employee has a conflict
	// the entire batch is rejected. If the DB write fails mid-batch, the single
	// outer transaction rolls everything back automatically.
	CreateRecordsBulk(ctx context.Context, records []*models.TravelRecord) error
	CreateRecordDirect(ctx context.Context, record *models.TravelRecord) error // For import: skips overlap check & WA notification
	InvalidateAllCache(ctx context.Context)                                    // Kept for interface compat, but safe
	GetRecords(ctx context.Context, filters map[string]interface{}) ([]models.TravelRecord, error)
	GetRecordByID(ctx context.Context, id uuid.UUID) (*models.TravelRecord, error)
	UpdateRecord(ctx context.Context, record *models.TravelRecord) error
	DeleteRecord(ctx context.Context, id uuid.UUID) error
	DeleteRecordsBySpd(ctx context.Context, spd string) error
	GetDashboardSummary(ctx context.Context, role string, userIDStr string) (*models.DashboardSummary, error)
	GetPaginatedRecords(ctx context.Context, params models.PaginatedParams) (*models.PaginatedResponse, error)
	GetFactualSequenceNumber(ctx context.Context, id uuid.UUID, createdAt time.Time) (int, error)
}

type service struct {
	repo        Repository
	cfg         *config.Config
	redisClient *redis.Client
}

func NewService(repo Repository, cfg *config.Config, rdb *redis.Client) Service {
	return &service{repo: repo, cfg: cfg, redisClient: rdb}
}

// GenerateSpdNumber returns the next SPD number in the format "ID-SPJ-NNN".
// It delegates entirely to the PostgreSQL spd_number_seq sequence via nextval(),
// which is atomic and safe under any level of concurrent load.
// No application-level locking, no SELECT MAX, no regex parsing.
func (s *service) GenerateSpdNumber(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	seq, err := s.repo.NextSpdNumber(ctx)
	if err != nil {
		return "", fmt.Errorf("GenerateSpdNumber: nextval failed: %w", err)
	}

	return fmt.Sprintf("ID-SPJ-%03d", seq), nil
}

// invalidateRecordCaches uses O(1) direct key deletion instead of the nuclear
// 'KEYS records:*' which wipes all unrelated individual record caches and blocks Redis.
func (s *service) invalidateRecordCaches(ctx context.Context, ids ...uuid.UUID) {
	updatedIDs, _ := s.repo.RecalculateSequenceNumbers(ctx)
	ids = append(ids, updatedIDs...)

	if s.redisClient == nil {
		return
	}

	keys := []string{
		"records:all",
		"records:status:Draft",
		"records:status:Pending",
		"records:status:Approved",
		"records:status:Rejected",
		"records:status:Revised",
		"dashboard:summary",
		"dashboard:summary:global",
	}

	for _, id := range ids {
		if id != uuid.Nil {
			keys = append(keys, fmt.Sprintf("records:id:%s", id.String()))
		}
	}

	// Always wipe SPD-related and paginated caches because any CRUD operation alters group content or lists
	iterSpd := s.redisClient.Scan(ctx, 0, "records:spd:*", 0).Iterator()
	for iterSpd.Next(ctx) {
		keys = append(keys, iterSpd.Val())
	}

	iterStatusSpd := s.redisClient.Scan(ctx, 0, "records:status:*:spd:*", 0).Iterator()
	for iterStatusSpd.Next(ctx) {
		keys = append(keys, iterStatusSpd.Val())
	}

	iterPaginated := s.redisClient.Scan(ctx, 0, "records:paginated:*", 0).Iterator()
	for iterPaginated.Next(ctx) {
		keys = append(keys, iterPaginated.Val())
	}

	if len(ids) == 0 {
		// If no specific IDs are passed, it implies a potentially larger wipe
		// We should clear ALL individual record caches as well using Scan.
		iter := s.redisClient.Scan(ctx, 0, "records:id:*", 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
	}

	iterDashboard := s.redisClient.Scan(ctx, 0, "dashboard:summary:user:*", 0).Iterator()
	for iterDashboard.Next(ctx) {
		keys = append(keys, iterDashboard.Val())
	}

	if len(keys) > 0 {
		// Remove duplicates before deletion for efficiency
		uniqueKeys := make(map[string]bool)
		finalKeys := make([]string, 0, len(keys))
		for _, k := range keys {
			if !uniqueKeys[k] {
				uniqueKeys[k] = true
				finalKeys = append(finalKeys, k)
			}
		}
		s.redisClient.Del(ctx, finalKeys...)
	}
}

func (s *service) InvalidateAllCache(ctx context.Context) {
	s.invalidateRecordCaches(ctx)
}

func (s *service) CreateRecord(ctx context.Context, record *models.TravelRecord) error {
	if len(record.Locations) == 0 {
		return errors.New("at least one location is required")
	}

	// Calculate overall start and end dates if not provided or to ensure accuracy
	var minStart, maxEnd time.Time
	for i, loc := range record.Locations {
		if loc.EndDate.Before(loc.StartDate) {
			return fmt.Errorf("location %s: end date cannot be before start date", loc.Location)
		}
		if i == 0 || loc.StartDate.Before(minStart) {
			minStart = loc.StartDate
		}
		if i == 0 || loc.EndDate.After(maxEnd) {
			maxEnd = loc.EndDate
		}
	}
	record.StartDate = minStart
	record.EndDate = maxEnd

	overlapping, err := s.repo.GetOverlappingRecords(ctx, record.EmployeeID, record.StartDate, record.EndDate)
	if err != nil {
		return err
	}
	if len(overlapping) > 0 {
		return errors.New("employee is already assigned to a trip during these dates")
	}

	record.Status = "Draft"
	record.ReportStatus = "Pending"

	// VALIDATION: Prevent duplicate SPD Number assignment manually.
	if record.SPDNumber != "" {
		existingRecords, errEx := s.repo.GetTravelRecords(ctx, map[string]interface{}{"spd": record.SPDNumber})
		if errEx == nil && len(existingRecords) > 0 {
			return fmt.Errorf("nomor SPJ %s sudah digunakan oleh perjalanan dinas lain", record.SPDNumber)
		}
	}

	if record.SPDNumber == "" {
		spd, errGen := s.GenerateSpdNumber(ctx)
		if errGen != nil {
			return fmt.Errorf("failed to generate SPD number: %w", errGen)
		}
		record.SPDNumber = spd
	}

	err = s.repo.CreateTravelRecord(ctx, record)
	if err == nil {
		s.invalidateRecordCaches(ctx, record.ID)

		// Send WhatsApp Notification
		go func() {
			// Using a background context here since the parent request context might be cancelled after response
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			user, errUsr := s.repo.GetUserByID(bgCtx, record.EmployeeID)
			if errUsr == nil && user != nil && user.NomorHP != "" {
				msg := fmt.Sprintf("*PEMBERITAHUAN PERJALANAN DINAS*\n\nHalo %s,\nAnda telah ditugaskan untuk perjalanan dinas baru.\n\n*Detail Penugasan:*\nNo. SPD: %s\nTujuan: %s, %s\nTanggal: %s s/d %s\nKeperluan: %s\n\nSilakan cek aplikasi Perjadin untuk detail selengkapnya dan mengunduh Surat Tugas.",
					user.Name,
					record.SPDNumber,
					record.Location,
					record.Province,
					record.StartDate.Format("02 Jan 2006"),
					record.EndDate.Format("02 Jan 2006"),
					record.Purpose,
				)
				
				if waErr := utils.SendWhatsAppMessage(s.cfg, user.NomorHP, msg); waErr != nil {
					fmt.Printf("Failed to send WhatsApp notification to %s: %v\n", user.NomorHP, waErr)
				} else {
					fmt.Printf("WhatsApp notification sent to %s for SPD %s\n", user.NomorHP, record.SPDNumber)
				}
			}
		}()
	}
	return err
}

// CreateRecordsBulk validates every record in the batch (overlap check, date
// sanity, location presence) BEFORE touching the database. Once all checks
// pass it delegates to the repository for a single atomic transaction.
// WhatsApp notifications are sent concurrently after a successful commit.
func (s *service) CreateRecordsBulk(ctx context.Context, records []*models.TravelRecord) error {
	if len(records) == 0 {
		return errors.New("CreateRecordsBulk: empty batch")
	}

	// --- Phase 1: validate all records before any DB write ---
	for _, r := range records {
		if len(r.Locations) == 0 {
			return fmt.Errorf("employee %s: at least one location is required", r.EmployeeID)
		}
		var minStart, maxEnd time.Time
		for i, loc := range r.Locations {
			if loc.EndDate.Before(loc.StartDate) {
				return fmt.Errorf("employee %s, location %s: end date cannot be before start date", r.EmployeeID, loc.Location)
			}
			if i == 0 || loc.StartDate.Before(minStart) {
				minStart = loc.StartDate
			}
			if i == 0 || loc.EndDate.After(maxEnd) {
				maxEnd = loc.EndDate
			}
		}
		r.StartDate = minStart
		r.EndDate = maxEnd

		overlapping, err := s.repo.GetOverlappingRecords(ctx, r.EmployeeID, r.StartDate, r.EndDate)
		if err != nil {
			return fmt.Errorf("overlap check for employee %s: %w", r.EmployeeID, err)
		}
		if len(overlapping) > 0 {
			return fmt.Errorf("employee %s is already assigned to a trip during these dates", r.EmployeeID)
		}

		r.Status = "Draft"
		r.ReportStatus = "Pending"
	}

	// Generate SPD number safely AFTER validation passes
	if records[0].SPDNumber == "" {
		spd, err := s.GenerateSpdNumber(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate SPD number: %w", err)
		}
		for _, r := range records {
			r.SPDNumber = spd
		}
	}

	// --- Phase 2: single atomic DB write ---
	if err := s.repo.CreateTravelRecordsBulk(ctx, records); err != nil {
		return err
	}

	ids := make([]uuid.UUID, 0, len(records))
	for _, r := range records {
		ids = append(ids, r.ID)
	}
	s.invalidateRecordCaches(ctx, ids...)

	// --- Phase 3: non-blocking WA notifications ---
	for _, r := range records {
		recordCopy := r // capture loop variable
		go func() {
			user, err := s.repo.GetUserByID(context.Background(), recordCopy.EmployeeID)
			if err != nil || user == nil || user.NomorHP == "" {
				return
			}
			msg := fmt.Sprintf(
				"*PEMBERITAHUAN PERJALANAN DINAS*\n\nHalo %s,\nAnda telah ditugaskan untuk perjalanan dinas baru.\n\n*Detail Penugasan:*\nNo. SPD: %s\nTujuan: %s, %s\nTanggal: %s s/d %s\nKeperluan: %s\n\nSilakan cek aplikasi Perjadin untuk detail selengkapnya dan mengunduh Surat Tugas.",
				user.Name, recordCopy.SPDNumber,
				recordCopy.Location, recordCopy.Province,
				recordCopy.StartDate.Format("02 Jan 2006"),
				recordCopy.EndDate.Format("02 Jan 2006"),
				recordCopy.Purpose,
			)
			if err := utils.SendWhatsAppMessage(s.cfg, user.NomorHP, msg); err != nil {
				fmt.Printf("WA notify failed for %s (SPD %s): %v\n", user.NomorHP, recordCopy.SPDNumber, err)
			}
		}()
	}

	return nil
}

// CreateRecordDirect creates a record without overlap checking or WhatsApp notification.
// Used for bulk import operations.
func (s *service) CreateRecordDirect(ctx context.Context, record *models.TravelRecord) error {
	if len(record.Locations) == 0 {
		return errors.New("at least one location is required")
	}

	// Calculate overall start and end dates
	var minStart, maxEnd time.Time
	for i, loc := range record.Locations {
		if i == 0 || loc.StartDate.Before(minStart) {
			minStart = loc.StartDate
		}
		if i == 0 || loc.EndDate.After(maxEnd) {
			maxEnd = loc.EndDate
		}
	}
	record.StartDate = minStart
	record.EndDate = maxEnd

	if record.Status == "" {
		record.Status = "Draft"
	}
	if record.ReportStatus == "" {
		record.ReportStatus = "Pending"
	}

	err := s.repo.CreateTravelRecord(ctx, record)
	if err == nil {
		s.InvalidateAllCache(ctx)
	}
	return err
}

func (s *service) GetRecords(ctx context.Context, filters map[string]interface{}) ([]models.TravelRecord, error) {
	var statusFilter string
	var spdFilter string
	if filters != nil {
		if status, ok := filters["status"]; ok && status != "" {
			statusFilter, _ = status.(string) //nolint:errcheck
		}
		if spd, ok := filters["spd"]; ok && spd != "" {
			spdFilter, _ = spd.(string) //nolint:errcheck
		}
	}

	cacheKey := "records:all"
	switch {
	case statusFilter != "" && spdFilter != "":
		cacheKey = fmt.Sprintf("records:status:%s:spd:%s", statusFilter, spdFilter)
	case statusFilter != "":
		cacheKey = fmt.Sprintf("records:status:%s", statusFilter)
	case spdFilter != "":
		cacheKey = fmt.Sprintf("records:spd:%s", spdFilter)
	}

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var records []models.TravelRecord
			if err := json.Unmarshal([]byte(cached), &records); err == nil {
				return records, nil
			}
		}
	}

	records, err := s.repo.GetTravelRecords(ctx, filters)
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		cacheBytes, err := json.Marshal(records)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheBytes, 15*time.Minute)
		}
	}

	return records, nil
}

func (s *service) GetRecordByID(ctx context.Context, id uuid.UUID) (*models.TravelRecord, error) {
	cacheKey := fmt.Sprintf("records:id:%s", id.String())

	if s.redisClient != nil {
		cached, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var record models.TravelRecord
			if err := json.Unmarshal([]byte(cached), &record); err == nil {
				return &record, nil
			}
		}
	}

	record, err := s.repo.GetTravelRecordByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		cacheBytes, err := json.Marshal(record)
		if err == nil {
			s.redisClient.Set(ctx, cacheKey, cacheBytes, 15*time.Minute)
		}
	}

	return record, nil
}

func (s *service) UpdateRecord(ctx context.Context, record *models.TravelRecord) error {
	if len(record.Locations) > 0 {
		// Calculate overall start and end dates
		var minStart, maxEnd time.Time
		for i, loc := range record.Locations {
			if loc.EndDate.Before(loc.StartDate) {
				return fmt.Errorf("location %s: end date cannot be before start date", loc.Location)
			}
			if i == 0 || loc.StartDate.Before(minStart) {
				minStart = loc.StartDate
			}
			if i == 0 || loc.EndDate.After(maxEnd) {
				maxEnd = loc.EndDate
			}
		}
		record.StartDate = minStart
		record.EndDate = maxEnd
	}

	// VALIDATION: Prevent duplicate SPD Number assignment manually.
	// If the user tries to assign an SPD number that already exists in the database
	// and doesn't belong to the current travel record or its group.
	if record.SPDNumber != "" {
		existingRecords, err := s.repo.GetTravelRecords(ctx, map[string]interface{}{"spd": record.SPDNumber})
		if err == nil && len(existingRecords) > 0 {
			// Check if we are trying to steal an SPD number from another trip
			for _, er := range existingRecords {
				// The record is trying to change its SPD to an existing one, but it wasn't part of that group originally
				if er.ID != record.ID {
					// We must fetch the original state of this record to see if its spd actually changed
					original, errOrig := s.repo.GetTravelRecordByID(ctx, record.ID)
					if errOrig == nil && original != nil && original.SPDNumber != record.SPDNumber {
						// Allow if the existing record is actually a sibling from the same trip 
						// (happens when the frontend loops and updates a group sequentially)
						if er.Purpose == record.Purpose && er.Location == record.Location {
							continue
						}
						return fmt.Errorf("nomor SPD %s sudah digunakan oleh perjalanan dinas lain", record.SPDNumber)
					}
				}
			}
		}
	}

	err := s.repo.UpdateTravelRecord(ctx, record)
	if err == nil {
		s.invalidateRecordCaches(ctx, record.ID)

		// Resynchronize the sequence safely to accommodate manual SPD number overrides
		go func(bgCtx context.Context) {
			if syncErr := s.repo.SyncSpdSequence(bgCtx); syncErr != nil {
				fmt.Printf("Warning: failed to sync SPD sequence after UpdateRecord: %v\n", syncErr)
			}
			// Invalidate cache again after sequence is updated so the frontend gets the new sorted order
			s.invalidateRecordCaches(bgCtx)
		}(context.Background())

		// NEW: If a report was updated, sync it to all other records in the same SPD group
		// using an O(1) SQL bulk update. Completely avoids O(N) memory fetching.
		if record.Report != nil && record.SPDNumber != "" {
			go func(ctx context.Context, spd string, rec *models.TravelRecord) {
				if syncErr := s.repo.SyncReportBySpd(ctx, spd, rec); syncErr != nil {
					fmt.Printf("Error syncing reports for SPD %s: %v\n", spd, syncErr)
				}
				// After DB sync, clear list caches again just in case.
				// For the other individual IDs, they will organically expire or be cleared when accessed.
				s.invalidateRecordCaches(ctx)
			}(context.Background(), record.SPDNumber, record)
		}
	}
	return err
}

func (s *service) DeleteRecord(ctx context.Context, id uuid.UUID) error {
	err := s.repo.DeleteTravelRecord(ctx, id)
	if err == nil {
		go func() {
			s.invalidateRecordCaches(context.Background(), id)
			if syncErr := s.repo.SyncSpdSequence(context.Background()); syncErr != nil {
				fmt.Printf("Warning: failed to sync SPD sequence after DeleteRecord: %v\n", syncErr)
			}
			s.invalidateRecordCaches(context.Background(), id)
		}()
	}
	return err
}

func (s *service) DeleteRecordsBySpd(ctx context.Context, spd string) error {
	err := s.repo.DeleteTravelRecordsBySpd(ctx, spd)
	if err == nil {
		go func() {
			s.invalidateRecordCaches(context.Background())
			if syncErr := s.repo.SyncSpdSequence(context.Background()); syncErr != nil {
				fmt.Printf("Warning: failed to sync SPD sequence after DeleteRecordsBySpd: %v\n", syncErr)
			}
			s.invalidateRecordCaches(context.Background())
		}()
	}
	return err
}

func (s *service) GetDashboardSummary(ctx context.Context, role string, userIDStr string) (*models.DashboardSummary, error) {
	cacheKey := "dashboard:summary:global"
	if role == "protokol" && userIDStr != "" {
		cacheKey = fmt.Sprintf("dashboard:summary:user:%s", userIDStr)
	}

	if s.redisClient != nil {
		val, err := s.redisClient.Get(ctx, cacheKey).Result()
		if err == nil && val != "" {
			var summary models.DashboardSummary
			if err := json.Unmarshal([]byte(val), &summary); err == nil {
				// Fire background revalidation (Stale-While-Revalidate)
				go func(r string, u string, key string) {
					bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
					defer cancel()
					fresh, err := s.repo.GetDashboardSummary(bgCtx, r, u)
					if err == nil {
						if data, err := json.Marshal(fresh); err == nil {
							s.redisClient.Set(bgCtx, key, data, 15*time.Minute)
						}
					}
				}(role, userIDStr, cacheKey)
				return &summary, nil
			}
		}
	}

	// Cache miss or error
	summary, err := s.repo.GetDashboardSummary(ctx, role, userIDStr)
	if err != nil {
		return nil, err
	}

	if s.redisClient != nil {
		if data, err := json.Marshal(summary); err == nil {
			s.redisClient.Set(ctx, cacheKey, data, 15*time.Minute)
		}
	}

	return summary, nil
}

func (s *service) GetPaginatedRecords(ctx context.Context, params models.PaginatedParams) (*models.PaginatedResponse, error) {
	if s.redisClient == nil {
		return s.repo.GetPaginatedRecords(ctx, params)
	}

	// Create a deterministic hash of the parameters for the cache key
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return s.repo.GetPaginatedRecords(ctx, params)
	}
	hash := sha256.Sum256(paramsBytes)
	cacheKey := fmt.Sprintf("records:paginated:%s", hex.EncodeToString(hash[:]))

	cached, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var response models.PaginatedResponse
		if umErr := json.Unmarshal([]byte(cached), &response); umErr == nil {
			return &response, nil
		}
	}

	response, err := s.repo.GetPaginatedRecords(ctx, params)
	if err != nil {
		return nil, err
	}

	if cacheBytes, err := json.Marshal(response); err == nil {
		s.redisClient.Set(ctx, cacheKey, cacheBytes, 15*time.Minute)
	}

	return response, nil
}

func (s *service) GetFactualSequenceNumber(ctx context.Context, id uuid.UUID, createdAt time.Time) (int, error) {
	return s.repo.GetFactualSequenceNumber(ctx, id, createdAt)
}
