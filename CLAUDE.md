# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SecuSense is a training portal with AI-powered course generation, video training (Synthesia), interactive quizzes, and certificate generation. It has a Go backend and Angular frontend.

## Common Commands

### Backend (Go)
```bash
cd backend
go mod download          # Install dependencies
go run cmd/api/main.go   # Run the API server (port 8080)
go build -o bin/api cmd/api/main.go  # Build binary
```

### Frontend (Angular 18)
```bash
cd frontend
npm install              # Install dependencies
npm start                # Dev server on http://localhost:4200
npm run build            # Production build
npm test                 # Run Karma tests
```

### Docker
```bash
./start.sh               # Interactive launcher (auto-detects native Ollama)
docker-compose up -d     # Start without Docker Ollama (uses native)
docker-compose --profile with-ollama up -d  # Start with Docker Ollama
```

### Database
PostgreSQL runs on port 5433 (not 5432). Migrations are in `backend/migrations/` and auto-run via Docker init.

## Architecture

### Backend - Clean Architecture
The backend follows Clean Architecture with dependency injection wired in `cmd/api/main.go`:

- **domain/** - Entities and repository interfaces (e.g., `Course`, `CourseRepository`)
- **usecase/** - Business logic, one package per domain (auth, course, enrollment, test, certificate, ai)
- **repository/postgres/** - Database implementations of domain interfaces
- **delivery/http/** - Chi router, handlers, middleware
- **infrastructure/** - External service clients (database, ollama, synthesia)
- **pkg/** - Shared utilities (jwt, pdf generation)

Request flow: Handler → UseCase → Repository → Database

### Frontend - Feature-based Angular
Angular 18 with standalone components and signal-based state:

- **core/** - Services, guards, interceptors (auth.service.ts manages JWT)
- **features/** - Lazy-loaded feature modules (auth, dashboard, courses, quiz, certificates, admin)
- **shared/** - Reusable components

Routes are defined in `app.routes.ts` with lazy loading via `loadChildren`.

### Key Integrations
- **Ollama**: Local LLM for course content generation (`infrastructure/ollama/`)
- **Synthesia**: Video generation API with webhook support (`infrastructure/synthesia/`)
- **Maroto**: PDF certificate generation (`pkg/pdf/`)

## Configuration

Backend config via `backend/config.yaml` or environment variables with `SECUSENSE_` prefix:
- `SECUSENSE_DATABASE_HOST`, `SECUSENSE_DATABASE_PORT`
- `SECUSENSE_JWT_SECRET`
- `SECUSENSE_OLLAMA_BASEURL`
- `SECUSENSE_SYNTHESIA_APIKEY`
