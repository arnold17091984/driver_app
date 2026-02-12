package handler

import (
	"encoding/json"
	"net/http"

	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/service"
	"github.com/kento/driver/backend/pkg/apperror"
)

type LocationHandler struct {
	locationSvc *service.LocationService
	vehicleSvc  *service.VehicleService
}

func NewLocationHandler(locationSvc *service.LocationService, vehicleSvc *service.VehicleService) *LocationHandler {
	return &LocationHandler{locationSvc: locationSvc, vehicleSvc: vehicleSvc}
}

type locationReportRequest struct {
	Points []model.LocationPoint `json:"points" validate:"required,min=1"`
}

func (h *LocationHandler) Report(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	vehicle, err := h.vehicleSvc.GetByDriverID(r.Context(), claims.UserID)
	if err != nil || vehicle == nil {
		apperror.WriteErrorMsg(w, 400, "NO_VEHICLE", "no vehicle assigned to this driver")
		return
	}

	var req locationReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if len(req.Points) == 0 {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "at least one location point is required")
		return
	}

	if err := h.locationSvc.ReportLocations(r.Context(), vehicle.ID, req.Points); err != nil {
		apperror.WriteError(w, apperror.ErrInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
