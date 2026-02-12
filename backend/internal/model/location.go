package model

import "time"

type VehicleLocation struct {
	ID         string    `db:"id" json:"id"`
	VehicleID  string    `db:"vehicle_id" json:"vehicle_id"`
	Latitude   float64   `db:"latitude" json:"latitude"`
	Longitude  float64   `db:"longitude" json:"longitude"`
	Heading    *float64  `db:"heading" json:"heading,omitempty"`
	Speed      *float64  `db:"speed" json:"speed,omitempty"`
	Accuracy   *float64  `db:"accuracy" json:"accuracy,omitempty"`
	RecordedAt time.Time `db:"recorded_at" json:"recorded_at"`
}

type VehicleLocationCurrent struct {
	VehicleID  string    `db:"vehicle_id" json:"vehicle_id"`
	Latitude   float64   `db:"latitude" json:"latitude"`
	Longitude  float64   `db:"longitude" json:"longitude"`
	Heading    *float64  `db:"heading" json:"heading,omitempty"`
	Speed      *float64  `db:"speed" json:"speed,omitempty"`
	Accuracy   *float64  `db:"accuracy" json:"accuracy,omitempty"`
	RecordedAt time.Time `db:"recorded_at" json:"recorded_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type LocationPoint struct {
	Latitude   float64   `json:"latitude" validate:"required,latitude"`
	Longitude  float64   `json:"longitude" validate:"required,longitude"`
	Heading    *float64  `json:"heading,omitempty"`
	Speed      *float64  `json:"speed,omitempty"`
	Accuracy   *float64  `json:"accuracy,omitempty"`
	RecordedAt time.Time `json:"recorded_at" validate:"required"`
}
