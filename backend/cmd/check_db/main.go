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
		"localhost", cfg.Database.User, cfg.Database.Password, cfg.Database.Name, "5433", cfg.Database.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';")
	if err != nil {
		return fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	fmt.Println("Tables in public schema:")
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		fmt.Println("-", tableName)
	}
	return nil
}
