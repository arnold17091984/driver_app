package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/middleware"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/pkg/apperror"
	"github.com/kento/driver/backend/pkg/jwt"
)

// ---------- helpers ----------

func withClaims(req *http.Request, userID, empID, role string) *http.Request {
	claims := &jwt.Claims{
		UserID:     userID,
		EmployeeID: empID,
		Role:       role,
		TokenType:  "access",
	}
	ctx := context.WithValue(req.Context(), middleware.ClaimsKey, claims)
	return req.WithContext(ctx)
}

func withChiParam(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func decodeError(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp.Error.Code
}

// ===================================================================
// Auth handler integration tests
// ===================================================================

func TestLogin_Success(t *testing.T) {
	mock := &mockAuthSvc{
		loginFn: func(_ context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
			return &dto.LoginResponse{
				AccessToken:  "access-tok",
				RefreshToken: "refresh-tok",
				User:         dto.UserInfo{ID: "u1", EmployeeID: req.EmployeeID, Name: "Admin", Role: "admin"},
			}, nil
		},
	}
	h := &AuthHandler{authSvc: mock}

	body := `{"employee_id":"admin001","password":"pass123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp dto.LoginResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.AccessToken != "access-tok" {
		t.Errorf("access_token = %q, want %q", resp.AccessToken, "access-tok")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mock := &mockAuthSvc{
		loginFn: func(_ context.Context, _ dto.LoginRequest) (*dto.LoginResponse, error) {
			return nil, apperror.ErrInvalidCredentials
		},
	}
	h := &AuthHandler{authSvc: mock}

	body := `{"employee_id":"admin001","password":"wrong"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if code := decodeError(t, rec); code != "INVALID_CREDENTIALS" {
		t.Errorf("error code = %q, want %q", code, "INVALID_CREDENTIALS")
	}
}

func TestRefresh_Success(t *testing.T) {
	mock := &mockAuthSvc{
		refreshFn: func(_ context.Context, _ string) (*dto.RefreshResponse, error) {
			return &dto.RefreshResponse{AccessToken: "new-access"}, nil
		},
	}
	h := &AuthHandler{authSvc: mock}

	body := `{"refresh_token":"old-refresh"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRefresh_ServiceError(t *testing.T) {
	mock := &mockAuthSvc{
		refreshFn: func(_ context.Context, _ string) (*dto.RefreshResponse, error) {
			return nil, apperror.ErrUnauthorized
		},
	}
	h := &AuthHandler{authSvc: mock}

	body := `{"refresh_token":"bad"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMe_Success(t *testing.T) {
	mock := &mockAuthSvc{
		getUserFn: func(_ context.Context, id string) (*model.User, error) {
			return &model.User{
				ID:         id,
				EmployeeID: "emp001",
				Name:       "Test User",
				Role:       model.RoleAdmin,
			}, nil
		},
	}
	h := &AuthHandler{authSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req = withClaims(req, "user-1", "emp001", "admin")
	rec := httptest.NewRecorder()
	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var resp dto.UserInfo
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Name != "Test User" {
		t.Errorf("name = %q, want %q", resp.Name, "Test User")
	}
}

func TestMe_UserNotFound(t *testing.T) {
	mock := &mockAuthSvc{
		getUserFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, nil
		},
	}
	h := &AuthHandler{authSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req = withClaims(req, "user-1", "emp001", "admin")
	rec := httptest.NewRecorder()
	h.Me(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// ===================================================================
// Dispatch handler integration tests
// ===================================================================

func TestDispatchCreate_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		createFn: func(_ context.Context, req dto.CreateDispatchRequest, rid string) (*model.Dispatch, error) {
			return &model.Dispatch{
				ID:            "d-1",
				RequesterID:   rid,
				Purpose:       req.Purpose,
				PickupAddress: req.PickupAddress,
				Status:        model.DispatchStatusPending,
			}, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	body := `{"purpose":"VIP transfer","pickup_address":"Tokyo Station"}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestDispatchCreate_MissingPurpose(t *testing.T) {
	h := &DispatchHandler{}

	body := `{"pickup_address":"Tokyo Station"}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatchCreate_MissingPickupAddress(t *testing.T) {
	h := &DispatchHandler{}

	body := `{"purpose":"VIP transfer"}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatchList_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		listFn: func(_ context.Context, _ string, _, _ int) ([]model.Dispatch, error) {
			return []model.Dispatch{{ID: "d-1"}, {ID: "d-2"}}, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/dispatches?status=pending", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDispatchGet_Found(t *testing.T) {
	mock := &mockDispatchSvc{
		getByIDFn: func(_ context.Context, id string) (*model.Dispatch, error) {
			return &model.Dispatch{ID: id, Purpose: "test"}, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/dispatches/d-1", nil)
	req = withChiParam(req, "id", "d-1")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDispatchGet_NotFound(t *testing.T) {
	mock := &mockDispatchSvc{
		getByIDFn: func(_ context.Context, _ string) (*model.Dispatch, error) {
			return nil, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/dispatches/nonexist", nil)
	req = withChiParam(req, "id", "nonexist")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDispatchAssign_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		assignFn: func(_ context.Context, _, _, _ string) error { return nil },
	}
	h := &DispatchHandler{dispatchSvc: mock}

	body := `{"vehicle_id":"v-1"}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches/d-1/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	req = withChiParam(req, "id", "d-1")
	rec := httptest.NewRecorder()
	h.Assign(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestDispatchAssign_ServiceError(t *testing.T) {
	mock := &mockDispatchSvc{
		assignFn: func(_ context.Context, _, _, _ string) error {
			return apperror.New(400, "INVALID_STATUS", "dispatch is not in pending status")
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	body := `{"vehicle_id":"v-1"}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches/d-1/assign", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	req = withChiParam(req, "id", "d-1")
	rec := httptest.NewRecorder()
	h.Assign(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatchCancel_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		cancelFn: func(_ context.Context, _, _, _ string) error { return nil },
	}
	h := &DispatchHandler{dispatchSvc: mock}

	body := `{"reason":"client cancelled"}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches/d-1/cancel", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	req = withChiParam(req, "id", "d-1")
	rec := httptest.NewRecorder()
	h.Cancel(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestDispatchCalculateETAs_ValidationError(t *testing.T) {
	h := &DispatchHandler{}

	body := `{"pickup_lat":999,"pickup_lng":0}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches/calculate-eta", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.CalculateETAs(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatchCalculateETAs_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		calculateETAsFn: func(_ context.Context, _, _ float64) ([]dto.VehicleETA, error) {
			return []dto.VehicleETA{{VehicleID: "v-1", DurationSec: 300}}, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	body := `{"pickup_lat":14.5547,"pickup_lng":121.0244}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches/calculate-eta", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.CalculateETAs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDispatchCurrentTrip_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		getCurrentTripFn: func(_ context.Context, did string) (*model.Dispatch, error) {
			return &model.Dispatch{ID: "d-1", Status: model.DispatchStatusEnRoute}, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/driver/trips/current", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.CurrentTrip(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDispatchCurrentTrip_NoTrip(t *testing.T) {
	mock := &mockDispatchSvc{
		getCurrentTripFn: func(_ context.Context, _ string) (*model.Dispatch, error) {
			return nil, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/driver/trips/current", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.CurrentTrip(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestDispatchAcceptTrip_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		updateStatusFn: func(_ context.Context, _ string, _ model.DispatchStatus, _ string) error {
			return nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	req := httptest.NewRequest("POST", "/api/v1/driver/trips/d-1/accept", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	req = withChiParam(req, "id", "d-1")
	rec := httptest.NewRecorder()
	h.AcceptTrip(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestDispatchQuickBoard_ValidationError(t *testing.T) {
	h := &DispatchHandler{}

	body := `{"vehicle_id":"","passenger_name":""}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches/quick-board", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	rec := httptest.NewRecorder()
	h.QuickBoard(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestDispatchQuickBoard_Success(t *testing.T) {
	mock := &mockDispatchSvc{
		quickBoardFn: func(_ context.Context, req dto.QuickBoardRequest, _ string) (*model.Dispatch, error) {
			return &model.Dispatch{ID: "d-1", Status: model.DispatchStatusEnRoute}, nil
		},
	}
	h := &DispatchHandler{dispatchSvc: mock}

	body := `{"vehicle_id":"v-1","passenger_name":"Tanaka"}`
	req := httptest.NewRequest("POST", "/api/v1/dispatches/quick-board", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	rec := httptest.NewRecorder()
	h.QuickBoard(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

// ===================================================================
// Vehicle handler integration tests
// ===================================================================

func TestVehicleList_Success(t *testing.T) {
	mock := &mockVehicleSvc{
		listWithStatusFn: func(_ context.Context) ([]model.VehicleWithStatus, error) {
			return []model.VehicleWithStatus{
				{ID: "v-1", Name: "Car A", Status: model.VehicleStatusAvailable},
			}, nil
		},
	}
	h := &VehicleHandler{vehicleSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/vehicles", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestVehicleGet_Found(t *testing.T) {
	mock := &mockVehicleSvc{
		getByIDFn: func(_ context.Context, id string) (*model.Vehicle, error) {
			return &model.Vehicle{ID: id, Name: "Car A"}, nil
		},
	}
	h := &VehicleHandler{vehicleSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/vehicles/v-1", nil)
	req = withChiParam(req, "id", "v-1")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestVehicleGet_NotFound(t *testing.T) {
	mock := &mockVehicleSvc{
		getByIDFn: func(_ context.Context, _ string) (*model.Vehicle, error) {
			return nil, nil
		},
	}
	h := &VehicleHandler{vehicleSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/vehicles/nonexist", nil)
	req = withChiParam(req, "id", "nonexist")
	rec := httptest.NewRecorder()
	h.Get(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestVehicleCreate_ValidationError(t *testing.T) {
	h := &VehicleHandler{}

	body := `{"name":"","license_plate":"","driver_id":""}`
	req := httptest.NewRequest("POST", "/api/v1/vehicles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "admin-1", "admin001", "admin")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestVehicleCreate_Success(t *testing.T) {
	mock := &mockVehicleSvc{
		createFn: func(_ context.Context, _, name, plate, did string) (*model.Vehicle, error) {
			return &model.Vehicle{ID: "v-new", Name: name, LicensePlate: plate, DriverID: did}, nil
		},
	}
	h := &VehicleHandler{vehicleSvc: mock}

	body := `{"name":"Car B","license_plate":"ABC-123","driver_id":"drv-1"}`
	req := httptest.NewRequest("POST", "/api/v1/vehicles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "admin-1", "admin001", "admin")
	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestVehicleUpdate_ValidationError(t *testing.T) {
	h := &VehicleHandler{}

	body := `{"name":"","license_plate":"","driver_id":""}`
	req := httptest.NewRequest("PUT", "/api/v1/vehicles/v-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "admin-1", "admin001", "admin")
	req = withChiParam(req, "id", "v-1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestVehicleUpdate_Success(t *testing.T) {
	mock := &mockVehicleSvc{
		updateFn: func(_ context.Context, _, _, _, _, _ string) error { return nil },
	}
	h := &VehicleHandler{vehicleSvc: mock}

	body := `{"name":"Car B","license_plate":"XYZ-999","driver_id":"drv-2"}`
	req := httptest.NewRequest("PUT", "/api/v1/vehicles/v-1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "admin-1", "admin001", "admin")
	req = withChiParam(req, "id", "v-1")
	rec := httptest.NewRecorder()
	h.Update(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestVehicleDelete_Success(t *testing.T) {
	mock := &mockVehicleSvc{
		deleteFn: func(_ context.Context, _, _ string) error { return nil },
	}
	h := &VehicleHandler{vehicleSvc: mock}

	req := httptest.NewRequest("DELETE", "/api/v1/vehicles/v-1", nil)
	req = withClaims(req, "admin-1", "admin001", "admin")
	req = withChiParam(req, "id", "v-1")
	rec := httptest.NewRecorder()
	h.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestVehicleListAvailable_Success(t *testing.T) {
	mock := &mockVehicleSvc{
		listAvailableFn: func(_ context.Context) ([]model.VehicleWithStatus, error) {
			return []model.VehicleWithStatus{}, nil
		},
	}
	h := &VehicleHandler{vehicleSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/vehicles/available", nil)
	rec := httptest.NewRecorder()
	h.ListAvailable(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestVehicleToggleMaintenance_Success(t *testing.T) {
	mock := &mockVehicleSvc{
		toggleMaintenanceFn: func(_ context.Context, _, _ string, _ bool) error { return nil },
	}
	h := &VehicleHandler{vehicleSvc: mock}

	body := `{"is_maintenance":true}`
	req := httptest.NewRequest("PATCH", "/api/v1/vehicles/v-1/maintenance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "user-1", "emp001", "dispatcher")
	req = withChiParam(req, "id", "v-1")
	rec := httptest.NewRecorder()
	h.ToggleMaintenance(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

// ===================================================================
// Attendance handler integration tests
// ===================================================================

func TestAttendanceClockIn_Success(t *testing.T) {
	mock := &mockAttendanceSvc{
		clockInFn: func(_ context.Context, did string) (*model.DriverAttendance, error) {
			return &model.DriverAttendance{ID: "att-1", DriverID: did, DriverStatus: model.DriverStatusActive}, nil
		},
	}
	h := &AttendanceHandler{attendanceSvc: mock}

	req := httptest.NewRequest("POST", "/api/v1/attendance/clock-in", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.ClockIn(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestAttendanceClockIn_AlreadyClockedIn(t *testing.T) {
	mock := &mockAttendanceSvc{
		clockInFn: func(_ context.Context, _ string) (*model.DriverAttendance, error) {
			return nil, apperror.New(400, "ALREADY_CLOCKED_IN", "driver is already clocked in")
		},
	}
	h := &AttendanceHandler{attendanceSvc: mock}

	req := httptest.NewRequest("POST", "/api/v1/attendance/clock-in", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.ClockIn(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if code := decodeError(t, rec); code != "ALREADY_CLOCKED_IN" {
		t.Errorf("error code = %q, want %q", code, "ALREADY_CLOCKED_IN")
	}
}

func TestAttendanceClockOut_Success(t *testing.T) {
	mock := &mockAttendanceSvc{
		clockOutFn: func(_ context.Context, _ string) error { return nil },
	}
	h := &AttendanceHandler{attendanceSvc: mock}

	req := httptest.NewRequest("POST", "/api/v1/attendance/clock-out", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.ClockOut(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}

func TestAttendanceUpdateDriverStatus_InvalidStatus(t *testing.T) {
	h := &AttendanceHandler{}

	body := `{"status":"sleeping"}`
	req := httptest.NewRequest("PUT", "/api/v1/driver/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.UpdateDriverStatus(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestAttendanceUpdateDriverStatus_Success(t *testing.T) {
	mock := &mockAttendanceSvc{
		updateStatusFn: func(_ context.Context, did string, s model.DriverStatus) (*model.DriverAttendance, error) {
			return &model.DriverAttendance{ID: "att-1", DriverID: did, DriverStatus: s}, nil
		},
	}
	h := &AttendanceHandler{attendanceSvc: mock}

	body := `{"status":"waiting"}`
	req := httptest.NewRequest("PUT", "/api/v1/driver/status", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.UpdateDriverStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAttendanceGetStatus_Success(t *testing.T) {
	mock := &mockAttendanceSvc{
		getStatusFn: func(_ context.Context, _ string) (*model.DriverAttendance, error) {
			return &model.DriverAttendance{ID: "att-1", DriverStatus: model.DriverStatusActive}, nil
		},
	}
	h := &AttendanceHandler{attendanceSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/attendance/status", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.GetStatus(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAttendanceGetHistory_Success(t *testing.T) {
	mock := &mockAttendanceSvc{
		getHistoryFn: func(_ context.Context, _ string, _ int) ([]model.DriverAttendance, error) {
			return []model.DriverAttendance{}, nil
		},
	}
	h := &AttendanceHandler{attendanceSvc: mock}

	req := httptest.NewRequest("GET", "/api/v1/attendance/history", nil)
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.GetHistory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// ===================================================================
// Location handler integration tests
// ===================================================================

func TestLocationReport_NoVehicle(t *testing.T) {
	vMock := &mockVehicleSvc{
		getByDriverIDFn: func(_ context.Context, _ string) (*model.Vehicle, error) {
			return nil, nil
		},
	}
	h := &LocationHandler{vehicleSvc: vMock}

	body := `{"points":[{"latitude":14.5,"longitude":121.0,"recorded_at":"2026-01-01T00:00:00Z"}]}`
	req := httptest.NewRequest("POST", "/api/v1/locations/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.Report(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if code := decodeError(t, rec); code != "NO_VEHICLE" {
		t.Errorf("error code = %q, want %q", code, "NO_VEHICLE")
	}
}

func TestLocationReport_EmptyPoints(t *testing.T) {
	vMock := &mockVehicleSvc{
		getByDriverIDFn: func(_ context.Context, _ string) (*model.Vehicle, error) {
			return &model.Vehicle{ID: "v-1"}, nil
		},
	}
	h := &LocationHandler{vehicleSvc: vMock}

	body := `{"points":[]}`
	req := httptest.NewRequest("POST", "/api/v1/locations/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.Report(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestLocationReport_Success(t *testing.T) {
	vMock := &mockVehicleSvc{
		getByDriverIDFn: func(_ context.Context, _ string) (*model.Vehicle, error) {
			return &model.Vehicle{ID: "v-1"}, nil
		},
	}
	lMock := &mockLocationSvc{
		reportFn: func(_ context.Context, _ string, _ []model.LocationPoint) error { return nil },
	}
	h := &LocationHandler{locationSvc: lMock, vehicleSvc: vMock}

	body := `{"points":[{"latitude":14.5547,"longitude":121.0244,"recorded_at":"2026-01-01T00:00:00Z"}]}`
	req := httptest.NewRequest("POST", "/api/v1/locations/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withClaims(req, "driver-1", "drv001", "driver")
	rec := httptest.NewRecorder()
	h.Report(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
}
