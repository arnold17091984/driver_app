package model

import (
	"encoding/json"
	"time"
)

type AuditLog struct {
	ID          string          `db:"id" json:"id"`
	ActorID     string          `db:"actor_id" json:"actor_id"`
	ActorName   string          `db:"actor_name" json:"actor_name,omitempty"`
	Action      string          `db:"action" json:"action"`
	TargetType  string          `db:"target_type" json:"target_type"`
	TargetID    string          `db:"target_id" json:"target_id"`
	BeforeState json.RawMessage `db:"before_state" json:"before_state,omitempty"`
	AfterState  json.RawMessage `db:"after_state" json:"after_state,omitempty"`
	Reason      *string         `db:"reason" json:"reason,omitempty"`
	IPAddress   *string         `db:"ip_address" json:"ip_address,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}
