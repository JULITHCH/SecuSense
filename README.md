# SecuSense Training Portal

A web-based training portal with AI-powered course generation, featuring video training, interactive quizzes, and certificate generation.

## Features

- **User Authentication**: JWT-based authentication with refresh tokens
- **Course Management**: Browse, enroll, and complete training courses
- **AI Course Generation**: Automatically generate courses from topics using Ollama LLM
- **Video Training**: Synthesia-powered video generation
- **Interactive Quizzes**: Multiple question types (multiple choice, drag & drop, fill-blank, matching, ordering)
- **Certificate System**: Generate and verify completion certificates

## Tech Stack

### Backend
- Go with Chi router (Clean Architecture)
- PostgreSQL database
- JWT authentication
- Maroto PDF generation

### Frontend
- Angular 18+ with standalone components
- PrimeNG UI components
- Reactive Forms
- Signal-based state management

### External Services
- Ollama (Local LLM for content generation)
- Synthesia (Video generation API)

## Getting Started

### Prerequisites
- Docker and Docker Compose
- Go 1.22+ (for local development)
- Node.js 20+ (for local development)
- PostgreSQL 16+ (for local development)

### Quick Start with Docker

```bash
# Start all services
docker-compose up -d

# The application will be available at:
# - Frontend: http://localhost
# - API: http://localhost:8080
# - Ollama: http://localhost:11434
```

### Local Development

#### Backend

```bash
cd backend

# Install dependencies
go mod download

# Run migrations (requires PostgreSQL)
# Create database first: createdb secusense

# Start the server
go run cmd/api/main.go
```

#### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm start

# Open http://localhost:4200
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - Login (returns JWT)
- `POST /api/v1/auth/refresh` - Refresh token
- `GET /api/v1/auth/me` - Current user profile

### Courses
- `GET /api/v1/courses` - List published courses
- `GET /api/v1/courses/:id` - Course details
- `POST /api/v1/courses/:id/enroll` - Enroll in course

### Enrollments
- `GET /api/v1/enrollments` - User's enrollments
- `PUT /api/v1/enrollments/:id/progress` - Update watch progress
- `POST /api/v1/enrollments/:id/complete-video` - Mark video watched

### Tests
- `GET /api/v1/courses/:courseId/test` - Get test
- `POST /api/v1/tests/:testId/attempts` - Start attempt
- `POST /api/v1/attempts/:attemptId/submit` - Submit test
- `GET /api/v1/attempts/:attemptId/results` - Get results

### Certificates
- `GET /api/v1/certificates` - User's certificates
- `GET /api/v1/certificates/:id/download` - Download PDF
- `GET /api/v1/certificates/verify/:hash` - Public verification

### Admin
- `POST /api/v1/admin/generate/course` - Generate course from topic
- `GET /api/v1/admin/generate/jobs/:id` - Check generation status
- CRUD for courses, tests, questions

## Project Structure

```
SecuSense/
├── backend/
│   ├── cmd/api/main.go              # Entry point
│   ├── config/                       # Configuration
│   ├── internal/
│   │   ├── domain/                   # Entities and interfaces
│   │   ├── usecase/                  # Business logic
│   │   ├── repository/postgres/      # Data access
│   │   └── delivery/http/            # HTTP handlers
│   ├── pkg/                          # Shared packages
│   ├── infrastructure/               # External services
│   └── migrations/                   # SQL migrations
├── frontend/
│   └── src/app/
│       ├── core/                     # Services, guards, interceptors
│       ├── shared/                   # Reusable components
│       └── features/                 # Feature modules
│           ├── auth/
│           ├── dashboard/
│           ├── courses/
│           ├── quiz/
│           ├── certificates/
│           └── admin/
└── docker-compose.yml
```

## Configuration

### Backend Configuration (config.yaml)

```yaml
server:
  port: "8080"
  environment: "development"
  allowOrigins:
    - "http://localhost:4200"

database:
  host: "localhost"
  port: 5433
  user: "secusense"
  password: "secusense"
  dbname: "secusense"

jwt:
  secret: "your-secret-key"
  accessExpiresIn: "15m"
  refreshExpiresIn: "168h"

ollama:
  baseUrl: "http://localhost:11434"
  model: "llama3.2"

synthesia:
  apiKey: "your-api-key"
  avatarId: "anna_costume1_cameraA"
```

### Environment Variables

All configuration can be overridden with environment variables prefixed with `SECUSENSE_`:

- `SECUSENSE_DATABASE_HOST`
- `SECUSENSE_DATABASE_PORT`
- `SECUSENSE_JWT_SECRET`
- `SECUSENSE_OLLAMA_BASEURL`
- `SECUSENSE_SYNTHESIA_APIKEY`

## License

MIT
