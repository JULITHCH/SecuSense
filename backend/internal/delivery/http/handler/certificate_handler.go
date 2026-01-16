package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/secusense/backend/internal/delivery/http/middleware"
	"github.com/secusense/backend/internal/usecase/certificate"
)

type CertificateHandler struct {
	certUC *certificate.UseCase
}

func NewCertificateHandler(certUC *certificate.UseCase) *CertificateHandler {
	return &CertificateHandler{
		certUC: certUC,
	}
}

func (h *CertificateHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	certs, err := h.certUC.GetByUserID(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list certificates")
		return
	}

	respondJSON(w, http.StatusOK, certs)
}

func (h *CertificateHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid certificate ID")
		return
	}

	cert, err := h.certUC.GetByID(id)
	if err != nil {
		switch err {
		case certificate.ErrCertificateNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to get certificate")
		}
		return
	}

	respondJSON(w, http.StatusOK, cert)
}

func (h *CertificateHandler) Generate(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	var req struct {
		CourseID  uuid.UUID `json:"courseId"`
		AttemptID uuid.UUID `json:"attemptId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cert, err := h.certUC.Generate(userID, req.CourseID, req.AttemptID)
	if err != nil {
		switch err {
		case certificate.ErrTestNotPassed:
			respondError(w, http.StatusBadRequest, err.Error())
		case certificate.ErrCertificateExists:
			respondError(w, http.StatusConflict, err.Error())
		case certificate.ErrAttemptNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to generate certificate")
		}
		return
	}

	respondJSON(w, http.StatusCreated, cert)
}

func (h *CertificateHandler) Download(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid certificate ID")
		return
	}

	userID := middleware.GetUserID(r.Context())

	pdfBytes, err := h.certUC.GeneratePDF(id, userID)
	if err != nil {
		switch err {
		case certificate.ErrCertificateNotFound:
			respondError(w, http.StatusNotFound, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "failed to generate PDF")
		}
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=certificate.pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}

func (h *CertificateHandler) Verify(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		respondError(w, http.StatusBadRequest, "verification hash required")
		return
	}

	verification, err := h.certUC.Verify(hash)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "verification failed")
		return
	}

	respondJSON(w, http.StatusOK, verification)
}
