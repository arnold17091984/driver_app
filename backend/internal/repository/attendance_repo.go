package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kento/driver/backend/internal/model"
)

type AttendanceRepo struct {
	db *sqlx.DB
}

func NewAttendanceRepo(db *sqlx.DB) *AttendanceRepo {
	return &AttendanceRepo{db: db}
}

func (r *AttendanceRepo) GetActiveByDriverID(ctx context.Context, driverID string) (*model.DriverAttendance, error) {
	var a model.DriverAttendance
	err := r.db.GetContext(ctx, &a,
		`SELECT id, driver_id, driver_status, clock_in_at, clock_out_at, created_at
		 FROM driver_attendance
		 WHERE driver_id = $1 AND clock_out_at IS NULL
		 ORDER BY clock_in_at DESC LIMIT 1`, driverID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &a, err
}

func (r *AttendanceRepo) ClockIn(ctx context.Context, driverID string) (*model.DriverAttendance, error) {
	var a model.DriverAttendance
	err := r.db.GetContext(ctx, &a,
		`INSERT INTO driver_attendance (driver_id) VALUES ($1)
		 RETURNING id, driver_id, driver_status, clock_in_at, clock_out_at, created_at`, driverID)
	return &a, err
}

func (r *AttendanceRepo) ClockOut(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE driver_attendance SET clock_out_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *AttendanceRepo) UpdateDriverStatus(ctx context.Context, id string, status model.DriverStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE driver_attendance SET driver_status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *AttendanceRepo) ListByDriverID(ctx context.Context, driverID string, limit int) ([]model.DriverAttendance, error) {
	var records []model.DriverAttendance
	err := r.db.SelectContext(ctx, &records,
		`SELECT id, driver_id, driver_status, clock_in_at, clock_out_at, created_at
		 FROM driver_attendance
		 WHERE driver_id = $1
		 ORDER BY clock_in_at DESC
		 LIMIT $2`, driverID, limit)
	return records, err
}
