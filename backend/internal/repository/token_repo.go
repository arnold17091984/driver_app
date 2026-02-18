package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type TokenRepo struct {
	db *sqlx.DB
}

func NewTokenRepo(db *sqlx.DB) *TokenRepo {
	return &TokenRepo{db: db}
}

// Blacklist adds a token JTI to the blacklist.
func (r *TokenRepo) Blacklist(ctx context.Context, jti, userID string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO token_blacklist (jti, user_id, expires_at) VALUES ($1, $2, $3) ON CONFLICT (jti) DO NOTHING`,
		jti, userID, expiresAt)
	return err
}

// IsBlacklisted checks whether a token JTI has been revoked.
func (r *TokenRepo) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE jti = $1)`, jti)
	return exists, err
}

// CleanExpired removes expired blacklist entries.
func (r *TokenRepo) CleanExpired(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM token_blacklist WHERE expires_at < $1`, time.Now())
	return err
}
