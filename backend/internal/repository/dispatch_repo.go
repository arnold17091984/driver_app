package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kento/driver/backend/internal/model"
)

type DispatchRepo struct {
	db *sqlx.DB
}

func NewDispatchRepo(db *sqlx.DB) *DispatchRepo {
	return &DispatchRepo{db: db}
}

func (r *DispatchRepo) Create(ctx context.Context, d *model.Dispatch) error {
	return r.db.GetContext(ctx, d, `
		INSERT INTO dispatches (
			requester_id, purpose, passenger_name, passenger_count, notes,
			pickup_address, pickup_location, dropoff_address, dropoff_location,
			estimated_end_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			CASE WHEN $7::float8 IS NOT NULL AND $8::float8 IS NOT NULL
				THEN ST_SetSRID(ST_MakePoint($8, $7), 4326)::geography
				ELSE NULL END,
			$9,
			CASE WHEN $10::float8 IS NOT NULL AND $11::float8 IS NOT NULL
				THEN ST_SetSRID(ST_MakePoint($11, $10), 4326)::geography
				ELSE NULL END,
			$12
		) RETURNING id, vehicle_id, requester_id, dispatcher_id, purpose, passenger_name,
		  passenger_count, notes, pickup_address,
		  ST_Y(pickup_location::geometry) AS pickup_lat, ST_X(pickup_location::geometry) AS pickup_lng,
		  dropoff_address,
		  ST_Y(dropoff_location::geometry) AS dropoff_lat, ST_X(dropoff_location::geometry) AS dropoff_lng,
		  status, estimated_duration_sec, estimated_distance_m, estimated_end_at,
		  assigned_at, accepted_at, en_route_at, arrived_at, completed_at, cancelled_at,
		  cancel_reason, created_at, updated_at`,
		d.RequesterID, d.Purpose, d.PassengerName, d.PassengerCount, d.Notes,
		d.PickupAddress, d.PickupLat, d.PickupLng,
		d.DropoffAddress, d.DropoffLat, d.DropoffLng, d.EstimatedEndAt)
}

func (r *DispatchRepo) GetByID(ctx context.Context, id string) (*model.Dispatch, error) {
	var d model.Dispatch
	err := r.db.GetContext(ctx, &d, `
		SELECT id, vehicle_id, requester_id, dispatcher_id, purpose, passenger_name,
			passenger_count, notes, pickup_address,
			ST_Y(pickup_location::geometry) AS pickup_lat,
			ST_X(pickup_location::geometry) AS pickup_lng,
			dropoff_address,
			ST_Y(dropoff_location::geometry) AS dropoff_lat,
			ST_X(dropoff_location::geometry) AS dropoff_lng,
			status, estimated_duration_sec, estimated_distance_m, estimated_end_at,
			assigned_at, accepted_at, en_route_at, arrived_at, completed_at, cancelled_at,
			cancel_reason, created_at, updated_at
		FROM dispatches WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &d, err
}

func (r *DispatchRepo) List(ctx context.Context, status string, limit, offset int) ([]model.Dispatch, error) {
	var dispatches []model.Dispatch
	query := `
		SELECT id, vehicle_id, requester_id, dispatcher_id, purpose, passenger_name,
			passenger_count, notes, pickup_address,
			ST_Y(pickup_location::geometry) AS pickup_lat,
			ST_X(pickup_location::geometry) AS pickup_lng,
			dropoff_address,
			ST_Y(dropoff_location::geometry) AS dropoff_lat,
			ST_X(dropoff_location::geometry) AS dropoff_lng,
			status, estimated_duration_sec, estimated_distance_m, estimated_end_at,
			assigned_at, accepted_at, en_route_at, arrived_at, completed_at, cancelled_at,
			cancel_reason, created_at, updated_at
		FROM dispatches`

	if status != "" {
		query += ` WHERE status = $3`
		query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		err := r.db.SelectContext(ctx, &dispatches, query, limit, offset, status)
		return dispatches, err
	}
	query += ` ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &dispatches, query, limit, offset)
	return dispatches, err
}

func (r *DispatchRepo) Assign(ctx context.Context, dispatchID, vehicleID, dispatcherID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE dispatches
		SET vehicle_id = $1, dispatcher_id = $2, status = 'assigned', assigned_at = NOW(), updated_at = NOW()
		WHERE id = $3 AND status = 'pending'`,
		vehicleID, dispatcherID, dispatchID)
	return err
}

func (r *DispatchRepo) UpdateStatus(ctx context.Context, id string, status model.DispatchStatus) error {
	var timestampCol string
	switch status {
	case model.DispatchStatusAccepted:
		timestampCol = "accepted_at"
	case model.DispatchStatusEnRoute:
		timestampCol = "en_route_at"
	case model.DispatchStatusArrived:
		timestampCol = "arrived_at"
	case model.DispatchStatusCompleted:
		timestampCol = "completed_at"
	default:
		_, err := r.db.ExecContext(ctx,
			`UPDATE dispatches SET status = $1, updated_at = NOW() WHERE id = $2`, status, id)
		return err
	}

	_, err := r.db.ExecContext(ctx,
		`UPDATE dispatches SET status = $1, `+timestampCol+` = NOW(), updated_at = NOW() WHERE id = $2`,
		status, id)
	return err
}

func (r *DispatchRepo) Cancel(ctx context.Context, id, reason string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE dispatches
		SET status = 'cancelled', cancelled_at = NOW(), cancel_reason = $1, updated_at = NOW()
		WHERE id = $2 AND status NOT IN ('completed', 'cancelled')`,
		reason, id)
	return err
}

func (r *DispatchRepo) GetActiveByVehicleID(ctx context.Context, vehicleID string) (*model.Dispatch, error) {
	var d model.Dispatch
	err := r.db.GetContext(ctx, &d, `
		SELECT id, vehicle_id, requester_id, dispatcher_id, purpose, passenger_name,
			passenger_count, notes, pickup_address,
			ST_Y(pickup_location::geometry) AS pickup_lat,
			ST_X(pickup_location::geometry) AS pickup_lng,
			dropoff_address,
			ST_Y(dropoff_location::geometry) AS dropoff_lat,
			ST_X(dropoff_location::geometry) AS dropoff_lng,
			status, estimated_duration_sec, estimated_distance_m, estimated_end_at,
			assigned_at, accepted_at, en_route_at, arrived_at, completed_at, cancelled_at,
			cancel_reason, created_at, updated_at
		FROM dispatches
		WHERE vehicle_id = $1 AND status IN ('assigned','accepted','en_route','arrived')
		ORDER BY created_at DESC LIMIT 1`, vehicleID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &d, err
}

func (r *DispatchRepo) GetActiveByDriverID(ctx context.Context, driverID string) (*model.Dispatch, error) {
	var d model.Dispatch
	err := r.db.GetContext(ctx, &d, `
		SELECT d.id, d.vehicle_id, d.requester_id, d.dispatcher_id, d.purpose, d.passenger_name,
			d.passenger_count, d.notes, d.pickup_address,
			ST_Y(d.pickup_location::geometry) AS pickup_lat,
			ST_X(d.pickup_location::geometry) AS pickup_lng,
			d.dropoff_address,
			ST_Y(d.dropoff_location::geometry) AS dropoff_lat,
			ST_X(d.dropoff_location::geometry) AS dropoff_lng,
			d.status, d.estimated_duration_sec, d.estimated_distance_m, d.estimated_end_at,
			d.assigned_at, d.accepted_at, d.en_route_at, d.arrived_at, d.completed_at, d.cancelled_at,
			d.cancel_reason, d.created_at, d.updated_at
		FROM dispatches d
		JOIN vehicles v ON v.id = d.vehicle_id
		WHERE v.driver_id = $1 AND d.status IN ('assigned','accepted','en_route','arrived')
		ORDER BY d.created_at DESC LIMIT 1`, driverID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &d, err
}

func (r *DispatchRepo) SaveETASnapshot(ctx context.Context, snap *model.DispatchETASnapshot) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dispatch_eta_snapshots (dispatch_id, vehicle_id, duration_sec, distance_m, origin_lat, origin_lng)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		snap.DispatchID, snap.VehicleID, snap.DurationSec, snap.DistanceM, 0.0, 0.0)
	return err
}

func (r *DispatchRepo) RateDispatch(ctx context.Context, dispatchID string, rating int, comment string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE dispatches SET rating = $1, rating_comment = $2, updated_at = NOW()
		WHERE id = $3`, rating, comment, dispatchID)
	return err
}

func (r *DispatchRepo) GetETASnapshots(ctx context.Context, dispatchID string) ([]model.DispatchETASnapshot, error) {
	var snapshots []model.DispatchETASnapshot
	err := r.db.SelectContext(ctx, &snapshots, `
		SELECT s.id, s.dispatch_id, s.vehicle_id, v.name AS vehicle_name,
			s.duration_sec, s.distance_m, s.calculated_at
		FROM dispatch_eta_snapshots s
		JOIN vehicles v ON v.id = s.vehicle_id
		WHERE s.dispatch_id = $1
		ORDER BY s.duration_sec ASC`, dispatchID)
	return snapshots, err
}
