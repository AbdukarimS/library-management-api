package handlers

import (
	"context"
	"encoding/json"
	"library-management-api/internal/middleware"
	"library-management-api/internal/models"
	"library-management-api/internal/response"
	"library-management-api/internal/services"
	"net/http"
	"time"
)

type AuthHandler struct {
	authService *services.AuthService
	loanService *services.LoanService
}

func NewAuthHandler(authService *services.AuthService, loanService *services.LoanService) *AuthHandler {
	return &AuthHandler{authService: authService, loanService: loanService}
}

func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "UP"})
}

func (h *AuthHandler) HealthDB(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.loanService.CheckDBHealth(ctx); err != nil {
		response.Error(w, http.StatusServiceUnavailable, "DB_UNAVAILABLE", "База данных недоступна")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"database": "CONNECTED"})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}
	user, err := h.authService.Register(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_JSON", "Некорректный формат JSON")
		return
	}
	res, err := h.authService.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, res)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.CtxUserID).(int)
	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, user)
}
