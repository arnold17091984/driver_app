package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/pkg/apperror"
)

type AuthHandler struct {
	authSvc authService
}

func NewAuthHandler(authSvc authService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.EmployeeID == "" || req.Password == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "employee_id and password are required")
		return
	}

	resp, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	resp, err := h.authSvc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, resp)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	user, err := h.authSvc.GetUser(r.Context(), claims.UserID)
	if err != nil || user == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	apperror.WriteSuccess(w, dto.UserInfo{
		ID:            user.ID,
		EmployeeID:    user.EmployeeID,
		Name:          user.Name,
		Role:          string(user.Role),
		PriorityLevel: user.PriorityLevel,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// For JWT-based auth, logout is handled client-side by discarding the token.
	// Refresh token invalidation would require a blocklist (future enhancement).
	w.WriteHeader(http.StatusNoContent)
}
