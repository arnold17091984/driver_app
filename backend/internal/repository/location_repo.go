package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kento/driver/backend/internal/model"
)

type LocationRepo struct {
	db *sqlx.DB
}

func NewLocationRepo(db *sqlx.DB) *LocationRepo {
	return &LocationRepo{db: db}
}

func (r *LocationRepo) BatchInsert(ctx context.Context, vehicleID string, points []model.LocationPoint) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, p := range points {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO vehicle_locations (vehicle_id, location, heading, speed, accuracy, recorded_at)
			VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4, $5, $6, $7)`,
			vehicleID, p.Longitude, p.Latitude, p.Heading, p.Speed, p.Accuracy, p.RecordedAt)
		if err != nil {
			return err
		}
	}

	// Upsert current location with the latest point
	latest := points[len(points)-1]
	_, err = tx.ExecContext(ctx, `
		INSERT INTO vehicle_location_current (vehicle_id, location, heading, speed, accuracy, recorded_at, updated_at)
		VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326)::geography, $4, $5, $6, $7, NOW())
		ON CONFLICT (vehicle_id)
		DO UPDATE SET
			location = EXCLUDED.location,
			heading = EXCLUDED.heading,
			speed = EXCLUDED.speed,
			accuracy = EXCLUDED.accuracy,
			recorded_at = EXCLUDED.recorded_at,
			updated_at = NOW()
		WHERE EXCLUDED.recorded_at >= vehicle_location_current.recorded_at`,
		vehicleID, latest.Longitude, latest.Latitude, latest.Heading, latest.Speed, latest.Accuracy, latest.RecordedAt)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *LocationRepo) GetHistory(ctx context.Context, vehicleID string, from, to time.Time) ([]model.VehicleLocation, error) {
	var locations []model.VehicleLocation
	err := r.db.SelectContext(ctx, &locations, `
		SELECT id, vehicle_id,
			ST_Y(location::geometry) AS latitude,
			ST_X(location::geometry) AS longitude,
			heading, speed, accuracy, recorded_at
		FROM vehicle_locations
		WHERE vehicle_id = $1 AND recorded_at BETWEEN $2 AND $3
		ORDER BY recorded_at ASC`, vehicleID, from, to)
	return locations, err
}

func (r *LocationRepo) GetCurrent(ctx context.Context, vehicleID string) (*model.VehicleLocationCurrent, error) {
	var loc model.VehicleLocationCurrent
	err := r.db.GetContext(ctx, &loc, `
		SELECT vehicle_id,
			ST_Y(location::geometry) AS latitude,
			ST_X(location::geometry) AS longitude,
			heading, speed, accuracy, recorded_at, updated_at
		FROM vehicle_location_current
		WHERE vehicle_id = $1`, vehicleID)
	if err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *LocationRepo) GetAllCurrent(ctx context.Context) ([]model.VehicleLocationCurrent, error) {
	var locations []model.VehicleLocationCurrent
	err := r.db.SelectContext(ctx, &locations, `
		SELECT vehicle_id,
			ST_Y(location::geometry) AS latitude,
			ST_X(location::geometry) AS longitude,
			heading, speed, accuracy, recorded_at, updated_at
		FROM vehicle_location_current`)
	return locations, err
}
