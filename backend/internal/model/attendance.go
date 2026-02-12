package model

import "time"

type DriverStatus string

const (
	DriverStatusActive  DriverStatus = "active"
	DriverStatusWaiting DriverStatus = "waiting"
)

type DriverAttendance struct {
	ID           string       `db:"id" json:"id"`
	DriverID     string       `db:"driver_id" json:"driver_id"`
	DriverStatus DriverStatus `db:"driver_status" json:"driver_status"`
	ClockInAt    time.Time    `db:"clock_in_at" json:"clock_in_at"`
	ClockOutAt   *time.Time   `db:"clock_out_at" json:"clock_out_at,omitempty"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
}
