package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type JobType string
type JobStatus string

const (
	JobTypeResearch JobType = "research"
	JobTypeContent  JobType = "content_generation"
	JobTypeVideo    JobType = "video_generation"
	JobTypeTest     JobType = "test_generation"

	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type SuggestionStatus string

const (
	SuggestionPending  SuggestionStatus = "pending"
	SuggestionApproved SuggestionStatus = "approved"
	SuggestionRejected SuggestionStatus = "rejected"
)

type OutputType string

const (
	OutputTypeVideo        OutputType = "video"
	OutputTypePresentation OutputType = "presentation"
)

type WorkflowStep string

const (
	StepResearch     WorkflowStep = "research"     // AI generating topic suggestions
	StepSelection    WorkflowStep = "selection"    // User reviewing suggestions
	StepRefinement   WorkflowStep = "refinement"   // AI refining approved topics
	StepScriptGen    WorkflowStep = "script"       // AI generating lesson scripts
	StepVideoGen     WorkflowStep = "video"        // Generating videos/presentations
	StepQuestionGen  WorkflowStep = "questions"    // AI generating quiz questions
	StepCompleted    WorkflowStep = "completed"    // Workflow complete
)

// CourseWorkflowSession tracks the multi-agency course creation workflow
type CourseWorkflowSession struct {
	ID               uuid.UUID          `db:"id" json:"id"`
	MainTopic        string             `db:"main_topic" json:"mainTopic"`
	TargetAudience   string             `db:"target_audience" json:"targetAudience"`
	DifficultyLevel  string             `db:"difficulty_level" json:"difficultyLevel"`
	Language         string             `db:"language" json:"language"`
	VideoDurationMin int                `db:"video_duration_min" json:"videoDurationMin"`
	CurrentStep      WorkflowStep       `db:"current_step" json:"currentStep"`
	Status           JobStatus          `db:"status" json:"status"`
	CourseID         *uuid.UUID         `db:"course_id" json:"courseId,omitempty"`
	CreatedAt        time.Time          `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time          `db:"updated_at" json:"updatedAt"`
	Suggestions      []TopicSuggestion  `db:"-" json:"suggestions"`
	RefinedTopics    []RefinedTopic     `db:"-" json:"refinedTopics"`
	LessonScripts    []LessonScript     `db:"-" json:"lessonScripts"`
}

// TopicSuggestion represents a suggested subtopic for a course (Step 1: Research)
type TopicSuggestion struct {
	ID          uuid.UUID        `db:"id" json:"id"`
	SessionID   uuid.UUID        `db:"session_id" json:"sessionId"`
	Title       string           `db:"title" json:"title"`
	Description string           `db:"description" json:"description"`
	IsCustom    bool             `db:"is_custom" json:"isCustom"`
	Status      SuggestionStatus `db:"status" json:"status"`
	SortOrder   int              `db:"sort_order" json:"sortOrder"`
	CreatedAt   time.Time        `db:"created_at" json:"createdAt"`
}

// RefinedTopic represents a refined topic with lessons (Step 2: Refinement)
type RefinedTopic struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	SessionID       uuid.UUID       `db:"session_id" json:"sessionId"`
	SuggestionID    uuid.UUID       `db:"suggestion_id" json:"suggestionId"`
	Title           string          `db:"title" json:"title"`
	Description     string          `db:"description" json:"description"`
	LearningGoals   json.RawMessage `db:"learning_goals" json:"learningGoals"`
	EstimatedTimeMin int            `db:"estimated_time_min" json:"estimatedTimeMin"`
	SortOrder       int             `db:"sort_order" json:"sortOrder"`
	CreatedAt       time.Time       `db:"created_at" json:"createdAt"`
}

// LessonScript represents a generated script for a lesson (Step 3: Script Generation)
type LessonScript struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	SessionID          uuid.UUID  `db:"session_id" json:"sessionId"`
	TopicID            uuid.UUID  `db:"topic_id" json:"topicId"`
	Title              string     `db:"title" json:"title"`
	Script             string     `db:"script" json:"script"`
	DurationMin        int        `db:"duration_min" json:"durationMin"`
	OutputType         OutputType `db:"output_type" json:"outputType"`
	VideoID            *string    `db:"video_id" json:"videoId,omitempty"`
	VideoURL           *string    `db:"video_url" json:"videoUrl,omitempty"`
	VideoStatus        *string    `db:"video_status" json:"videoStatus,omitempty"`
	PresentationStatus *string    `db:"presentation_status" json:"presentationStatus,omitempty"`
	SortOrder          int        `db:"sort_order" json:"sortOrder"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
}

// PresentationSlide represents a single slide in a presentation
type PresentationSlide struct {
	Title         string `json:"title"`
	Content       string `json:"content"`       // HTML/Markdown content for the slide
	Script        string `json:"script"`        // TTS narration text
	AudioURL      string `json:"audioUrl"`      // URL to the generated audio file
	ImageURL      string `json:"imageUrl"`      // URL to stock image (from Unsplash)
	ImageAlt      string `json:"imageAlt"`      // Alt text for the image
	ImageKeywords string `json:"imageKeywords"` // Keywords used to search for the image
}

// LessonPresentation represents a RevealJS presentation for a lesson
type LessonPresentation struct {
	ID        uuid.UUID           `db:"id" json:"id"`
	LessonID  uuid.UUID           `db:"lesson_id" json:"lessonId"`
	Slides    []PresentationSlide `db:"-" json:"slides"`
	SlidesRaw json.RawMessage     `db:"slides" json:"-"`
	Status    string              `db:"status" json:"status"`
	CreatedAt time.Time           `db:"created_at" json:"createdAt"`
}

// SetOutputTypeRequest for changing lesson output type
type SetOutputTypeRequest struct {
	OutputType OutputType `json:"outputType" validate:"required,oneof=video presentation"`
}

// StartResearchRequest initiates the research phase
type StartResearchRequest struct {
	Topic            string `json:"topic" validate:"required,min=3,max=500"`
	TargetAudience   string `json:"targetAudience,omitempty"`
	DifficultyLevel  string `json:"difficultyLevel,omitempty" validate:"omitempty,oneof=beginner intermediate advanced"`
	Language         string `json:"language" validate:"required,oneof=en de fr es it pt"`
	VideoDurationMin int    `json:"videoDurationMin,omitempty" validate:"omitempty,min=1,max=60"`
}

// UpdateSuggestionRequest updates a suggestion's status
type UpdateSuggestionRequest struct {
	Status SuggestionStatus `json:"status" validate:"required,oneof=pending approved rejected"`
}

// AddCustomTopicRequest adds a custom topic
type AddCustomTopicRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Description string `json:"description,omitempty" validate:"max=1000"`
}

// UpdateRefinedTopicRequest for editing refined topic content
type UpdateRefinedTopicRequest struct {
	Title            string   `json:"title" validate:"required,min=3,max=200"`
	Description      string   `json:"description" validate:"required,max=2000"`
	LearningGoals    []string `json:"learningGoals" validate:"required,min=1,max=10"`
	EstimatedTimeMin int      `json:"estimatedTimeMin" validate:"required,min=1,max=60"`
}

// ReorderRefinedTopicsRequest for updating sort orders
type ReorderRefinedTopicsRequest struct {
	TopicOrders []TopicOrder `json:"topicOrders" validate:"required,min=1"`
}

// UpdateLessonScriptRequest for editing lesson script content
type UpdateLessonScriptRequest struct {
	Title  string `json:"title,omitempty" validate:"omitempty,min=3,max=200"`
	Script string `json:"script" validate:"required,min=10"`
}

// TopicOrder represents a topic ID and its new sort order
type TopicOrder struct {
	TopicID   string `json:"topicId" validate:"required,uuid"`
	SortOrder int    `json:"sortOrder" validate:"min=0"`
}

// GeneratedTopicSuggestions from AI
type GeneratedTopicSuggestions struct {
	Suggestions []struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"suggestions"`
}

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
