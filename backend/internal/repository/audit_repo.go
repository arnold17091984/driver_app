package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kento/driver/backend/internal/model"
)

type AuditRepo struct {
	db *sqlx.DB
}

func NewAuditRepo(db *sqlx.DB) *AuditRepo {
	return &AuditRepo{db: db}
}

func (r *AuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_logs (actor_id, action, target_type, target_id, before_state, after_state, reason, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ActorID, log.Action, log.TargetType, log.TargetID,
		log.BeforeState, log.AfterState, log.Reason, log.IPAddress)
	return err
}

func (r *AuditRepo) List(ctx context.Context, actorID, action, targetType string, from, to time.Time, limit, offset int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	query := `
		SELECT a.id, a.actor_id, u.name AS actor_name, a.action, a.target_type, a.target_id,
			a.before_state, a.after_state, a.reason, a.ip_address, a.created_at
		FROM audit_logs a
		JOIN users u ON u.id = a.actor_id
		WHERE 1=1`

	args := []interface{}{}
	idx := 1

	if actorID != "" {
		query += ` AND a.actor_id = $` + intToStr(idx)
		args = append(args, actorID)
		idx++
	}
	if action != "" {
		query += ` AND a.action = $` + intToStr(idx)
		args = append(args, action)
		idx++
	}
	if targetType != "" {
		query += ` AND a.target_type = $` + intToStr(idx)
		args = append(args, targetType)
		idx++
	}
	if !from.IsZero() {
		query += ` AND a.created_at >= $` + intToStr(idx)
		args = append(args, from)
		idx++
	}
	if !to.IsZero() {
		query += ` AND a.created_at <= $` + intToStr(idx)
		args = append(args, to)
		idx++
	}

	query += ` ORDER BY a.created_at DESC LIMIT $` + intToStr(idx) + ` OFFSET $` + intToStr(idx+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &logs, query, args...)
	return logs, err
}

func (r *AuditRepo) GetByID(ctx context.Context, id string) (*model.AuditLog, error) {
	var log model.AuditLog
	err := r.db.GetContext(ctx, &log, `
		SELECT a.id, a.actor_id, u.name AS actor_name, a.action, a.target_type, a.target_id,
			a.before_state, a.after_state, a.reason, a.ip_address, a.created_at
		FROM audit_logs a
		JOIN users u ON u.id = a.actor_id
		WHERE a.id = $1`, id)
	return &log, err
}

func intToStr(i int) string {
	return strconv.Itoa(i)
}

func ToJSON(v interface{}) json.RawMessage {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
