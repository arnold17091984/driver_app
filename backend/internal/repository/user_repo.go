package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kento/driver/backend/internal/model"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByEmployeeID(ctx context.Context, employeeID string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user,
		`SELECT id, employee_id, password_hash, name, role, priority_level, phone_number, fcm_token, is_active, created_at, updated_at
		 FROM users WHERE employee_id = $1`, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user,
		`SELECT id, employee_id, password_hash, name, role, priority_level, phone_number, fcm_token, is_active, created_at, updated_at
		 FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) List(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.SelectContext(ctx, &users,
		`SELECT id, employee_id, password_hash, name, role, priority_level, phone_number, fcm_token, is_active, created_at, updated_at
		 FROM users ORDER BY role, name`)
	return users, err
}

func (r *UserRepo) UpdateRole(ctx context.Context, id string, role model.Role) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`, role, id)
	return err
}

func (r *UserRepo) UpdatePriority(ctx context.Context, id string, priority int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET priority_level = $1, updated_at = NOW() WHERE id = $2`, priority, id)
	return err
}

func (r *UserRepo) UpdateFCMToken(ctx context.Context, id string, token string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET fcm_token = $1, updated_at = NOW() WHERE id = $2`, token, id)
	return err
}

func (r *UserRepo) GetDriversByVehicleIDs(ctx context.Context, vehicleIDs []string) ([]model.User, error) {
	query, args, err := sqlx.In(
		`SELECT u.id, u.employee_id, u.password_hash, u.name, u.role, u.priority_level, u.phone_number, u.fcm_token, u.is_active, u.created_at, u.updated_at
		 FROM users u
		 JOIN vehicles v ON v.driver_id = u.id
		 WHERE v.id IN (?)`, vehicleIDs)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var users []model.User
	err = r.db.SelectContext(ctx, &users, query, args...)
	return users, err
}

func (r *UserRepo) GetByPhoneNumber(ctx context.Context, phone string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user,
		`SELECT id, employee_id, password_hash, name, role, priority_level, phone_number, fcm_token, is_active, created_at, updated_at
		 FROM users WHERE phone_number = $1`, phone)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) CreatePassenger(ctx context.Context, phoneNumber, passwordHash, name string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user,
		`INSERT INTO users (employee_id, password_hash, name, role, phone_number)
		 VALUES ($1, $2, $3, 'passenger', $4)
		 RETURNING id, employee_id, password_hash, name, role, priority_level, phone_number, fcm_token, is_active, created_at, updated_at`,
		phoneNumber, passwordHash, name, phoneNumber)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) GetByRole(ctx context.Context, roles ...model.Role) ([]model.User, error) {
	roleStrings := make([]interface{}, len(roles))
	for i, role := range roles {
		roleStrings[i] = string(role)
	}
	query, args, err := sqlx.In(
		`SELECT id, employee_id, password_hash, name, role, priority_level, phone_number, fcm_token, is_active, created_at, updated_at
		 FROM users WHERE role IN (?) AND is_active = true`, roleStrings)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)
	var users []model.User
	err = r.db.SelectContext(ctx, &users, query, args...)
	return users, err
}
