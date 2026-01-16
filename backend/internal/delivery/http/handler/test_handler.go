package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/secusense/backend/internal/delivery/http/middleware"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/internal/usecase/test"
)

type TestHandler struct {
	testUC   *test.UseCase
	validate *validator.Validate
}

func NewTestHandler(testUC *test.UseCase) *TestHandler {
	return &TestHandler{
		testUC:   testUC,
		validate: validator.New(),
	}
}

func (h *TestHandler) GetByCourseID(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "courseId")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	testObj, err := h.testUC.GetByCourseID(courseID)
	if err != nil {
		switch err {
		case test.ErrTestNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get test")
		}
		return
	}

	// Strip correct answers from question data for the student
	for _, q := range testObj.Questions {
		q.QuestionData = stripCorrectAnswers(q.QuestionType, q.QuestionData)
	}

	respondJSON(w, http.StatusOK, testObj)
}

func (h *TestHandler) StartAttempt(w http.ResponseWriter, r *http.Request) {
	testIDStr := chi.URLParam(r, "testId")
	testID, err := uuid.Parse(testIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid test ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	attempt, err := h.testUC.StartAttempt(userID, testID)
	if err != nil {
		switch err {
		case test.ErrTestNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case test.ErrNotEnrolled:
			respondError(w, http.StatusForbidden, err.Error())
		case test.ErrVideoNotWatched:
			respondError(w, http.StatusForbidden, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to start attempt")
		}
		return
	}

	respondJSON(w, http.StatusCreated, attempt)
}

func (h *TestHandler) SubmitAttempt(w http.ResponseWriter, r *http.Request) {
	attemptIDStr := chi.URLParam(r, "attemptId")
	attemptID, err := uuid.Parse(attemptIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid attempt ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req domain.SubmitTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.testUC.SubmitAttempt(attemptID, userID, &req)
	if err != nil {
		switch err {
		case test.ErrAttemptNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case test.ErrAttemptCompleted:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to submit attempt")
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (h *TestHandler) GetAttemptResults(w http.ResponseWriter, r *http.Request) {
	attemptIDStr := chi.URLParam(r, "attemptId")
	attemptID, err := uuid.Parse(attemptIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid attempt ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	result, err := h.testUC.GetAttemptResults(attemptID, userID)
	if err != nil {
		switch err {
		case test.ErrAttemptNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get results")
		}
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (h *TestHandler) CreateTest(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	testObj, err := h.testUC.CreateTest(&req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create test")
		return
	}

	respondJSON(w, http.StatusCreated, testObj)
}

func (h *TestHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	question, err := h.testUC.CreateQuestion(&req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create question")
		return
	}

	respondJSON(w, http.StatusCreated, question)
}

func stripCorrectAnswers(qType domain.QuestionType, data json.RawMessage) json.RawMessage {
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Remove answer keys based on question type
	switch qType {
	case domain.QuestionTypeMultipleChoice:
		delete(result, "correctIndices")
		delete(result, "explanation")
	case domain.QuestionTypeDragDrop:
		delete(result, "correctMapping")
		delete(result, "explanation")
	case domain.QuestionTypeFillBlank:
		delete(result, "blanks")
		delete(result, "explanation")
	case domain.QuestionTypeMatching:
		delete(result, "correctPairs")
		delete(result, "explanation")
	case domain.QuestionTypeOrdering:
		delete(result, "correctOrder")
		delete(result, "explanation")
	}

	stripped, _ := json.Marshal(result)
	return stripped
}
