package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type WorkflowRepository struct {
	db *sqlx.DB
}

func NewWorkflowRepository(db *sqlx.DB) *WorkflowRepository {
	return &WorkflowRepository{db: db}
}

// Session operations

func (r *WorkflowRepository) CreateSession(session *domain.CourseWorkflowSession) error {
	query := `
		INSERT INTO course_workflow_sessions (id, main_topic, target_audience, difficulty_level, language, video_duration_min, current_step, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`
	return r.db.QueryRow(query,
		session.ID, session.MainTopic, session.TargetAudience, session.DifficultyLevel,
		session.Language, session.VideoDurationMin, session.CurrentStep, session.Status,
	).Scan(&session.CreatedAt, &session.UpdatedAt)
}

func (r *WorkflowRepository) GetSessionByID(id uuid.UUID) (*domain.CourseWorkflowSession, error) {
	var session domain.CourseWorkflowSession
	query := `SELECT * FROM course_workflow_sessions WHERE id = $1`
	err := r.db.Get(&session, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load suggestions (ensure non-nil slice for JSON serialization)
	suggestions, err := r.GetSuggestionsBySessionID(id)
	if err != nil {
		return nil, err
	}
	if suggestions == nil {
		suggestions = []domain.TopicSuggestion{}
	}
	session.Suggestions = suggestions

	// Load refined topics (ensure non-nil slice for JSON serialization)
	topics, err := r.GetRefinedTopicsBySessionID(id)
	if err != nil {
		return nil, err
	}
	if topics == nil {
		topics = []domain.RefinedTopic{}
	}
	session.RefinedTopics = topics

	// Load lesson scripts (ensure non-nil slice for JSON serialization)
	scripts, err := r.GetLessonScriptsBySessionID(id)
	if err != nil {
		return nil, err
	}
	if scripts == nil {
		scripts = []domain.LessonScript{}
	}
	session.LessonScripts = scripts

	return &session, nil
}

func (r *WorkflowRepository) GetSessionByCourseID(courseID uuid.UUID) (*domain.CourseWorkflowSession, error) {
	var session domain.CourseWorkflowSession
	query := `SELECT * FROM course_workflow_sessions WHERE course_id = $1 ORDER BY created_at DESC LIMIT 1`
	err := r.db.Get(&session, query, courseID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Load lesson scripts
	scripts, err := r.GetLessonScriptsBySessionID(session.ID)
	if err != nil {
		return nil, err
	}
	if scripts == nil {
		scripts = []domain.LessonScript{}
	}
	session.LessonScripts = scripts

	return &session, nil
}

func (r *WorkflowRepository) UpdateSession(session *domain.CourseWorkflowSession) error {
	query := `
		UPDATE course_workflow_sessions
		SET current_step = $2, status = $3, course_id = $4, updated_at = NOW()
		WHERE id = $1`
	_, err := r.db.Exec(query, session.ID, session.CurrentStep, session.Status, session.CourseID)
	return err
}

// Topic Suggestions operations

func (r *WorkflowRepository) CreateSuggestion(suggestion *domain.TopicSuggestion) error {
	query := `
		INSERT INTO topic_suggestions (id, session_id, title, description, is_custom, status, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`
	return r.db.QueryRow(query,
		suggestion.ID, suggestion.SessionID, suggestion.Title, suggestion.Description,
		suggestion.IsCustom, suggestion.Status, suggestion.SortOrder,
	).Scan(&suggestion.CreatedAt)
}

func (r *WorkflowRepository) CreateSuggestionsBatch(suggestions []domain.TopicSuggestion) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO topic_suggestions (id, session_id, title, description, is_custom, status, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	for _, s := range suggestions {
		_, err := tx.Exec(query, s.ID, s.SessionID, s.Title, s.Description, s.IsCustom, s.Status, s.SortOrder)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *WorkflowRepository) GetSuggestionsBySessionID(sessionID uuid.UUID) ([]domain.TopicSuggestion, error) {
	var suggestions []domain.TopicSuggestion
	query := `SELECT * FROM topic_suggestions WHERE session_id = $1 ORDER BY sort_order, created_at`
	err := r.db.Select(&suggestions, query, sessionID)
	return suggestions, err
}

func (r *WorkflowRepository) GetApprovedSuggestions(sessionID uuid.UUID) ([]domain.TopicSuggestion, error) {
	var suggestions []domain.TopicSuggestion
	query := `SELECT * FROM topic_suggestions WHERE session_id = $1 AND status = 'approved' ORDER BY sort_order, created_at`
	err := r.db.Select(&suggestions, query, sessionID)
	return suggestions, err
}

func (r *WorkflowRepository) UpdateSuggestionStatus(id uuid.UUID, status domain.SuggestionStatus) error {
	query := `UPDATE topic_suggestions SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}

func (r *WorkflowRepository) GetSuggestionByID(id uuid.UUID) (*domain.TopicSuggestion, error) {
	var suggestion domain.TopicSuggestion
	query := `SELECT * FROM topic_suggestions WHERE id = $1`
	err := r.db.Get(&suggestion, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &suggestion, err
}

// Refined Topics operations

func (r *WorkflowRepository) CreateRefinedTopic(topic *domain.RefinedTopic) error {
	query := `
		INSERT INTO refined_topics (id, session_id, suggestion_id, title, description, learning_goals, estimated_time_min, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`
	return r.db.QueryRow(query,
		topic.ID, topic.SessionID, topic.SuggestionID, topic.Title, topic.Description,
		topic.LearningGoals, topic.EstimatedTimeMin, topic.SortOrder,
	).Scan(&topic.CreatedAt)
}

func (r *WorkflowRepository) CreateRefinedTopicsBatch(topics []domain.RefinedTopic) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO refined_topics (id, session_id, suggestion_id, title, description, learning_goals, estimated_time_min, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	for _, t := range topics {
		_, err := tx.Exec(query, t.ID, t.SessionID, t.SuggestionID, t.Title, t.Description, t.LearningGoals, t.EstimatedTimeMin, t.SortOrder)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *WorkflowRepository) GetRefinedTopicsBySessionID(sessionID uuid.UUID) ([]domain.RefinedTopic, error) {
	var topics []domain.RefinedTopic
	query := `SELECT * FROM refined_topics WHERE session_id = $1 ORDER BY sort_order, created_at`
	err := r.db.Select(&topics, query, sessionID)
	return topics, err
}

func (r *WorkflowRepository) GetRefinedTopicByID(id uuid.UUID) (*domain.RefinedTopic, error) {
	var topic domain.RefinedTopic
	query := `SELECT * FROM refined_topics WHERE id = $1`
	err := r.db.Get(&topic, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &topic, err
}

func (r *WorkflowRepository) UpdateRefinedTopic(topic *domain.RefinedTopic) error {
	query := `
		UPDATE refined_topics
		SET title = $2, description = $3, learning_goals = $4, estimated_time_min = $5
		WHERE id = $1`
	_, err := r.db.Exec(query, topic.ID, topic.Title, topic.Description,
		topic.LearningGoals, topic.EstimatedTimeMin)
	return err
}

func (r *WorkflowRepository) UpdateRefinedTopicSortOrders(orders map[uuid.UUID]int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE refined_topics SET sort_order = $2 WHERE id = $1`
	for id, sortOrder := range orders {
		_, err := tx.Exec(query, id, sortOrder)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Lesson Scripts operations

func (r *WorkflowRepository) CreateLessonScript(script *domain.LessonScript) error {
	query := `
		INSERT INTO lesson_scripts (id, session_id, topic_id, title, script, duration_min, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`
	return r.db.QueryRow(query,
		script.ID, script.SessionID, script.TopicID, script.Title, script.Script,
		script.DurationMin, script.SortOrder,
	).Scan(&script.CreatedAt)
}

func (r *WorkflowRepository) CreateLessonScriptsBatch(scripts []domain.LessonScript) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO lesson_scripts (id, session_id, topic_id, title, script, duration_min, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	for _, s := range scripts {
		_, err := tx.Exec(query, s.ID, s.SessionID, s.TopicID, s.Title, s.Script, s.DurationMin, s.SortOrder)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *WorkflowRepository) GetLessonScriptsBySessionID(sessionID uuid.UUID) ([]domain.LessonScript, error) {
	var scripts []domain.LessonScript
	query := `SELECT * FROM lesson_scripts WHERE session_id = $1 ORDER BY sort_order, created_at`
	err := r.db.Select(&scripts, query, sessionID)
	return scripts, err
}

func (r *WorkflowRepository) UpdateLessonScriptVideo(id uuid.UUID, videoID, videoURL, videoStatus string) error {
	query := `UPDATE lesson_scripts SET video_id = $2, video_url = $3, video_status = $4 WHERE id = $1`
	_, err := r.db.Exec(query, id, videoID, videoURL, videoStatus)
	return err
}

func (r *WorkflowRepository) UpdateLessonScriptOutputType(id uuid.UUID, outputType domain.OutputType) error {
	query := `UPDATE lesson_scripts SET output_type = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, outputType)
	return err
}

func (r *WorkflowRepository) UpdateLessonScriptPresentationStatus(id uuid.UUID, status string) error {
	query := `UPDATE lesson_scripts SET presentation_status = $2 WHERE id = $1`
	_, err := r.db.Exec(query, id, status)
	return err
}

func (r *WorkflowRepository) GetLessonScriptByID(id uuid.UUID) (*domain.LessonScript, error) {
	var script domain.LessonScript
	query := `SELECT * FROM lesson_scripts WHERE id = $1`
	err := r.db.Get(&script, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &script, err
}

func (r *WorkflowRepository) UpdateLessonScript(script *domain.LessonScript) error {
	query := `UPDATE lesson_scripts SET title = $2, script = $3, duration_min = $4 WHERE id = $1`
	_, err := r.db.Exec(query, script.ID, script.Title, script.Script, script.DurationMin)
	return err
}
