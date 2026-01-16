package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/secusense/backend/internal/delivery/http/middleware"
	"github.com/secusense/backend/internal/domain"
	"github.com/secusense/backend/internal/usecase/auth"
)

type AuthHandler struct {
	authUC   *auth.UseCase
	validate *validator.Validate
}

func NewAuthHandler(authUC *auth.UseCase) *AuthHandler {
	return &AuthHandler{
		authUC:   authUC,
		validate: validator.New(),
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.authUC.Register(&req)
	if err != nil {
		switch err {
		case auth.ErrUserAlreadyExists:
			respondError(w, http.StatusConflict, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "registration failed")
		}
		return
	}

	respondJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.authUC.Login(&req)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			respondError(w, http.StatusUnauthorized, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	response, err := h.authUC.RefreshToken(req.RefreshToken)
	if err != nil {
		switch err {
		case auth.ErrInvalidRefreshToken:
			respondError(w, http.StatusUnauthorized, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, "token refresh failed")
		}
		return
	}

	respondJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())

	user, err := h.authUC.GetCurrentUser(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get user")
		return
	}
	if user == nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, user)
}
