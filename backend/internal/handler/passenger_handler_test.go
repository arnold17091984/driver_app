package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/model"
)

func TestPassenger_Register_MissingFields(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	body := `{"phone_number":"","password":"","name":""}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPassenger_Register_InvalidPhone(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	body := `{"phone_number":"abc","password":"pass123","name":"Test"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPassenger_Register_Success(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	body := `{"phone_number":"+639123456789","password":"pass123","name":"Test"}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestPassenger_Login_MissingFields(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	body := `{"phone_number":"","password":""}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPassenger_RequestRide_InvalidGPS(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	body := `{"pickup_address":"123 St","pickup_lat":999,"pickup_lng":0}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	h.RequestRide(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPassenger_RequestRide_MissingAddress(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	body := `{"pickup_address":"","pickup_lat":14.5,"pickup_lng":121.0}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	h.RequestRide(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPassenger_RateRide_InvalidRating(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	// Need to set up a dispatch that belongs to this user
	h.dispatchSvc = &mockDispatchSvc{
		getByIDFn: func(_ context.Context, id string) (*model.Dispatch, error) {
			return nil, nil
		},
	}
	body := `{"rating":6,"comment":"great"}`
	req := httptest.NewRequest("POST", "/rides/d1/rate", strings.NewReader(body))
	req = withChiParam(req, "id", "d1")
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	// GetByID returns nil, so we get 404
	h.RateRide(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPassenger_GetCurrentRide_NoClaims(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	req := httptest.NewRequest("GET", "/rides/current", nil)
	rec := httptest.NewRecorder()

	h.GetCurrentRide(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPassenger_GetRideHistory_InvalidLimit(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})
	req := httptest.NewRequest("GET", "/rides/history?limit=abc", nil)
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	h.GetRideHistory(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPassenger_GetNearbyVehicles_Success(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{
		calculateETAsFn: func(_ context.Context, lat, lng float64) ([]dto.VehicleETA, error) {
			return []dto.VehicleETA{
				{VehicleID: "v-1", VehicleName: "Car A", DriverName: "Driver 1", DurationSec: 300, IsAvailable: true},
			}, nil
		},
	}, &mockLocationSvc{}, &mockBookingSvc{})

	body := `{"pickup_lat":35.6812,"pickup_lng":139.7671}`
	req := httptest.NewRequest("POST", "/passenger/rides/nearby-vehicles", strings.NewReader(body))
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	h.GetNearbyVehicles(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var vehicles []dto.VehicleETA
	json.NewDecoder(rec.Body).Decode(&vehicles)
	if len(vehicles) != 1 {
		t.Errorf("got %d vehicles, want 1", len(vehicles))
	}
}

func TestPassenger_GetNearbyVehicles_InvalidCoords(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})

	body := `{"pickup_lat":999,"pickup_lng":0}`
	req := httptest.NewRequest("POST", "/passenger/rides/nearby-vehicles", strings.NewReader(body))
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	h.GetNearbyVehicles(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPassenger_GetNearbyVehicles_NoClaims(t *testing.T) {
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{})

	body := `{"pickup_lat":35.6812,"pickup_lng":139.7671}`
	req := httptest.NewRequest("POST", "/passenger/rides/nearby-vehicles", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.GetNearbyVehicles(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPassenger_RequestRide_WithVehicleID(t *testing.T) {
	var capturedReq dto.UnifiedBookingRequest
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{
		createBookingFn: func(_ context.Context, req dto.UnifiedBookingRequest, _ string, _ int) (*dto.UnifiedBookingResponse, error) {
			capturedReq = req
			return &dto.UnifiedBookingResponse{Type: "dispatch", Dispatch: &model.Dispatch{ID: "d1"}}, nil
		},
	})

	body := `{"pickup_address":"123 St","pickup_lat":35.6812,"pickup_lng":139.7671,"vehicle_id":"v-1"}`
	req := httptest.NewRequest("POST", "/passenger/rides", strings.NewReader(body))
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	h.RequestRide(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if capturedReq.Mode != "specific" {
		t.Errorf("mode = %q, want %q", capturedReq.Mode, "specific")
	}
	if capturedReq.VehicleID == nil || *capturedReq.VehicleID != "v-1" {
		t.Errorf("vehicle_id = %v, want v-1", capturedReq.VehicleID)
	}
}

func TestPassenger_RequestRide_WithoutVehicleID(t *testing.T) {
	var capturedReq dto.UnifiedBookingRequest
	h := NewPassengerHandler(&mockPassengerAuthSvc{}, &mockDispatchSvc{}, &mockLocationSvc{}, &mockBookingSvc{
		createBookingFn: func(_ context.Context, req dto.UnifiedBookingRequest, _ string, _ int) (*dto.UnifiedBookingResponse, error) {
			capturedReq = req
			return &dto.UnifiedBookingResponse{Type: "dispatch"}, nil
		},
	})

	body := `{"pickup_address":"123 St","pickup_lat":35.6812,"pickup_lng":139.7671}`
	req := httptest.NewRequest("POST", "/passenger/rides", strings.NewReader(body))
	req = withClaims(req, "user1", "pass1", "passenger")
	rec := httptest.NewRecorder()

	h.RequestRide(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if capturedReq.Mode != "any" {
		t.Errorf("mode = %q, want %q", capturedReq.Mode, "any")
	}
	if capturedReq.VehicleID != nil {
		t.Errorf("vehicle_id = %v, want nil", capturedReq.VehicleID)
	}
}
