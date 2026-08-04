package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/pkg/cache"
	"github.com/kemnaker/perjadin-backend/pkg/database"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	cfg := config.LoadConfig()

	db, err := database.NewPostgresDB(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Generating 2000 travel records...")

	// Get a random user for creator_id and employee_id
	var userID uuid.UUID
	err = db.QueryRow("SELECT id FROM users LIMIT 1").Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get a user: %v. Make sure you have run the normal seeder first", err)
	}

	// Get all provinces
	rows, err := db.Query("SELECT name FROM provinces")
	if err != nil {
		return fmt.Errorf("failed to get provinces: %v", err)
	}
	var provinces []string
	for rows.Next() {
		var name string
		if scanErr := rows.Scan(&name); scanErr != nil {
			rows.Close()
			return fmt.Errorf("failed to scan province: %v", scanErr)
		}
		provinces = append(provinces, name)
	}
	rows.Close()

	if len(provinces) == 0 {
		return fmt.Errorf("no provinces found. Run normal seeder first")
	}

	purposes := []string{
		"Koordinasi Pelaksanaan Program",
		"Monitoring dan Evaluasi Kegiatan",
		"Pendampingan Kunjungan Kerja",
		"Rapat Koordinasi Teknis",
		"Sosialisasi Peraturan Baru",
		"Bimbingan Teknis Aplikasi",
	}

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for i := 1; i <= 2000; i++ {
		recordID := uuid.New()
		spdNumber := fmt.Sprintf("ID-SPJ-DUMMY-%04d", i)
		purpose := purposes[rand.Intn(len(purposes))]
		province := provinces[rand.Intn(len(provinces))]
		
		startDate := time.Now().AddDate(0, 0, rand.Intn(30)-15)
		endDate := startDate.AddDate(0, 0, rand.Intn(3)+1)

		// Insert Travel Record
		_, err = tx.ExecContext(ctx, `
			INSERT INTO travel_records (
				id, spd_number, employee_id, creator_id, start_date, end_date, 
				location, province, type, purpose, status, report_status, payment_status,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())`,
			recordID, spdNumber, userID, userID, startDate, endDate,
			"Pusat Kota "+province, province, "Luar_Kota", purpose, "Approved", "Pending", "Unpaid",
		)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("rollback failed: %v", rbErr)
			}
			return fmt.Errorf("failed to insert record %d: %v", i, err)
		}

		// Insert 1 Location for each record
		_, err = tx.ExecContext(ctx, `
			INSERT INTO travel_locations (
				id, travel_record_id, location, province, start_date, end_date, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
			uuid.New(), recordID, "Pusat Kota "+province, province, startDate, endDate,
		)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("rollback failed: %v", rbErr)
			}
			return fmt.Errorf("failed to insert location for record %d: %v", i, err)
		}

		if i%100 == 0 {
			log.Printf("Inserted %d records...", i)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Println("Successfully generated 2000 dummy travel records.")

	// Flush Redis Cache
	rdb, err := cache.NewRedisClient(cfg)
	if err == nil && rdb != nil {
		defer rdb.Close()
		rdb.FlushDB(ctx)
		log.Println("Redis cache cleared after massive seeding.")
	} else {
		log.Printf("Could not connect to Redis to flush cache: %v", err)
	}

	return nil
}
