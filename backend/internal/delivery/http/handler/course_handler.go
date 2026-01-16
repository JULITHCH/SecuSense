package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/internal/usecase/course"
)

type CourseHandler struct {
	courseUC *course.UseCase
	validate *validator.Validate
}

func NewCourseHandler(courseUC *course.UseCase) *CourseHandler {
	return &CourseHandler{
		courseUC: courseUC,
		validate: validator.New(),
	}
}

func (h *CourseHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	courses, total, err := h.courseUC.List(page, pageSize, true) // Only published for public
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list courses")
		return
	}

	respondPaginated(w, courses, total, page, pageSize)
}

func (h *CourseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	courseObj, err := h.courseUC.GetByID(id)
	if err != nil {
		switch err {
		case course.ErrCourseNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get course")
		}
		return
	}

	respondJSON(w, http.StatusOK, courseObj)
}

func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	courseObj, err := h.courseUC.Create(&req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create course")
		return
	}

	respondJSON(w, http.StatusCreated, courseObj)
}

func (h *CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	var req domain.UpdateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	courseObj, err := h.courseUC.Update(id, &req)
	if err != nil {
		switch err {
		case course.ErrCourseNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to update course")
		}
		return
	}

	respondJSON(w, http.StatusOK, courseObj)
}

func (h *CourseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	if err := h.courseUC.Delete(id); err != nil {
		switch err {
		case course.ErrCourseNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to delete course")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	courses, total, err := h.courseUC.List(page, pageSize, false) // All courses for admin
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list courses")
		return
	}

	respondPaginated(w, courses, total, page, pageSize)
}

func (h *CourseHandler) Publish(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	if err := h.courseUC.Publish(id); err != nil {
		switch err {
		case course.ErrCourseNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to publish course")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"published": true})
}

func (h *CourseHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid course ID")
		return
	}

	if err := h.courseUC.Unpublish(id); err != nil {
		switch err {
		case course.ErrCourseNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to unpublish course")
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"published": false})
}
