package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kento/driver/backend/internal/model"
)

type VehicleRepo struct {
	db *sqlx.DB
}

func NewVehicleRepo(db *sqlx.DB) *VehicleRepo {
	return &VehicleRepo{db: db}
}

func (r *VehicleRepo) ListWithStatus(ctx context.Context, staleThreshold time.Duration) ([]model.VehicleWithStatus, error) {
	var vehicles []model.VehicleWithStatus
	err := r.db.SelectContext(ctx, &vehicles, `
		SELECT
			v.id,
			v.name,
			v.license_plate,
			v.driver_id,
			u.name AS driver_name,
			v.is_maintenance,
			v.photo_url,
			(da.id IS NOT NULL) AS is_clocked_in,
			ST_Y(vlc.location::geometry) AS latitude,
			ST_X(vlc.location::geometry) AS longitude,
			vlc.heading,
			vlc.speed,
			vlc.recorded_at AS location_at,
			CASE
				WHEN v.is_maintenance THEN 'maintenance'
				WHEN da.id IS NULL THEN 'driver_absent'
				WHEN d.id IS NOT NULL THEN 'in_trip'
				WHEN res.id IS NOT NULL THEN 'reserved'
				WHEN vlc.recorded_at < NOW() - $1::interval THEN 'stale_location'
				WHEN da.driver_status = 'waiting' THEN 'waiting'
				ELSE 'available'
			END AS computed_status
		FROM vehicles v
		JOIN users u ON u.id = v.driver_id
		LEFT JOIN vehicle_location_current vlc ON vlc.vehicle_id = v.id
		LEFT JOIN driver_attendance da ON da.driver_id = v.driver_id AND da.clock_out_at IS NULL
		LEFT JOIN dispatches d ON d.vehicle_id = v.id
			AND d.status IN ('assigned','accepted','en_route','arrived')
		LEFT JOIN reservations res ON res.vehicle_id = v.id
			AND res.status = 'confirmed'
			AND NOW() BETWEEN res.start_time AND res.end_time
		ORDER BY v.name
	`, staleThreshold.String())
	return vehicles, err
}

func (r *VehicleRepo) GetByID(ctx context.Context, id string) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.GetContext(ctx, &v,
		`SELECT id, name, license_plate, driver_id, is_maintenance, photo_url, created_at, updated_at
		 FROM vehicles WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, err
}

func (r *VehicleRepo) GetByDriverID(ctx context.Context, driverID string) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.GetContext(ctx, &v,
		`SELECT id, name, license_plate, driver_id, is_maintenance, photo_url, created_at, updated_at
		 FROM vehicles WHERE driver_id = $1`, driverID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &v, err
}

func (r *VehicleRepo) Create(ctx context.Context, name, licensePlate, driverID string) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.GetContext(ctx, &v,
		`INSERT INTO vehicles (name, license_plate, driver_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, license_plate, driver_id, is_maintenance, photo_url, created_at, updated_at`,
		name, licensePlate, driverID)
	return &v, err
}

func (r *VehicleRepo) Update(ctx context.Context, id, name, licensePlate, driverID string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicles SET name = $1, license_plate = $2, driver_id = $3, updated_at = NOW() WHERE id = $4`,
		name, licensePlate, driverID, id)
	return err
}

func (r *VehicleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM vehicles WHERE id = $1`, id)
	return err
}

func (r *VehicleRepo) UpdatePhotoURL(ctx context.Context, id string, photoURL *string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicles SET photo_url = $1, updated_at = NOW() WHERE id = $2`,
		photoURL, id)
	return err
}

func (r *VehicleRepo) ToggleMaintenance(ctx context.Context, id string, maintenance bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE vehicles SET is_maintenance = $1, updated_at = NOW() WHERE id = $2`,
		maintenance, id)
	return err
}

func (r *VehicleRepo) ListAvailable(ctx context.Context, staleThreshold time.Duration) ([]model.VehicleWithStatus, error) {
	all, err := r.ListWithStatus(ctx, staleThreshold)
	if err != nil {
		return nil, err
	}
	var available []model.VehicleWithStatus
	for _, v := range all {
		if v.Status == model.VehicleStatusAvailable || v.Status == model.VehicleStatusWaiting {
			available = append(available, v)
		}
	}
	return available, nil
}
