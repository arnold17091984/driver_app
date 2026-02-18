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

type PassengerHandler struct {
	authSvc      passengerAuthService
	dispatchSvc  dispatchService
	locationSvc  locationService
	bookingSvc   bookingService
	loginLimiter interface {
		IsLocked(account string) bool
		RecordFailure(account string)
		RecordSuccess(account string)
	}
}

func NewPassengerHandler(authSvc passengerAuthService, dispatchSvc dispatchService, locationSvc locationService, bookingSvc bookingService, ll interface {
	IsLocked(account string) bool
	RecordFailure(account string)
	RecordSuccess(account string)
}) *PassengerHandler {
	return &PassengerHandler{
		authSvc:      authSvc,
		dispatchSvc:  dispatchSvc,
		locationSvc:  locationSvc,
		bookingSvc:   bookingSvc,
		loginLimiter: ll,
	}
}

func (h *PassengerHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.PassengerRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.PhoneNumber == "" || req.Password == "" || req.Name == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "phone_number, password, and name are required")
		return
	}
	if !isValidPhoneNumber(req.PhoneNumber) {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "phone_number must be 7-15 digits, optionally starting with +")
		return
	}
	if len(req.Password) < 8 {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "password must be at least 8 characters")
		return
	}
	if len(req.Password) > 72 {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "password must be at most 72 characters")
		return
	}

	resp, err := h.authSvc.RegisterPassenger(r.Context(), req)
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

func (h *PassengerHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.PassengerLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.PhoneNumber == "" || req.Password == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "phone_number and password are required")
		return
	}

	// Check brute-force lockout
	if h.loginLimiter != nil && h.loginLimiter.IsLocked(req.PhoneNumber) {
		apperror.WriteErrorMsg(w, http.StatusTooManyRequests, "ACCOUNT_LOCKED", "too many failed login attempts; try again later")
		return
	}

	resp, err := h.authSvc.LoginByPhone(r.Context(), req)
	if err != nil {
		if h.loginLimiter != nil {
			h.loginLimiter.RecordFailure(req.PhoneNumber)
		}
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	if h.loginLimiter != nil {
		h.loginLimiter.RecordSuccess(req.PhoneNumber)
	}
	apperror.WriteSuccess(w, resp)
}

func (h *PassengerHandler) GetNearbyVehicles(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	var req dto.CalculateETARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if !isValidGPSCoord(req.PickupLat, req.PickupLng) {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "pickup_lat and pickup_lng must be valid GPS coordinates")
		return
	}

	vehicles, err := h.dispatchSvc.CalculateETAs(r.Context(), req.PickupLat, req.PickupLng)
	if err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, vehicles)
}

func (h *PassengerHandler) RequestRide(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	var req dto.PassengerRideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.PickupAddress == "" {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "pickup_address is required")
		return
	}
	if !isValidGPSCoord(req.PickupLat, req.PickupLng) {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "pickup_lat and pickup_lng must be valid GPS coordinates")
		return
	}

	// Use unified booking service to create an immediate dispatch
	pickupLat := &req.PickupLat
	pickupLng := &req.PickupLng
	var destinations []string
	if req.DropoffAddress != "" {
		destinations = []string{req.DropoffAddress}
	}

	mode := "any"
	if req.VehicleID != nil {
		mode = "specific"
	}

	bookingReq := dto.UnifiedBookingRequest{
		Mode:          mode,
		VehicleID:     req.VehicleID,
		IsNow:         true,
		PickupAddress: req.PickupAddress,
		PickupLat:     pickupLat,
		PickupLng:     pickupLng,
		PassengerName: &req.PassengerName,
		Purpose:       "passenger_request",
		Destinations:  destinations,
	}

	resp, err := h.bookingSvc.CreateBooking(r.Context(), bookingReq, claims.UserID, 0)
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

func (h *PassengerHandler) GetCurrentRide(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	// Query dispatches by this specific requester only
	dispatches, err := h.dispatchSvc.ListByRequester(r.Context(), claims.UserID, "", 50, 0)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	// Find the current active ride for this passenger
	for _, d := range dispatches {
		if d.Status != "completed" && d.Status != "cancelled" {
			apperror.WriteSuccess(w, d)
			return
		}
	}

	// No current ride
	apperror.WriteSuccess(w, nil)
}

func (h *PassengerHandler) GetRideHistory(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	limit, ok := parseIntParam(w, r, "limit", 20)
	if !ok {
		return
	}
	offset, ok := parseIntParam(w, r, "offset", 0)
	if !ok {
		return
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Query only this passenger's dispatches using the proper filtered query
	dispatches, err := h.dispatchSvc.ListByRequester(r.Context(), claims.UserID, "", limit, offset)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, dispatches)
}

func (h *PassengerHandler) CancelRide(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	dispatchID := chi.URLParam(r, "id")

	// Verify this dispatch belongs to the passenger
	dispatch, err := h.dispatchSvc.GetByID(r.Context(), dispatchID)
	if err != nil || dispatch == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	if dispatch.RequesterID != claims.UserID {
		apperror.WriteError(w, apperror.ErrForbidden)
		return
	}

	// Only allow cancellation for pending/assigned status
	if dispatch.Status != "pending" && dispatch.Status != "assigned" {
		apperror.WriteErrorMsg(w, 400, "CANNOT_CANCEL", "ride can only be cancelled when pending or assigned")
		return
	}

	if err := h.dispatchSvc.Cancel(r.Context(), dispatchID, "cancelled by passenger", claims.UserID); err != nil {
		if appErr, ok := err.(*apperror.AppError); ok {
			apperror.WriteError(w, appErr)
			return
		}
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PassengerHandler) GetDriverLocation(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	dispatchID := chi.URLParam(r, "id")

	dispatch, err := h.dispatchSvc.GetByID(r.Context(), dispatchID)
	if err != nil || dispatch == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	if dispatch.RequesterID != claims.UserID {
		apperror.WriteError(w, apperror.ErrForbidden)
		return
	}

	if dispatch.VehicleID == nil {
		apperror.WriteErrorMsg(w, 404, "NO_VEHICLE", "no vehicle assigned yet")
		return
	}

	// Get the latest location for this vehicle (last 1 minute)
	to := time.Now()
	from := to.Add(-1 * time.Minute)
	locations, err := h.locationSvc.GetHistory(r.Context(), *dispatch.VehicleID, from, to)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	if len(locations) == 0 {
		apperror.WriteSuccess(w, nil)
		return
	}

	// Return the most recent location
	apperror.WriteSuccess(w, locations[len(locations)-1])
}

func (h *PassengerHandler) RateRide(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		apperror.WriteError(w, apperror.ErrUnauthorized)
		return
	}

	dispatchID := chi.URLParam(r, "id")

	dispatch, err := h.dispatchSvc.GetByID(r.Context(), dispatchID)
	if err != nil || dispatch == nil {
		apperror.WriteError(w, apperror.ErrNotFound)
		return
	}

	if dispatch.RequesterID != claims.UserID {
		apperror.WriteError(w, apperror.ErrForbidden)
		return
	}

	var req struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "rating must be between 1 and 5")
		return
	}

	if err := h.dispatchSvc.RateDispatch(r.Context(), dispatchID, req.Rating, req.Comment); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	apperror.WriteSuccess(w, map[string]string{"status": "rated"})
}
