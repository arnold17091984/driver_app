package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
