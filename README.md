# Perjadin Protokol Kemnaker RI

This repository contains the official Perjadin Protokol system for Kementerian Ketenagakerjaan Republik Indonesia. It is an enterprise-grade web application designed to manage official travel records, expense reporting, and protocol activities.

## Architecture & Tech Stack

The project follows a Monorepo structure separating the frontend and backend, orchestrated via containerization.

| Component | Technology | Description |
| :--- | :--- | :--- |
| **Frontend** | [SvelteKit](https://kit.svelte.dev/) | SSR/CSR hybrid web application using Node.js adapter. |
| **Backend** | [Go (Golang)](https://go.dev/) | RESTful API using [Echo](https://echo.labstack.com/) framework. |
| **Database** | [PostgreSQL](https://www.postgresql.org/) | Primary relational database with keyset pagination indexing. |
| **Cache** | [Redis](https://redis.io/) | Session storage and caching layer. |
| **PDF Engine** | [Gotenberg](https://gotenberg.dev/) | Microservice for converting documents to PDF. |
| **Infra & Deployment** | [Podman](https://podman.io/) / Systemd | Container runtime, orchestrated using user-level systemd services. |
| **CI/CD Security** | [Trivy](https://aquasecurity.github.io/trivy/) | Automated vulnerability scanning for Go modules and npm packages. |

---

## Prerequisites

To contribute to this project, you must have the following tools installed on your local machine:

1.  **[Git](https://git-scm.com/)** - Version control.
2.  **[Go](https://go.dev/dl/)** - Version 1.24+.
3.  **[Node.js](https://nodejs.org/)** - LTS Version (v20+).
4.  **[Podman Desktop](https://podman-desktop.io/)** (or Docker Desktop) - Container runtime.

---

## Quick Start

Follow these steps to set up your local development environment.

### 1. Setup Project
```bash
# Clone the repository
git clone <repository-url>
cd perjadin-kemnaker-ri
```

### 2. Start Infrastructure
Start the necessary infrastructure containers in the background using Compose.
```bash
podman-compose up -d redis db
```

### 3. Initialize Database
Apply the database migrations to your local Postgres instance. Ensure you have `golang-migrate` installed.
```bash
cd backend
migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/perjadin_db?sslmode=disable" up
```

### 4. Run Backend
Open a new terminal. Run the Go backend locally. Ensure your `.env` file is properly configured.
```bash
cd backend
go run cmd/api/main.go
# Server starts at http://localhost:8081
```

### 5. Run Frontend
Open another terminal. Install dependencies and start the SvelteKit dev server.
```bash
cd frontend
npm install
npm run dev
# App starts at http://localhost:3000
```

---

## Development Workflow

### Database Changes
We use a Schema-First approach. Do not modify Go structs manually for DB tables.

1.  **Create Migration:**
    ```bash
    migrate create -ext sql -dir db/migrations -seq <name_of_change>
    ```
2.  **Edit SQL:** Modify the generated `.up.sql` and `.down.sql` files in `backend/db/migrations/`.
3.  **Apply Migration:**
    ```bash
    migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/perjadin_db?sslmode=disable" up
    ```
4.  **Update Queries:** If you changed queries, edit `.sql` files in `backend/db/queries/`.
5.  **Generate Go Code:**
    ```bash
    cd backend
    sqlc generate
    ```

### Running Tests and Linting
```bash
cd backend
go test -v ./...
golangci-lint run
```

### Full Stack Simulation
To test the production build locally (including multi-stage container builds):
```bash
podman-compose up --build
```

---

## Project Structure

```
├── .github/            # CI/CD Workflows (Trivy Security Scans, etc.)
├── backend/            # Go Backend
│   ├── cmd/            # Entry points (api, seeder, etc.)
│   ├── db/             # Migrations and SQL queries
│   ├── internal/       # Private application logic (handlers, services)
│   ├── pkg/            # Public libraries
│   └── sqlc.yaml       # SQLC Configuration
├── frontend/           # SvelteKit Frontend
│   ├── src/lib/        # Shared components and utilities
│   ├── src/routes/     # App pages and API routes
│   └── static/         # Public assets
├── docker-compose.yml  # Container orchestration
└── README.md           # You are here
```

## Project Metrics (Lines of Code)

The codebase consists of approximately **29,800 lines of pure code** (excluding comments, blanks, and external dependencies) distributed across **235 unique files**. 

- **Frontend (Svelte, TypeScript, etc):** ~11,600 LoC
- **Backend (Go, SQL):** ~10,500 LoC 
- **Configuration (JSON, YAML, etc):** ~7,700 LoC

The architecture maintains a very healthy and balanced distribution between the frontend user interface and backend business logic.

---

## Troubleshooting

**Q: Connection refused to Database?**
A: Ensure your local Postgres container/service is running and port 5432 is available. Check your `.env` secrets for `DB_HOST` (should be `localhost` when running native Go, or `host.docker.internal` when running inside Docker/Podman).

**Q: Permission denied on `go install`?**
A: Check your `$GOPATH` and `$PATH`. Ensure `Go/bin` is in your system PATH.

---
Maintained by Application Development Team
