package model

import (
	"time"

	"github.com/lib/pq"
)

type ReservationStatus string

const (
	ReservationStatusConfirmed       ReservationStatus = "confirmed"
	ReservationStatusPendingConflict ReservationStatus = "pending_conflict"
	ReservationStatusPendingDriver   ReservationStatus = "pending_driver"
	ReservationStatusDriverDeclined  ReservationStatus = "driver_declined"
	ReservationStatusCancelled       ReservationStatus = "cancelled"
	ReservationStatusCompleted       ReservationStatus = "completed"
)

type Reservation struct {
	ID                  string            `db:"id" json:"id"`
	VehicleID           string            `db:"vehicle_id" json:"vehicle_id"`
	RequesterID         string            `db:"requester_id" json:"requester_id"`
	StartTime           time.Time         `db:"start_time" json:"start_time"`
	EndTime             time.Time         `db:"end_time" json:"end_time"`
	Purpose             string            `db:"purpose" json:"purpose"`
	Destinations        pq.StringArray    `db:"destinations" json:"destinations"`
	Notes               *string           `db:"notes" json:"notes,omitempty"`
	PassengerName       *string           `db:"passenger_name" json:"passenger_name,omitempty"`
	PickupAddress       *string           `db:"pickup_address" json:"pickup_address,omitempty"`
	PickupLat           *float64          `db:"pickup_lat" json:"pickup_lat,omitempty"`
	PickupLng           *float64          `db:"pickup_lng" json:"pickup_lng,omitempty"`
	PriorityLevel       int               `db:"priority_level" json:"priority_level"`
	Status              ReservationStatus `db:"status" json:"status"`
	CancelReason        *string           `db:"cancel_reason" json:"cancel_reason,omitempty"`
	CancelledBy         *string           `db:"cancelled_by" json:"cancelled_by,omitempty"`
	DeclinedByDriverIDs pq.StringArray    `db:"declined_by_driver_ids" json:"declined_by_driver_ids,omitempty"`
	CreatedAt           time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time         `db:"updated_at" json:"updated_at"`
}

type ReservationWithDetails struct {
	Reservation
	VehicleName   string `db:"vehicle_name" json:"vehicle_name"`
	RequesterName string `db:"requester_name" json:"requester_name"`
	DriverName    string `db:"driver_name" json:"driver_name,omitempty"`
}

type ConflictStatus string

const (
	ConflictStatusPending           ConflictStatus = "pending"
	ConflictStatusResolvedReassign  ConflictStatus = "resolved_reassign"
	ConflictStatusResolvedChanged   ConflictStatus = "resolved_changed"
	ConflictStatusResolvedCancelled ConflictStatus = "resolved_cancelled"
	ConflictStatusForceAssigned     ConflictStatus = "force_assigned"
)

type ReservationConflict struct {
	ID                   string         `db:"id" json:"id"`
	WinningReservationID string         `db:"winning_reservation_id" json:"winning_reservation_id"`
	LosingReservationID  string         `db:"losing_reservation_id" json:"losing_reservation_id"`
	Status               ConflictStatus `db:"status" json:"status"`
	ResolvedBy           *string        `db:"resolved_by" json:"resolved_by,omitempty"`
	ResolutionReason     *string        `db:"resolution_reason" json:"resolution_reason,omitempty"`
	ResolvedAt           *time.Time     `db:"resolved_at" json:"resolved_at,omitempty"`
	CreatedAt            time.Time      `db:"created_at" json:"created_at"`
}
