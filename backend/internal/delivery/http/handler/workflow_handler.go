package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/internal/usecase/workflow"
)

type WorkflowHandler struct {
	workflowUC *workflow.UseCase
	validate   *validator.Validate
}

func NewWorkflowHandler(workflowUC *workflow.UseCase) *WorkflowHandler {
	return &WorkflowHandler{
		workflowUC: workflowUC,
		validate:   validator.New(),
	}
}

// StartResearch initiates a new course workflow with the Research Agency
func (h *WorkflowHandler) StartResearch(w http.ResponseWriter, r *http.Request) {
	var req domain.StartResearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	session, err := h.workflowUC.StartResearch(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to start research: "+err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, session)
}

// GetSession retrieves the current state of a workflow session
func (h *WorkflowHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	session, err := h.workflowUC.GetSession(id)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get session")
		}
		return
	}

	respondJSON(w, http.StatusOK, session)
}

// UpdateSuggestionStatus approves or rejects a topic suggestion
func (h *WorkflowHandler) UpdateSuggestionStatus(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	suggestionID, err := uuid.Parse(chi.URLParam(r, "suggestionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid suggestion ID")
		return
	}

	var req domain.UpdateSuggestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.workflowUC.UpdateSuggestionStatus(sessionID, suggestionID, req.Status); err != nil {
		switch err {
		case workflow.ErrSessionNotFound, workflow.ErrSuggestionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to update suggestion")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// AddCustomTopic adds a user-defined topic to the session
func (h *WorkflowHandler) AddCustomTopic(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	var req domain.AddCustomTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	suggestion, err := h.workflowUC.AddCustomTopic(sessionID, &req)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to add custom topic")
		}
		return
	}

	respondJSON(w, http.StatusCreated, suggestion)
}

// GenerateMoreSuggestions generates additional topic suggestions
func (h *WorkflowHandler) GenerateMoreSuggestions(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	if err := h.workflowUC.GenerateMoreSuggestions(r.Context(), sessionID); err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to generate more suggestions: "+err.Error())
		}
		return
	}

	// Return updated session
	session, _ := h.workflowUC.GetSession(sessionID)
	respondJSON(w, http.StatusOK, session)
}

// ProceedToRefinement moves the workflow to the refinement step
func (h *WorkflowHandler) ProceedToRefinement(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	session, err := h.workflowUC.ProceedToRefinement(r.Context(), sessionID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep, workflow.ErrNoApprovedTopics:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to proceed: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusAccepted, session)
}

// ProceedToScriptGeneration moves the workflow to script generation
func (h *WorkflowHandler) ProceedToScriptGeneration(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	session, err := h.workflowUC.ProceedToScriptGeneration(r.Context(), sessionID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to proceed: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusAccepted, session)
}

// ProceedToVideoGeneration starts video generation for all scripts
func (h *WorkflowHandler) ProceedToVideoGeneration(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	session, err := h.workflowUC.ProceedToVideoGeneration(r.Context(), sessionID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to proceed: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusAccepted, session)
}

// UpdateRefinedTopic updates a refined topic's content
func (h *WorkflowHandler) UpdateRefinedTopic(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	topicID, err := uuid.Parse(chi.URLParam(r, "topicId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid topic ID")
		return
	}

	var req domain.UpdateRefinedTopicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	topic, err := h.workflowUC.UpdateRefinedTopic(sessionID, topicID, &req)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound, workflow.ErrTopicNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to update topic")
		}
		return
	}

	respondJSON(w, http.StatusOK, topic)
}

// RegenerateTopic regenerates a single refined topic
func (h *WorkflowHandler) RegenerateTopic(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	topicID, err := uuid.Parse(chi.URLParam(r, "topicId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid topic ID")
		return
	}

	topic, err := h.workflowUC.RegenerateSingleTopic(r.Context(), sessionID, topicID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound, workflow.ErrTopicNotFound, workflow.ErrSuggestionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to regenerate topic: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, topic)
}

// ReorderRefinedTopics updates the sort order of refined topics
func (h *WorkflowHandler) ReorderRefinedTopics(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	var req domain.ReorderRefinedTopicsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.workflowUC.ReorderRefinedTopics(sessionID, req.TopicOrders); err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to reorder topics")
		}
		return
	}

	// Return updated session
	session, _ := h.workflowUC.GetSession(sessionID)
	respondJSON(w, http.StatusOK, session)
}

// SetOutputType sets the output type for a lesson (video or presentation)
func (h *WorkflowHandler) SetOutputType(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	var req domain.SetOutputTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.workflowUC.SetLessonOutputType(sessionID, lessonID, req.OutputType); err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "session not found")
		case workflow.ErrLessonNotFound:
			respondError(w, http.StatusNotFound, "lesson not found")
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to set output type")
		}
		return
	}

	// Return updated session
	session, _ := h.workflowUC.GetSession(sessionID)
	respondJSON(w, http.StatusOK, session)
}

// GeneratePresentation generates a presentation for a lesson
func (h *WorkflowHandler) GeneratePresentation(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	presentation, err := h.workflowUC.GeneratePresentation(r.Context(), sessionID, lessonID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "session not found")
		case workflow.ErrLessonNotFound:
			respondError(w, http.StatusNotFound, "lesson not found")
		default:
			respondError(w, http.StatusInternalServerError, "failed to generate presentation")
		}
		return
	}

	respondJSON(w, http.StatusAccepted, presentation)
}

// GetPresentation retrieves a presentation by lesson ID
func (h *WorkflowHandler) GetPresentation(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	presentation, err := h.workflowUC.GetPresentation(sessionID, lessonID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "session not found")
		case workflow.ErrLessonNotFound:
			respondError(w, http.StatusNotFound, "lesson not found")
		default:
			respondError(w, http.StatusInternalServerError, "failed to get presentation")
		}
		return
	}

	if presentation == nil {
		respondError(w, http.StatusNotFound, "presentation not found")
		return
	}

	respondJSON(w, http.StatusOK, presentation)
}

// RegenerateAudio regenerates audio files for a presentation
func (h *WorkflowHandler) RegenerateAudio(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	lessonIDStr := chi.URLParam(r, "lessonId")
	lessonID, err := uuid.Parse(lessonIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	presentation, err := h.workflowUC.RegenerateAudio(r.Context(), sessionID, lessonID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "session not found")
		case workflow.ErrLessonNotFound:
			respondError(w, http.StatusNotFound, "lesson not found")
		default:
			respondError(w, http.StatusInternalServerError, "failed to regenerate audio: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, presentation)
}

// UpdateLessonScript updates a lesson script's content
func (h *WorkflowHandler) UpdateLessonScript(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	var req domain.UpdateLessonScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	script, err := h.workflowUC.UpdateLessonScript(sessionID, lessonID, &req)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "session not found")
		case workflow.ErrLessonNotFound:
			respondError(w, http.StatusNotFound, "lesson not found")
		default:
			respondError(w, http.StatusInternalServerError, "failed to update script")
		}
		return
	}

	respondJSON(w, http.StatusOK, script)
}

// RegenerateScript regenerates a lesson script using AI
func (h *WorkflowHandler) RegenerateScript(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	script, err := h.workflowUC.RegenerateScript(r.Context(), sessionID, lessonID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, "session not found")
		case workflow.ErrLessonNotFound:
			respondError(w, http.StatusNotFound, "lesson not found")
		case workflow.ErrTopicNotFound:
			respondError(w, http.StatusNotFound, "topic not found")
		default:
			respondError(w, http.StatusInternalServerError, "failed to regenerate script: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, script)
}

// GetCourseLessons retrieves all lessons for a course (public endpoint)
func (h *WorkflowHandler) GetCourseLessons(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "courseId")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	lessons, err := h.workflowUC.GetCourseLessons(courseID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no lessons found for this course")
		return
	}

	respondJSON(w, http.StatusOK, lessons)
}

// ProceedToQuestionGeneration starts quiz question generation
func (h *WorkflowHandler) ProceedToQuestionGeneration(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	session, err := h.workflowUC.ProceedToQuestionGeneration(r.Context(), sessionID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case workflow.ErrInvalidStep:
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to proceed: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusAccepted, session)
}

// PreviewQuestions generates questions for preview without saving them
func (h *WorkflowHandler) PreviewQuestions(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session ID")
		return
	}

	preview, err := h.workflowUC.PreviewQuestions(r.Context(), sessionID)
	if err != nil {
		switch err {
		case workflow.ErrSessionNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to preview questions: "+err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, preview)
}

// ======== Course-level lesson management handlers ========

// UpdateLessonScriptForCourse updates a lesson script for a completed course
func (h *WorkflowHandler) UpdateLessonScriptForCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	var req domain.UpdateLessonScriptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	lesson, err := h.workflowUC.UpdateLessonScriptForCourse(r.Context(), courseID, lessonID, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update lesson: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, lesson)
}

// RegenerateLessonScriptForCourse regenerates a lesson script using AI for a completed course
func (h *WorkflowHandler) RegenerateLessonScriptForCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	lesson, err := h.workflowUC.RegenerateLessonScriptForCourse(r.Context(), courseID, lessonID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to regenerate lesson: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, lesson)
}

// RegeneratePresentationForCourse regenerates a presentation for a lesson in a completed course
func (h *WorkflowHandler) RegeneratePresentationForCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := uuid.Parse(chi.URLParam(r, "courseId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	lessonID, err := uuid.Parse(chi.URLParam(r, "lessonId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid lesson ID")
		return
	}

	presentation, err := h.workflowUC.RegeneratePresentationForCourse(r.Context(), courseID, lessonID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to regenerate presentation: "+err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, presentation)
}
