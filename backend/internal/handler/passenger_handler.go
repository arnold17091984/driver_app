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
}

func NewPassengerHandler(authSvc passengerAuthService, dispatchSvc dispatchService, locationSvc locationService, bookingSvc bookingService) *PassengerHandler {
	return &PassengerHandler{
		authSvc:     authSvc,
		dispatchSvc: dispatchSvc,
		locationSvc: locationSvc,
		bookingSvc:  bookingSvc,
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

	resp, err := h.authSvc.LoginByPhone(r.Context(), req)
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
	bookingReq := dto.UnifiedBookingRequest{
		Mode:          "any",
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

	// List dispatches by requester_id (passenger's user ID) with active statuses
	dispatches, err := h.dispatchSvc.List(r.Context(), "", 50, 0)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	// Find the current active ride for this passenger
	for _, d := range dispatches {
		if d.RequesterID == claims.UserID &&
			d.Status != "completed" && d.Status != "cancelled" {
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

	dispatches, err := h.dispatchSvc.List(r.Context(), "", limit, offset)
	if err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	// Filter to this passenger's rides
	var rides []interface{}
	for _, d := range dispatches {
		if d.RequesterID == claims.UserID {
			rides = append(rides, d)
		}
	}

	apperror.WriteSuccess(w, rides)
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
