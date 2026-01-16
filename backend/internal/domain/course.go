package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type VideoStatus string

const (
	VideoStatusPending    VideoStatus = "pending"
	VideoStatusProcessing VideoStatus = "processing"
	VideoStatusCompleted  VideoStatus = "completed"
	VideoStatusFailed     VideoStatus = "failed"
)

type Course struct {
	ID               uuid.UUID    `db:"id" json:"id"`
	Title            string       `db:"title" json:"title"`
	Description      string       `db:"description" json:"description"`
	VideoURL         *string      `db:"video_url" json:"videoUrl,omitempty"`
	SynthesiaVideoID *string      `db:"synthesia_video_id" json:"synthesiaVideoId,omitempty"`
	VideoStatus      *VideoStatus `db:"video_status" json:"videoStatus,omitempty"`
	VideoError       *string      `db:"video_error" json:"videoError,omitempty"`
	ThumbnailURL     *string      `db:"thumbnail_url" json:"thumbnailUrl,omitempty"`
	PassPercentage   int          `db:"pass_percentage" json:"passPercentage"`
	IsPublished      bool         `db:"is_published" json:"isPublished"`
	CreatedAt        time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time    `db:"updated_at" json:"updatedAt"`
}

type CourseOutlineChapter struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Topics      []string `json:"topics"`
}

type CourseContent struct {
	ID                 uuid.UUID              `db:"id" json:"id"`
	CourseID           uuid.UUID              `db:"course_id" json:"courseId"`
	Outline            json.RawMessage        `db:"outline" json:"outline"`
	VideoScript        string                 `db:"video_script" json:"videoScript"`
	LearningObjectives []string               `db:"-" json:"learningObjectives"`
	GenerationPrompt   string                 `db:"generation_prompt" json:"generationPrompt"`
	CreatedAt          time.Time              `db:"created_at" json:"createdAt"`
}

type CreateCourseRequest struct {
	Title          string `json:"title" validate:"required,min=1,max=255"`
	Description    string `json:"description" validate:"required,min=1"`
	PassPercentage int    `json:"passPercentage" validate:"required,min=0,max=100"`
}

type UpdateCourseRequest struct {
	Title          *string `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Description    *string `json:"description,omitempty"`
	VideoURL       *string `json:"videoUrl,omitempty"`
	ThumbnailURL   *string `json:"thumbnailUrl,omitempty"`
	PassPercentage *int    `json:"passPercentage,omitempty" validate:"omitempty,min=0,max=100"`
	IsPublished    *bool   `json:"isPublished,omitempty"`
}

type CourseRepository interface {
	Create(course *Course) error
	GetByID(id uuid.UUID) (*Course, error)
	Update(course *Course) error
	Delete(id uuid.UUID) error
	List(limit, offset int, publishedOnly bool) ([]*Course, error)
	Count(publishedOnly bool) (int, error)
	GetByVideoStatus(status VideoStatus) ([]*Course, error)
	GetBySynthesiaVideoID(videoID string) (*Course, error)
}

type CourseContentRepository interface {
	Create(content *CourseContent) error
	GetByCourseID(courseID uuid.UUID) (*CourseContent, error)
	Update(content *CourseContent) error
}
