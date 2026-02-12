package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/pkg/apperror"
)

type AttendanceHandler struct {
	attendanceSvc attendanceService
}

func NewAttendanceHandler(attendanceSvc attendanceService) *AttendanceHandler {
	return &AttendanceHandler{attendanceSvc: attendanceSvc}
}

func (h *AttendanceHandler) ClockIn(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	attendance, err := h.attendanceSvc.ClockIn(r.Context(), claims.UserID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteCreated(w, attendance)
}

func (h *AttendanceHandler) ClockOut(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	if err := h.attendanceSvc.ClockOut(r.Context(), claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AttendanceHandler) UpdateDriverStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	status := model.DriverStatus(req.Status)
	if status != model.DriverStatusActive && status != model.DriverStatusWaiting {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "status must be 'active' or 'waiting'")
		return
	}

	attendance, err := h.attendanceSvc.UpdateDriverStatus(r.Context(), claims.UserID, status)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, attendance)
}

func (h *AttendanceHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	attendance, err := h.attendanceSvc.GetStatus(r.Context(), claims.UserID)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, map[string]interface{}{
		"clocked_in": attendance != nil,
		"attendance": attendance,
	})
}

func (h *AttendanceHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	records, err := h.attendanceSvc.GetHistory(r.Context(), claims.UserID, 50)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, records)
}
