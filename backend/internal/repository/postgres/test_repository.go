package postgres

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/secusense/backend/internal/domain"
)

type TestRepository struct {
	db *sqlx.DB
}

func NewTestRepository(db *sqlx.DB) *TestRepository {
	return &TestRepository{db: db}
}

func (r *TestRepository) Create(test *domain.Test) error {
	query := `
		INSERT INTO tests (id, course_id, title, description, time_limit_minutes, passing_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at`

	if test.ID == uuid.Nil {
		test.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		test.ID, test.CourseID, test.Title, test.Description, test.TimeLimitMinutes, test.PassingScore,
	).Scan(&test.CreatedAt, &test.UpdatedAt)
}

func (r *TestRepository) GetByID(id uuid.UUID) (*domain.Test, error) {
	var test domain.Test
	query := `SELECT id, course_id, title, description, time_limit_minutes, passing_score, created_at, updated_at
			  FROM tests WHERE id = $1`

	err := r.db.Get(&test, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (r *TestRepository) GetByCourseID(courseID uuid.UUID) (*domain.Test, error) {
	var test domain.Test
	query := `SELECT id, course_id, title, description, time_limit_minutes, passing_score, created_at, updated_at
			  FROM tests WHERE course_id = $1`

	err := r.db.Get(&test, query, courseID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &test, nil
}

func (r *TestRepository) Update(test *domain.Test) error {
	query := `
		UPDATE tests
		SET title = $1, description = $2, time_limit_minutes = $3, passing_score = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at`

	return r.db.QueryRow(
		query,
		test.Title, test.Description, test.TimeLimitMinutes, test.PassingScore, test.ID,
	).Scan(&test.UpdatedAt)
}

func (r *TestRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM tests WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

type QuestionRepository struct {
	db *sqlx.DB
}

func NewQuestionRepository(db *sqlx.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) Create(question *domain.Question) error {
	query := `
		INSERT INTO questions (id, test_id, question_type, question_text, question_data, points, order_index, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		RETURNING created_at`

	if question.ID == uuid.Nil {
		question.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		question.ID, question.TestID, question.QuestionType, question.QuestionText,
		question.QuestionData, question.Points, question.OrderIndex,
	).Scan(&question.CreatedAt)
}

func (r *QuestionRepository) GetByID(id uuid.UUID) (*domain.Question, error) {
	var question domain.Question
	query := `SELECT id, test_id, question_type, question_text, question_data, points, order_index, created_at
			  FROM questions WHERE id = $1`

	err := r.db.Get(&question, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &question, nil
}

func (r *QuestionRepository) GetByTestID(testID uuid.UUID) ([]*domain.Question, error) {
	var questions []*domain.Question
	query := `SELECT id, test_id, question_type, question_text, question_data, points, order_index, created_at
			  FROM questions WHERE test_id = $1 ORDER BY order_index ASC`

	err := r.db.Select(&questions, query, testID)
	if err != nil {
		return nil, err
	}
	return questions, nil
}

func (r *QuestionRepository) Update(question *domain.Question) error {
	query := `
		UPDATE questions
		SET question_type = $1, question_text = $2, question_data = $3, points = $4, order_index = $5
		WHERE id = $6`

	_, err := r.db.Exec(query, question.QuestionType, question.QuestionText, question.QuestionData,
		question.Points, question.OrderIndex, question.ID)
	return err
}

func (r *QuestionRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM questions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *QuestionRepository) DeleteByTestID(testID uuid.UUID) error {
	query := `DELETE FROM questions WHERE test_id = $1`
	_, err := r.db.Exec(query, testID)
	return err
}

type TestAttemptRepository struct {
	db *sqlx.DB
}

func NewTestAttemptRepository(db *sqlx.DB) *TestAttemptRepository {
	return &TestAttemptRepository{db: db}
}

func (r *TestAttemptRepository) Create(attempt *domain.TestAttempt) error {
	query := `
		INSERT INTO test_attempts (id, user_id, test_id, started_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING started_at`

	if attempt.ID == uuid.Nil {
		attempt.ID = uuid.New()
	}

	return r.db.QueryRow(query, attempt.ID, attempt.UserID, attempt.TestID).Scan(&attempt.StartedAt)
}

func (r *TestAttemptRepository) GetByID(id uuid.UUID) (*domain.TestAttempt, error) {
	var attempt domain.TestAttempt
	query := `SELECT id, user_id, test_id, started_at, completed_at, score, max_score, percentage, passed
			  FROM test_attempts WHERE id = $1`

	err := r.db.Get(&attempt, query, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *TestAttemptRepository) GetByUserAndTest(userID, testID uuid.UUID) ([]*domain.TestAttempt, error) {
	var attempts []*domain.TestAttempt
	query := `SELECT id, user_id, test_id, started_at, completed_at, score, max_score, percentage, passed
			  FROM test_attempts WHERE user_id = $1 AND test_id = $2 ORDER BY started_at DESC`

	err := r.db.Select(&attempts, query, userID, testID)
	if err != nil {
		return nil, err
	}
	return attempts, nil
}

func (r *TestAttemptRepository) GetLatestByUserAndTest(userID, testID uuid.UUID) (*domain.TestAttempt, error) {
	var attempt domain.TestAttempt
	query := `SELECT id, user_id, test_id, started_at, completed_at, score, max_score, percentage, passed
			  FROM test_attempts WHERE user_id = $1 AND test_id = $2 ORDER BY started_at DESC LIMIT 1`

	err := r.db.Get(&attempt, query, userID, testID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *TestAttemptRepository) Update(attempt *domain.TestAttempt) error {
	query := `
		UPDATE test_attempts
		SET completed_at = $1, score = $2, max_score = $3, percentage = $4, passed = $5
		WHERE id = $6`

	_, err := r.db.Exec(query, attempt.CompletedAt, attempt.Score, attempt.MaxScore,
		attempt.Percentage, attempt.Passed, attempt.ID)
	return err
}

type UserAnswerRepository struct {
	db *sqlx.DB
}

func NewUserAnswerRepository(db *sqlx.DB) *UserAnswerRepository {
	return &UserAnswerRepository{db: db}
}

func (r *UserAnswerRepository) Create(answer *domain.UserAnswer) error {
	query := `
		INSERT INTO user_answers (id, attempt_id, question_id, answer_data, is_correct, points_awarded, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		RETURNING created_at`

	if answer.ID == uuid.Nil {
		answer.ID = uuid.New()
	}

	return r.db.QueryRow(
		query,
		answer.ID, answer.AttemptID, answer.QuestionID, answer.AnswerData, answer.IsCorrect, answer.PointsAwarded,
	).Scan(&answer.CreatedAt)
}

func (r *UserAnswerRepository) CreateBatch(answers []*domain.UserAnswer) error {
	query := `
		INSERT INTO user_answers (id, attempt_id, question_id, answer_data, is_correct, points_awarded, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())`

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}

	for _, answer := range answers {
		if answer.ID == uuid.Nil {
			answer.ID = uuid.New()
		}
		_, err = tx.Exec(query, answer.ID, answer.AttemptID, answer.QuestionID,
			answer.AnswerData, answer.IsCorrect, answer.PointsAwarded)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *UserAnswerRepository) GetByAttemptID(attemptID uuid.UUID) ([]*domain.UserAnswer, error) {
	var answers []*domain.UserAnswer
	query := `SELECT id, attempt_id, question_id, answer_data, is_correct, points_awarded, created_at
			  FROM user_answers WHERE attempt_id = $1`

	err := r.db.Select(&answers, query, attemptID)
	if err != nil {
		return nil, err
	}
	return answers, nil
}
