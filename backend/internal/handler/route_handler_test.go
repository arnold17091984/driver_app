package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoute_ComputeRoute_InvalidJSON(t *testing.T) {
	h := NewRouteHandler(&mockRouteComputer{})
	req := httptest.NewRequest("POST", "/routes/compute", strings.NewReader("bad"))
	rec := httptest.NewRecorder()

	h.ComputeRoute(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRoute_ComputeRoute_InvalidGPS(t *testing.T) {
	h := NewRouteHandler(&mockRouteComputer{})
	body := `{"origin":{"lat":999,"lng":0},"destination":{"lat":14.5,"lng":121.0}}`
	req := httptest.NewRequest("POST", "/routes/compute", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.ComputeRoute(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRoute_ComputeRoute_ValidZeroCoords(t *testing.T) {
	h := NewRouteHandler(&mockRouteComputer{})
	body := `{"origin":{"lat":0,"lng":0},"destination":{"lat":0,"lng":0}}`
	req := httptest.NewRequest("POST", "/routes/compute", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.ComputeRoute(rec, req)

	// (0,0) should be accepted
	if rec.Code == http.StatusBadRequest {
		t.Errorf("(0,0) coordinates should be valid")
	}
}

func TestRoute_ComputeRoute_Success(t *testing.T) {
	h := NewRouteHandler(&mockRouteComputer{})
	body := `{"origin":{"lat":14.5547,"lng":121.0244},"destination":{"lat":14.56,"lng":121.03}}`
	req := httptest.NewRequest("POST", "/routes/compute", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.ComputeRoute(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
