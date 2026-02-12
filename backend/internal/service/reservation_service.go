package service

import (
	"context"
	"time"

	"github.com/kento/driver/backend/internal/dto"
	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/pkg/apperror"
)

type ReservationService struct {
	repo         *repository.ReservationRepo
	conflictRepo *repository.ConflictRepo
	auditSvc     *AuditService
}

func NewReservationService(repo *repository.ReservationRepo, conflictRepo *repository.ConflictRepo, auditSvc *AuditService) *ReservationService {
	return &ReservationService{repo: repo, conflictRepo: conflictRepo, auditSvc: auditSvc}
}

func (s *ReservationService) Create(ctx context.Context, req dto.CreateReservationRequest, requesterID string, priorityLevel int) (*model.Reservation, error) {
	if req.EndTime.Before(req.StartTime) || req.EndTime.Equal(req.StartTime) {
		return nil, apperror.New(400, "INVALID_TIME_RANGE", "end_time must be after start_time")
	}
	if req.StartTime.Before(time.Now()) {
		return nil, apperror.New(400, "PAST_TIME", "cannot create reservation in the past")
	}

	// Check for overlapping reservations
	overlaps, err := s.repo.FindOverlapping(ctx, req.VehicleID, req.StartTime, req.EndTime, "")
	if err != nil {
		return nil, err
	}

	reservation := &model.Reservation{
		VehicleID:     req.VehicleID,
		RequesterID:   requesterID,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		Purpose:       req.Purpose,
		Destinations:  req.Destinations,
		Notes:         req.Notes,
		PriorityLevel: priorityLevel,
		Status:        model.ReservationStatusConfirmed,
	}

	// Apply priority rules for conflicts
	for _, existing := range overlaps {
		if priorityLevel > existing.PriorityLevel {
			// New reservation has higher priority: both become pending_conflict
			reservation.Status = model.ReservationStatusPendingConflict
			// Mark existing as pending_conflict too
			_ = s.repo.UpdateStatus(ctx, existing.ID, model.ReservationStatusPendingConflict)
		} else if priorityLevel == existing.PriorityLevel {
			// Same priority: first-come-first-served, new one loses
			reservation.Status = model.ReservationStatusPendingConflict
		} else {
			// New reservation has lower priority: it loses
			reservation.Status = model.ReservationStatusPendingConflict
		}
	}

	if err := s.repo.Create(ctx, reservation); err != nil {
		return nil, err
	}

	// Create conflict records
	for _, existing := range overlaps {
		if priorityLevel > existing.PriorityLevel {
			// New wins
			_, _ = s.conflictRepo.Create(ctx, reservation.ID, existing.ID)
		} else {
			// Existing wins
			_, _ = s.conflictRepo.Create(ctx, existing.ID, reservation.ID)
		}
	}

	s.auditSvc.Log(ctx, requesterID, "reservation.create", "reservation", reservation.ID, nil, reservation, "")
	return reservation, nil
}

func (s *ReservationService) GetByID(ctx context.Context, id string) (*model.Reservation, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ReservationService) List(ctx context.Context, vehicleID string, from, to time.Time, status string, limit, offset int) ([]model.ReservationWithDetails, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(ctx, vehicleID, from, to, status, limit, offset)
}

func (s *ReservationService) Cancel(ctx context.Context, id, cancelledBy, reason string) error {
	before, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if before == nil {
		return apperror.ErrNotFound
	}

	if err := s.repo.Cancel(ctx, id, cancelledBy, reason); err != nil {
		return err
	}

	s.auditSvc.Log(ctx, cancelledBy, "reservation.cancel", "reservation", id, before, nil, reason)
	return nil
}

func (s *ReservationService) Update(ctx context.Context, id string, req dto.UpdateReservationRequest, actorID string) (*model.Reservation, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.ErrNotFound
	}

	before := *existing

	if req.VehicleID != nil {
		existing.VehicleID = *req.VehicleID
	}
	if req.StartTime != nil {
		existing.StartTime = *req.StartTime
	}
	if req.EndTime != nil {
		existing.EndTime = *req.EndTime
	}
	if req.Purpose != nil {
		existing.Purpose = *req.Purpose
	}
	if req.Destinations != nil {
		existing.Destinations = req.Destinations
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.auditSvc.Log(ctx, actorID, "reservation.update", "reservation", id, before, existing, "")
	return existing, nil
}

func (s *ReservationService) CheckAvailability(ctx context.Context, vehicleID string, startTime, endTime time.Time) ([]model.Reservation, error) {
	return s.repo.FindOverlapping(ctx, vehicleID, startTime, endTime, "")
}

func (s *ReservationService) GetUpcomingReminders(ctx context.Context, minutesBefore int) ([]model.ReservationWithDetails, error) {
	return s.repo.GetUpcomingReminders(ctx, minutesBefore)
}

func (s *ReservationService) AutoCompleteExpired(ctx context.Context) (int64, error) {
	return s.repo.AutoCompleteExpired(ctx)
}
