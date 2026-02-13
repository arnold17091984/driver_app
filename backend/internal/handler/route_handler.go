package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/kento/driver/backend/internal/maps"
	"github.com/kento/driver/backend/pkg/apperror"
)

type RouteHandler struct {
	mapsClient routeComputer
}

func NewRouteHandler(mapsClient routeComputer) *RouteHandler {
	return &RouteHandler{mapsClient: mapsClient}
}

type computeRouteRequest struct {
	Origin       maps.LatLng   `json:"origin"`
	Destination  maps.LatLng   `json:"destination"`
	Intermediates []maps.LatLng `json:"intermediates"`
}

func (h *RouteHandler) ComputeRoute(w http.ResponseWriter, r *http.Request) {
	var req computeRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.WriteError(w, apperror.ErrBadRequest)
		return
	}

	if !isValidGPSCoord(req.Origin.Lat, req.Origin.Lng) || !isValidGPSCoord(req.Destination.Lat, req.Destination.Lng) {
		apperror.WriteErrorMsg(w, 400, "VALIDATION_ERROR", "origin and destination must have valid GPS coordinates")
		return
	}

	result, err := h.mapsClient.ComputeRoute(r.Context(), req.Origin, req.Destination, req.Intermediates)
	if err != nil {
		log.Printf("[routes] ComputeRoute error: %v", err)
		apperror.WriteErrorMsg(w, 502, "ROUTES_API_ERROR", "failed to compute route")
		return
	}

	apperror.WriteSuccess(w, result)
}
