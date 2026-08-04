package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/kemnaker/perjadin-backend/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port, cfg.Database.SSLMode)
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Println("Connected to database. Clearing travel data...")

	tables := []string{
		"travel_reports",
		"travel_costs",
		"travel_locations",
		"travel_records",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s", table)
		res, err := db.Exec(query)
		if err != nil {
			log.Printf("Failed to clear table %s: %v", table, err)
			continue
		}
		rows, err := res.RowsAffected()
		if err != nil {
			log.Printf("Failed to get rows affected for table %s: %v", table, err)
			continue
		}
		log.Printf("Cleared %d rows from %s", rows, table)
	}

	log.Println("Travel data cleared successfully.")
	return nil
}
