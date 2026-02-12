package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/internal/service"
	"github.com/kento/driver/backend/pkg/apperror"
)

type AdminHandler struct {
	userRepo *repository.UserRepo
	auditSvc *service.AuditService
}

func NewAdminHandler(userRepo *repository.UserRepo, auditSvc *service.AuditService) *AdminHandler {
	return &AdminHandler{userRepo: userRepo, auditSvc: auditSvc}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.List(r.Context())
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	apperror.WriteSuccess(w, users)
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

func (h *AdminHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	role := model.Role(req.Role)
	if !role.IsValid() {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "invalid role")
		return
	}

	before, _ := h.userRepo.GetByID(r.Context(), id)
	if err := h.userRepo.UpdateRole(r.Context(), id, role); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	after, _ := h.userRepo.GetByID(r.Context(), id)

	h.auditSvc.Log(r.Context(), claims.UserID, "user.role_change", "user", id, before, after, "")
	w.WriteHeader(http.StatusNoContent)
}

type updatePriorityRequest struct {
	PriorityLevel int `json:"priority_level"`
}

func (h *AdminHandler) UpdatePriority(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req updatePriorityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	before, _ := h.userRepo.GetByID(r.Context(), id)
	if err := h.userRepo.UpdatePriority(r.Context(), id, req.PriorityLevel); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	after, _ := h.userRepo.GetByID(r.Context(), id)

	h.auditSvc.Log(r.Context(), claims.UserID, "user.priority_change", "user", id, before, after, "")
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	actorID := r.URL.Query().Get("actor_id")
	action := r.URL.Query().Get("action")
	targetType := r.URL.Query().Get("target_type")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 50
	}

	var from, to time.Time
	if fromStr != "" {
		from, _ = time.Parse(time.RFC3339, fromStr)
	}
	if toStr != "" {
		to, _ = time.Parse(time.RFC3339, toStr)
	}

	logs, err := h.auditSvc.List(r.Context(), actorID, action, targetType, from, to, limit, offset)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, logs)
}

func (h *AdminHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	log, err := h.auditSvc.GetByID(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	apperror.WriteSuccess(w, log)
}
