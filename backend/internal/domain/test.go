package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type QuestionType string

const (
	QuestionTypeMultipleChoice QuestionType = "multiple_choice"
	QuestionTypeDragDrop       QuestionType = "drag_drop"
	QuestionTypeFillBlank      QuestionType = "fill_blank"
	QuestionTypeMatching       QuestionType = "matching"
	QuestionTypeOrdering       QuestionType = "ordering"
)

type Test struct {
	ID               uuid.UUID `db:"id" json:"id"`
	CourseID         uuid.UUID `db:"course_id" json:"courseId"`
	Title            string    `db:"title" json:"title"`
	Description      string    `db:"description" json:"description"`
	TimeLimitMinutes *int      `db:"time_limit_minutes" json:"timeLimitMinutes,omitempty"`
	PassingScore     int       `db:"passing_score" json:"passingScore"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time `db:"updated_at" json:"updatedAt"`

	// Joined fields
	Questions []*Question `db:"-" json:"questions,omitempty"`
}

type Question struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	TestID       uuid.UUID       `db:"test_id" json:"testId"`
	QuestionType QuestionType    `db:"question_type" json:"questionType"`
	QuestionText string          `db:"question_text" json:"questionText"`
	QuestionData json.RawMessage `db:"question_data" json:"questionData"`
	Points       int             `db:"points" json:"points"`
	OrderIndex   int             `db:"order_index" json:"orderIndex"`
	CreatedAt    time.Time       `db:"created_at" json:"createdAt"`
}

// Multiple Choice question data structure
type MultipleChoiceData struct {
	Options        []string `json:"options"`
	CorrectIndices []int    `json:"correctIndices"`
	Explanation    string   `json:"explanation,omitempty"`
}

// Drag and Drop question data structure
type DragDropData struct {
	Items          []string          `json:"items"`
	DropZones      []string          `json:"dropZones"`
	CorrectMapping map[string]string `json:"correctMapping"`
	Explanation    string            `json:"explanation,omitempty"`
}

// Fill in the Blank question data structure
type FillBlankData struct {
	Template    string   `json:"template"` // Use {{blank}} for placeholders
	Blanks      []string `json:"blanks"`   // Correct answers for each blank
	Explanation string   `json:"explanation,omitempty"`
}

// Matching question data structure
type MatchingData struct {
	LeftItems    []string          `json:"leftItems"`
	RightItems   []string          `json:"rightItems"`
	CorrectPairs map[string]string `json:"correctPairs"`
	Explanation  string            `json:"explanation,omitempty"`
}

// Ordering question data structure
type OrderingData struct {
	Items        []string `json:"items"`
	CorrectOrder []int    `json:"correctOrder"`
	Explanation  string   `json:"explanation,omitempty"`
}

type TestAttempt struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	UserID      uuid.UUID  `db:"user_id" json:"userId"`
	TestID      uuid.UUID  `db:"test_id" json:"testId"`
	StartedAt   time.Time  `db:"started_at" json:"startedAt"`
	CompletedAt *time.Time `db:"completed_at" json:"completedAt,omitempty"`
	Score       *int       `db:"score" json:"score,omitempty"`
	MaxScore    *int       `db:"max_score" json:"maxScore,omitempty"`
	Percentage  *float64   `db:"percentage" json:"percentage,omitempty"`
	Passed      *bool      `db:"passed" json:"passed,omitempty"`

	// Joined fields
	Answers []*UserAnswer `db:"-" json:"answers,omitempty"`
}

type UserAnswer struct {
	ID         uuid.UUID       `db:"id" json:"id"`
	AttemptID  uuid.UUID       `db:"attempt_id" json:"attemptId"`
	QuestionID uuid.UUID       `db:"question_id" json:"questionId"`
	AnswerData json.RawMessage `db:"answer_data" json:"answerData"`
	IsCorrect  bool            `db:"is_correct" json:"isCorrect"`
	PointsAwarded int          `db:"points_awarded" json:"pointsAwarded"`
	CreatedAt  time.Time       `db:"created_at" json:"createdAt"`
}

type SubmitAnswerRequest struct {
	QuestionID uuid.UUID       `json:"questionId" validate:"required"`
	AnswerData json.RawMessage `json:"answerData" validate:"required"`
}

type SubmitTestRequest struct {
	Answers []SubmitAnswerRequest `json:"answers" validate:"required,dive"`
}

type TestResult struct {
	AttemptID   uuid.UUID       `json:"attemptId"`
	Score       int             `json:"score"`
	MaxScore    int             `json:"maxScore"`
	Percentage  float64         `json:"percentage"`
	Passed      bool            `json:"passed"`
	Answers     []*AnswerResult `json:"answers"`
}

type AnswerResult struct {
	QuestionID    uuid.UUID       `json:"questionId"`
	IsCorrect     bool            `json:"isCorrect"`
	PointsAwarded int             `json:"pointsAwarded"`
	MaxPoints     int             `json:"maxPoints"`
	Explanation   string          `json:"explanation,omitempty"`
}

type CreateTestRequest struct {
	CourseID         uuid.UUID `json:"courseId" validate:"required"`
	Title            string    `json:"title" validate:"required,min=1,max=255"`
	Description      string    `json:"description"`
	TimeLimitMinutes *int      `json:"timeLimitMinutes" validate:"omitempty,min=1"`
	PassingScore     int       `json:"passingScore" validate:"required,min=0,max=100"`
}

type CreateQuestionRequest struct {
	TestID       uuid.UUID       `json:"testId" validate:"required"`
	QuestionType QuestionType    `json:"questionType" validate:"required"`
	QuestionText string          `json:"questionText" validate:"required,min=1"`
	QuestionData json.RawMessage `json:"questionData" validate:"required"`
	Points       int             `json:"points" validate:"required,min=1"`
	OrderIndex   int             `json:"orderIndex"`
}

type TestRepository interface {
	Create(test *Test) error
	GetByID(id uuid.UUID) (*Test, error)
	GetByCourseID(courseID uuid.UUID) (*Test, error)
	Update(test *Test) error
	Delete(id uuid.UUID) error
}

type QuestionRepository interface {
	Create(question *Question) error
	GetByID(id uuid.UUID) (*Question, error)
	GetByTestID(testID uuid.UUID) ([]*Question, error)
	Update(question *Question) error
	Delete(id uuid.UUID) error
	DeleteByTestID(testID uuid.UUID) error
}

type TestAttemptRepository interface {
	Create(attempt *TestAttempt) error
	GetByID(id uuid.UUID) (*TestAttempt, error)
	GetByUserAndTest(userID, testID uuid.UUID) ([]*TestAttempt, error)
	Update(attempt *TestAttempt) error
	GetLatestByUserAndTest(userID, testID uuid.UUID) (*TestAttempt, error)
}

type UserAnswerRepository interface {
	Create(answer *UserAnswer) error
	CreateBatch(answers []*UserAnswer) error
	GetByAttemptID(attemptID uuid.UUID) ([]*UserAnswer, error)
}
