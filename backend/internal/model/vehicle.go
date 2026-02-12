package model

import "time"

type VehicleStatus string

const (
	VehicleStatusAvailable    VehicleStatus = "available"
	VehicleStatusDriverAbsent VehicleStatus = "driver_absent"
	VehicleStatusReserved     VehicleStatus = "reserved"
	VehicleStatusInTrip       VehicleStatus = "in_trip"
	VehicleStatusMaintenance  VehicleStatus = "maintenance"
	VehicleStatusStale        VehicleStatus = "stale_location"
	VehicleStatusWaiting      VehicleStatus = "waiting"
)

type Vehicle struct {
	ID            string    `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	LicensePlate  string    `db:"license_plate" json:"license_plate"`
	DriverID      string    `db:"driver_id" json:"driver_id"`
	IsMaintenance bool      `db:"is_maintenance" json:"is_maintenance"`
	PhotoURL      *string   `db:"photo_url" json:"photo_url,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type VehicleWithStatus struct {
	ID             string        `db:"id" json:"id"`
	Name           string        `db:"name" json:"name"`
	LicensePlate   string        `db:"license_plate" json:"license_plate"`
	DriverID       string        `db:"driver_id" json:"driver_id"`
	DriverName     string        `db:"driver_name" json:"driver_name"`
	IsMaintenance  bool          `db:"is_maintenance" json:"is_maintenance"`
	IsClockedIn    bool          `db:"is_clocked_in" json:"is_clocked_in"`
	PhotoURL       *string       `db:"photo_url" json:"photo_url,omitempty"`
	Status         VehicleStatus `db:"computed_status" json:"status"`
	Latitude       *float64      `db:"latitude" json:"latitude,omitempty"`
	Longitude      *float64      `db:"longitude" json:"longitude,omitempty"`
	Heading        *float64      `db:"heading" json:"heading,omitempty"`
	Speed          *float64      `db:"speed" json:"speed,omitempty"`
	LocationAt     *time.Time    `db:"location_at" json:"location_at,omitempty"`
}
