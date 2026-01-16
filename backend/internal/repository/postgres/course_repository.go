package postgres

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type CourseRepository struct {
	db *sqlx.DB
}

func NewCourseRepository(db *sqlx.DB) *CourseRepository {
	return &CourseRepository{db: db}
}

func (r *CourseRepository) Create(course *domain.Course) error {
	query := `
		INSERT INTO courses (id, title, description, video_url, synthesia_video_id, video_status, video_error, thumbnail_url, pass_percentage, is_published, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING created_at, updated_at`

	if course.ID == uuid.Nil {
		course.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		course.ID, course.Title, course.Description, course.VideoURL, course.SynthesiaVideoID,
		course.VideoStatus, course.VideoError, course.ThumbnailURL, course.PassPercentage, course.IsPublished,
	).Scan(&course.CreatedAt, &course.UpdatedAt)
}

func (r *CourseRepository) GetByID(id uuid.UUID) (*domain.Course, error) {
	var course domain.Course
	query := `SELECT id, title, description, video_url, synthesia_video_id, video_status, video_error, thumbnail_url, pass_percentage, is_published, created_at, updated_at
			  FROM courses WHERE id = $1`

	err := r.db.Get(&course, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) Update(course *domain.Course) error {
	query := `
		UPDATE courses
		SET title = $1, description = $2, video_url = $3, synthesia_video_id = $4,
		    video_status = $5, video_error = $6, thumbnail_url = $7, pass_percentage = $8, is_published = $9, updated_at = NOW()
		WHERE id = $10
		RETURNING updated_at`

	return r.db.QueryRow(
		query,
		course.Title, course.Description, course.VideoURL, course.SynthesiaVideoID,
		course.VideoStatus, course.VideoError, course.ThumbnailURL, course.PassPercentage, course.IsPublished, course.ID,
	).Scan(&course.UpdatedAt)
}

func (r *CourseRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM courses WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *CourseRepository) List(limit, offset int, publishedOnly bool) ([]*domain.Course, error) {
	var courses []*domain.Course
	var query string

	if publishedOnly {
		query = `SELECT id, title, description, video_url, synthesia_video_id, video_status, video_error, thumbnail_url, pass_percentage, is_published, created_at, updated_at
				 FROM courses WHERE is_published = true ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	} else {
		query = `SELECT id, title, description, video_url, synthesia_video_id, video_status, video_error, thumbnail_url, pass_percentage, is_published, created_at, updated_at
				 FROM courses ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	}

	err := r.db.Select(&courses, query, limit, offset)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// GetByVideoStatus returns courses with a specific video status
func (r *CourseRepository) GetByVideoStatus(status domain.VideoStatus) ([]*domain.Course, error) {
	var courses []*domain.Course
	query := `SELECT id, title, description, video_url, synthesia_video_id, video_status, video_error, thumbnail_url, pass_percentage, is_published, created_at, updated_at
			  FROM courses WHERE video_status = $1 ORDER BY updated_at ASC`

	err := r.db.Select(&courses, query, status)
	if err != nil {
		return nil, err
	}
	return courses, nil
}

// GetBySynthesiaVideoID returns a course by its Synthesia video ID
func (r *CourseRepository) GetBySynthesiaVideoID(videoID string) (*domain.Course, error) {
	var course domain.Course
	query := `SELECT id, title, description, video_url, synthesia_video_id, video_status, video_error, thumbnail_url, pass_percentage, is_published, created_at, updated_at
			  FROM courses WHERE synthesia_video_id = $1`

	err := r.db.Get(&course, query, videoID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &course, nil
}

func (r *CourseRepository) Count(publishedOnly bool) (int, error) {
	var count int
	var query string

	if publishedOnly {
		query = `SELECT COUNT(*) FROM courses WHERE is_published = true`
	} else {
		query = `SELECT COUNT(*) FROM courses`
	}

	err := r.db.Get(&count, query)
	return count, err
}

type CourseContentRepository struct {
	db *sqlx.DB
}

func NewCourseContentRepository(db *sqlx.DB) *CourseContentRepository {
	return &CourseContentRepository{db: db}
}

func (r *CourseContentRepository) Create(content *domain.CourseContent) error {
	query := `
		INSERT INTO course_content (id, course_id, outline, video_script, generation_prompt, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING created_at`

	if content.ID == uuid.Nil {
		content.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		content.ID, content.CourseID, content.Outline, content.VideoScript, content.GenerationPrompt,
	).Scan(&content.CreatedAt)
}

func (r *CourseContentRepository) GetByCourseID(courseID uuid.UUID) (*domain.CourseContent, error) {
	var content domain.CourseContent
	query := `SELECT id, course_id, outline, video_script, generation_prompt, created_at
			  FROM course_content WHERE course_id = $1`

	err := r.db.Get(&content, query, courseID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func (r *CourseContentRepository) Update(content *domain.CourseContent) error {
	query := `
		UPDATE course_content
		SET outline = $1, video_script = $2, generation_prompt = $3
		WHERE id = $4`

	_, err := r.db.Exec(query, content.Outline, content.VideoScript, content.GenerationPrompt, content.ID)
	return err
}
