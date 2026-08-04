package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func NewPostgresDB(host, port, user, password, name, sslMode string) (*sql.DB, error) {
	// First connect to default 'postgres' db to check and create the target db
	dsnDefault := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		host, user, password, port, sslMode)

	dbDefault, err := sql.Open("postgres", dsnDefault)
	if err != nil {
		log.Printf("Warning: failed to open default postgres database: %v", err)
	} else {
		// Check if database exists
		var checkDb int
		err = dbDefault.QueryRow("SELECT count(*) FROM pg_database WHERE datname = $1", name).Scan(&checkDb)
		if err == nil && checkDb == 0 {
			log.Printf("Database %s does not exist. Creating...", name)
			// We cannot use parameterized queries for CREATE DATABASE
			_, err = dbDefault.Exec(fmt.Sprintf("CREATE DATABASE \"%s\"", name))
			if err != nil {
				log.Printf("Failed to create database: %v", err)
			} else {
				log.Printf("Database %s created successfully.", name)
			}
		}
		dbDefault.Close()
	}

	// Now connect to the target database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, name, port, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open target database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping target database: %w", err)
	}

	// Connection Pooling Configuration
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)

	log.Println("Connected to PostgreSQL database with connection pooling enabled")
	return db, nil
}
