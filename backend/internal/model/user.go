package model

import "time"

type Role string

const (
	RoleAdmin      Role = "admin"
	RoleDispatcher Role = "dispatcher"
	RoleViewer     Role = "viewer"
	RoleDriver    Role = "driver"
	RolePassenger Role = "passenger"
)

type User struct {
	ID            string    `db:"id" json:"id"`
	EmployeeID    string    `db:"employee_id" json:"employee_id"`
	PasswordHash  string    `db:"password_hash" json:"-"`
	Name          string    `db:"name" json:"name"`
	Role          Role      `db:"role" json:"role"`
	PriorityLevel int       `db:"priority_level" json:"priority_level"`
	PhoneNumber   *string   `db:"phone_number" json:"phone_number,omitempty"`
	FCMToken      *string   `db:"fcm_token" json:"-"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleDispatcher, RoleViewer, RoleDriver, RolePassenger:
		return true
	}
	return false
}

func (r Role) CanDispatch() bool {
	return r == RoleAdmin || r == RoleDispatcher
}

func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}
