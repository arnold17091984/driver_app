package handler

import (
	"context"
	"time"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/maps"
	"github.com/kento/driver/backend/internal/model"
)

// ── Mock: authService ──

type mockAuthSvc struct {
	loginFn    func(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	refreshFn  func(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error)
	getUserFn  func(ctx context.Context, userID string) (*model.User, error)
}

func (m *mockAuthSvc) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthSvc) Refresh(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error) {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, refreshToken)
	}
	return nil, nil
}

func (m *mockAuthSvc) GetUser(ctx context.Context, userID string) (*model.User, error) {
	if m.getUserFn != nil {
		return m.getUserFn(ctx, userID)
	}
	return nil, nil
}

// ── Mock: dispatchService ──

type mockDispatchSvc struct {
	createFn            func(ctx context.Context, req dto.CreateDispatchRequest, requesterID string) (*model.Dispatch, error)
	quickBoardFn        func(ctx context.Context, req dto.QuickBoardRequest, dispatcherID string) (*model.Dispatch, error)
	getByIDFn           func(ctx context.Context, id string) (*model.Dispatch, error)
	listFn              func(ctx context.Context, status string, limit, offset int) ([]model.Dispatch, error)
	listByRequesterFn   func(ctx context.Context, requesterID, status string, limit, offset int) ([]model.Dispatch, error)
	assignFn            func(ctx context.Context, dispatchID, vehicleID, dispatcherID string) error
	updateStatusFn      func(ctx context.Context, dispatchID string, status model.DispatchStatus, actorID string) error
	cancelFn            func(ctx context.Context, dispatchID, reason, actorID string) error
	getCurrentTripFn    func(ctx context.Context, driverID string) (*model.Dispatch, error)
	getETASnapshotsFn   func(ctx context.Context, dispatchID string) ([]model.DispatchETASnapshot, error)
	calculateETAsFn     func(ctx context.Context, pickupLat, pickupLng float64) ([]dto.VehicleETA, error)
	rateDispatchFn      func(ctx context.Context, dispatchID string, rating int, comment string) error
}

func (m *mockDispatchSvc) Create(ctx context.Context, req dto.CreateDispatchRequest, requesterID string) (*model.Dispatch, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req, requesterID)
	}
	return &model.Dispatch{ID: "d1"}, nil
}

func (m *mockDispatchSvc) QuickBoard(ctx context.Context, req dto.QuickBoardRequest, dispatcherID string) (*model.Dispatch, error) {
	if m.quickBoardFn != nil {
		return m.quickBoardFn(ctx, req, dispatcherID)
	}
	return &model.Dispatch{ID: "d1"}, nil
}

func (m *mockDispatchSvc) GetByID(ctx context.Context, id string) (*model.Dispatch, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockDispatchSvc) List(ctx context.Context, status string, limit, offset int) ([]model.Dispatch, error) {
	if m.listFn != nil {
		return m.listFn(ctx, status, limit, offset)
	}
	return nil, nil
}

func (m *mockDispatchSvc) ListByRequester(ctx context.Context, requesterID, status string, limit, offset int) ([]model.Dispatch, error) {
	if m.listByRequesterFn != nil {
		return m.listByRequesterFn(ctx, requesterID, status, limit, offset)
	}
	return nil, nil
}

func (m *mockDispatchSvc) Assign(ctx context.Context, dispatchID, vehicleID, dispatcherID string) error {
	if m.assignFn != nil {
		return m.assignFn(ctx, dispatchID, vehicleID, dispatcherID)
	}
	return nil
}

func (m *mockDispatchSvc) UpdateStatus(ctx context.Context, dispatchID string, status model.DispatchStatus, actorID string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, dispatchID, status, actorID)
	}
	return nil
}

func (m *mockDispatchSvc) Cancel(ctx context.Context, dispatchID, reason, actorID string) error {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, dispatchID, reason, actorID)
	}
	return nil
}

func (m *mockDispatchSvc) GetCurrentTripByDriverID(ctx context.Context, driverID string) (*model.Dispatch, error) {
	if m.getCurrentTripFn != nil {
		return m.getCurrentTripFn(ctx, driverID)
	}
	return nil, nil
}

func (m *mockDispatchSvc) GetETASnapshots(ctx context.Context, dispatchID string) ([]model.DispatchETASnapshot, error) {
	if m.getETASnapshotsFn != nil {
		return m.getETASnapshotsFn(ctx, dispatchID)
	}
	return nil, nil
}

func (m *mockDispatchSvc) CalculateETAs(ctx context.Context, pickupLat, pickupLng float64) ([]dto.VehicleETA, error) {
	if m.calculateETAsFn != nil {
		return m.calculateETAsFn(ctx, pickupLat, pickupLng)
	}
	return nil, nil
}

func (m *mockDispatchSvc) RateDispatch(ctx context.Context, dispatchID string, rating int, comment string) error {
	if m.rateDispatchFn != nil {
		return m.rateDispatchFn(ctx, dispatchID, rating, comment)
	}
	return nil
}

// ── Mock: vehicleService ──

type mockVehicleSvc struct {
	listWithStatusFn    func(ctx context.Context) ([]model.VehicleWithStatus, error)
	getByIDFn           func(ctx context.Context, id string) (*model.Vehicle, error)
	getByDriverIDFn     func(ctx context.Context, driverID string) (*model.Vehicle, error)
	listAvailableFn     func(ctx context.Context) ([]model.VehicleWithStatus, error)
	createFn            func(ctx context.Context, actorID, name, licensePlate, driverID string) (*model.Vehicle, error)
	updateFn            func(ctx context.Context, actorID, vehicleID, name, licensePlate, driverID string) error
	deleteFn            func(ctx context.Context, actorID, vehicleID string) error
	updatePhotoURLFn    func(ctx context.Context, vehicleID string, photoURL *string) error
	toggleMaintenanceFn func(ctx context.Context, actorID, vehicleID string, maintenance bool) error
}

func (m *mockVehicleSvc) ListWithStatus(ctx context.Context) ([]model.VehicleWithStatus, error) {
	if m.listWithStatusFn != nil {
		return m.listWithStatusFn(ctx)
	}
	return nil, nil
}

func (m *mockVehicleSvc) GetByID(ctx context.Context, id string) (*model.Vehicle, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockVehicleSvc) GetByDriverID(ctx context.Context, driverID string) (*model.Vehicle, error) {
	if m.getByDriverIDFn != nil {
		return m.getByDriverIDFn(ctx, driverID)
	}
	return nil, nil
}

func (m *mockVehicleSvc) ListAvailable(ctx context.Context) ([]model.VehicleWithStatus, error) {
	if m.listAvailableFn != nil {
		return m.listAvailableFn(ctx)
	}
	return nil, nil
}

func (m *mockVehicleSvc) Create(ctx context.Context, actorID, name, licensePlate, driverID string) (*model.Vehicle, error) {
	if m.createFn != nil {
		return m.createFn(ctx, actorID, name, licensePlate, driverID)
	}
	return &model.Vehicle{ID: "v1"}, nil
}

func (m *mockVehicleSvc) Update(ctx context.Context, actorID, vehicleID, name, licensePlate, driverID string) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, actorID, vehicleID, name, licensePlate, driverID)
	}
	return nil
}

func (m *mockVehicleSvc) Delete(ctx context.Context, actorID, vehicleID string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, actorID, vehicleID)
	}
	return nil
}

func (m *mockVehicleSvc) UpdatePhotoURL(ctx context.Context, vehicleID string, photoURL *string) error {
	if m.updatePhotoURLFn != nil {
		return m.updatePhotoURLFn(ctx, vehicleID, photoURL)
	}
	return nil
}

func (m *mockVehicleSvc) ToggleMaintenance(ctx context.Context, actorID, vehicleID string, maintenance bool) error {
	if m.toggleMaintenanceFn != nil {
		return m.toggleMaintenanceFn(ctx, actorID, vehicleID, maintenance)
	}
	return nil
}

// ── Mock: locationService ──

type mockLocationSvc struct {
	reportFn    func(ctx context.Context, vehicleID string, points []model.LocationPoint) error
	getHistoryFn func(ctx context.Context, vehicleID string, from, to time.Time) ([]model.VehicleLocation, error)
}

func (m *mockLocationSvc) ReportLocations(ctx context.Context, vehicleID string, points []model.LocationPoint) error {
	if m.reportFn != nil {
		return m.reportFn(ctx, vehicleID, points)
	}
	return nil
}

func (m *mockLocationSvc) GetHistory(ctx context.Context, vehicleID string, from, to time.Time) ([]model.VehicleLocation, error) {
	if m.getHistoryFn != nil {
		return m.getHistoryFn(ctx, vehicleID, from, to)
	}
	return nil, nil
}

// ── Mock: attendanceService ──

type mockAttendanceSvc struct {
	clockInFn      func(context.Context, string) (*model.DriverAttendance, error)
	clockOutFn     func(context.Context, string) error
	updateStatusFn func(context.Context, string, model.DriverStatus) (*model.DriverAttendance, error)
	getStatusFn    func(context.Context, string) (*model.DriverAttendance, error)
	getHistoryFn   func(context.Context, string, int) ([]model.DriverAttendance, error)
}

func (m *mockAttendanceSvc) ClockIn(ctx context.Context, did string) (*model.DriverAttendance, error) {
	if m.clockInFn != nil {
		return m.clockInFn(ctx, did)
	}
	return &model.DriverAttendance{}, nil
}

func (m *mockAttendanceSvc) ClockOut(ctx context.Context, did string) error {
	if m.clockOutFn != nil {
		return m.clockOutFn(ctx, did)
	}
	return nil
}

func (m *mockAttendanceSvc) UpdateDriverStatus(ctx context.Context, did string, s model.DriverStatus) (*model.DriverAttendance, error) {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, did, s)
	}
	return &model.DriverAttendance{}, nil
}

func (m *mockAttendanceSvc) GetStatus(ctx context.Context, did string) (*model.DriverAttendance, error) {
	if m.getStatusFn != nil {
		return m.getStatusFn(ctx, did)
	}
	return nil, nil
}

func (m *mockAttendanceSvc) GetHistory(ctx context.Context, did string, limit int) ([]model.DriverAttendance, error) {
	if m.getHistoryFn != nil {
		return m.getHistoryFn(ctx, did, limit)
	}
	return nil, nil
}

// ── Mock: userRepository ──

type mockUserRepo struct {
	listFn          func(ctx context.Context) ([]model.User, error)
	getByIDFn       func(ctx context.Context, id string) (*model.User, error)
	updateRoleFn    func(ctx context.Context, id string, role model.Role) error
	updatePriorityFn func(ctx context.Context, id string, priority int) error
	updateFCMTokenFn func(ctx context.Context, id string, token string) error
}

func (m *mockUserRepo) List(ctx context.Context) ([]model.User, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, id string, role model.Role) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, id, role)
	}
	return nil
}

func (m *mockUserRepo) UpdatePriority(ctx context.Context, id string, priority int) error {
	if m.updatePriorityFn != nil {
		return m.updatePriorityFn(ctx, id, priority)
	}
	return nil
}

func (m *mockUserRepo) UpdateFCMToken(ctx context.Context, id string, token string) error {
	if m.updateFCMTokenFn != nil {
		return m.updateFCMTokenFn(ctx, id, token)
	}
	return nil
}

// ── Mock: auditLogger ──

type mockAuditSvc struct {
	logFn     func(ctx context.Context, actorID, action, targetType, targetID string, before, after interface{}, reason string)
	listFn    func(ctx context.Context, actorID, action, targetType string, from, to time.Time, limit, offset int) ([]model.AuditLog, error)
	getByIDFn func(ctx context.Context, id string) (*model.AuditLog, error)
}

func (m *mockAuditSvc) Log(ctx context.Context, actorID, action, targetType, targetID string, before, after interface{}, reason string) {
	if m.logFn != nil {
		m.logFn(ctx, actorID, action, targetType, targetID, before, after, reason)
	}
}

func (m *mockAuditSvc) List(ctx context.Context, actorID, action, targetType string, from, to time.Time, limit, offset int) ([]model.AuditLog, error) {
	if m.listFn != nil {
		return m.listFn(ctx, actorID, action, targetType, from, to, limit, offset)
	}
	return nil, nil
}

func (m *mockAuditSvc) GetByID(ctx context.Context, id string) (*model.AuditLog, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

// ── Mock: routeComputer ──

type mockRouteComputer struct {
	computeFn func(ctx context.Context, origin, destination maps.LatLng, intermediates []maps.LatLng) (*maps.RouteResult, error)
}

func (m *mockRouteComputer) ComputeRoute(ctx context.Context, origin, destination maps.LatLng, intermediates []maps.LatLng) (*maps.RouteResult, error) {
	if m.computeFn != nil {
		return m.computeFn(ctx, origin, destination, intermediates)
	}
	return &maps.RouteResult{}, nil
}

// ── Mock: bookingService ──

type mockBookingSvc struct {
	createBookingFn         func(ctx context.Context, req dto.UnifiedBookingRequest, requesterID string, priorityLevel int) (*dto.UnifiedBookingResponse, error)
	driverAcceptFn          func(ctx context.Context, reservationID, driverID string) error
	driverDeclineFn         func(ctx context.Context, reservationID, driverID, reason string) error
	getVehicleTimelineFn    func(ctx context.Context, vehicleID string, date time.Time) ([]model.ReservationWithDetails, error)
	getPendingByDriverIDFn  func(ctx context.Context, driverID string) ([]model.ReservationWithDetails, error)
}

func (m *mockBookingSvc) CreateBooking(ctx context.Context, req dto.UnifiedBookingRequest, requesterID string, priorityLevel int) (*dto.UnifiedBookingResponse, error) {
	if m.createBookingFn != nil {
		return m.createBookingFn(ctx, req, requesterID, priorityLevel)
	}
	return &dto.UnifiedBookingResponse{}, nil
}

func (m *mockBookingSvc) DriverAcceptReservation(ctx context.Context, reservationID, driverID string) error {
	if m.driverAcceptFn != nil {
		return m.driverAcceptFn(ctx, reservationID, driverID)
	}
	return nil
}

func (m *mockBookingSvc) DriverDeclineReservation(ctx context.Context, reservationID, driverID, reason string) error {
	if m.driverDeclineFn != nil {
		return m.driverDeclineFn(ctx, reservationID, driverID, reason)
	}
	return nil
}

func (m *mockBookingSvc) GetVehicleTimeline(ctx context.Context, vehicleID string, date time.Time) ([]model.ReservationWithDetails, error) {
	if m.getVehicleTimelineFn != nil {
		return m.getVehicleTimelineFn(ctx, vehicleID, date)
	}
	return nil, nil
}

func (m *mockBookingSvc) GetPendingByDriverID(ctx context.Context, driverID string) ([]model.ReservationWithDetails, error) {
	if m.getPendingByDriverIDFn != nil {
		return m.getPendingByDriverIDFn(ctx, driverID)
	}
	return nil, nil
}

// ── Mock: tokenService ──

type mockTokenSvc struct {
	blacklistFn    func(ctx context.Context, jti, userID string, expiresAt time.Time) error
	isBlacklistedFn func(ctx context.Context, jti string) (bool, error)
}

func (m *mockTokenSvc) Blacklist(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	if m.blacklistFn != nil {
		return m.blacklistFn(ctx, jti, userID, expiresAt)
	}
	return nil
}

func (m *mockTokenSvc) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	if m.isBlacklistedFn != nil {
		return m.isBlacklistedFn(ctx, jti)
	}
	return false, nil
}

// ── Mock: loginLimiter ──

type mockLoginLimiter struct {
	isLockedFn      func(account string) bool
	recordFailureFn func(account string)
	recordSuccessFn func(account string)
}

func (m *mockLoginLimiter) IsLocked(account string) bool {
	if m.isLockedFn != nil {
		return m.isLockedFn(account)
	}
	return false
}

func (m *mockLoginLimiter) RecordFailure(account string) {
	if m.recordFailureFn != nil {
		m.recordFailureFn(account)
	}
}

func (m *mockLoginLimiter) RecordSuccess(account string) {
	if m.recordSuccessFn != nil {
		m.recordSuccessFn(account)
	}
}

// ── Mock: passengerAuthService ──

type mockPassengerAuthSvc struct {
	registerFn func(ctx context.Context, req dto.PassengerRegisterRequest) (*dto.LoginResponse, error)
	loginFn    func(ctx context.Context, req dto.PassengerLoginRequest) (*dto.LoginResponse, error)
}

func (m *mockPassengerAuthSvc) RegisterPassenger(ctx context.Context, req dto.PassengerRegisterRequest) (*dto.LoginResponse, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, req)
	}
	return &dto.LoginResponse{}, nil
}

func (m *mockPassengerAuthSvc) LoginByPhone(ctx context.Context, req dto.PassengerLoginRequest) (*dto.LoginResponse, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, req)
	}
	return &dto.LoginResponse{}, nil
}
