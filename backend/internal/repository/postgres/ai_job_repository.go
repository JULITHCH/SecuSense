package postgres

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type AIGenerationJobRepository struct {
	db *sqlx.DB
}

func NewAIGenerationJobRepository(db *sqlx.DB) *AIGenerationJobRepository {
	return &AIGenerationJobRepository{db: db}
}

func (r *AIGenerationJobRepository) Create(job *domain.AIGenerationJob) error {
	query := `
		INSERT INTO ai_generation_jobs (id, course_id, job_type, status, input_data, output_data, error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING created_at, updated_at`

	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}

	// Handle nil OutputData - use nil for SQL NULL instead of empty json.RawMessage
	var outputData interface{}
	if len(job.OutputData) > 0 {
		outputData = job.OutputData
	}

	return r.db.QueryRow(
		query,
		job.ID, job.CourseID, job.JobType, job.Status, job.InputData, outputData, job.Error,
	).Scan(&job.CreatedAt, &job.UpdatedAt)
}

func (r *AIGenerationJobRepository) GetByID(id uuid.UUID) (*domain.AIGenerationJob, error) {
	var job domain.AIGenerationJob
	query := `SELECT id, course_id, job_type, status, input_data,
			  COALESCE(output_data, '{}') as output_data,
			  error, created_at, updated_at, completed_at
			  FROM ai_generation_jobs WHERE id = $1`

	err := r.db.Get(&job, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *AIGenerationJobRepository) GetByCourseID(courseID uuid.UUID) ([]*domain.AIGenerationJob, error) {
	var jobs []*domain.AIGenerationJob
	query := `SELECT id, course_id, job_type, status, input_data, output_data, error, created_at, updated_at, completed_at
			  FROM ai_generation_jobs WHERE course_id = $1 ORDER BY created_at DESC`

	err := r.db.Select(&jobs, query, courseID)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *AIGenerationJobRepository) Update(job *domain.AIGenerationJob) error {
	query := `
		UPDATE ai_generation_jobs
		SET course_id = $1, status = $2, output_data = $3, error = $4, updated_at = NOW(), completed_at = $5
		WHERE id = $6
		RETURNING updated_at`

	return r.db.QueryRow(
		query,
		job.CourseID, job.Status, job.OutputData, job.Error, job.CompletedAt, job.ID,
	).Scan(&job.UpdatedAt)
}

func (r *AIGenerationJobRepository) ListPending(limit int) ([]*domain.AIGenerationJob, error) {
	var jobs []*domain.AIGenerationJob
	query := `SELECT id, course_id, job_type, status, input_data, output_data, error, created_at, updated_at, completed_at
			  FROM ai_generation_jobs WHERE status = 'pending' ORDER BY created_at ASC LIMIT $1`

	err := r.db.Select(&jobs, query, limit)
	if err != nil {
		return nil, err
	}
	return jobs, nil
}
