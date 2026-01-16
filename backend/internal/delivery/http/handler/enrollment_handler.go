package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/secusense/backend/internal/delivery/http/middleware"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/internal/usecase/enrollment"
)

type EnrollmentHandler struct {
	enrollmentUC *enrollment.UseCase
}

func NewEnrollmentHandler(enrollmentUC *enrollment.UseCase) *EnrollmentHandler {
	return &EnrollmentHandler{
		enrollmentUC: enrollmentUC,
	}
}

func (h *EnrollmentHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	enrollmentObj, err := h.enrollmentUC.Enroll(userID, courseID)
	if err != nil {
		switch err {
		case enrollment.ErrCourseNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case enrollment.ErrCourseNotPublished:
			respondError(w, http.StatusBadRequest, err.Error())
		case enrollment.ErrAlreadyEnrolled:
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to enroll")
		}
		return
	}

	respondJSON(w, http.StatusCreated, enrollmentObj)
}

func (h *EnrollmentHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	enrollments, err := h.enrollmentUC.ListByUser(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list enrollments")
		return
	}

	respondJSON(w, http.StatusOK, enrollments)
}

func (h *EnrollmentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid enrollment ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	enrollmentObj, err := h.enrollmentUC.GetByID(id)
	if err != nil {
		switch err {
		case enrollment.ErrEnrollmentNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get enrollment")
		}
		return
	}

	if enrollmentObj.UserID != userID {
		respondError(w, http.StatusNotFound, "enrollment not found")
		return
	}

	respondJSON(w, http.StatusOK, enrollmentObj)
}

func (h *EnrollmentHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid enrollment ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	var req domain.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	enrollmentObj, err := h.enrollmentUC.UpdateProgress(id, userID, req.ProgressPercentage)
	if err != nil {
		switch err {
		case enrollment.ErrEnrollmentNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case enrollment.ErrNotOwner:
			respondError(w, http.StatusForbidden, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to update progress")
		}
		return
	}

	respondJSON(w, http.StatusOK, enrollmentObj)
}

func (h *EnrollmentHandler) CompleteVideo(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid enrollment ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	enrollmentObj, err := h.enrollmentUC.CompleteVideo(id, userID)
	if err != nil {
		switch err {
		case enrollment.ErrEnrollmentNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		case enrollment.ErrNotOwner:
			respondError(w, http.StatusForbidden, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to complete video")
		}
		return
	}

	respondJSON(w, http.StatusOK, enrollmentObj)
}
