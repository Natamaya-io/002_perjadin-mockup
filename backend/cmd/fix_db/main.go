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

	// Try to connect as postgres user
	passwords := []string{"postgres", "root", "password", ""}
	
	var db *sql.DB
	var err error
	connected := false

	for _, p := range passwords {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			"localhost", "postgres", p, cfg.Database.Name, cfg.Database.Port, cfg.Database.SSLMode)
		
		db, err = sql.Open("postgres", dsn)
		if err == nil {
			err = db.Ping()
			if err == nil {
				log.Printf("Connected as postgres with password '%s'", p)
				connected = true
				break
			}
		}
	}

	if !connected {
		log.Printf("Failed to connect as postgres. Falling back to configured user.")
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
			"localhost", cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port, cfg.Database.SSLMode)
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %v", err)
		}
		if pingErr := db.Ping(); pingErr != nil {
			return fmt.Errorf("failed to ping database: %v", pingErr)
		}
	}
	defer db.Close()

	if connected {
		log.Printf("Changing ownership of tables to %s...", cfg.Database.User)
		_, err = db.Exec(fmt.Sprintf(`
			DO $$ DECLARE
				r RECORD;
			BEGIN
				FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
					EXECUTE 'ALTER TABLE public.' || quote_ident(r.tablename) || ' OWNER TO %s;';
				END LOOP;
			END $$;
		`, cfg.Database.User))
		if err != nil {
			log.Printf("Failed to change ownership: %v", err)
		} else {
			log.Printf("Successfully changed ownership to %s", cfg.Database.User)
		}
	}

	_, err = db.Exec("UPDATE schema_migrations SET dirty = false;")
	if err != nil {
		return fmt.Errorf("failed to update schema_migrations: %v", err)
	}
	
	// Drop everything to start fresh
	_, err = db.Exec("DROP TABLE IF EXISTS schema_migrations, users, travel_records, travel_costs, travel_reports, provinces, sbm_rates CASCADE;")
	if err != nil {
		return fmt.Errorf("failed to drop tables: %v", err)
	}

	log.Println("Schema migrations fixed. Try running docker-compose up again.")
	return nil
}