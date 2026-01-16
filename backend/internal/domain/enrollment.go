package domain

import (
	"time"

	"github.com/google/uuid"
)

type EnrollmentStatus string

const (
	EnrollmentStatusActive    EnrollmentStatus = "active"
	EnrollmentStatusCompleted EnrollmentStatus = "completed"
	EnrollmentStatusDropped   EnrollmentStatus = "dropped"
)

type Enrollment struct {
	ID                 uuid.UUID        `db:"id" json:"id"`
	UserID             uuid.UUID        `db:"user_id" json:"userId"`
	CourseID           uuid.UUID        `db:"course_id" json:"courseId"`
	Status             EnrollmentStatus `db:"status" json:"status"`
	ProgressPercentage int              `db:"progress_percentage" json:"progressPercentage"`
	VideoWatched       bool             `db:"video_watched" json:"videoWatched"`
	EnrolledAt         time.Time        `db:"enrolled_at" json:"enrolledAt"`
	CompletedAt        *time.Time       `db:"completed_at" json:"completedAt,omitempty"`
	UpdatedAt          time.Time        `db:"updated_at" json:"updatedAt"`

	// Joined fields
	Course *Course `db:"-" json:"course,omitempty"`
}

type EnrollmentWithCourse struct {
	Enrollment
	CourseTitle       string  `db:"course_title" json:"courseTitle"`
	CourseDescription string  `db:"course_description" json:"courseDescription"`
	CourseThumbnail   *string `db:"course_thumbnail_url" json:"courseThumbnailUrl,omitempty"`
}

type UpdateProgressRequest struct {
	ProgressPercentage int `json:"progressPercentage" validate:"required,min=0,max=100"`
}

type EnrollmentRepository interface {
	Create(enrollment *Enrollment) error
	GetByID(id uuid.UUID) (*Enrollment, error)
	GetByUserAndCourse(userID, courseID uuid.UUID) (*Enrollment, error)
	Update(enrollment *Enrollment) error
	ListByUser(userID uuid.UUID) ([]*EnrollmentWithCourse, error)
	ListByCourse(courseID uuid.UUID) ([]*Enrollment, error)
	Delete(id uuid.UUID) error
}
