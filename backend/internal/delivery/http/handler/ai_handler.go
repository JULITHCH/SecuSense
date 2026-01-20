package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/internal/usecase/ai"
)

type AIHandler struct {
	aiUC     *ai.UseCase
	validate *validator.Validate
}

func NewAIHandler(aiUC *ai.UseCase) *AIHandler {
	return &AIHandler{
		aiUC:     aiUC,
		validate: validator.New(),
	}
}

func (h *AIHandler) GenerateCourse(w http.ResponseWriter, r *http.Request) {
	var req domain.GenerateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	job, err := h.aiUC.GenerateCourse(r.Context(), &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to start course generation: "+err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, job)
}

func (h *AIHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid job ID")
		return
	}

	job, err := h.aiUC.GetJob(id)
	if err != nil {
		switch err {
		case ai.ErrJobNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get job")
		}
		return
	}

	respondJSON(w, http.StatusOK, job)
}

func (h *AIHandler) SynthesiaWebhook(w http.ResponseWriter, r *http.Request) {
	var payload domain.SynthesiaWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondError(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	if err := h.aiUC.HandleSynthesiaWebhook(&payload); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to process webhook")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AIHandler) RefreshVideoStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	course, err := h.aiUC.RefreshVideoStatus(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, course)
}

func (h *AIHandler) PollPendingVideos(w http.ResponseWriter, r *http.Request) {
	if err := h.aiUC.PollPendingVideos(r.Context()); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "polling complete"})
}

func (h *AIHandler) GetVideoURL(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	videoURL, err := h.aiUC.GetFreshVideoURL(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"videoUrl": videoURL})
}
