package service

import (
	"context"
	"time"

	"github.com/kento/driver/backend/internal/repository"
)

type TokenService struct {
	repo *repository.TokenRepo
}

func NewTokenService(repo *repository.TokenRepo) *TokenService {
	return &TokenService{repo: repo}
}

// Blacklist adds a token to the blacklist.
func (s *TokenService) Blacklist(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	return s.repo.Blacklist(ctx, jti, userID, expiresAt)
}

// IsBlacklisted checks whether a token has been revoked.
func (s *TokenService) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	return s.repo.IsBlacklisted(ctx, jti)
}

// CleanExpired removes expired blacklist entries.
func (s *TokenService) CleanExpired(ctx context.Context) error {
	return s.repo.CleanExpired(ctx)
}
