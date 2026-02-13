package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kento/driver/backend/internal/model"
)

func TestDispatch_Create_MissingFields(t *testing.T) {
	h := NewDispatchHandler(&mockDispatchSvc{}, &mockVehicleSvc{})
	body := `{"purpose":"","pickup_address":""}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req = withClaims(req, "user1", "emp1", "dispatcher")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatch_Create_Success(t *testing.T) {
	h := NewDispatchHandler(&mockDispatchSvc{}, &mockVehicleSvc{})
	body := `{"purpose":"test","pickup_address":"123 St","passenger_count":2}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	req = withClaims(req, "user1", "emp1", "dispatcher")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestDispatch_List_InvalidLimit(t *testing.T) {
	h := NewDispatchHandler(&mockDispatchSvc{}, &mockVehicleSvc{})
	req := httptest.NewRequest("GET", "/dispatches?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatch_List_Success(t *testing.T) {
	svc := &mockDispatchSvc{
		listFn: func(ctx context.Context, status string, limit, offset int) ([]model.Dispatch, error) {
			return []model.Dispatch{{ID: "d1"}}, nil
		},
	}
	h := NewDispatchHandler(svc, &mockVehicleSvc{})
	req := httptest.NewRequest("GET", "/dispatches?limit=10", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDispatch_Get_NotFound(t *testing.T) {
	h := NewDispatchHandler(&mockDispatchSvc{}, &mockVehicleSvc{})
	req := httptest.NewRequest("GET", "/dispatches/xxx", nil)
	req = withChiParam(req, "id", "xxx")
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDispatch_CalculateETAs_InvalidGPS(t *testing.T) {
	h := NewDispatchHandler(&mockDispatchSvc{}, &mockVehicleSvc{})
	body := `{"pickup_lat":999,"pickup_lng":0}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.CalculateETAs(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatch_CalculateETAs_ValidZeroCoord(t *testing.T) {
	svc := &mockDispatchSvc{}
	h := NewDispatchHandler(svc, &mockVehicleSvc{})
	body := `{"pickup_lat":0,"pickup_lng":0}`
	req := httptest.NewRequest("POST", "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.CalculateETAs(rec, req)

	// (0,0) is valid — should not return 400
	if rec.Code == http.StatusBadRequest {
		t.Errorf("status = %d, (0,0) should be valid GPS coordinates", rec.Code)
	}
}
