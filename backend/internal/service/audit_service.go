package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kento/driver/backend/internal/model"
	"github.com/kento/driver/backend/internal/repository"
)

type AuditService struct {
	repo *repository.AuditRepo
}

func NewAuditService(repo *repository.AuditRepo) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, actorID, action, targetType, targetID string, before, after interface{}, reason string) {
	var beforeJSON, afterJSON json.RawMessage
	if before != nil {
		beforeJSON, _ = json.Marshal(before)
	}
	if after != nil {
		afterJSON, _ = json.Marshal(after)
	}

	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}

	entry := &model.AuditLog{
		ActorID:     actorID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		BeforeState: beforeJSON,
		AfterState:  afterJSON,
		Reason:      reasonPtr,
	}
	// Fire and forget - audit log failure should not block operations
	_ = s.repo.Create(ctx, entry)
}

func (s *AuditService) List(ctx context.Context, actorID, action, targetType string, from, to time.Time, limit, offset int) ([]model.AuditLog, error) {
	return s.repo.List(ctx, actorID, action, targetType, from, to, limit, offset)
}

func (s *AuditService) GetByID(ctx context.Context, id string) (*model.AuditLog, error) {
	return s.repo.GetByID(ctx, id)
}
