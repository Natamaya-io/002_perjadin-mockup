# Perjadin Backend API

This is a production-grade Go backend for the Perjadin application, built with **Echo**, **GORM**, and **PostgreSQL**.

## Tech Stack

- **Language:** Go 1.22+
- **Framework:** Echo v4
- **Database:** PostgreSQL
- **ORM:** GORM
- **Auth:** JWT (JSON Web Tokens)
- **Config:** Viper
- **Logging:** Zap

## Setup

1.  **Prerequisites:**
    - Go 1.22 or later installed.
    - PostgreSQL database running.

2.  **Configuration:**
    - Use Doppler for environment variable management. Ensure Doppler CLI is installed and configured for the project.

3.  **Run the Application:**

    ```bash
    cd backend
    doppler run -- go run cmd/api/main.go
    ```

    The server will start on port `8080` (or as configured in Doppler).

## API Endpoints

### Auth
- `POST /api/v1/auth/register` - Register a new user.
- `POST /api/v1/auth/login` - Login and get JWT token.

### Travel Records (Protected)
- `GET /api/v1/records` - Get all travel records.
- `POST /api/v1/records` - Create a new travel record.
- `GET /api/v1/records/:id` - Get a specific travel record.

## Project Structure

- `cmd/api/main.go`: Application entry point.
- `internal/config`: Configuration loading.
- `internal/handlers`: HTTP request handlers.
- `internal/middleware`: Custom middleware (Auth, Logging).
- `internal/models`: Database models (Structs).
- `internal/repository`: Data access layer.
- `internal/services`: Business logic.
- `pkg/database`: Database connection logic.
