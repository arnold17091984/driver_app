package dto

type CreateDispatchRequest struct {
	Purpose        string   `json:"purpose" validate:"required"`
	PassengerName  *string  `json:"passenger_name,omitempty"`
	PassengerCount int      `json:"passenger_count" validate:"min=1"`
	Notes          *string  `json:"notes,omitempty"`
	PickupAddress  string   `json:"pickup_address" validate:"required"`
	PickupLat      *float64 `json:"pickup_lat,omitempty"`
	PickupLng      *float64 `json:"pickup_lng,omitempty"`
	DropoffAddress *string  `json:"dropoff_address,omitempty"`
	DropoffLat     *float64 `json:"dropoff_lat,omitempty"`
	DropoffLng     *float64 `json:"dropoff_lng,omitempty"`
}

type QuickBoardRequest struct {
	VehicleID        string  `json:"vehicle_id" validate:"required,uuid"`
	PassengerName    string  `json:"passenger_name" validate:"required"`
	PassengerCount   int     `json:"passenger_count"`
	Purpose          string  `json:"purpose"`
	Notes            *string `json:"notes,omitempty"`
	EstimatedMinutes int     `json:"estimated_minutes,omitempty"`
}

type AssignDispatchRequest struct {
	VehicleID string `json:"vehicle_id" validate:"required,uuid"`
}

type CancelDispatchRequest struct {
	Reason string `json:"reason"`
}

type CalculateETARequest struct {
	PickupLat float64 `json:"pickup_lat" validate:"required"`
	PickupLng float64 `json:"pickup_lng" validate:"required"`
}

type VehicleETA struct {
	VehicleID   string  `json:"vehicle_id"`
	VehicleName string  `json:"vehicle_name"`
	DriverName  string  `json:"driver_name"`
	Plate       string  `json:"plate"`
	Status      string  `json:"status"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	DistanceM   int     `json:"distance_m"`
	DurationSec int     `json:"duration_sec"`
	IsAvailable bool    `json:"is_available"`
}
