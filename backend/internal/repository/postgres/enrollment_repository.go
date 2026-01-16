package postgres

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type EnrollmentRepository struct {
	db *sqlx.DB
}

func NewEnrollmentRepository(db *sqlx.DB) *EnrollmentRepository {
	return &EnrollmentRepository{db: db}
}

func (r *EnrollmentRepository) Create(enrollment *domain.Enrollment) error {
	query := `
		INSERT INTO enrollments (id, user_id, course_id, status, progress_percentage, video_watched, enrolled_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING enrolled_at, updated_at`

	if enrollment.ID == uuid.Nil {
		enrollment.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		enrollment.ID, enrollment.UserID, enrollment.CourseID, enrollment.Status,
		enrollment.ProgressPercentage, enrollment.VideoWatched,
	).Scan(&enrollment.EnrolledAt, &enrollment.UpdatedAt)
}

func (r *EnrollmentRepository) GetByID(id uuid.UUID) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	query := `SELECT id, user_id, course_id, status, progress_percentage, video_watched, enrolled_at, completed_at, updated_at
			  FROM enrollments WHERE id = $1`

	err := r.db.Get(&enrollment, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}

func (r *EnrollmentRepository) GetByUserAndCourse(userID, courseID uuid.UUID) (*domain.Enrollment, error) {
	var enrollment domain.Enrollment
	query := `SELECT id, user_id, course_id, status, progress_percentage, video_watched, enrolled_at, completed_at, updated_at
			  FROM enrollments WHERE user_id = $1 AND course_id = $2`

	err := r.db.Get(&enrollment, query, userID, courseID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &enrollment, nil
}

func (r *EnrollmentRepository) Update(enrollment *domain.Enrollment) error {
	query := `
		UPDATE enrollments
		SET status = $1, progress_percentage = $2, video_watched = $3, completed_at = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at`

	return r.db.QueryRow(
		query,
		enrollment.Status, enrollment.ProgressPercentage, enrollment.VideoWatched,
		enrollment.CompletedAt, enrollment.ID,
	).Scan(&enrollment.UpdatedAt)
}

func (r *EnrollmentRepository) ListByUser(userID uuid.UUID) ([]*domain.EnrollmentWithCourse, error) {
	var enrollments []*domain.EnrollmentWithCourse
	query := `
		SELECT e.id, e.user_id, e.course_id, e.status, e.progress_percentage, e.video_watched,
		       e.enrolled_at, e.completed_at, e.updated_at,
		       c.title as course_title, c.description as course_description, c.thumbnail_url as course_thumbnail_url
		FROM enrollments e
		JOIN courses c ON e.course_id = c.id
		WHERE e.user_id = $1
		ORDER BY e.enrolled_at DESC`

	err := r.db.Select(&enrollments, query, userID)
	if err != nil {
		return nil, err
	}
	return enrollments, nil
}

func (r *EnrollmentRepository) ListByCourse(courseID uuid.UUID) ([]*domain.Enrollment, error) {
	var enrollments []*domain.Enrollment
	query := `SELECT id, user_id, course_id, status, progress_percentage, video_watched, enrolled_at, completed_at, updated_at
			  FROM enrollments WHERE course_id = $1 ORDER BY enrolled_at DESC`

	err := r.db.Select(&enrollments, query, courseID)
	if err != nil {
		return nil, err
	}
	return enrollments, nil
}

func (r *EnrollmentRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM enrollments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
