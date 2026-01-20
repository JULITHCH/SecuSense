package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type PresentationRepository struct {
	db *sqlx.DB
}

func NewPresentationRepository(db *sqlx.DB) *PresentationRepository {
	return &PresentationRepository{db: db}
}

// Create creates a new presentation
func (r *PresentationRepository) Create(p *domain.LessonPresentation) error {
	slidesJSON, err := json.Marshal(p.Slides)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO lesson_presentations (id, lesson_id, slides, status)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`
	return r.db.QueryRow(query, p.ID, p.LessonID, slidesJSON, p.Status).Scan(&p.CreatedAt)
}

// GetByLessonID retrieves a presentation by lesson ID
func (r *PresentationRepository) GetByLessonID(lessonID uuid.UUID) (*domain.LessonPresentation, error) {
	var p domain.LessonPresentation
	query := `SELECT id, lesson_id, slides, status, created_at FROM lesson_presentations WHERE lesson_id = $1`
	err := r.db.QueryRow(query, lessonID).Scan(&p.ID, &p.LessonID, &p.SlidesRaw, &p.Status, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal slides from JSON
	if err := json.Unmarshal(p.SlidesRaw, &p.Slides); err != nil {
		return nil, err
	}

	return &p, nil
}

// GetByID retrieves a presentation by its ID
func (r *PresentationRepository) GetByID(id uuid.UUID) (*domain.LessonPresentation, error) {
	var p domain.LessonPresentation
	query := `SELECT id, lesson_id, slides, status, created_at FROM lesson_presentations WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.LessonID, &p.SlidesRaw, &p.Status, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal slides from JSON
	if err := json.Unmarshal(p.SlidesRaw, &p.Slides); err != nil {
		return nil, err
	}

	return &p, nil
}

// UpdateSlides updates the slides for a presentation
func (r *PresentationRepository) UpdateSlides(id uuid.UUID, slides []domain.PresentationSlide) error {
	slidesJSON, err := json.Marshal(slides)
	if err != nil {
		return err
	}

	query := `UPDATE lesson_presentations SET slides = $2 WHERE id = $1`
	_, err = r.db.Exec(query, id, slidesJSON)
	return err
}

// UpdateStatus updates the status of a presentation
func (r *PresentationRepository) UpdateStatus(id uuid.UUID, status string) error {
	query := `UPDATE lesson_presentations SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}

// Delete deletes a presentation by ID
func (r *PresentationRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM lesson_presentations WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteByLessonID deletes all presentations for a lesson
func (r *PresentationRepository) DeleteByLessonID(lessonID uuid.UUID) error {
	query := `DELETE FROM lesson_presentations WHERE lesson_id = $1`
	_, err := r.db.Exec(query, lessonID)
	return err
}
