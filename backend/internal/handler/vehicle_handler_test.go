package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kento/driver/backend/internal/model"
)

func TestVehicle_List_Success(t *testing.T) {
	svc := &mockVehicleSvc{
		listWithStatusFn: func(ctx context.Context) ([]model.VehicleWithStatus, error) {
			return []model.VehicleWithStatus{{ID: "v1"}}, nil
		},
	}
	h := NewVehicleHandler(svc, &mockLocationSvc{}, "/tmp/test-uploads")
	req := httptest.NewRequest("GET", "/vehicles", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestVehicle_Get_NotFound(t *testing.T) {
	h := NewVehicleHandler(&mockVehicleSvc{}, &mockLocationSvc{}, "/tmp/test-uploads")
	req := httptest.NewRequest("GET", "/vehicles/xxx", nil)
	req = withChiParam(req, "id", "xxx")
	rec := httptest.NewRecorder()

	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestVehicle_Create_MissingFields(t *testing.T) {
	h := NewVehicleHandler(&mockVehicleSvc{}, &mockLocationSvc{}, "/tmp/test-uploads")
	body := `{"name":"","license_plate":"","driver_id":""}`
	req := httptest.NewRequest("POST", "/vehicles", strings.NewReader(body))
	req = withClaims(req, "admin1", "adm1", "admin")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestVehicle_Create_Success(t *testing.T) {
	svc := &mockVehicleSvc{
		createFn: func(ctx context.Context, actorID, name, licensePlate, driverID string) (*model.Vehicle, error) {
			return &model.Vehicle{ID: "v1", Name: name}, nil
		},
	}
	h := NewVehicleHandler(svc, &mockLocationSvc{}, "/tmp/test-uploads")
	body := `{"name":"Van 1","license_plate":"ABC-123","driver_id":"d1"}`
	req := httptest.NewRequest("POST", "/vehicles", strings.NewReader(body))
	req = withClaims(req, "admin1", "adm1", "admin")
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestVehicle_Update_InvalidJSON(t *testing.T) {
	h := NewVehicleHandler(&mockVehicleSvc{}, &mockLocationSvc{}, "/tmp/test-uploads")
	req := httptest.NewRequest("PUT", "/vehicles/v1", strings.NewReader("bad"))
	req = withChiParam(req, "id", "v1")
	req = withClaims(req, "admin1", "adm1", "admin")
	rec := httptest.NewRecorder()

	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestValidateMagicBytes_JPEG(t *testing.T) {
	// Valid JPEG header
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	ok := validateMagicBytes(bytes.NewReader(data), ".jpg")
	if !ok {
		t.Error("expected valid JPEG magic bytes")
	}
}

func TestValidateMagicBytes_PNG(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	ok := validateMagicBytes(bytes.NewReader(data), ".png")
	if !ok {
		t.Error("expected valid PNG magic bytes")
	}
}

func TestValidateMagicBytes_FakeJPEG(t *testing.T) {
	// PNG magic bytes but claiming .jpg extension
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	ok := validateMagicBytes(bytes.NewReader(data), ".jpg")
	if ok {
		t.Error("should reject PNG data with .jpg extension")
	}
}

func TestSafeDeleteOldPhoto_Traversal(t *testing.T) {
	h := &VehicleHandler{uploadDir: "/tmp/test-uploads"}
	// Should not panic or delete anything for traversal paths
	h.safeDeleteOldPhoto("../../../etc/passwd")
	h.safeDeleteOldPhoto("../../..")
	h.safeDeleteOldPhoto(".")
	h.safeDeleteOldPhoto("/")
}
