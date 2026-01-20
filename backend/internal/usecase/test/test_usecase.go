package test

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
)

var (
	ErrTestNotFound       = errors.New("test not found")
	ErrAttemptNotFound    = errors.New("attempt not found")
	ErrAttemptCompleted   = errors.New("attempt already completed")
	ErrVideoNotWatched    = errors.New("video must be watched before taking the test")
	ErrNotEnrolled        = errors.New("not enrolled in this course")
)

type UseCase struct {
	testRepo       domain.TestRepository
	questionRepo   domain.QuestionRepository
	attemptRepo    domain.TestAttemptRepository
	answerRepo     domain.UserAnswerRepository
	enrollmentRepo domain.EnrollmentRepository
	courseRepo     domain.CourseRepository
}

func NewUseCase(
	testRepo domain.TestRepository,
	questionRepo domain.QuestionRepository,
	attemptRepo domain.TestAttemptRepository,
	answerRepo domain.UserAnswerRepository,
	enrollmentRepo domain.EnrollmentRepository,
	courseRepo domain.CourseRepository,
) *UseCase {
	return &UseCase{
		testRepo:       testRepo,
		questionRepo:   questionRepo,
		attemptRepo:    attemptRepo,
		answerRepo:     answerRepo,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
	}
}

func (uc *UseCase) GetByCourseID(courseID uuid.UUID) (*domain.Test, error) {
	test, err := uc.testRepo.GetByCourseID(courseID)
	if err != nil {
		return nil, err
	}
	if test == nil {
		return nil, ErrTestNotFound
	}

	// Load questions
	questions, err := uc.questionRepo.GetByTestID(test.ID)
	if err != nil {
		return nil, err
	}
	test.Questions = questions

	return test, nil
}

func (uc *UseCase) CreateTest(req *domain.CreateTestRequest) (*domain.Test, error) {
	test := &domain.Test{
		ID:               uuid.New(),
		CourseID:         req.CourseID,
		Title:            req.Title,
		Description:      req.Description,
		TimeLimitMinutes: req.TimeLimitMinutes,
		PassingScore:     req.PassingScore,
	}

	if err := uc.testRepo.Create(test); err != nil {
		return nil, err
	}

	return test, nil
}

func (uc *UseCase) CreateQuestion(req *domain.CreateQuestionRequest) (*domain.Question, error) {
	question := &domain.Question{
		ID:           uuid.New(),
		TestID:       req.TestID,
		QuestionType: req.QuestionType,
		QuestionText: req.QuestionText,
		QuestionData: req.QuestionData,
		Points:       req.Points,
		OrderIndex:   req.OrderIndex,
	}

	if err := uc.questionRepo.Create(question); err != nil {
		return nil, err
	}

	return question, nil
}

func (uc *UseCase) StartAttempt(userID, testID uuid.UUID) (*domain.TestAttempt, error) {
	test, err := uc.testRepo.GetByID(testID)
	if err != nil {
		return nil, err
	}
	if test == nil {
		return nil, ErrTestNotFound
	}

	// Check if user is enrolled
	enrollment, err := uc.enrollmentRepo.GetByUserAndCourse(userID, test.CourseID)
	if err != nil {
		return nil, err
	}
	if enrollment == nil {
		return nil, ErrNotEnrolled
	}

	// Only require video to be watched if the course has a video
	course, err := uc.courseRepo.GetByID(test.CourseID)
	if err != nil {
		return nil, err
	}
	if course != nil && course.VideoURL != nil && *course.VideoURL != "" {
		if !enrollment.VideoWatched {
			return nil, ErrVideoNotWatched
		}
	}

	attempt := &domain.TestAttempt{
		ID:     uuid.New(),
		UserID: userID,
		TestID: testID,
	}

	if err := uc.attemptRepo.Create(attempt); err != nil {
		return nil, err
	}

	return attempt, nil
}

func (uc *UseCase) SubmitAttempt(attemptID uuid.UUID, userID uuid.UUID, submission *domain.SubmitTestRequest) (*domain.TestResult, error) {
	attempt, err := uc.attemptRepo.GetByID(attemptID)
	if err != nil {
		return nil, err
	}
	if attempt == nil {
		return nil, ErrAttemptNotFound
	}
	if attempt.UserID != userID {
		return nil, ErrAttemptNotFound
	}
	if attempt.CompletedAt != nil {
		return nil, ErrAttemptCompleted
	}

	// Get test and questions
	test, err := uc.testRepo.GetByID(attempt.TestID)
	if err != nil {
		return nil, err
	}

	questions, err := uc.questionRepo.GetByTestID(test.ID)
	if err != nil {
		return nil, err
	}

	// Build question map
	questionMap := make(map[uuid.UUID]*domain.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	// Grade answers
	var totalScore, maxScore int
	var answers []*domain.UserAnswer
	var answerResults []*domain.AnswerResult

	for _, sub := range submission.Answers {
		question, exists := questionMap[sub.QuestionID]
		if !exists {
			continue
		}

		isCorrect, points, explanation := uc.gradeAnswer(question, sub.AnswerData)
		maxScore += question.Points
		totalScore += points

		answer := &domain.UserAnswer{
			ID:            uuid.New(),
			AttemptID:     attemptID,
			QuestionID:    sub.QuestionID,
			AnswerData:    sub.AnswerData,
			IsCorrect:     isCorrect,
			PointsAwarded: points,
		}
		answers = append(answers, answer)

		answerResults = append(answerResults, &domain.AnswerResult{
			QuestionID:    sub.QuestionID,
			IsCorrect:     isCorrect,
			PointsAwarded: points,
			MaxPoints:     question.Points,
			Explanation:   explanation,
		})
	}

	// Save answers
	if err := uc.answerRepo.CreateBatch(answers); err != nil {
		return nil, err
	}

	// Calculate percentage
	var percentage float64
	if maxScore > 0 {
		percentage = float64(totalScore) / float64(maxScore) * 100
	}
	passed := percentage >= float64(test.PassingScore)

	// Update attempt
	now := time.Now()
	attempt.CompletedAt = &now
	attempt.Score = &totalScore
	attempt.MaxScore = &maxScore
	attempt.Percentage = &percentage
	attempt.Passed = &passed

	if err := uc.attemptRepo.Update(attempt); err != nil {
		return nil, err
	}

	return &domain.TestResult{
		AttemptID:  attemptID,
		Score:      totalScore,
		MaxScore:   maxScore,
		Percentage: percentage,
		Passed:     passed,
		Answers:    answerResults,
	}, nil
}

func (uc *UseCase) gradeAnswer(question *domain.Question, answerData json.RawMessage) (bool, int, string) {
	switch question.QuestionType {
	case domain.QuestionTypeMultipleChoice:
		return uc.gradeMultipleChoice(question, answerData)
	case domain.QuestionTypeDragDrop:
		return uc.gradeDragDrop(question, answerData)
	case domain.QuestionTypeFillBlank:
		return uc.gradeFillBlank(question, answerData)
	case domain.QuestionTypeMatching:
		return uc.gradeMatching(question, answerData)
	case domain.QuestionTypeOrdering:
		return uc.gradeOrdering(question, answerData)
	default:
		return false, 0, ""
	}
}

func (uc *UseCase) gradeMultipleChoice(question *domain.Question, answerData json.RawMessage) (bool, int, string) {
	var data domain.MultipleChoiceData
	if err := json.Unmarshal(question.QuestionData, &data); err != nil {
		return false, 0, ""
	}

	var selected []int
	if err := json.Unmarshal(answerData, &selected); err != nil {
		return false, 0, data.Explanation
	}

	if len(selected) != len(data.CorrectIndices) {
		return false, 0, data.Explanation
	}

	correctSet := make(map[int]bool)
	for _, idx := range data.CorrectIndices {
		correctSet[idx] = true
	}

	for _, idx := range selected {
		if !correctSet[idx] {
			return false, 0, data.Explanation
		}
	}

	return true, question.Points, data.Explanation
}

func (uc *UseCase) gradeDragDrop(question *domain.Question, answerData json.RawMessage) (bool, int, string) {
	var data domain.DragDropData
	if err := json.Unmarshal(question.QuestionData, &data); err != nil {
		return false, 0, ""
	}

	var mapping map[string]string
	if err := json.Unmarshal(answerData, &mapping); err != nil {
		return false, 0, data.Explanation
	}

	correct := 0
	for item, zone := range data.CorrectMapping {
		if mapping[item] == zone {
			correct++
		}
	}

	if correct == len(data.CorrectMapping) {
		return true, question.Points, data.Explanation
	}

	// Partial credit
	partialPoints := (question.Points * correct) / len(data.CorrectMapping)
	return false, partialPoints, data.Explanation
}

func (uc *UseCase) gradeFillBlank(question *domain.Question, answerData json.RawMessage) (bool, int, string) {
	var data domain.FillBlankData
	if err := json.Unmarshal(question.QuestionData, &data); err != nil {
		return false, 0, ""
	}

	var answers []string
	if err := json.Unmarshal(answerData, &answers); err != nil {
		return false, 0, data.Explanation
	}

	if len(answers) != len(data.Blanks) {
		return false, 0, data.Explanation
	}

	correct := 0
	for i, answer := range answers {
		if answer == data.Blanks[i] {
			correct++
		}
	}

	if correct == len(data.Blanks) {
		return true, question.Points, data.Explanation
	}

	partialPoints := (question.Points * correct) / len(data.Blanks)
	return false, partialPoints, data.Explanation
}

func (uc *UseCase) gradeMatching(question *domain.Question, answerData json.RawMessage) (bool, int, string) {
	var data domain.MatchingData
	if err := json.Unmarshal(question.QuestionData, &data); err != nil {
		return false, 0, ""
	}

	var pairs map[string]string
	if err := json.Unmarshal(answerData, &pairs); err != nil {
		return false, 0, data.Explanation
	}

	correct := 0
	for left, right := range data.CorrectPairs {
		if pairs[left] == right {
			correct++
		}
	}

	if correct == len(data.CorrectPairs) {
		return true, question.Points, data.Explanation
	}

	partialPoints := (question.Points * correct) / len(data.CorrectPairs)
	return false, partialPoints, data.Explanation
}

func (uc *UseCase) gradeOrdering(question *domain.Question, answerData json.RawMessage) (bool, int, string) {
	var data domain.OrderingData
	if err := json.Unmarshal(question.QuestionData, &data); err != nil {
		return false, 0, ""
	}

	var order []int
	if err := json.Unmarshal(answerData, &order); err != nil {
		return false, 0, data.Explanation
	}

	if len(order) != len(data.CorrectOrder) {
		return false, 0, data.Explanation
	}

	correct := 0
	for i, pos := range order {
		if pos == data.CorrectOrder[i] {
			correct++
		}
	}

	if correct == len(data.CorrectOrder) {
		return true, question.Points, data.Explanation
	}

	partialPoints := (question.Points * correct) / len(data.CorrectOrder)
	return false, partialPoints, data.Explanation
}

func (uc *UseCase) GetAttemptResults(attemptID uuid.UUID, userID uuid.UUID) (*domain.TestResult, error) {
	attempt, err := uc.attemptRepo.GetByID(attemptID)
	if err != nil {
		return nil, err
	}
	if attempt == nil || attempt.UserID != userID {
		return nil, ErrAttemptNotFound
	}
	if attempt.CompletedAt == nil {
		return nil, errors.New("attempt not completed")
	}

	answers, err := uc.answerRepo.GetByAttemptID(attemptID)
	if err != nil {
		return nil, err
	}

	// Get questions for explanations
	questions, err := uc.questionRepo.GetByTestID(attempt.TestID)
	if err != nil {
		return nil, err
	}

	questionMap := make(map[uuid.UUID]*domain.Question)
	for _, q := range questions {
		questionMap[q.ID] = q
	}

	var answerResults []*domain.AnswerResult
	for _, a := range answers {
		q := questionMap[a.QuestionID]
		explanation := uc.getExplanation(q)
		answerResults = append(answerResults, &domain.AnswerResult{
			QuestionID:    a.QuestionID,
			IsCorrect:     a.IsCorrect,
			PointsAwarded: a.PointsAwarded,
			MaxPoints:     q.Points,
			Explanation:   explanation,
		})
	}

	return &domain.TestResult{
		AttemptID:  attemptID,
		Score:      *attempt.Score,
		MaxScore:   *attempt.MaxScore,
		Percentage: *attempt.Percentage,
		Passed:     *attempt.Passed,
		Answers:    answerResults,
	}, nil
}

func (uc *UseCase) getExplanation(question *domain.Question) string {
	switch question.QuestionType {
	case domain.QuestionTypeMultipleChoice:
		var data domain.MultipleChoiceData
		json.Unmarshal(question.QuestionData, &data)
		return data.Explanation
	case domain.QuestionTypeDragDrop:
		var data domain.DragDropData
		json.Unmarshal(question.QuestionData, &data)
		return data.Explanation
	case domain.QuestionTypeFillBlank:
		var data domain.FillBlankData
		json.Unmarshal(question.QuestionData, &data)
		return data.Explanation
	case domain.QuestionTypeMatching:
		var data domain.MatchingData
		json.Unmarshal(question.QuestionData, &data)
		return data.Explanation
	case domain.QuestionTypeOrdering:
		var data domain.OrderingData
		json.Unmarshal(question.QuestionData, &data)
		return data.Explanation
	default:
		return ""
	}
}

// UpdateQuestion updates an existing question
func (uc *UseCase) UpdateQuestion(questionID uuid.UUID, req *domain.UpdateQuestionRequest) (*domain.Question, error) {
	question, err := uc.questionRepo.GetByID(questionID)
	if err != nil {
		return nil, err
	}
	if question == nil {
		return nil, errors.New("question not found")
	}

	question.QuestionType = req.QuestionType
	question.QuestionText = req.QuestionText
	question.QuestionData = req.QuestionData
	question.Points = req.Points

	if err := uc.questionRepo.Update(question); err != nil {
		return nil, err
	}

	return question, nil
}

// DeleteQuestion deletes a question
func (uc *UseCase) DeleteQuestion(questionID uuid.UUID) error {
	question, err := uc.questionRepo.GetByID(questionID)
	if err != nil {
		return err
	}
	if question == nil {
		return errors.New("question not found")
	}

	return uc.questionRepo.Delete(questionID)
}

// GetQuestionsByTestID retrieves all questions for a test
func (uc *UseCase) GetQuestionsByTestID(testID uuid.UUID) ([]*domain.Question, error) {
	return uc.questionRepo.GetByTestID(testID)
}

// GetTestByCourseIDWithQuestions retrieves a test with all questions for editing
func (uc *UseCase) GetTestByCourseIDWithQuestions(courseID uuid.UUID) (*domain.Test, error) {
	test, err := uc.testRepo.GetByCourseID(courseID)
	if err != nil {
		return nil, err
	}
	if test == nil {
		return nil, ErrTestNotFound
	}

	questions, err := uc.questionRepo.GetByTestID(test.ID)
	if err != nil {
		return nil, err
	}
	test.Questions = questions

	return test, nil
}
