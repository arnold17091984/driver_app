package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/pkg/apperror"
)

type ReservationHandler struct {
	reservationSvc reservationService
	userSvc        authService
}

func NewReservationHandler(reservationSvc reservationService, userSvc authService) *ReservationHandler {
	return &ReservationHandler{reservationSvc: reservationSvc, userSvc: userSvc}
}

func (h *ReservationHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req dto.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.VehicleID == "" || req.Purpose == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "vehicle_id, start_time, end_time, and purpose are required")
		return
	}

	user, err := h.userSvc.GetUser(r.Context(), claims.UserID)
	if err != nil || user == nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	reservation, err := h.reservationSvc.Create(r.Context(), req, claims.UserID, user.PriorityLevel)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteCreated(w, reservation)
}

func (h *ReservationHandler) List(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.URL.Query().Get("vehicle_id")
	status := r.URL.Query().Get("status")

	limit, ok := parseIntParam(w, r, "limit", 0)
	if !ok {
		return
	}
	offset, ok := parseIntParam(w, r, "offset", 0)
	if !ok {
		return
	}
	from, ok := parseTimeParam(w, r, "from")
	if !ok {
		return
	}
	to, ok := parseTimeParam(w, r, "to")
	if !ok {
		return
	}

	reservations, err := h.reservationSvc.List(r.Context(), vehicleID, from, to, status, limit, offset)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, reservations)
}

func (h *ReservationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	reservation, err := h.reservationSvc.GetByID(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	if reservation == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	apperror.WriteSuccess(w, reservation)
}

func (h *ReservationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.UpdateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	reservation, err := h.reservationSvc.Update(r.Context(), id, req, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, reservation)
}

func (h *ReservationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.CancelReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.reservationSvc.Cancel(r.Context(), id, claims.UserID, req.Reason); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ReservationHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start_time")
	endStr := r.URL.Query().Get("end_time")
	vehicleID := r.URL.Query().Get("vehicle_id")

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "invalid start_time format")
		return
	}
	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "invalid end_time format")
		return
	}

	overlaps, err := h.reservationSvc.CheckAvailability(r.Context(), vehicleID, startTime, endTime)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	available := len(overlaps) == 0
	apperror.WriteSuccess(w, map[string]interface{}{
		"available":    available,
		"conflicts":    overlaps,
		"conflict_count": len(overlaps),
	})
}
