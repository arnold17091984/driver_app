package service

import (
	"context"
	"time"

	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
)

type VehicleService struct {
	repo           *repository.VehicleRepo
	staleThreshold time.Duration
	auditSvc       *AuditService
}

func NewVehicleService(repo *repository.VehicleRepo, staleThreshold time.Duration, auditSvc *AuditService) *VehicleService {
	return &VehicleService{
		repo:           repo,
		staleThreshold: staleThreshold,
		auditSvc:       auditSvc,
	}
}

func (s *VehicleService) ListWithStatus(ctx context.Context) ([]model.VehicleWithStatus, error) {
	return s.repo.ListWithStatus(ctx, s.staleThreshold)
}

func (s *VehicleService) GetByID(ctx context.Context, id string) (*model.Vehicle, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *VehicleService) GetByDriverID(ctx context.Context, driverID string) (*model.Vehicle, error) {
	return s.repo.GetByDriverID(ctx, driverID)
}

func (s *VehicleService) ListAvailable(ctx context.Context) ([]model.VehicleWithStatus, error) {
	return s.repo.ListAvailable(ctx, s.staleThreshold)
}

func (s *VehicleService) Create(ctx context.Context, actorID, name, licensePlate, driverID string) (*model.Vehicle, error) {
	v, err := s.repo.Create(ctx, name, licensePlate, driverID)
	if err != nil {
		return nil, err
	}
	s.auditSvc.Log(ctx, actorID, "vehicle.create", "vehicle", v.ID, nil, v, "")
	return v, nil
}

func (s *VehicleService) Update(ctx context.Context, actorID, vehicleID, name, licensePlate, driverID string) error {
	before, _ := s.repo.GetByID(ctx, vehicleID)
	err := s.repo.Update(ctx, vehicleID, name, licensePlate, driverID)
	if err != nil {
		return err
	}
	after, _ := s.repo.GetByID(ctx, vehicleID)
	s.auditSvc.Log(ctx, actorID, "vehicle.update", "vehicle", vehicleID, before, after, "")
	return nil
}

func (s *VehicleService) Delete(ctx context.Context, actorID, vehicleID string) error {
	before, _ := s.repo.GetByID(ctx, vehicleID)
	err := s.repo.Delete(ctx, vehicleID)
	if err != nil {
		return err
	}
	s.auditSvc.Log(ctx, actorID, "vehicle.delete", "vehicle", vehicleID, before, nil, "")
	return nil
}

func (s *VehicleService) UpdatePhotoURL(ctx context.Context, vehicleID string, photoURL *string) error {
	return s.repo.UpdatePhotoURL(ctx, vehicleID, photoURL)
}

func (s *VehicleService) ToggleMaintenance(ctx context.Context, actorID, vehicleID string, maintenance bool) error {
	before, _ := s.repo.GetByID(ctx, vehicleID)
	err := s.repo.ToggleMaintenance(ctx, vehicleID, maintenance)
	if err != nil {
		return err
	}
	after, _ := s.repo.GetByID(ctx, vehicleID)
	s.auditSvc.Log(ctx, actorID, "vehicle.maintenance_toggle", "vehicle", vehicleID, before, after, "")
	return nil
}
