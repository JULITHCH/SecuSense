package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/secusense/backend/infrastructure/ollama"
	"github.com/secusense/backend/infrastructure/synthesia"
	"github.com/secusense/backend/infrastructure/tts"
	"github.com/secusense/backend/infrastructure/unsplash"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/internal/repository/postgres"
)

var (
	ErrSessionNotFound    = errors.New("workflow session not found")
	ErrSuggestionNotFound = errors.New("suggestion not found")
	ErrTopicNotFound      = errors.New("refined topic not found")
	ErrLessonNotFound     = errors.New("lesson script not found")
	ErrInvalidStep        = errors.New("invalid workflow step for this operation")
	ErrNoApprovedTopics   = errors.New("no approved topics to proceed")
)

type UseCase struct {
	workflowRepo     *postgres.WorkflowRepository
	presentationRepo *postgres.PresentationRepository
	courseRepo       *postgres.CourseRepository
	testRepo         domain.TestRepository
	questionRepo     domain.QuestionRepository
	ollamaClient     *ollama.Client
	synthesiaClient  *synthesia.Client
	ttsClient        *tts.Client
	unsplashClient   *unsplash.Client
}

func NewUseCase(
	workflowRepo *postgres.WorkflowRepository,
	presentationRepo *postgres.PresentationRepository,
	courseRepo *postgres.CourseRepository,
	testRepo domain.TestRepository,
	questionRepo domain.QuestionRepository,
	ollamaClient *ollama.Client,
	synthesiaClient *synthesia.Client,
	ttsClient *tts.Client,
	unsplashClient *unsplash.Client,
) *UseCase {
	return &UseCase{
		workflowRepo:     workflowRepo,
		presentationRepo: presentationRepo,
		courseRepo:       courseRepo,
		testRepo:         testRepo,
		questionRepo:     questionRepo,
		ollamaClient:     ollamaClient,
		synthesiaClient:  synthesiaClient,
		ttsClient:        ttsClient,
		unsplashClient:   unsplashClient,
	}
}

// StartResearch begins a new workflow by generating topic suggestions
func (uc *UseCase) StartResearch(ctx context.Context, req *domain.StartResearchRequest) (*domain.CourseWorkflowSession, error) {
	// Default language to English if not specified
	language := req.Language
	if language == "" {
		language = "en"
	}

	// Create session
	session := &domain.CourseWorkflowSession{
		ID:               uuid.New(),
		MainTopic:        req.Topic,
		TargetAudience:   req.TargetAudience,
		DifficultyLevel:  req.DifficultyLevel,
		Language:         language,
		VideoDurationMin: req.VideoDurationMin,
		CurrentStep:      domain.StepResearch,
		Status:           domain.JobStatusProcessing,
	}

	if session.VideoDurationMin == 0 {
		session.VideoDurationMin = 5
	}

	if err := uc.workflowRepo.CreateSession(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Generate suggestions async
	go uc.generateSuggestionsAsync(session.ID, req.Topic, req.TargetAudience, req.DifficultyLevel, language)

	return session, nil
}

func (uc *UseCase) generateSuggestionsAsync(sessionID uuid.UUID, topic, audience, difficulty, language string) {
	ctx := context.Background()

	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return
	}

	// Call Research Agency
	suggestions, err := uc.ollamaClient.ResearchTopicSuggestions(ctx, topic, audience, difficulty, language, 6)
	if err != nil {
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	// Save suggestions
	var domainSuggestions []domain.TopicSuggestion
	for i, s := range suggestions {
		domainSuggestions = append(domainSuggestions, domain.TopicSuggestion{
			ID:          uuid.New(),
			SessionID:   sessionID,
			Title:       s.Title,
			Description: s.Description,
			IsCustom:    false,
			Status:      domain.SuggestionPending,
			SortOrder:   i,
		})
	}

	if err := uc.workflowRepo.CreateSuggestionsBatch(domainSuggestions); err != nil {
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	// Update session to selection step
	session.CurrentStep = domain.StepSelection
	session.Status = domain.JobStatusCompleted
	uc.workflowRepo.UpdateSession(session)
}

// GetSession retrieves a workflow session with all its data
func (uc *UseCase) GetSession(sessionID uuid.UUID) (*domain.CourseWorkflowSession, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

// UpdateSuggestionStatus updates the approval status of a suggestion
func (uc *UseCase) UpdateSuggestionStatus(sessionID, suggestionID uuid.UUID, status domain.SuggestionStatus) error {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return ErrSessionNotFound
	}

	if session.CurrentStep != domain.StepSelection {
		return ErrInvalidStep
	}

	suggestion, err := uc.workflowRepo.GetSuggestionByID(suggestionID)
	if err != nil || suggestion == nil {
		return ErrSuggestionNotFound
	}

	if suggestion.SessionID != sessionID {
		return ErrSuggestionNotFound
	}

	return uc.workflowRepo.UpdateSuggestionStatus(suggestionID, status)
}

// AddCustomTopic adds a user-defined topic to the session
func (uc *UseCase) AddCustomTopic(sessionID uuid.UUID, req *domain.AddCustomTopicRequest) (*domain.TopicSuggestion, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	if session.CurrentStep != domain.StepSelection {
		return nil, ErrInvalidStep
	}

	suggestion := &domain.TopicSuggestion{
		ID:          uuid.New(),
		SessionID:   sessionID,
		Title:       req.Title,
		Description: req.Description,
		IsCustom:    true,
		Status:      domain.SuggestionApproved, // Custom topics are auto-approved
		SortOrder:   len(session.Suggestions),
	}

	if err := uc.workflowRepo.CreateSuggestion(suggestion); err != nil {
		return nil, err
	}

	return suggestion, nil
}

// GenerateMoreSuggestions generates additional topic suggestions
func (uc *UseCase) GenerateMoreSuggestions(ctx context.Context, sessionID uuid.UUID) error {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return ErrSessionNotFound
	}

	if session.CurrentStep != domain.StepSelection {
		return ErrInvalidStep
	}

	// Generate more suggestions
	suggestions, err := uc.ollamaClient.ResearchTopicSuggestions(
		ctx,
		session.MainTopic,
		session.TargetAudience,
		session.DifficultyLevel,
		session.Language,
		4, // Generate 4 more
	)
	if err != nil {
		return err
	}

	// Save new suggestions
	existingCount := len(session.Suggestions)
	var domainSuggestions []domain.TopicSuggestion
	for i, s := range suggestions {
		domainSuggestions = append(domainSuggestions, domain.TopicSuggestion{
			ID:          uuid.New(),
			SessionID:   sessionID,
			Title:       s.Title,
			Description: s.Description,
			IsCustom:    false,
			Status:      domain.SuggestionPending,
			SortOrder:   existingCount + i,
		})
	}

	return uc.workflowRepo.CreateSuggestionsBatch(domainSuggestions)
}

// ProceedToRefinement moves to the refinement step and processes approved topics
func (uc *UseCase) ProceedToRefinement(ctx context.Context, sessionID uuid.UUID) (*domain.CourseWorkflowSession, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Allow proceeding from selection, or retrying a failed refinement
	canProceed := session.CurrentStep == domain.StepSelection ||
		(session.CurrentStep == domain.StepRefinement && session.Status == domain.JobStatusFailed)
	if !canProceed {
		return nil, ErrInvalidStep
	}

	// Get approved suggestions
	approved, err := uc.workflowRepo.GetApprovedSuggestions(sessionID)
	if err != nil {
		return nil, err
	}

	if len(approved) == 0 {
		return nil, ErrNoApprovedTopics
	}

	// Update status to processing
	session.CurrentStep = domain.StepRefinement
	session.Status = domain.JobStatusProcessing
	if err := uc.workflowRepo.UpdateSession(session); err != nil {
		return nil, err
	}

	// Process refinement async
	go uc.processRefinementAsync(session, approved)

	return session, nil
}

func (uc *UseCase) processRefinementAsync(session *domain.CourseWorkflowSession, approved []domain.TopicSuggestion) {
	ctx := context.Background()

	// Convert to Ollama format
	var suggestions []ollama.TopicSuggestionResult
	for _, a := range approved {
		suggestions = append(suggestions, ollama.TopicSuggestionResult{
			Title:       a.Title,
			Description: a.Description,
		})
	}

	// Call Refinement Agency
	refined, err := uc.ollamaClient.RefineTopics(ctx, session.MainTopic, suggestions, session.TargetAudience, session.DifficultyLevel, session.Language)
	if err != nil {
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	// Save refined topics
	var domainTopics []domain.RefinedTopic
	for i, r := range refined {
		// Find matching suggestion
		var suggestionID uuid.UUID
		for _, a := range approved {
			if a.Title == r.OriginalTitle {
				suggestionID = a.ID
				break
			}
		}
		if suggestionID == uuid.Nil {
			suggestionID = approved[i].ID // Fallback to index
		}

		goalsJSON, _ := json.Marshal(r.LearningGoals)

		domainTopics = append(domainTopics, domain.RefinedTopic{
			ID:               uuid.New(),
			SessionID:        session.ID,
			SuggestionID:     suggestionID,
			Title:            r.Title,
			Description:      r.Description,
			LearningGoals:    goalsJSON,
			EstimatedTimeMin: r.EstimatedTimeMin,
			SortOrder:        i,
		})
	}

	if err := uc.workflowRepo.CreateRefinedTopicsBatch(domainTopics); err != nil {
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	// Move to script generation step
	session.CurrentStep = domain.StepScriptGen
	session.Status = domain.JobStatusCompleted
	uc.workflowRepo.UpdateSession(session)
}

// ProceedToScriptGeneration generates scripts for all refined topics
func (uc *UseCase) ProceedToScriptGeneration(ctx context.Context, sessionID uuid.UUID) (*domain.CourseWorkflowSession, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Allow when step is script_gen (completed refinement or retrying failed script gen)
	if session.CurrentStep != domain.StepScriptGen {
		return nil, ErrInvalidStep
	}
	// Don't allow if already processing
	if session.Status == domain.JobStatusProcessing {
		return nil, errors.New("script generation already in progress")
	}

	if len(session.RefinedTopics) == 0 {
		return nil, errors.New("no refined topics available")
	}

	// Update status
	session.Status = domain.JobStatusProcessing
	if err := uc.workflowRepo.UpdateSession(session); err != nil {
		return nil, err
	}

	// Process scripts async
	go uc.processScriptGenerationAsync(session)

	return session, nil
}

func (uc *UseCase) processScriptGenerationAsync(session *domain.CourseWorkflowSession) {
	ctx := context.Background()

	// Convert refined topics to Ollama format
	var topics []ollama.RefinedTopicResult
	for _, t := range session.RefinedTopics {
		var goals []string
		json.Unmarshal(t.LearningGoals, &goals)

		topics = append(topics, ollama.RefinedTopicResult{
			Title:            t.Title,
			Description:      t.Description,
			LearningGoals:    goals,
			EstimatedTimeMin: t.EstimatedTimeMin,
		})
	}

	// Call Script Agency
	scripts, err := uc.ollamaClient.GenerateLessonScripts(ctx, session.MainTopic, topics, session.TargetAudience, session.DifficultyLevel, session.Language)
	if err != nil {
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	// Save scripts
	var domainScripts []domain.LessonScript
	for i, s := range scripts {
		// Find matching topic
		var topicID uuid.UUID
		for _, t := range session.RefinedTopics {
			if t.Title == s.TopicTitle {
				topicID = t.ID
				break
			}
		}
		if topicID == uuid.Nil && i < len(session.RefinedTopics) {
			topicID = session.RefinedTopics[i].ID
		}

		domainScripts = append(domainScripts, domain.LessonScript{
			ID:          uuid.New(),
			SessionID:   session.ID,
			TopicID:     topicID,
			Title:       s.Title,
			Script:      s.Script,
			DurationMin: s.DurationMin,
			SortOrder:   i,
		})
	}

	if err := uc.workflowRepo.CreateLessonScriptsBatch(domainScripts); err != nil {
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	// Move to video generation step
	session.CurrentStep = domain.StepVideoGen
	session.Status = domain.JobStatusCompleted
	uc.workflowRepo.UpdateSession(session)
}

// ProceedToVideoGeneration starts video generation for all scripts
func (uc *UseCase) ProceedToVideoGeneration(ctx context.Context, sessionID uuid.UUID) (*domain.CourseWorkflowSession, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Allow when step is video_gen (completed scripts or retrying failed video gen)
	if session.CurrentStep != domain.StepVideoGen {
		return nil, ErrInvalidStep
	}
	// Don't allow if already processing
	if session.Status == domain.JobStatusProcessing {
		return nil, errors.New("video generation already in progress")
	}

	if len(session.LessonScripts) == 0 {
		return nil, errors.New("no scripts available for video generation")
	}

	if uc.synthesiaClient == nil {
		return nil, errors.New("video generation not configured")
	}

	// Update status
	session.Status = domain.JobStatusProcessing
	if err := uc.workflowRepo.UpdateSession(session); err != nil {
		return nil, err
	}

	// Generate videos async
	go uc.processVideoGenerationAsync(session)

	return session, nil
}

func (uc *UseCase) processVideoGenerationAsync(session *domain.CourseWorkflowSession) {
	ctx := context.Background()

	for _, script := range session.LessonScripts {
		// Skip lessons that are set to presentation output type
		if script.OutputType == domain.OutputTypePresentation {
			log.Printf("[VideoGen] Skipping lesson %s - output type is presentation", script.ID)
			continue
		}

		// Generate video
		log.Printf("[VideoGen] Generating video for lesson %s: %s", script.ID, script.Title)
		videoID, err := uc.synthesiaClient.CreateVideo(ctx, script.Script, script.Title)
		if err != nil {
			log.Printf("[VideoGen] ERROR: Failed to generate video for lesson %s: %v", script.ID, err)
			// Mark this script as failed but continue with others
			uc.workflowRepo.UpdateLessonScriptVideo(script.ID, "", "", "failed")
			continue
		}

		log.Printf("[VideoGen] Video created for lesson %s: videoID=%s", script.ID, videoID)
		// Update script with video ID
		uc.workflowRepo.UpdateLessonScriptVideo(script.ID, videoID, "", "pending")
	}

	// Move to question generation step
	session.CurrentStep = domain.StepQuestionGen
	session.Status = domain.JobStatusCompleted
	uc.workflowRepo.UpdateSession(session)
	log.Printf("[VideoGen] Video/presentation generation complete, moving to question generation for session %s", session.ID)
}

// createCourseFromWorkflow creates a Course from a completed workflow session
func (uc *UseCase) createCourseFromWorkflow(session *domain.CourseWorkflowSession) (*domain.Course, error) {
	if uc.courseRepo == nil {
		return nil, errors.New("course repository not configured")
	}

	// Build description from the refined topics
	description := fmt.Sprintf("Course generated from AI Workflow.\n\nTopics covered:\n")
	for i, topic := range session.RefinedTopics {
		description += fmt.Sprintf("%d. %s\n", i+1, topic.Title)
	}

	// Get video URL from first lesson with video (if any)
	var videoURL *string
	var synthesiaVideoID *string
	var videoStatus *domain.VideoStatus
	for _, lesson := range session.LessonScripts {
		if lesson.OutputType == domain.OutputTypeVideo && lesson.VideoURL != nil {
			videoURL = lesson.VideoURL
			synthesiaVideoID = lesson.VideoID
			if lesson.VideoStatus != nil {
				status := domain.VideoStatus(*lesson.VideoStatus)
				videoStatus = &status
			}
			break
		}
	}

	course := &domain.Course{
		ID:               uuid.New(),
		Title:            session.MainTopic,
		Description:      description,
		VideoURL:         videoURL,
		SynthesiaVideoID: synthesiaVideoID,
		VideoStatus:      videoStatus,
		PassPercentage:   70, // Default pass percentage
		IsPublished:      false, // Not published by default - admin can review and publish
	}

	if err := uc.courseRepo.Create(course); err != nil {
		return nil, fmt.Errorf("failed to create course: %w", err)
	}

	return course, nil
}

// UpdateRefinedTopic updates the content of a refined topic
func (uc *UseCase) UpdateRefinedTopic(sessionID, topicID uuid.UUID, req *domain.UpdateRefinedTopicRequest) (*domain.RefinedTopic, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Allow editing only in script step (after refinement is complete but before scripts are generated)
	if session.CurrentStep != domain.StepScriptGen {
		return nil, ErrInvalidStep
	}

	topic, err := uc.workflowRepo.GetRefinedTopicByID(topicID)
	if err != nil || topic == nil {
		return nil, ErrTopicNotFound
	}

	if topic.SessionID != sessionID {
		return nil, ErrTopicNotFound
	}

	// Update the topic
	goalsJSON, _ := json.Marshal(req.LearningGoals)
	topic.Title = req.Title
	topic.Description = req.Description
	topic.LearningGoals = goalsJSON
	topic.EstimatedTimeMin = req.EstimatedTimeMin

	if err := uc.workflowRepo.UpdateRefinedTopic(topic); err != nil {
		return nil, err
	}

	return topic, nil
}

// RegenerateSingleTopic regenerates a single refined topic using Ollama
func (uc *UseCase) RegenerateSingleTopic(ctx context.Context, sessionID, topicID uuid.UUID) (*domain.RefinedTopic, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	if session.CurrentStep != domain.StepScriptGen {
		return nil, ErrInvalidStep
	}

	topic, err := uc.workflowRepo.GetRefinedTopicByID(topicID)
	if err != nil || topic == nil {
		return nil, ErrTopicNotFound
	}

	if topic.SessionID != sessionID {
		return nil, ErrTopicNotFound
	}

	// Get the original suggestion to use as context
	suggestion, err := uc.workflowRepo.GetSuggestionByID(topic.SuggestionID)
	if err != nil || suggestion == nil {
		return nil, ErrSuggestionNotFound
	}

	// Create a single-item slice for refinement
	singleSuggestion := []ollama.TopicSuggestionResult{{
		Title:       suggestion.Title,
		Description: suggestion.Description,
	}}

	// Regenerate using Ollama
	refined, err := uc.ollamaClient.RefineTopics(ctx, session.MainTopic, singleSuggestion,
		session.TargetAudience, session.DifficultyLevel, session.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to regenerate topic: %w", err)
	}

	if len(refined) == 0 {
		return nil, errors.New("no refined content generated")
	}

	// Update the topic with regenerated content
	goalsJSON, _ := json.Marshal(refined[0].LearningGoals)
	topic.Title = refined[0].Title
	topic.Description = refined[0].Description
	topic.LearningGoals = goalsJSON
	topic.EstimatedTimeMin = refined[0].EstimatedTimeMin

	if err := uc.workflowRepo.UpdateRefinedTopic(topic); err != nil {
		return nil, err
	}

	return topic, nil
}

// ReorderRefinedTopics updates the sort order of refined topics
func (uc *UseCase) ReorderRefinedTopics(sessionID uuid.UUID, orders []domain.TopicOrder) error {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return ErrSessionNotFound
	}

	if session.CurrentStep != domain.StepScriptGen {
		return ErrInvalidStep
	}

	// Convert to map for repository
	orderMap := make(map[uuid.UUID]int)
	for _, o := range orders {
		topicID, err := uuid.Parse(o.TopicID)
		if err != nil {
			return fmt.Errorf("invalid topic ID: %s", o.TopicID)
		}
		orderMap[topicID] = o.SortOrder
	}

	return uc.workflowRepo.UpdateRefinedTopicSortOrders(orderMap)
}

// UpdateLessonScript updates the content of a lesson script
func (uc *UseCase) UpdateLessonScript(sessionID, lessonID uuid.UUID, req *domain.UpdateLessonScriptRequest) (*domain.LessonScript, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Allow editing in video step
	if session.CurrentStep != domain.StepVideoGen && session.CurrentStep != domain.StepCompleted {
		return nil, ErrInvalidStep
	}

	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != sessionID {
		return nil, ErrLessonNotFound
	}

	// Update the lesson script
	if req.Title != "" {
		lesson.Title = req.Title
	}
	lesson.Script = req.Script

	if err := uc.workflowRepo.UpdateLessonScript(lesson); err != nil {
		return nil, fmt.Errorf("failed to update script: %w", err)
	}

	return lesson, nil
}

// RegenerateScript regenerates a single lesson script using AI
func (uc *UseCase) RegenerateScript(ctx context.Context, sessionID, lessonID uuid.UUID) (*domain.LessonScript, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Allow regenerating in video step
	if session.CurrentStep != domain.StepVideoGen && session.CurrentStep != domain.StepCompleted {
		return nil, ErrInvalidStep
	}

	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != sessionID {
		return nil, ErrLessonNotFound
	}

	// Get the refined topic for this lesson
	topic, err := uc.workflowRepo.GetRefinedTopicByID(lesson.TopicID)
	if err != nil || topic == nil {
		return nil, ErrTopicNotFound
	}

	// Parse learning goals
	var goals []string
	json.Unmarshal(topic.LearningGoals, &goals)

	// Create a single-item slice for script generation
	singleTopic := []ollama.RefinedTopicResult{{
		Title:            topic.Title,
		Description:      topic.Description,
		LearningGoals:    goals,
		EstimatedTimeMin: topic.EstimatedTimeMin,
	}}

	// Regenerate using Ollama
	scripts, err := uc.ollamaClient.GenerateLessonScripts(ctx, session.MainTopic, singleTopic,
		session.TargetAudience, session.DifficultyLevel, session.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to regenerate script: %w", err)
	}

	if len(scripts) == 0 {
		return nil, errors.New("no script generated")
	}

	// Update the lesson script
	lesson.Title = scripts[0].Title
	lesson.Script = scripts[0].Script
	lesson.DurationMin = scripts[0].DurationMin

	if err := uc.workflowRepo.UpdateLessonScript(lesson); err != nil {
		return nil, fmt.Errorf("failed to update script: %w", err)
	}

	log.Printf("[RegenerateScript] Successfully regenerated script for lesson %s", lesson.ID)
	return lesson, nil
}

// SetLessonOutputType changes the output type for a lesson (video or presentation)
func (uc *UseCase) SetLessonOutputType(sessionID, lessonID uuid.UUID, outputType domain.OutputType) error {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return ErrSessionNotFound
	}

	// Allow changing output type in script or video step
	if session.CurrentStep != domain.StepScriptGen && session.CurrentStep != domain.StepVideoGen {
		return ErrInvalidStep
	}

	// Verify lesson belongs to this session
	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return ErrLessonNotFound
	}
	if lesson.SessionID != sessionID {
		return ErrLessonNotFound
	}

	return uc.workflowRepo.UpdateLessonScriptOutputType(lessonID, outputType)
}

// GeneratePresentation creates a presentation for a lesson
func (uc *UseCase) GeneratePresentation(ctx context.Context, sessionID, lessonID uuid.UUID) (*domain.LessonPresentation, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Get the lesson script
	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != sessionID {
		return nil, ErrLessonNotFound
	}

	// Set status to processing
	uc.workflowRepo.UpdateLessonScriptPresentationStatus(lessonID, "processing")

	// Create presentation record
	presentation := &domain.LessonPresentation{
		ID:       uuid.New(),
		LessonID: lessonID,
		Status:   "processing",
		Slides:   []domain.PresentationSlide{},
	}

	if err := uc.presentationRepo.Create(presentation); err != nil {
		return nil, fmt.Errorf("failed to create presentation: %w", err)
	}

	// Generate presentation async
	go uc.generatePresentationAsync(presentation.ID, lesson, session.Language)

	return presentation, nil
}

// generatePresentationAsync handles the async presentation generation
func (uc *UseCase) generatePresentationAsync(presentationID uuid.UUID, lesson *domain.LessonScript, language string) {
	ctx := context.Background()

	log.Printf("[Presentation] Starting generation for lesson %s (presentation %s)", lesson.ID, presentationID)

	// Generate slides using Ollama
	log.Printf("[Presentation] Calling Ollama to generate slides for lesson: %s", lesson.Title)
	slideResults, err := uc.ollamaClient.GeneratePresentationSlides(ctx, lesson.Title, lesson.Script, language)
	if err != nil {
		log.Printf("[Presentation] ERROR: Failed to generate slides for lesson %s: %v", lesson.ID, err)
		uc.presentationRepo.UpdateStatus(presentationID, "failed")
		uc.workflowRepo.UpdateLessonScriptPresentationStatus(lesson.ID, "failed")
		return
	}

	log.Printf("[Presentation] Ollama generated %d slides for lesson %s", len(slideResults), lesson.ID)

	// Convert to domain slides and generate audio/images for each
	slides := make([]domain.PresentationSlide, len(slideResults))
	for i, sr := range slideResults {
		slide := domain.PresentationSlide{
			Title:         sr.Title,
			Content:       sr.Content,
			Script:        sr.Script,
			ImageKeywords: sr.ImageKeywords,
		}

		// Fetch stock image from Unsplash if available
		if uc.unsplashClient != nil && uc.unsplashClient.IsAvailable() && sr.ImageKeywords != "" {
			log.Printf("[Presentation] Fetching stock image for slide %d/%d: keywords=%s", i+1, len(slideResults), sr.ImageKeywords)
			photo, err := uc.unsplashClient.GetPhotoForKeywords(ctx, sr.ImageKeywords)
			if err != nil {
				log.Printf("[Presentation] WARNING: Failed to fetch image for slide %d: %v", i, err)
			} else if photo != nil {
				slide.ImageURL = photo.GetImageURL()
				slide.ImageAlt = photo.GetAltText()
				log.Printf("[Presentation] Stock image fetched: %s", slide.ImageURL)
			}
		}

		// Generate TTS audio if client is available
		if uc.ttsClient != nil && uc.ttsClient.IsAvailable() {
			log.Printf("[Presentation] Generating TTS audio for slide %d/%d: %s", i+1, len(slideResults), sr.Title)
			audioURL, err := uc.ttsClient.GenerateAudio(ctx, sr.Script, language)
			if err != nil {
				log.Printf("[Presentation] WARNING: Failed to generate audio for slide %d: %v", i, err)
				// Continue without audio
			} else {
				slide.AudioURL = audioURL
				log.Printf("[Presentation] Audio generated: %s", audioURL)
			}
		} else {
			log.Printf("[Presentation] TTS client not available, skipping audio generation")
		}

		slides[i] = slide
	}

	// Update presentation with slides
	log.Printf("[Presentation] Saving %d slides to database", len(slides))
	if err := uc.presentationRepo.UpdateSlides(presentationID, slides); err != nil {
		log.Printf("[Presentation] ERROR: Failed to update slides for presentation %s: %v", presentationID, err)
		uc.presentationRepo.UpdateStatus(presentationID, "failed")
		uc.workflowRepo.UpdateLessonScriptPresentationStatus(lesson.ID, "failed")
		return
	}

	// Mark as completed
	log.Printf("[Presentation] Marking presentation %s as completed", presentationID)
	uc.presentationRepo.UpdateStatus(presentationID, "completed")
	uc.workflowRepo.UpdateLessonScriptPresentationStatus(lesson.ID, "completed")
	log.Printf("[Presentation] Successfully completed presentation generation for lesson %s", lesson.ID)
}

// GetPresentation retrieves a presentation by lesson ID
func (uc *UseCase) GetPresentation(sessionID, lessonID uuid.UUID) (*domain.LessonPresentation, error) {
	// Verify session exists
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Verify lesson belongs to session
	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != sessionID {
		return nil, ErrLessonNotFound
	}

	return uc.presentationRepo.GetByLessonID(lessonID)
}

// RegenerateAudio regenerates audio files for an existing presentation
func (uc *UseCase) RegenerateAudio(ctx context.Context, sessionID, lessonID uuid.UUID) (*domain.LessonPresentation, error) {
	// Verify session exists
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Verify lesson belongs to session
	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != sessionID {
		return nil, ErrLessonNotFound
	}

	// Get existing presentation
	presentation, err := uc.presentationRepo.GetByLessonID(lessonID)
	if err != nil || presentation == nil {
		return nil, errors.New("presentation not found")
	}

	if uc.ttsClient == nil || !uc.ttsClient.IsAvailable() {
		return nil, errors.New("TTS client not available")
	}

	log.Printf("[RegenerateAudio] Regenerating audio for presentation %s", presentation.ID)

	// Regenerate audio for each slide
	for i := range presentation.Slides {
		slide := &presentation.Slides[i]
		if slide.Script == "" {
			continue
		}

		log.Printf("[RegenerateAudio] Generating audio for slide %d: %s", i+1, slide.Title)
		audioURL, err := uc.ttsClient.GenerateAudio(ctx, slide.Script, session.Language)
		if err != nil {
			log.Printf("[RegenerateAudio] WARNING: Failed to generate audio for slide %d: %v", i, err)
			continue
		}
		slide.AudioURL = audioURL
		log.Printf("[RegenerateAudio] Audio generated: %s", audioURL)
	}

	// Update presentation with new audio URLs
	if err := uc.presentationRepo.UpdateSlides(presentation.ID, presentation.Slides); err != nil {
		return nil, fmt.Errorf("failed to update slides: %w", err)
	}

	log.Printf("[RegenerateAudio] Successfully regenerated audio for presentation %s", presentation.ID)
	return presentation, nil
}

// CourseLessonWithPresentation combines lesson data with presentation
type CourseLessonWithPresentation struct {
	ID                 uuid.UUID                  `json:"id"`
	Title              string                     `json:"title"`
	OutputType         domain.OutputType          `json:"outputType"`
	VideoURL           *string                    `json:"videoUrl,omitempty"`
	VideoStatus        *string                    `json:"videoStatus,omitempty"`
	PresentationStatus *string                    `json:"presentationStatus,omitempty"`
	Presentation       *domain.LessonPresentation `json:"presentation,omitempty"`
}

// GetCourseLessons retrieves all lessons for a course with their presentations
func (uc *UseCase) GetCourseLessons(courseID uuid.UUID) ([]CourseLessonWithPresentation, error) {
	// Find workflow session by course ID
	session, err := uc.workflowRepo.GetSessionByCourseID(courseID)
	if err != nil || session == nil {
		return nil, errors.New("no lessons found for this course")
	}

	// Get all lessons for this session
	lessons := session.LessonScripts
	result := make([]CourseLessonWithPresentation, 0, len(lessons))

	for _, lesson := range lessons {
		item := CourseLessonWithPresentation{
			ID:                 lesson.ID,
			Title:              lesson.Title,
			OutputType:         lesson.OutputType,
			VideoURL:           lesson.VideoURL,
			VideoStatus:        lesson.VideoStatus,
			PresentationStatus: lesson.PresentationStatus,
		}

		// If it's a presentation, get the presentation data
		if lesson.OutputType == domain.OutputTypePresentation && lesson.PresentationStatus != nil && *lesson.PresentationStatus == "completed" {
			presentation, err := uc.presentationRepo.GetByLessonID(lesson.ID)
			if err == nil && presentation != nil {
				item.Presentation = presentation
			}
		}

		result = append(result, item)
	}

	return result, nil
}

// ProceedToQuestionGeneration generates quiz questions for the course
func (uc *UseCase) ProceedToQuestionGeneration(ctx context.Context, sessionID uuid.UUID) (*domain.CourseWorkflowSession, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Allow when step is question_gen (after video/presentation)
	if session.CurrentStep != domain.StepQuestionGen {
		return nil, ErrInvalidStep
	}
	// Don't allow if already processing
	if session.Status == domain.JobStatusProcessing {
		return nil, errors.New("question generation already in progress")
	}

	if len(session.LessonScripts) == 0 {
		return nil, errors.New("no lesson scripts available for question generation")
	}

	// Update status
	session.Status = domain.JobStatusProcessing
	if err := uc.workflowRepo.UpdateSession(session); err != nil {
		return nil, err
	}

	// Process question generation async
	go uc.processQuestionGenerationAsync(session)

	return session, nil
}

func (uc *UseCase) processQuestionGenerationAsync(session *domain.CourseWorkflowSession) {
	ctx := context.Background()

	log.Printf("[QuestionGen] Starting question generation for session %s", session.ID)

	// Collect lesson scripts
	var scripts []string
	for _, lesson := range session.LessonScripts {
		scripts = append(scripts, lesson.Script)
	}

	// Generate questions using Ollama
	questions, err := uc.ollamaClient.GenerateQuizQuestions(ctx, session.MainTopic, scripts, session.Language, 10)
	if err != nil {
		log.Printf("[QuestionGen] ERROR: Failed to generate questions: %v", err)
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	log.Printf("[QuestionGen] Generated %d questions for session %s", len(questions), session.ID)

	// Create the course from the workflow session
	course, err := uc.createCourseFromWorkflow(session)
	if err != nil {
		log.Printf("[QuestionGen] ERROR: Failed to create course from workflow: %v", err)
		session.Status = domain.JobStatusFailed
		uc.workflowRepo.UpdateSession(session)
		return
	}

	log.Printf("[QuestionGen] Course created: %s (%s)", course.Title, course.ID)
	session.CourseID = &course.ID

	// Create test for the course
	if uc.testRepo != nil && uc.questionRepo != nil {
		test := &domain.Test{
			ID:           uuid.New(),
			CourseID:     course.ID,
			Title:        fmt.Sprintf("%s - Assessment", course.Title),
			Description:  "Test your knowledge of the course material",
			PassingScore: 70,
		}

		if err := uc.testRepo.Create(test); err != nil {
			log.Printf("[QuestionGen] ERROR: Failed to create test: %v", err)
		} else {
			log.Printf("[QuestionGen] Test created: %s", test.ID)

			// Create questions for the test
			for i, q := range questions {
				question := &domain.Question{
					ID:           uuid.New(),
					TestID:       test.ID,
					QuestionType: q.QuestionType,
					QuestionText: q.QuestionText,
					QuestionData: q.QuestionData,
					Points:       q.Points,
					OrderIndex:   i,
				}

				if err := uc.questionRepo.Create(question); err != nil {
					log.Printf("[QuestionGen] WARNING: Failed to create question %d: %v", i, err)
				}
			}
			log.Printf("[QuestionGen] Created %d questions for test %s", len(questions), test.ID)
		}
	} else {
		log.Printf("[QuestionGen] WARNING: Test/Question repos not available, skipping test creation")
	}

	// Mark workflow as completed
	session.CurrentStep = domain.StepCompleted
	session.Status = domain.JobStatusCompleted
	uc.workflowRepo.UpdateSession(session)
	log.Printf("[QuestionGen] Workflow completed for session %s", session.ID)
}

// GeneratedQuestionsPreview returns preview of questions that would be generated
type GeneratedQuestionsPreview struct {
	Questions []domain.GeneratedQuestion `json:"questions"`
}

// PreviewQuestions generates questions for preview without saving them
func (uc *UseCase) PreviewQuestions(ctx context.Context, sessionID uuid.UUID) (*GeneratedQuestionsPreview, error) {
	session, err := uc.workflowRepo.GetSessionByID(sessionID)
	if err != nil || session == nil {
		return nil, ErrSessionNotFound
	}

	// Collect lesson scripts
	var scripts []string
	for _, lesson := range session.LessonScripts {
		scripts = append(scripts, lesson.Script)
	}

	// Generate questions using Ollama
	questions, err := uc.ollamaClient.GenerateQuizQuestions(ctx, session.MainTopic, scripts, session.Language, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	return &GeneratedQuestionsPreview{
		Questions: questions,
	}, nil
}

// ======== Course-level lesson management (for completed courses) ========

// UpdateLessonScriptForCourse updates a lesson script for a completed course
func (uc *UseCase) UpdateLessonScriptForCourse(ctx context.Context, courseID, lessonID uuid.UUID, req *domain.UpdateLessonScriptRequest) (*domain.LessonScript, error) {
	// Find workflow session by course ID
	session, err := uc.workflowRepo.GetSessionByCourseID(courseID)
	if err != nil || session == nil {
		return nil, errors.New("no workflow session found for this course")
	}

	// Find and update the lesson
	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != session.ID {
		return nil, ErrLessonNotFound
	}

	// Update the lesson script
	if req.Title != "" {
		lesson.Title = req.Title
	}
	lesson.Script = req.Script

	if err := uc.workflowRepo.UpdateLessonScript(lesson); err != nil {
		return nil, fmt.Errorf("failed to update script: %w", err)
	}

	return lesson, nil
}

// RegenerateLessonScriptForCourse regenerates a lesson script using AI for a completed course
func (uc *UseCase) RegenerateLessonScriptForCourse(ctx context.Context, courseID, lessonID uuid.UUID) (*domain.LessonScript, error) {
	// Find workflow session by course ID
	session, err := uc.workflowRepo.GetSessionByCourseID(courseID)
	if err != nil || session == nil {
		return nil, errors.New("no workflow session found for this course")
	}

	// Get the lesson script
	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != session.ID {
		return nil, ErrLessonNotFound
	}

	// Get the refined topic for this lesson
	topic, err := uc.workflowRepo.GetRefinedTopicByID(lesson.TopicID)
	if err != nil || topic == nil {
		return nil, ErrTopicNotFound
	}

	// Parse learning goals
	var goals []string
	json.Unmarshal(topic.LearningGoals, &goals)

	// Create a single-item slice for script generation
	singleTopic := []ollama.RefinedTopicResult{{
		Title:            topic.Title,
		Description:      topic.Description,
		LearningGoals:    goals,
		EstimatedTimeMin: topic.EstimatedTimeMin,
	}}

	// Regenerate using Ollama
	scripts, err := uc.ollamaClient.GenerateLessonScripts(ctx, session.MainTopic, singleTopic,
		session.TargetAudience, session.DifficultyLevel, session.Language)
	if err != nil {
		return nil, fmt.Errorf("failed to regenerate script: %w", err)
	}

	if len(scripts) == 0 {
		return nil, errors.New("no script generated")
	}

	// Update the lesson script
	lesson.Title = scripts[0].Title
	lesson.Script = scripts[0].Script
	lesson.DurationMin = scripts[0].DurationMin

	if err := uc.workflowRepo.UpdateLessonScript(lesson); err != nil {
		return nil, fmt.Errorf("failed to update script: %w", err)
	}

	log.Printf("[RegenerateLessonForCourse] Successfully regenerated script for lesson %s", lesson.ID)
	return lesson, nil
}

// RegeneratePresentationForCourse regenerates a presentation for a lesson in a completed course
func (uc *UseCase) RegeneratePresentationForCourse(ctx context.Context, courseID, lessonID uuid.UUID) (*domain.LessonPresentation, error) {
	// Find workflow session by course ID
	session, err := uc.workflowRepo.GetSessionByCourseID(courseID)
	if err != nil || session == nil {
		return nil, errors.New("no workflow session found for this course")
	}

	// Get the lesson script
	lesson, err := uc.workflowRepo.GetLessonScriptByID(lessonID)
	if err != nil || lesson == nil {
		return nil, ErrLessonNotFound
	}
	if lesson.SessionID != session.ID {
		return nil, ErrLessonNotFound
	}

	// Check if lesson is set to presentation type
	if lesson.OutputType != domain.OutputTypePresentation {
		return nil, errors.New("lesson is not set to presentation output type")
	}

	// Set status to processing
	uc.workflowRepo.UpdateLessonScriptPresentationStatus(lessonID, "processing")

	// Get or create presentation record
	presentation, _ := uc.presentationRepo.GetByLessonID(lessonID)
	if presentation == nil {
		presentation = &domain.LessonPresentation{
			ID:       uuid.New(),
			LessonID: lessonID,
			Status:   "processing",
			Slides:   []domain.PresentationSlide{},
		}
		if err := uc.presentationRepo.Create(presentation); err != nil {
			return nil, fmt.Errorf("failed to create presentation: %w", err)
		}
	} else {
		uc.presentationRepo.UpdateStatus(presentation.ID, "processing")
	}

	// Generate presentation async
	go uc.generatePresentationAsync(presentation.ID, lesson, session.Language)

	return presentation, nil
}
