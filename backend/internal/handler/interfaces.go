package handler

import (
	"context"
	"time"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/maps"
	"github.com/kento/driver/backend/internal/model"
)

// Service interfaces used by handlers.
// Concrete types in the service and repository packages satisfy these implicitly.

type authService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	Refresh(ctx context.Context, refreshToken string) (*dto.RefreshResponse, error)
	GetUser(ctx context.Context, userID string) (*model.User, error)
}

type dispatchService interface {
	Create(ctx context.Context, req dto.CreateDispatchRequest, requesterID string) (*model.Dispatch, error)
	QuickBoard(ctx context.Context, req dto.QuickBoardRequest, dispatcherID string) (*model.Dispatch, error)
	GetByID(ctx context.Context, id string) (*model.Dispatch, error)
	List(ctx context.Context, status string, limit, offset int) ([]model.Dispatch, error)
	Assign(ctx context.Context, dispatchID, vehicleID, dispatcherID string) error
	UpdateStatus(ctx context.Context, dispatchID string, status model.DispatchStatus, actorID string) error
	Cancel(ctx context.Context, dispatchID, reason, actorID string) error
	GetCurrentTripByDriverID(ctx context.Context, driverID string) (*model.Dispatch, error)
	GetETASnapshots(ctx context.Context, dispatchID string) ([]model.DispatchETASnapshot, error)
	CalculateETAs(ctx context.Context, pickupLat, pickupLng float64) ([]dto.VehicleETA, error)
}

type vehicleService interface {
	ListWithStatus(ctx context.Context) ([]model.VehicleWithStatus, error)
	GetByID(ctx context.Context, id string) (*model.Vehicle, error)
	GetByDriverID(ctx context.Context, driverID string) (*model.Vehicle, error)
	ListAvailable(ctx context.Context) ([]model.VehicleWithStatus, error)
	Create(ctx context.Context, actorID, name, licensePlate, driverID string) (*model.Vehicle, error)
	Update(ctx context.Context, actorID, vehicleID, name, licensePlate, driverID string) error
	Delete(ctx context.Context, actorID, vehicleID string) error
	UpdatePhotoURL(ctx context.Context, vehicleID string, photoURL *string) error
	ToggleMaintenance(ctx context.Context, actorID, vehicleID string, maintenance bool) error
}

type locationService interface {
	ReportLocations(ctx context.Context, vehicleID string, points []model.LocationPoint) error
	GetHistory(ctx context.Context, vehicleID string, from, to time.Time) ([]model.VehicleLocation, error)
}

type attendanceService interface {
	ClockIn(ctx context.Context, driverID string) (*model.DriverAttendance, error)
	ClockOut(ctx context.Context, driverID string) error
	UpdateDriverStatus(ctx context.Context, driverID string, status model.DriverStatus) (*model.DriverAttendance, error)
	GetStatus(ctx context.Context, driverID string) (*model.DriverAttendance, error)
	GetHistory(ctx context.Context, driverID string, limit int) ([]model.DriverAttendance, error)
}

type bookingService interface {
	CreateBooking(ctx context.Context, req dto.UnifiedBookingRequest, requesterID string, priorityLevel int) (*dto.UnifiedBookingResponse, error)
	DriverAcceptReservation(ctx context.Context, reservationID, driverID string) error
	DriverDeclineReservation(ctx context.Context, reservationID, driverID, reason string) error
	GetVehicleTimeline(ctx context.Context, vehicleID string, date time.Time) ([]model.ReservationWithDetails, error)
	GetPendingByDriverID(ctx context.Context, driverID string) ([]model.ReservationWithDetails, error)
}

type reservationService interface {
	Create(ctx context.Context, req dto.CreateReservationRequest, requesterID string, priorityLevel int) (*model.Reservation, error)
	GetByID(ctx context.Context, id string) (*model.Reservation, error)
	List(ctx context.Context, vehicleID string, from, to time.Time, status string, limit, offset int) ([]model.ReservationWithDetails, error)
	Cancel(ctx context.Context, id, cancelledBy, reason string) error
	Update(ctx context.Context, id string, req dto.UpdateReservationRequest, actorID string) (*model.Reservation, error)
	CheckAvailability(ctx context.Context, vehicleID string, startTime, endTime time.Time) ([]model.Reservation, error)
}

type conflictService interface {
	ListPending(ctx context.Context) ([]model.ReservationConflict, error)
	GetByID(ctx context.Context, id string) (*model.ReservationConflict, error)
	ResolveReassign(ctx context.Context, conflictID, newVehicleID, resolvedBy, reason string) error
	ResolveChangeTime(ctx context.Context, conflictID, resolvedBy, reason string, losingRes *model.Reservation) error
	ResolveCancel(ctx context.Context, conflictID, resolvedBy, reason string) error
	ForceAssign(ctx context.Context, conflictID, resolvedBy, reason string) error
}

type userRepository interface {
	List(ctx context.Context) ([]model.User, error)
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdateRole(ctx context.Context, id string, role model.Role) error
	UpdatePriority(ctx context.Context, id string, priority int) error
	UpdateFCMToken(ctx context.Context, id string, token string) error
}

type auditLogger interface {
	Log(ctx context.Context, actorID, action, targetType, targetID string, before, after interface{}, reason string)
	List(ctx context.Context, actorID, action, targetType string, from, to time.Time, limit, offset int) ([]model.AuditLog, error)
	GetByID(ctx context.Context, id string) (*model.AuditLog, error)
}

type routeComputer interface {
	ComputeRoute(ctx context.Context, origin, destination maps.LatLng, intermediates []maps.LatLng) (*maps.RouteResult, error)
}
