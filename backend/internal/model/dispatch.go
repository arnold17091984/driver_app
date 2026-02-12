package model

import "time"

type DispatchStatus string

const (
	DispatchStatusPending   DispatchStatus = "pending"
	DispatchStatusAssigned  DispatchStatus = "assigned"
	DispatchStatusAccepted  DispatchStatus = "accepted"
	DispatchStatusEnRoute   DispatchStatus = "en_route"
	DispatchStatusArrived   DispatchStatus = "arrived"
	DispatchStatusCompleted DispatchStatus = "completed"
	DispatchStatusCancelled DispatchStatus = "cancelled"
)

type Dispatch struct {
	ID                   string         `db:"id" json:"id"`
	VehicleID            *string        `db:"vehicle_id" json:"vehicle_id,omitempty"`
	RequesterID          string         `db:"requester_id" json:"requester_id"`
	DispatcherID         *string        `db:"dispatcher_id" json:"dispatcher_id,omitempty"`
	Purpose              string         `db:"purpose" json:"purpose"`
	PassengerName        *string        `db:"passenger_name" json:"passenger_name,omitempty"`
	PassengerCount       int            `db:"passenger_count" json:"passenger_count"`
	Notes                *string        `db:"notes" json:"notes,omitempty"`
	PickupAddress        string         `db:"pickup_address" json:"pickup_address"`
	PickupLat            *float64       `db:"pickup_lat" json:"pickup_lat,omitempty"`
	PickupLng            *float64       `db:"pickup_lng" json:"pickup_lng,omitempty"`
	DropoffAddress       *string        `db:"dropoff_address" json:"dropoff_address,omitempty"`
	DropoffLat           *float64       `db:"dropoff_lat" json:"dropoff_lat,omitempty"`
	DropoffLng           *float64       `db:"dropoff_lng" json:"dropoff_lng,omitempty"`
	Status               DispatchStatus `db:"status" json:"status"`
	EstimatedDurationSec *int           `db:"estimated_duration_sec" json:"estimated_duration_sec,omitempty"`
	EstimatedDistanceM   *int           `db:"estimated_distance_m" json:"estimated_distance_m,omitempty"`
	AssignedAt           *time.Time     `db:"assigned_at" json:"assigned_at,omitempty"`
	AcceptedAt           *time.Time     `db:"accepted_at" json:"accepted_at,omitempty"`
	EnRouteAt            *time.Time     `db:"en_route_at" json:"en_route_at,omitempty"`
	ArrivedAt            *time.Time     `db:"arrived_at" json:"arrived_at,omitempty"`
	CompletedAt          *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
	CancelledAt          *time.Time     `db:"cancelled_at" json:"cancelled_at,omitempty"`
	EstimatedEndAt       *time.Time     `db:"estimated_end_at" json:"estimated_end_at,omitempty"`
	CancelReason         *string        `db:"cancel_reason" json:"cancel_reason,omitempty"`
	CreatedAt            time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at" json:"updated_at"`
}

type DispatchETASnapshot struct {
	ID           string    `db:"id" json:"id"`
	DispatchID   string    `db:"dispatch_id" json:"dispatch_id"`
	VehicleID    string    `db:"vehicle_id" json:"vehicle_id"`
	VehicleName  string    `db:"vehicle_name" json:"vehicle_name,omitempty"`
	DurationSec  int       `db:"duration_sec" json:"duration_sec"`
	DistanceM    int       `db:"distance_m" json:"distance_m"`
	CalculatedAt time.Time `db:"calculated_at" json:"calculated_at"`
}
