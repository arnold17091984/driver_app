package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/pkg/apperror"
)

type DispatchHandler struct {
	dispatchSvc dispatchService
	vehicleSvc  vehicleService
}

func NewDispatchHandler(dispatchSvc dispatchService, vehicleSvc vehicleService) *DispatchHandler {
	return &DispatchHandler{dispatchSvc: dispatchSvc, vehicleSvc: vehicleSvc}
}

func (h *DispatchHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req dto.CreateDispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.Purpose == "" || req.PickupAddress == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "purpose and pickup_address are required")
		return
	}
	if req.PassengerCount <= 0 {
		req.PassengerCount = 1
	}

	dispatch, err := h.dispatchSvc.Create(r.Context(), req, claims.UserID)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteCreated(w, dispatch)
}

func (h *DispatchHandler) QuickBoard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req dto.QuickBoardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.VehicleID == "" || req.PassengerName == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "vehicle_id and passenger_name are required")
		return
	}

	dispatch, err := h.dispatchSvc.QuickBoard(r.Context(), req, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteCreated(w, dispatch)
}

func (h *DispatchHandler) Alight(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	if err := h.dispatchSvc.UpdateStatus(r.Context(), id, model.DispatchStatusCompleted, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DispatchHandler) DriverBoard(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req struct {
		PassengerName    string  `json:"passenger_name"`
		PassengerCount   int     `json:"passenger_count"`
		Purpose          string  `json:"purpose"`
		Notes            *string `json:"notes,omitempty"`
		EstimatedMinutes int     `json:"estimated_minutes,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.PassengerName == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "passenger_name is required")
		return
	}

	// Find the driver's vehicle
	vehicle, err := h.vehicleSvc.GetByDriverID(r.Context(), claims.UserID)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	if vehicle == nil {
		apperror.WriteErrorMsg(w, 404, "NO_VEHICLE", "no vehicle assigned to this driver")
		return
	}

	boardReq := dto.QuickBoardRequest{
		VehicleID:        vehicle.ID,
		PassengerName:    req.PassengerName,
		PassengerCount:   req.PassengerCount,
		Purpose:          req.Purpose,
		Notes:            req.Notes,
		EstimatedMinutes: req.EstimatedMinutes,
	}

	dispatch, err := h.dispatchSvc.QuickBoard(r.Context(), boardReq, claims.UserID)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteCreated(w, dispatch)
}

func (h *DispatchHandler) List(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit, ok := parseIntParam(w, r, "limit", 0)
	if !ok {
		return
	}
	offset, ok := parseIntParam(w, r, "offset", 0)
	if !ok {
		return
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	dispatches, err := h.dispatchSvc.List(r.Context(), status, limit, offset)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, dispatches)
}

func (h *DispatchHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	dispatch, err := h.dispatchSvc.GetByID(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	if dispatch == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	apperror.WriteSuccess(w, dispatch)
}

func (h *DispatchHandler) Assign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.AssignDispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.dispatchSvc.Assign(r.Context(), id, req.VehicleID, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DispatchHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	var req dto.CancelDispatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if err := h.dispatchSvc.Cancel(r.Context(), id, req.Reason, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DispatchHandler) CalculateETAs(w http.ResponseWriter, r *http.Request) {
	var req dto.CalculateETARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if !isValidGPSCoord(req.PickupLat, req.PickupLng) {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "pickup_lat and pickup_lng must be valid GPS coordinates")
		return
	}

	etas, err := h.dispatchSvc.CalculateETAs(r.Context(), req.PickupLat, req.PickupLng)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, etas)
}

func (h *DispatchHandler) GetETASnapshots(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	snapshots, err := h.dispatchSvc.GetETASnapshots(r.Context(), id)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, snapshots)
}

// Driver endpoints
func (h *DispatchHandler) CurrentTrip(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	dispatch, err := h.dispatchSvc.GetCurrentTripByDriverID(r.Context(), claims.UserID)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}
	if dispatch == nil {
		apperror.WriteSuccess(w, nil)
		return
	}

	apperror.WriteSuccess(w, dispatch)
}

func (h *DispatchHandler) AcceptTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	if err := h.dispatchSvc.UpdateStatus(r.Context(), id, model.DispatchStatusAccepted, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DispatchHandler) EnRouteTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	if err := h.dispatchSvc.UpdateStatus(r.Context(), id, model.DispatchStatusEnRoute, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DispatchHandler) ArriveTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	if err := h.dispatchSvc.UpdateStatus(r.Context(), id, model.DispatchStatusArrived, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *DispatchHandler) CompleteTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	claims := middleware.GetClaims(r.Context())

	if err := h.dispatchSvc.UpdateStatus(r.Context(), id, model.DispatchStatusCompleted, claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
