package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
)

var (
	ErrJobNotFound = errors.New("job not found")
	ErrJobFailed   = errors.New("job failed")
)

type OllamaClient interface {
	GenerateCourseContent(ctx context.Context, req *domain.GenerateCourseRequest) (*domain.GeneratedCourseContent, error)
}

type SynthesiaClient interface {
	CreateVideo(ctx context.Context, script string, title string) (string, error)
	GetVideoStatus(ctx context.Context, videoID string) (string, string, error)
}

type UseCase struct {
	jobRepo        domain.AIGenerationJobRepository
	courseRepo     domain.CourseRepository
	contentRepo    domain.CourseContentRepository
	testRepo       domain.TestRepository
	questionRepo   domain.QuestionRepository
	ollamaClient   OllamaClient
	synthesiaClient SynthesiaClient
}

func NewUseCase(
	jobRepo domain.AIGenerationJobRepository,
	courseRepo domain.CourseRepository,
	contentRepo domain.CourseContentRepository,
	testRepo domain.TestRepository,
	questionRepo domain.QuestionRepository,
	ollamaClient OllamaClient,
	synthesiaClient SynthesiaClient,
) *UseCase {
	return &UseCase{
		jobRepo:         jobRepo,
		courseRepo:      courseRepo,
		contentRepo:     contentRepo,
		testRepo:        testRepo,
		questionRepo:    questionRepo,
		ollamaClient:    ollamaClient,
		synthesiaClient: synthesiaClient,
	}
}

func (uc *UseCase) GenerateCourse(ctx context.Context, req *domain.GenerateCourseRequest) (*domain.AIGenerationJob, error) {
	inputData, _ := json.Marshal(req)

	job := &domain.AIGenerationJob{
		ID:        uuid.New(),
		JobType:   domain.JobTypeContent,
		Status:    domain.JobStatusPending,
		InputData: inputData,
	}

	if err := uc.jobRepo.Create(job); err != nil {
		return nil, err
	}

	// Process job asynchronously
	go uc.processContentGeneration(context.Background(), job.ID, req)

	return job, nil
}

func (uc *UseCase) processContentGeneration(ctx context.Context, jobID uuid.UUID, req *domain.GenerateCourseRequest) {
	job, _ := uc.jobRepo.GetByID(jobID)
	if job == nil {
		return
	}

	// Update status to processing
	job.Status = domain.JobStatusProcessing
	uc.jobRepo.Update(job)

	// Generate content with Ollama
	content, err := uc.ollamaClient.GenerateCourseContent(ctx, req)
	if err != nil {
		errStr := err.Error()
		job.Status = domain.JobStatusFailed
		job.Error = &errStr
		uc.jobRepo.Update(job)
		return
	}

	// Create course
	course := &domain.Course{
		ID:             uuid.New(),
		Title:          content.Title,
		Description:    content.Description,
		PassPercentage: 70,
		IsPublished:    false,
	}
	if err := uc.courseRepo.Create(course); err != nil {
		errStr := err.Error()
		job.Status = domain.JobStatusFailed
		job.Error = &errStr
		uc.jobRepo.Update(job)
		return
	}

	// Save course content
	outlineJSON, _ := json.Marshal(content.Outline)
	courseContent := &domain.CourseContent{
		ID:               uuid.New(),
		CourseID:         course.ID,
		Outline:          outlineJSON,
		VideoScript:      content.VideoScript,
		GenerationPrompt: req.Topic,
	}
	if err := uc.contentRepo.Create(courseContent); err != nil {
		errStr := err.Error()
		job.Status = domain.JobStatusFailed
		job.Error = &errStr
		uc.jobRepo.Update(job)
		return
	}

	// Create test
	test := &domain.Test{
		ID:           uuid.New(),
		CourseID:     course.ID,
		Title:        content.Title + " Assessment",
		Description:  "Test your knowledge of " + content.Title,
		PassingScore: 70,
	}
	if err := uc.testRepo.Create(test); err != nil {
		errStr := err.Error()
		job.Status = domain.JobStatusFailed
		job.Error = &errStr
		uc.jobRepo.Update(job)
		return
	}

	// Create questions
	questionsCreated := 0
	for i, q := range content.Questions {
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
			// Log error but continue with other questions
			errStr := fmt.Sprintf("failed to create question %d: %v", i, err)
			println(errStr)
			continue
		}
		questionsCreated++
	}
	println(fmt.Sprintf("Created %d/%d questions for test %s", questionsCreated, len(content.Questions), test.ID))

	// Update job with course ID
	job.CourseID = &course.ID

	// Request video generation if synthesia client is available
	if uc.synthesiaClient != nil && content.VideoScript != "" {
		videoID, err := uc.synthesiaClient.CreateVideo(ctx, content.VideoScript, content.Title)
		if err == nil {
			course.SynthesiaVideoID = &videoID
			status := domain.VideoStatusPending
			course.VideoStatus = &status
			println(fmt.Sprintf("Started video generation for course %s, Synthesia ID: %s", course.ID, videoID))
		} else {
			status := domain.VideoStatusFailed
			errStr := err.Error()
			course.VideoStatus = &status
			course.VideoError = &errStr
			println(fmt.Sprintf("Failed to start video generation: %v", err))
		}
		uc.courseRepo.Update(course)
	}

	// Complete job
	outputData, _ := json.Marshal(map[string]interface{}{
		"courseId": course.ID,
		"testId":   test.ID,
	})
	job.Status = domain.JobStatusCompleted
	job.OutputData = outputData
	now := time.Now()
	job.CompletedAt = &now
	uc.jobRepo.Update(job)
}

func (uc *UseCase) GetJob(id uuid.UUID) (*domain.AIGenerationJob, error) {
	job, err := uc.jobRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, ErrJobNotFound
	}
	return job, nil
}

func (uc *UseCase) HandleSynthesiaWebhook(payload *domain.SynthesiaWebhookPayload) error {
	// Find course with this Synthesia video ID
	course, err := uc.courseRepo.GetBySynthesiaVideoID(payload.VideoID)
	if err != nil {
		return err
	}
	if course == nil {
		return fmt.Errorf("course not found for video ID: %s", payload.VideoID)
	}

	// Update video status based on webhook payload
	switch payload.Status {
	case "complete":
		status := domain.VideoStatusCompleted
		course.VideoStatus = &status
		course.VideoURL = &payload.URL
		course.VideoError = nil
		println(fmt.Sprintf("Video completed for course %s: %s", course.ID, payload.URL))
	case "failed":
		status := domain.VideoStatusFailed
		course.VideoStatus = &status
		errStr := "Video generation failed"
		course.VideoError = &errStr
		println(fmt.Sprintf("Video failed for course %s", course.ID))
	case "in_progress":
		status := domain.VideoStatusProcessing
		course.VideoStatus = &status
	}

	return uc.courseRepo.Update(course)
}

// RefreshVideoStatus checks the status of a video with Synthesia and updates the course
func (uc *UseCase) RefreshVideoStatus(ctx context.Context, courseID uuid.UUID) (*domain.Course, error) {
	course, err := uc.courseRepo.GetByID(courseID)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, fmt.Errorf("course not found")
	}
	if course.SynthesiaVideoID == nil || *course.SynthesiaVideoID == "" {
		return nil, fmt.Errorf("no video generation in progress for this course")
	}

	// Skip if already completed or no synthesia client
	if course.VideoStatus != nil && *course.VideoStatus == domain.VideoStatusCompleted {
		return course, nil
	}
	if uc.synthesiaClient == nil {
		return nil, fmt.Errorf("synthesia client not configured")
	}

	// Check video status with Synthesia
	status, downloadURL, err := uc.synthesiaClient.GetVideoStatus(ctx, *course.SynthesiaVideoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video status: %w", err)
	}

	// Update course based on status
	switch status {
	case "complete":
		videoStatus := domain.VideoStatusCompleted
		course.VideoStatus = &videoStatus
		course.VideoURL = &downloadURL
		course.VideoError = nil
		println(fmt.Sprintf("Video completed for course %s: %s", course.ID, downloadURL))
	case "in_progress":
		videoStatus := domain.VideoStatusProcessing
		course.VideoStatus = &videoStatus
	case "failed":
		videoStatus := domain.VideoStatusFailed
		course.VideoStatus = &videoStatus
		errStr := "Video generation failed"
		course.VideoError = &errStr
	}

	if err := uc.courseRepo.Update(course); err != nil {
		return nil, err
	}

	return course, nil
}

// PollPendingVideos checks status of all pending video generations
func (uc *UseCase) PollPendingVideos(ctx context.Context) error {
	if uc.synthesiaClient == nil {
		return nil
	}

	// Get courses with pending or processing video status
	pendingCourses, err := uc.courseRepo.GetByVideoStatus(domain.VideoStatusPending)
	if err != nil {
		return err
	}
	processingCourses, err := uc.courseRepo.GetByVideoStatus(domain.VideoStatusProcessing)
	if err != nil {
		return err
	}

	allCourses := append(pendingCourses, processingCourses...)

	for _, course := range allCourses {
		if course.SynthesiaVideoID == nil {
			continue
		}

		_, err := uc.RefreshVideoStatus(ctx, course.ID)
		if err != nil {
			println(fmt.Sprintf("Failed to refresh video status for course %s: %v", course.ID, err))
		}
	}

	return nil
}

// GetFreshVideoURL fetches a fresh signed URL from Synthesia for a course's video
func (uc *UseCase) GetFreshVideoURL(ctx context.Context, courseID uuid.UUID) (string, error) {
	course, err := uc.courseRepo.GetByID(courseID)
	if err != nil {
		return "", err
	}
	if course == nil {
		return "", fmt.Errorf("course not found")
	}
	if course.SynthesiaVideoID == nil || *course.SynthesiaVideoID == "" {
		return "", fmt.Errorf("no video available for this course")
	}
	if uc.synthesiaClient == nil {
		return "", fmt.Errorf("synthesia client not configured")
	}

	// Get fresh URL from Synthesia
	status, downloadURL, err := uc.synthesiaClient.GetVideoStatus(ctx, *course.SynthesiaVideoID)
	if err != nil {
		return "", fmt.Errorf("failed to get video URL: %w", err)
	}

	if status != "complete" {
		return "", fmt.Errorf("video is not ready yet (status: %s)", status)
	}

	// Update the stored URL
	course.VideoURL = &downloadURL
	if err := uc.courseRepo.Update(course); err != nil {
		// Non-fatal, just log
		println(fmt.Sprintf("Warning: Failed to update video URL in database: %v", err))
	}

	return downloadURL, nil
}
