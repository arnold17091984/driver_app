package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/internal/service"
	"github.com/kento/driver/backend/pkg/apperror"
)

type BookingHandler struct {
	bookingSvc *service.BookingService
	authSvc    *service.AuthService
}

func NewBookingHandler(bookingSvc *service.BookingService, authSvc *service.AuthService) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc, authSvc: authSvc}
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req dto.UnifiedBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.Purpose == "" || req.PickupAddress == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "purpose and pickup_address are required")
		return
	}
	if req.Mode != "specific" && req.Mode != "any" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "mode must be 'specific' or 'any'")
		return
	}

	user, err := h.authSvc.GetUser(r.Context(), claims.UserID)
	if err != nil || user == nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	resp, err := h.bookingSvc.CreateBooking(r.Context(), req, claims.UserID, user.PriorityLevel)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteCreated(w, resp)
}

func (h *BookingHandler) GetVehicleTimeline(w http.ResponseWriter, r *http.Request) {
	vehicleID := chi.URLParam(r, "id")
	dateStr := r.URL.Query().Get("date")

	var date time.Time
	if dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "invalid date format, use YYYY-MM-DD")
			return
		}
	} else {
		date = time.Now()
	}

	reservations, err := h.bookingSvc.GetVehicleTimeline(r.Context(), vehicleID, date)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, reservations)
}

func (h *BookingHandler) AcceptReservation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	if err := h.bookingSvc.DriverAcceptReservation(r.Context(), id, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BookingHandler) DeclineReservation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.DriverReservationDeclineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.bookingSvc.DriverDeclineReservation(r.Context(), id, claims.UserID, req.Reason); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BookingHandler) PendingReservations(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	reservations, err := h.bookingSvc.GetPendingByDriverID(r.Context(), claims.UserID)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, reservations)
}
