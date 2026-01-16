package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobType string
type JobStatus string

const (
	JobTypeContent JobType = "content_generation"
	JobTypeVideo   JobType = "video_generation"
	JobTypeTest    JobType = "test_generation"

	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type AIGenerationJob struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	CourseID   *uuid.UUID      `db:"course_id" json:"courseId,omitempty"`
	JobType    JobType         `db:"job_type" json:"jobType"`
	Status     JobStatus       `db:"status" json:"status"`
	InputData  json.RawMessage `db:"input_data" json:"inputData"`
	OutputData json.RawMessage `db:"output_data" json:"outputData,omitempty"`
	Error      *string         `db:"error" json:"error,omitempty"`
	CreatedAt  time.Time       `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time       `db:"updated_at" json:"updatedAt"`
	CompletedAt *time.Time     `db:"completed_at" json:"completedAt,omitempty"`
}

type GenerateCourseRequest struct {
	Topic            string `json:"topic" validate:"required,min=5,max=500"`
	TargetAudience   string `json:"targetAudience,omitempty"`
	DifficultyLevel  string `json:"difficultyLevel,omitempty" validate:"omitempty,oneof=beginner intermediate advanced"`
	VideoDurationMin int    `json:"videoDurationMin,omitempty" validate:"omitempty,min=1,max=60"`
}

type GeneratedCourseContent struct {
	Title              string                  `json:"title"`
	Description        string                  `json:"description"`
	LearningObjectives []string                `json:"learningObjectives"`
	Outline            []CourseOutlineChapter  `json:"outline"`
	VideoScript        string                  `json:"videoScript"`
	Questions          []GeneratedQuestion     `json:"questions"`
}

type GeneratedQuestion struct {
	QuestionType QuestionType    `json:"questionType"`
	QuestionText string          `json:"questionText"`
	QuestionData json.RawMessage `json:"questionData"`
	Points       int             `json:"points"`
}

type SynthesiaWebhookPayload struct {
	VideoID string `json:"id"`
	Status  string `json:"status"`
	URL     string `json:"download,omitempty"`
}

type AIGenerationJobRepository interface {
	Create(job *AIGenerationJob) error
	GetByID(id uuid.UUID) (*AIGenerationJob, error)
	GetByCourseID(courseID uuid.UUID) ([]*AIGenerationJob, error)
	Update(job *AIGenerationJob) error
	ListPending(limit int) ([]*AIGenerationJob, error)
}
