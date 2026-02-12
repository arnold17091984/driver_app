package dto

import "time"

type CreateReservationRequest struct {
	VehicleID    string    `json:"vehicle_id" validate:"required,uuid"`
	StartTime    time.Time `json:"start_time" validate:"required"`
	EndTime      time.Time `json:"end_time" validate:"required"`
	Purpose      string    `json:"purpose" validate:"required"`
	Destinations []string  `json:"destinations,omitempty"`
	Notes        *string   `json:"notes,omitempty"`
}

type UpdateReservationRequest struct {
	VehicleID    *string    `json:"vehicle_id,omitempty"`
	StartTime    *time.Time `json:"start_time,omitempty"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	Purpose      *string    `json:"purpose,omitempty"`
	Destinations []string   `json:"destinations,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}

type CancelReservationRequest struct {
	Reason string `json:"reason"`
}

type AvailabilityRequest struct {
	StartTime time.Time `json:"start_time" validate:"required"`
	EndTime   time.Time `json:"end_time" validate:"required"`
}

type ResolveConflictReassignRequest struct {
	NewVehicleID string `json:"new_vehicle_id" validate:"required,uuid"`
	Reason       string `json:"reason"`
}

type ResolveConflictChangeTimeRequest struct {
	NewStartTime time.Time `json:"new_start_time" validate:"required"`
	NewEndTime   time.Time `json:"new_end_time" validate:"required"`
	Reason       string    `json:"reason"`
}

type ResolveConflictCancelRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type ForceAssignRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// Unified booking flow DTOs

type UnifiedBookingRequest struct {
	Mode          string     `json:"mode" validate:"required,oneof=specific any"` // "specific" or "any"
	VehicleID     *string    `json:"vehicle_id,omitempty"`                        // required when mode=specific
	IsNow         bool       `json:"is_now"`
	StartTime     *time.Time `json:"start_time,omitempty"` // required when !is_now
	EndTime       *time.Time `json:"end_time,omitempty"`   // required when !is_now
	PickupAddress string     `json:"pickup_address" validate:"required"`
	PickupLat     *float64   `json:"pickup_lat,omitempty"`
	PickupLng     *float64   `json:"pickup_lng,omitempty"`
	Purpose       string     `json:"purpose" validate:"required"`
	Destinations  []string   `json:"destinations,omitempty"`
	PassengerName *string    `json:"passenger_name,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
}

type UnifiedBookingResponse struct {
	Type        string      `json:"type"` // "dispatch" or "reservation"
	Dispatch    interface{} `json:"dispatch,omitempty"`
	Reservation interface{} `json:"reservation,omitempty"`
}

type DriverReservationDeclineRequest struct {
	Reason string `json:"reason"`
}
