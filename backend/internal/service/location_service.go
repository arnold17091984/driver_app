package service

import (
	"context"
	"time"

	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
)

type LocationService struct {
	repo *repository.LocationRepo
}

func NewLocationService(repo *repository.LocationRepo) *LocationService {
	return &LocationService{repo: repo}
}

func (s *LocationService) ReportLocations(ctx context.Context, vehicleID string, points []model.LocationPoint) error {
	if len(points) == 0 {
		return nil
	}
	return s.repo.BatchInsert(ctx, vehicleID, points)
}

func (s *LocationService) GetHistory(ctx context.Context, vehicleID string, from, to time.Time) ([]model.VehicleLocation, error) {
	return s.repo.GetHistory(ctx, vehicleID, from, to)
}

func (s *LocationService) GetCurrent(ctx context.Context, vehicleID string) (*model.VehicleLocationCurrent, error) {
	return s.repo.GetCurrent(ctx, vehicleID)
}
