package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/pkg/apperror"
)

type AuthHandler struct {
	authSvc      authService
	tokenSvc     tokenService
	loginLimiter loginLimiter
}

type loginLimiter interface {
	IsLocked(account string) bool
	RecordFailure(account string)
	RecordSuccess(account string)
}

func NewAuthHandler(authSvc authService, tokenSvc tokenService, ll loginLimiter) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, tokenSvc: tokenSvc, loginLimiter: ll}
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

	// Check brute-force lockout
	if h.loginLimiter != nil && h.loginLimiter.IsLocked(req.EmployeeID) {
		apperror.WriteErrorMsg(w, http.StatusTooManyRequests, "ACCOUNT_LOCKED", "too many failed login attempts; try again later")
		return
	}

	resp, err := h.authSvc.Login(r.Context(), req)
	if err != nil {
		if h.loginLimiter != nil {
			h.loginLimiter.RecordFailure(req.EmployeeID)
		}
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	if h.loginLimiter != nil {
		h.loginLimiter.RecordSuccess(req.EmployeeID)
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
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Blacklist the current access token
	if claims.ID != "" && claims.ExpiresAt != nil {
		_ = h.tokenSvc.Blacklist(r.Context(), claims.ID, claims.UserID, claims.ExpiresAt.Time)
	}

	w.WriteHeader(http.StatusNoContent)
}
