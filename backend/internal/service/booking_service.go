package service

import (
	"context"
	"time"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/notify"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/pkg/apperror"
)

type BookingService struct {
	dispatchSvc     *DispatchService
	reservationSvc  *ReservationService
	vehicleRepo     *repository.VehicleRepo
	reservationRepo *repository.ReservationRepo
	auditSvc        *AuditService
	fcmSvc          *notify.FCMService
}

func NewBookingService(
	dispatchSvc *DispatchService,
	reservationSvc *ReservationService,
	vehicleRepo *repository.VehicleRepo,
	reservationRepo *repository.ReservationRepo,
	auditSvc *AuditService,
	fcmSvc *notify.FCMService,
) *BookingService {
	return &BookingService{
		dispatchSvc:     dispatchSvc,
		reservationSvc:  reservationSvc,
		vehicleRepo:     vehicleRepo,
		reservationRepo: reservationRepo,
		auditSvc:        auditSvc,
		fcmSvc:          fcmSvc,
	}
}

func (s *BookingService) CreateBooking(ctx context.Context, req dto.UnifiedBookingRequest, requesterID string, priorityLevel int) (*dto.UnifiedBookingResponse, error) {
	if req.IsNow {
		return s.createNowBooking(ctx, req, requesterID)
	}
	return s.createFutureBooking(ctx, req, requesterID, priorityLevel)
}

func (s *BookingService) createNowBooking(ctx context.Context, req dto.UnifiedBookingRequest, requesterID string) (*dto.UnifiedBookingResponse, error) {
	// For immediate dispatch, use the first destination as the dropoff address
	var dropoff *string
	if len(req.Destinations) > 0 {
		dropoff = &req.Destinations[0]
	}
	dispatchReq := dto.CreateDispatchRequest{
		Purpose:        req.Purpose,
		PassengerName:  req.PassengerName,
		PickupAddress:  req.PickupAddress,
		PickupLat:      req.PickupLat,
		PickupLng:      req.PickupLng,
		DropoffAddress: dropoff,
		Notes:          req.Notes,
		PassengerCount: 1,
	}

	dispatch, err := s.dispatchSvc.Create(ctx, dispatchReq, requesterID)
	if err != nil {
		return nil, err
	}

	// If specific vehicle requested, assign immediately
	if req.Mode == "specific" && req.VehicleID != nil {
		if err := s.dispatchSvc.Assign(ctx, dispatch.ID, *req.VehicleID, requesterID); err != nil {
			return nil, err
		}
		// Re-fetch to get updated state
		dispatch, _ = s.dispatchSvc.GetByID(ctx, dispatch.ID)
	}

	return &dto.UnifiedBookingResponse{
		Type:     "dispatch",
		Dispatch: dispatch,
	}, nil
}

func (s *BookingService) createFutureBooking(ctx context.Context, req dto.UnifiedBookingRequest, requesterID string, priorityLevel int) (*dto.UnifiedBookingResponse, error) {
	if req.StartTime == nil || req.EndTime == nil {
		return nil, apperror.New(400, "MISSING_TIME", "start_time and end_time are required for future bookings")
	}
	if req.Mode == "specific" && (req.VehicleID == nil || *req.VehicleID == "") {
		return nil, apperror.New(400, "MISSING_VEHICLE", "vehicle_id is required when mode is specific")
	}

	var vehicleID string
	if req.Mode == "specific" {
		vehicleID = *req.VehicleID
	} else {
		// Find an available vehicle for the time slot
		vehicleIDs, err := s.reservationRepo.FindAvailableVehicleForSlot(ctx, *req.StartTime, *req.EndTime, nil)
		if err != nil {
			return nil, err
		}
		if len(vehicleIDs) == 0 {
			return nil, apperror.New(404, "NO_VEHICLE_AVAILABLE", "no vehicles available for this time slot")
		}
		vehicleID = vehicleIDs[0]
	}

	reservation := &model.Reservation{
		VehicleID:     vehicleID,
		RequesterID:   requesterID,
		StartTime:     *req.StartTime,
		EndTime:       *req.EndTime,
		Purpose:       req.Purpose,
		Destinations:  req.Destinations,
		Notes:         req.Notes,
		PassengerName: req.PassengerName,
		PickupAddress: &req.PickupAddress,
		PickupLat:     req.PickupLat,
		PickupLng:     req.PickupLng,
		PriorityLevel: priorityLevel,
		Status:        model.ReservationStatusPendingDriver,
	}

	if err := s.reservationRepo.Create(ctx, reservation); err != nil {
		return nil, err
	}

	s.auditSvc.Log(ctx, requesterID, "reservation.create", "reservation", reservation.ID, nil, reservation, "")

	// Notify the driver of the assigned vehicle
	go s.fcmSvc.NotifyVehicleDriver(ctx, vehicleID, "Reservation Pending", reservation.Purpose, map[string]string{
		"type": "reservation_pending", "reservation_id": reservation.ID,
	})

	return &dto.UnifiedBookingResponse{
		Type:        "reservation",
		Reservation: reservation,
	}, nil
}

func (s *BookingService) DriverAcceptReservation(ctx context.Context, reservationID, driverID string) error {
	res, err := s.reservationRepo.GetByID(ctx, reservationID)
	if err != nil {
		return err
	}
	if res == nil {
		return apperror.ErrNotFound
	}
	if res.Status != model.ReservationStatusPendingDriver {
		return apperror.New(400, "INVALID_STATUS", "reservation is not pending driver acceptance")
	}

	// Verify the vehicle belongs to this driver
	vehicle, err := s.vehicleRepo.GetByDriverID(ctx, driverID)
	if err != nil {
		return err
	}
	if vehicle == nil || vehicle.ID != res.VehicleID {
		return apperror.New(403, "NOT_YOUR_VEHICLE", "this reservation is not assigned to your vehicle")
	}

	if err := s.reservationRepo.UpdateStatus(ctx, reservationID, model.ReservationStatusConfirmed); err != nil {
		return err
	}

	s.auditSvc.Log(ctx, driverID, "reservation.driver_accept", "reservation", reservationID, res, nil, "")

	// Notify the requester that their reservation was accepted
	go s.fcmSvc.NotifyUser(ctx, res.RequesterID, "Reservation Confirmed", res.Purpose, map[string]string{
		"type": "reservation_confirmed", "reservation_id": reservationID,
	})

	return nil
}

func (s *BookingService) DriverDeclineReservation(ctx context.Context, reservationID, driverID, reason string) error {
	res, err := s.reservationRepo.GetByID(ctx, reservationID)
	if err != nil {
		return err
	}
	if res == nil {
		return apperror.ErrNotFound
	}
	if res.Status != model.ReservationStatusPendingDriver {
		return apperror.New(400, "INVALID_STATUS", "reservation is not pending driver acceptance")
	}

	// Verify the vehicle belongs to this driver
	vehicle, err := s.vehicleRepo.GetByDriverID(ctx, driverID)
	if err != nil {
		return err
	}
	if vehicle == nil || vehicle.ID != res.VehicleID {
		return apperror.New(403, "NOT_YOUR_VEHICLE", "this reservation is not assigned to your vehicle")
	}

	// Record the declined vehicle ID for exclusion in auto-reassignment.
	// We store vehicle.ID (not driverID) because FindAvailableVehicleForSlot
	// excludes by vehicle ID.
	if err := s.reservationRepo.AddDeclinedDriver(ctx, reservationID, vehicle.ID); err != nil {
		return err
	}

	s.auditSvc.Log(ctx, driverID, "reservation.driver_decline", "reservation", reservationID, res, nil, reason)

	// Auto-reassign to next available vehicle
	return s.autoReassign(ctx, reservationID, res)
}

func (s *BookingService) autoReassign(ctx context.Context, reservationID string, res *model.Reservation) error {
	// Build exclude list: current vehicle + all previously declined vehicles
	excludeIDs := []string{res.VehicleID}
	for _, id := range res.DeclinedByDriverIDs {
		excludeIDs = append(excludeIDs, id)
	}

	vehicleIDs, err := s.reservationRepo.FindAvailableVehicleForSlot(ctx, res.StartTime, res.EndTime, excludeIDs)
	if err != nil {
		return err
	}

	if len(vehicleIDs) == 0 {
		// No vehicles available - mark as driver_declined (terminal)
		return s.reservationRepo.UpdateStatus(ctx, reservationID, model.ReservationStatusDriverDeclined)
	}

	// Reassign to the first available vehicle
	if err := s.reservationRepo.UpdateVehicle(ctx, reservationID, vehicleIDs[0]); err != nil {
		return err
	}

	// Notify the new driver
	go s.fcmSvc.NotifyVehicleDriver(ctx, vehicleIDs[0], "Reservation Pending", res.Purpose, map[string]string{
		"type": "reservation_pending", "reservation_id": reservationID,
	})

	return nil
}

// GetVehicleTimeline returns reservations for a vehicle on a given date.
func (s *BookingService) GetVehicleTimeline(ctx context.Context, vehicleID string, date time.Time) ([]model.ReservationWithDetails, error) {
	return s.reservationRepo.GetDayReservations(ctx, vehicleID, date)
}

// GetPendingByDriverID returns pending reservations for a driver's vehicle.
func (s *BookingService) GetPendingByDriverID(ctx context.Context, driverID string) ([]model.ReservationWithDetails, error) {
	return s.reservationRepo.FindPendingByDriverID(ctx, driverID)
}
