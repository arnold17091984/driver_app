package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/pkg/apperror"
)

type ConflictHandler struct {
	conflictSvc    conflictService
	reservationSvc reservationService
}

func NewConflictHandler(conflictSvc conflictService, reservationSvc reservationService) *ConflictHandler {
	return &ConflictHandler{conflictSvc: conflictSvc, reservationSvc: reservationSvc}
}

func (h *ConflictHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	conflicts, err := h.conflictSvc.ListPending(r.Context())
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	apperror.WriteSuccess(w, conflicts)
}

func (h *ConflictHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	conflict, err := h.conflictSvc.GetByID(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	if conflict == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	// Get both reservations for detail view
	winning, _ := h.reservationSvc.GetByID(r.Context(), conflict.WinningReservationID)
	losing, _ := h.reservationSvc.GetByID(r.Context(), conflict.LosingReservationID)

	apperror.WriteSuccess(w, map[string]interface{}{
		"conflict":            conflict,
		"winning_reservation": winning,
		"losing_reservation":  losing,
	})
}

func (h *ConflictHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.ResolveConflictReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.conflictSvc.ResolveReassign(r.Context(), id, req.NewVehicleID, claims.UserID, req.Reason); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConflictHandler) ChangeTime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.ResolveConflictChangeTimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	conflict, err := h.conflictSvc.GetByID(r.Context(), id)
	if err != nil || conflict == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	losingRes, err := h.reservationSvc.GetByID(r.Context(), conflict.LosingReservationID)
	if err != nil || losingRes == nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	losingRes.StartTime = req.NewStartTime
	losingRes.EndTime = req.NewEndTime

	if err := h.conflictSvc.ResolveChangeTime(r.Context(), id, claims.UserID, req.Reason, losingRes); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConflictHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.ResolveConflictCancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.conflictSvc.ResolveCancel(r.Context(), id, claims.UserID, req.Reason); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ConflictHandler) ForceAssign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.ForceAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.conflictSvc.ForceAssign(r.Context(), id, claims.UserID, req.Reason); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
