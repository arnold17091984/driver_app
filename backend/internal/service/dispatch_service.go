package service

import (
	"context"
	"math"
	"math/rand"
	"time"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/notify"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/pkg/apperror"
)

type DispatchService struct {
	repo        *repository.DispatchRepo
	vehicleRepo *repository.VehicleRepo
	auditSvc    *AuditService
	staleThr    time.Duration
	fcmSvc      *notify.FCMService
}

func NewDispatchService(repo *repository.DispatchRepo, vehicleRepo *repository.VehicleRepo, auditSvc *AuditService, staleThr time.Duration, fcmSvc *notify.FCMService) *DispatchService {
	return &DispatchService{repo: repo, vehicleRepo: vehicleRepo, auditSvc: auditSvc, staleThr: staleThr, fcmSvc: fcmSvc}
}

func (s *DispatchService) Create(ctx context.Context, req dto.CreateDispatchRequest, requesterID string) (*model.Dispatch, error) {
	d := &model.Dispatch{
		RequesterID:    requesterID,
		Purpose:        req.Purpose,
		PassengerName:  req.PassengerName,
		PassengerCount: req.PassengerCount,
		Notes:          req.Notes,
		PickupAddress:  req.PickupAddress,
		PickupLat:      req.PickupLat,
		PickupLng:      req.PickupLng,
		DropoffAddress: req.DropoffAddress,
		DropoffLat:     req.DropoffLat,
		DropoffLng:     req.DropoffLng,
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}

	s.auditSvc.Log(ctx, requesterID, "dispatch.create", "dispatch", d.ID, nil, d, "")

	// Notify dispatchers of new dispatch
	go s.fcmSvc.NotifyRole(ctx, "New Dispatch", d.Purpose, map[string]string{
		"type": "dispatch_created", "dispatch_id": d.ID,
	}, model.RoleAdmin, model.RoleDispatcher)

	return d, nil
}

func (s *DispatchService) QuickBoard(ctx context.Context, req dto.QuickBoardRequest, dispatcherID string) (*model.Dispatch, error) {
	// Verify vehicle is not already on a trip
	vehicles, err := s.vehicleRepo.ListWithStatus(ctx, s.staleThr)
	if err != nil {
		return nil, err
	}
	var found bool
	for _, v := range vehicles {
		if v.ID == req.VehicleID {
			if v.Status == model.VehicleStatusInTrip {
				return nil, apperror.New(400, "VEHICLE_BUSY", "vehicle already has an active trip")
			}
			found = true
			break
		}
	}
	if !found {
		return nil, apperror.ErrNotFound
	}

	purpose := req.Purpose
	if purpose == "" {
		purpose = "乗車"
	}
	count := req.PassengerCount
	if count <= 0 {
		count = 1
	}

	d := &model.Dispatch{
		RequesterID:    dispatcherID,
		Purpose:        purpose,
		PassengerName:  &req.PassengerName,
		PassengerCount: count,
		Notes:          req.Notes,
		PickupAddress:  "（ルート未定）",
	}
	if req.EstimatedMinutes > 0 {
		endAt := time.Now().Add(time.Duration(req.EstimatedMinutes) * time.Minute)
		d.EstimatedEndAt = &endAt
	}

	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}

	// Assign + advance to en_route immediately
	if err := s.repo.Assign(ctx, d.ID, req.VehicleID, dispatcherID); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateStatus(ctx, d.ID, model.DispatchStatusAccepted); err != nil {
		return nil, err
	}
	if err := s.repo.UpdateStatus(ctx, d.ID, model.DispatchStatusEnRoute); err != nil {
		return nil, err
	}

	result, _ := s.repo.GetByID(ctx, d.ID)
	s.auditSvc.Log(ctx, dispatcherID, "dispatch.quick_board", "dispatch", d.ID, nil, result, "")
	return result, nil
}

func (s *DispatchService) GetByID(ctx context.Context, id string) (*model.Dispatch, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DispatchService) List(ctx context.Context, status string, limit, offset int) ([]model.Dispatch, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(ctx, status, limit, offset)
}

func (s *DispatchService) Assign(ctx context.Context, dispatchID, vehicleID, dispatcherID string) error {
	before, err := s.repo.GetByID(ctx, dispatchID)
	if err != nil {
		return err
	}
	if before == nil {
		return apperror.ErrNotFound
	}
	if before.Status != model.DispatchStatusPending {
		return apperror.New(400, "INVALID_STATUS", "dispatch is not in pending status")
	}

	if err := s.repo.Assign(ctx, dispatchID, vehicleID, dispatcherID); err != nil {
		return err
	}

	after, _ := s.repo.GetByID(ctx, dispatchID)
	s.auditSvc.Log(ctx, dispatcherID, "dispatch.assign", "dispatch", dispatchID, before, after, "")

	// Notify the driver of the assigned vehicle
	go s.fcmSvc.NotifyVehicleDriver(ctx, vehicleID, "Trip Assigned", "You have been assigned a new trip", map[string]string{
		"type": "dispatch_assigned", "dispatch_id": dispatchID,
	})

	return nil
}

func (s *DispatchService) UpdateStatus(ctx context.Context, dispatchID string, status model.DispatchStatus, actorID string) error {
	before, err := s.repo.GetByID(ctx, dispatchID)
	if err != nil {
		return err
	}
	if before == nil {
		return apperror.ErrNotFound
	}

	if err := s.repo.UpdateStatus(ctx, dispatchID, status); err != nil {
		return err
	}

	after, _ := s.repo.GetByID(ctx, dispatchID)
	s.auditSvc.Log(ctx, actorID, "dispatch.status_change", "dispatch", dispatchID, before, after, "")
	return nil
}

func (s *DispatchService) Cancel(ctx context.Context, dispatchID, reason, actorID string) error {
	before, err := s.repo.GetByID(ctx, dispatchID)
	if err != nil {
		return err
	}
	if before == nil {
		return apperror.ErrNotFound
	}

	if err := s.repo.Cancel(ctx, dispatchID, reason); err != nil {
		return err
	}

	after, _ := s.repo.GetByID(ctx, dispatchID)
	s.auditSvc.Log(ctx, actorID, "dispatch.cancel", "dispatch", dispatchID, before, after, reason)
	return nil
}

func (s *DispatchService) GetCurrentTripByDriverID(ctx context.Context, driverID string) (*model.Dispatch, error) {
	return s.repo.GetActiveByDriverID(ctx, driverID)
}

func (s *DispatchService) GetETASnapshots(ctx context.Context, dispatchID string) ([]model.DispatchETASnapshot, error) {
	return s.repo.GetETASnapshots(ctx, dispatchID)
}

// CalculateETAs returns ETA estimates for all available vehicles to a given pickup point.
// Uses haversine distance with simulated Manila traffic speed (~18 km/h avg).
func (s *DispatchService) CalculateETAs(ctx context.Context, pickupLat, pickupLng float64) ([]dto.VehicleETA, error) {
	vehicles, err := s.vehicleRepo.ListWithStatus(ctx, s.staleThr)
	if err != nil {
		return nil, err
	}

	// Manila default center for vehicles without GPS
	defaultLat, defaultLng := 14.5547, 121.0244
	offsets := [][2]float64{{0.005, 0.003}, {-0.003, 0.008}, {0.008, -0.004}, {-0.006, -0.006}, {0.002, 0.012}}

	var results []dto.VehicleETA
	for i, v := range vehicles {
		vLat := v.Latitude
		vLng := v.Longitude
		if vLat == nil {
			lat := defaultLat + offsets[i%len(offsets)][0]
			vLat = &lat
		}
		if vLng == nil {
			lng := defaultLng + offsets[i%len(offsets)][1]
			vLng = &lng
		}

		distM := haversineDistance(*vLat, *vLng, pickupLat, pickupLng)
		// Manila average traffic speed: 15-25 km/h, simulate with some randomness
		speedKmh := 15.0 + rand.Float64()*10.0
		durationSec := int(math.Round(distM / (speedKmh * 1000 / 3600)))
		if durationSec < 60 {
			durationSec = 60 + rand.Intn(120) // minimum 1-3 min
		}

		results = append(results, dto.VehicleETA{
			VehicleID:   v.ID,
			VehicleName: v.Name,
			DriverName:  v.DriverName,
			Plate:       v.LicensePlate,
			Status:      string(v.Status),
			Latitude:    *vLat,
			Longitude:   *vLng,
			DistanceM:   int(math.Round(distM)),
			DurationSec: durationSec,
			IsAvailable: string(v.Status) == "available",
		})
	}

	return results, nil
}

func haversineDistance(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}
