package repository

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kento/driver/backend/internal/model"
)

type ReservationRepo struct {
	db *sqlx.DB
}

func NewReservationRepo(db *sqlx.DB) *ReservationRepo {
	return &ReservationRepo{db: db}
}

const reservationColumns = `id, vehicle_id, requester_id, start_time, end_time, purpose, destinations, notes,
	passenger_name, pickup_address,
	ST_Y(pickup_location::geometry) AS pickup_lat, ST_X(pickup_location::geometry) AS pickup_lng,
	priority_level, status, cancel_reason, cancelled_by, declined_by_driver_ids,
	created_at, updated_at`

func (r *ReservationRepo) Create(ctx context.Context, res *model.Reservation) error {
	return r.db.GetContext(ctx, res, `
		INSERT INTO reservations (vehicle_id, requester_id, start_time, end_time, purpose, destinations, notes,
			passenger_name, pickup_address, pickup_location, priority_level, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,
			CASE WHEN $10::float8 IS NOT NULL AND $11::float8 IS NOT NULL
				THEN ST_SetSRID(ST_MakePoint($11, $10), 4326)::geography
				ELSE NULL END,
			$12, $13)
		RETURNING `+reservationColumns,
		res.VehicleID, res.RequesterID, res.StartTime, res.EndTime,
		res.Purpose, pq.Array(res.Destinations), res.Notes,
		res.PassengerName, res.PickupAddress, res.PickupLat, res.PickupLng,
		res.PriorityLevel, res.Status)
}

func (r *ReservationRepo) GetByID(ctx context.Context, id string) (*model.Reservation, error) {
	var res model.Reservation
	err := r.db.GetContext(ctx, &res, `
		SELECT `+reservationColumns+`
		FROM reservations WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &res, err
}

func (r *ReservationRepo) List(ctx context.Context, vehicleID string, from, to time.Time, status string, limit, offset int) ([]model.ReservationWithDetails, error) {
	var reservations []model.ReservationWithDetails
	query := `
		SELECT r.id, r.vehicle_id, r.requester_id, r.start_time, r.end_time, r.purpose,
			r.destinations, r.notes, r.passenger_name, r.pickup_address,
			ST_Y(r.pickup_location::geometry) AS pickup_lat, ST_X(r.pickup_location::geometry) AS pickup_lng,
			r.priority_level, r.status, r.cancel_reason, r.cancelled_by, r.declined_by_driver_ids,
			r.created_at, r.updated_at,
			v.name AS vehicle_name, u.name AS requester_name
		FROM reservations r
		JOIN vehicles v ON v.id = r.vehicle_id
		JOIN users u ON u.id = r.requester_id
		WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if vehicleID != "" {
		query += ` AND r.vehicle_id = $` + itoa(argIdx)
		args = append(args, vehicleID)
		argIdx++
	}
	if !from.IsZero() {
		query += ` AND r.end_time >= $` + itoa(argIdx)
		args = append(args, from)
		argIdx++
	}
	if !to.IsZero() {
		query += ` AND r.start_time <= $` + itoa(argIdx)
		args = append(args, to)
		argIdx++
	}
	if status != "" {
		query += ` AND r.status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}

	query += ` ORDER BY r.start_time ASC LIMIT $` + itoa(argIdx) + ` OFFSET $` + itoa(argIdx+1)
	args = append(args, limit, offset)

	err := r.db.SelectContext(ctx, &reservations, query, args...)
	return reservations, err
}

func (r *ReservationRepo) FindOverlapping(ctx context.Context, vehicleID string, startTime, endTime time.Time, excludeID string) ([]model.Reservation, error) {
	var reservations []model.Reservation
	query := `
		SELECT ` + reservationColumns + `
		FROM reservations
		WHERE vehicle_id = $1
			AND status IN ('confirmed', 'pending_conflict', 'pending_driver')
			AND start_time < $3
			AND end_time > $2`

	args := []interface{}{vehicleID, startTime, endTime}

	if excludeID != "" {
		query += ` AND id != $4`
		args = append(args, excludeID)
	}

	err := r.db.SelectContext(ctx, &reservations, query, args...)
	return reservations, err
}

func (r *ReservationRepo) UpdateStatus(ctx context.Context, id string, status model.ReservationStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE reservations SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
	return err
}

func (r *ReservationRepo) Cancel(ctx context.Context, id, cancelledBy, reason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reservations
		SET status = 'cancelled', cancelled_by = $1, cancel_reason = $2, updated_at = NOW()
		WHERE id = $3`, cancelledBy, reason, id)
	return err
}

func (r *ReservationRepo) Update(ctx context.Context, res *model.Reservation) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reservations
		SET vehicle_id = $1, start_time = $2, end_time = $3, purpose = $4,
			destinations = $5, notes = $6, updated_at = NOW()
		WHERE id = $7`,
		res.VehicleID, res.StartTime, res.EndTime, res.Purpose,
		pq.Array(res.Destinations), res.Notes, res.ID)
	return err
}

func (r *ReservationRepo) UpdateVehicle(ctx context.Context, id, vehicleID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reservations SET vehicle_id = $1, status = 'pending_driver', updated_at = NOW()
		WHERE id = $2`, vehicleID, id)
	return err
}

func (r *ReservationRepo) GetUpcomingReminders(ctx context.Context, minutesBefore int) ([]model.ReservationWithDetails, error) {
	var reservations []model.ReservationWithDetails
	err := r.db.SelectContext(ctx, &reservations, `
		SELECT r.id, r.vehicle_id, r.requester_id, r.start_time, r.end_time, r.purpose,
			r.destinations, r.notes, r.passenger_name, r.pickup_address,
			ST_Y(r.pickup_location::geometry) AS pickup_lat, ST_X(r.pickup_location::geometry) AS pickup_lng,
			r.priority_level, r.status, r.cancel_reason, r.cancelled_by, r.declined_by_driver_ids,
			r.created_at, r.updated_at,
			v.name AS vehicle_name, u.name AS requester_name
		FROM reservations r
		JOIN vehicles v ON v.id = r.vehicle_id
		JOIN users u ON u.id = r.requester_id
		WHERE r.status = 'confirmed'
			AND r.start_time BETWEEN NOW() + ($1 - 1) * INTERVAL '1 minute'
			AND NOW() + $1 * INTERVAL '1 minute'`, minutesBefore)
	return reservations, err
}

func (r *ReservationRepo) AutoCompleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE reservations
		SET status = 'completed', updated_at = NOW()
		WHERE status = 'confirmed' AND end_time < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// FindPendingByDriverID returns pending_driver reservations for vehicles assigned to a driver.
func (r *ReservationRepo) FindPendingByDriverID(ctx context.Context, driverID string) ([]model.ReservationWithDetails, error) {
	var reservations []model.ReservationWithDetails
	err := r.db.SelectContext(ctx, &reservations, `
		SELECT r.id, r.vehicle_id, r.requester_id, r.start_time, r.end_time, r.purpose,
			r.destinations, r.notes, r.passenger_name, r.pickup_address,
			ST_Y(r.pickup_location::geometry) AS pickup_lat, ST_X(r.pickup_location::geometry) AS pickup_lng,
			r.priority_level, r.status, r.cancel_reason, r.cancelled_by, r.declined_by_driver_ids,
			r.created_at, r.updated_at,
			v.name AS vehicle_name, u.name AS requester_name
		FROM reservations r
		JOIN vehicles v ON v.id = r.vehicle_id
		JOIN users u ON u.id = r.requester_id
		WHERE v.driver_id = $1
			AND r.status = 'pending_driver'
		ORDER BY r.start_time ASC`, driverID)
	return reservations, err
}

// FindAvailableVehicleForSlot returns vehicle IDs available during a time slot, excluding given IDs.
func (r *ReservationRepo) FindAvailableVehicleForSlot(ctx context.Context, startTime, endTime time.Time, excludeVehicleIDs []string) ([]string, error) {
	var vehicleIDs []string
	err := r.db.SelectContext(ctx, &vehicleIDs, `
		SELECT v.id
		FROM vehicles v
		JOIN users u ON u.id = v.driver_id
		WHERE v.is_maintenance = false
			AND v.id != ALL($3::uuid[])
			AND NOT EXISTS (
				SELECT 1 FROM reservations res
				WHERE res.vehicle_id = v.id
					AND res.status IN ('confirmed', 'pending_driver')
					AND res.start_time < $2
					AND res.end_time > $1
			)
			AND NOT EXISTS (
				SELECT 1 FROM dispatches d
				WHERE d.vehicle_id = v.id
					AND d.status IN ('assigned','accepted','en_route','arrived')
			)
		ORDER BY v.name`, startTime, endTime, pq.Array(excludeVehicleIDs))
	return vehicleIDs, err
}

// AddDeclinedDriver appends a driver_id to the declined_by_driver_ids array.
func (r *ReservationRepo) AddDeclinedDriver(ctx context.Context, reservationID, driverID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE reservations
		SET declined_by_driver_ids = array_append(declined_by_driver_ids, $1::uuid), updated_at = NOW()
		WHERE id = $2`, driverID, reservationID)
	return err
}

// GetDayReservations returns all non-cancelled reservations for a vehicle on a given date.
func (r *ReservationRepo) GetDayReservations(ctx context.Context, vehicleID string, date time.Time) ([]model.ReservationWithDetails, error) {
	var reservations []model.ReservationWithDetails
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	err := r.db.SelectContext(ctx, &reservations, `
		SELECT r.id, r.vehicle_id, r.requester_id, r.start_time, r.end_time, r.purpose,
			r.destinations, r.notes, r.passenger_name, r.pickup_address,
			ST_Y(r.pickup_location::geometry) AS pickup_lat, ST_X(r.pickup_location::geometry) AS pickup_lng,
			r.priority_level, r.status, r.cancel_reason, r.cancelled_by, r.declined_by_driver_ids,
			r.created_at, r.updated_at,
			v.name AS vehicle_name, u.name AS requester_name
		FROM reservations r
		JOIN vehicles v ON v.id = r.vehicle_id
		JOIN users u ON u.id = r.requester_id
		WHERE r.vehicle_id = $1
			AND r.status IN ('confirmed', 'pending_driver', 'pending_conflict')
			AND r.start_time < $3
			AND r.end_time > $2
		ORDER BY r.start_time ASC`, vehicleID, dayStart, dayEnd)
	return reservations, err
}

// Helper for building parameterized queries
func itoa(i int) string {
	return strconv.Itoa(i)
}
