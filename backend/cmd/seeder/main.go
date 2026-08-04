package main

import (
	"database/sql"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kemnaker/perjadin-backend/internal/config"
	"github.com/kemnaker/perjadin-backend/internal/seeder"
	"github.com/kemnaker/perjadin-backend/pkg/cache"
	"github.com/kemnaker/perjadin-backend/pkg/database"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.NewPostgresDB(cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database.")

	rdb, err := cache.NewRedisClient(cfg)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v. Cache will not be cleared.", err)
	} else {
		defer rdb.Close()
	}

	// Ensure Migrations are up to date before seeding
	log.Println("Running Migrations...")
	runMigrations(db)

	// Call the shared seeder logic
	seeder.Seed(db, rdb)
}

func runMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create postgres driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("Migration init failed: %v", err)
	}
	
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully.")
}
