package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/kemnaker/perjadin-backend/internal/config"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg := config.LoadConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.User, cfg.Database.Password, cfg.Database.Name, cfg.Database.Port, cfg.Database.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if pingErr := db.Ping(); pingErr != nil {
		log.Printf("Failed to ping database: %v", pingErr)
		return
	}

	email := "ramvito@kemnaker.go.id"
	password := "vito123"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		return
	}

	id := uuid.New()
	_, err = db.Exec(`
		INSERT INTO users (id, email, password, name, role, demo_password)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, email, string(hashedPassword), "Vito", "super_admin", password)

	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}

	log.Println("User ramvito@kemnaker.go.id created successfully.")
}
