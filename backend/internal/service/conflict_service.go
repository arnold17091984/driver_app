package service

import (
	"context"

	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
	"github.com/kento/driver/backend/pkg/apperror"
)

type ConflictService struct {
	conflictRepo    *repository.ConflictRepo
	reservationRepo *repository.ReservationRepo
	auditSvc        *AuditService
}

func NewConflictService(conflictRepo *repository.ConflictRepo, reservationRepo *repository.ReservationRepo, auditSvc *AuditService) *ConflictService {
	return &ConflictService{
		conflictRepo:    conflictRepo,
		reservationRepo: reservationRepo,
		auditSvc:        auditSvc,
	}
}

func (s *ConflictService) ListPending(ctx context.Context) ([]model.ReservationConflict, error) {
	return s.conflictRepo.ListPending(ctx)
}

func (s *ConflictService) GetByID(ctx context.Context, id string) (*model.ReservationConflict, error) {
	return s.conflictRepo.GetByID(ctx, id)
}

func (s *ConflictService) ResolveReassign(ctx context.Context, conflictID, newVehicleID, resolvedBy, reason string) error {
	conflict, err := s.conflictRepo.GetByID(ctx, conflictID)
	if err != nil {
		return err
	}
	if conflict == nil {
		return apperror.ErrNotFound
	}
	if conflict.Status != model.ConflictStatusPending {
		return apperror.New(400, "ALREADY_RESOLVED", "conflict is already resolved")
	}

	losingRes, err := s.reservationRepo.GetByID(ctx, conflict.LosingReservationID)
	if err != nil {
		return err
	}

	// Update the losing reservation's vehicle
	losingRes.VehicleID = newVehicleID
	if err := s.reservationRepo.Update(ctx, losingRes); err != nil {
		return err
	}

	// Mark losing reservation as confirmed
	if err := s.reservationRepo.UpdateStatus(ctx, conflict.LosingReservationID, model.ReservationStatusConfirmed); err != nil {
		return err
	}

	// Mark winning reservation as confirmed (if it was pending_conflict)
	_ = s.reservationRepo.UpdateStatus(ctx, conflict.WinningReservationID, model.ReservationStatusConfirmed)

	// Resolve the conflict
	if err := s.conflictRepo.Resolve(ctx, conflictID, resolvedBy, reason, model.ConflictStatusResolvedReassign); err != nil {
		return err
	}

	s.auditSvc.Log(ctx, resolvedBy, "conflict.resolve_reassign", "conflict", conflictID, conflict, nil, reason)
	return nil
}

func (s *ConflictService) ResolveChangeTime(ctx context.Context, conflictID, resolvedBy, reason string, losingRes *model.Reservation) error {
	conflict, err := s.conflictRepo.GetByID(ctx, conflictID)
	if err != nil {
		return err
	}
	if conflict == nil {
		return apperror.ErrNotFound
	}
	if conflict.Status != model.ConflictStatusPending {
		return apperror.New(400, "ALREADY_RESOLVED", "conflict is already resolved")
	}

	if err := s.reservationRepo.Update(ctx, losingRes); err != nil {
		return err
	}

	if err := s.reservationRepo.UpdateStatus(ctx, conflict.LosingReservationID, model.ReservationStatusConfirmed); err != nil {
		return err
	}

	_ = s.reservationRepo.UpdateStatus(ctx, conflict.WinningReservationID, model.ReservationStatusConfirmed)

	if err := s.conflictRepo.Resolve(ctx, conflictID, resolvedBy, reason, model.ConflictStatusResolvedChanged); err != nil {
		return err
	}

	s.auditSvc.Log(ctx, resolvedBy, "conflict.resolve_change_time", "conflict", conflictID, conflict, losingRes, reason)
	return nil
}

func (s *ConflictService) ResolveCancel(ctx context.Context, conflictID, resolvedBy, reason string) error {
	conflict, err := s.conflictRepo.GetByID(ctx, conflictID)
	if err != nil {
		return err
	}
	if conflict == nil {
		return apperror.ErrNotFound
	}
	if conflict.Status != model.ConflictStatusPending {
		return apperror.New(400, "ALREADY_RESOLVED", "conflict is already resolved")
	}

	if err := s.reservationRepo.Cancel(ctx, conflict.LosingReservationID, resolvedBy, reason); err != nil {
		return err
	}

	_ = s.reservationRepo.UpdateStatus(ctx, conflict.WinningReservationID, model.ReservationStatusConfirmed)

	if err := s.conflictRepo.Resolve(ctx, conflictID, resolvedBy, reason, model.ConflictStatusResolvedCancelled); err != nil {
		return err
	}

	s.auditSvc.Log(ctx, resolvedBy, "conflict.resolve_cancel", "conflict", conflictID, conflict, nil, reason)
	return nil
}

func (s *ConflictService) ForceAssign(ctx context.Context, conflictID, resolvedBy, reason string) error {
	if reason == "" {
		return apperror.New(400, "REASON_REQUIRED", "reason is required for force assign")
	}

	conflict, err := s.conflictRepo.GetByID(ctx, conflictID)
	if err != nil {
		return err
	}
	if conflict == nil {
		return apperror.ErrNotFound
	}
	if conflict.Status != model.ConflictStatusPending {
		return apperror.New(400, "ALREADY_RESOLVED", "conflict is already resolved")
	}

	// Force: keep losing reservation's original slot, cancel winning
	if err := s.reservationRepo.UpdateStatus(ctx, conflict.LosingReservationID, model.ReservationStatusConfirmed); err != nil {
		return err
	}
	if err := s.reservationRepo.Cancel(ctx, conflict.WinningReservationID, resolvedBy, "force assigned: "+reason); err != nil {
		return err
	}

	if err := s.conflictRepo.Resolve(ctx, conflictID, resolvedBy, reason, model.ConflictStatusForceAssigned); err != nil {
		return err
	}

	s.auditSvc.Log(ctx, resolvedBy, "conflict.force_assign", "conflict", conflictID, conflict, nil, reason)
	return nil
}
