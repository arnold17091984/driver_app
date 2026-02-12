package service

import (
	"context"

	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/pkg/apperror"
)

type AttendanceService struct {
	repo     *repository.AttendanceRepo
	auditSvc *AuditService
}

func NewAttendanceService(repo *repository.AttendanceRepo, auditSvc *AuditService) *AttendanceService {
	return &AttendanceService{repo: repo, auditSvc: auditSvc}
}

func (s *AttendanceService) ClockIn(ctx context.Context, driverID string) (*model.DriverAttendance, error) {
	existing, err := s.repo.GetActiveByDriverID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.New(400, "ALREADY_CLOCKED_IN", "driver is already clocked in")
	}

	a, err := s.repo.ClockIn(ctx, driverID)
	if err != nil {
		return nil, err
	}
	s.auditSvc.Log(ctx, driverID, "attendance.clock_in", "attendance", a.ID, nil, a, "")
	return a, nil
}

func (s *AttendanceService) ClockOut(ctx context.Context, driverID string) error {
	existing, err := s.repo.GetActiveByDriverID(ctx, driverID)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.New(400, "NOT_CLOCKED_IN", "driver is not clocked in")
	}

	err = s.repo.ClockOut(ctx, existing.ID)
	if err != nil {
		return err
	}
	s.auditSvc.Log(ctx, driverID, "attendance.clock_out", "attendance", existing.ID, existing, nil, "")
	return nil
}

func (s *AttendanceService) UpdateDriverStatus(ctx context.Context, driverID string, status model.DriverStatus) (*model.DriverAttendance, error) {
	existing, err := s.repo.GetActiveByDriverID(ctx, driverID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.New(400, "NOT_CLOCKED_IN", "driver is not clocked in")
	}

	oldStatus := existing.DriverStatus
	if err := s.repo.UpdateDriverStatus(ctx, existing.ID, status); err != nil {
		return nil, err
	}
	existing.DriverStatus = status

	s.auditSvc.Log(ctx, driverID, "attendance.update_status", "attendance", existing.ID,
		map[string]interface{}{"driver_status": oldStatus},
		map[string]interface{}{"driver_status": status}, "")
	return existing, nil
}

func (s *AttendanceService) GetStatus(ctx context.Context, driverID string) (*model.DriverAttendance, error) {
	return s.repo.GetActiveByDriverID(ctx, driverID)
}

func (s *AttendanceService) GetHistory(ctx context.Context, driverID string, limit int) ([]model.DriverAttendance, error) {
	return s.repo.ListByDriverID(ctx, driverID, limit)
}
