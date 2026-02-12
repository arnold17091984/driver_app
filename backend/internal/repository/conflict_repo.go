package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kento/driver/backend/internal/model"
)

type ConflictRepo struct {
	db *sqlx.DB
}

func NewConflictRepo(db *sqlx.DB) *ConflictRepo {
	return &ConflictRepo{db: db}
}

func (r *ConflictRepo) Create(ctx context.Context, winningID, losingID string) (*model.ReservationConflict, error) {
	var c model.ReservationConflict
	err := r.db.GetContext(ctx, &c, `
		INSERT INTO reservation_conflicts (winning_reservation_id, losing_reservation_id)
		VALUES ($1, $2)
		RETURNING id, winning_reservation_id, losing_reservation_id, status, resolved_by,
			resolution_reason, resolved_at, created_at`,
		winningID, losingID)
	return &c, err
}

func (r *ConflictRepo) GetByID(ctx context.Context, id string) (*model.ReservationConflict, error) {
	var c model.ReservationConflict
	err := r.db.GetContext(ctx, &c, `
		SELECT id, winning_reservation_id, losing_reservation_id, status, resolved_by,
			resolution_reason, resolved_at, created_at
		FROM reservation_conflicts WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &c, err
}

func (r *ConflictRepo) ListPending(ctx context.Context) ([]model.ReservationConflict, error) {
	var conflicts []model.ReservationConflict
	err := r.db.SelectContext(ctx, &conflicts, `
		SELECT id, winning_reservation_id, losing_reservation_id, status, resolved_by,
			resolution_reason, resolved_at, created_at
		FROM reservation_conflicts
		WHERE status = 'pending'
		ORDER BY created_at DESC`)
	return conflicts, err
}

func (r *ConflictRepo) Resolve(ctx context.Context, id, resolvedBy, reason string, status model.ConflictStatus) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reservation_conflicts
		SET status = $1, resolved_by = $2, resolution_reason = $3, resolved_at = NOW()
		WHERE id = $4`,
		status, resolvedBy, reason, id)
	return err
}
