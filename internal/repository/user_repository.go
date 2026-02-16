package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/golang-project/internal/model"
)

// UserRepository provides CRUD access to users in Postgres. It uses BaseRepository to reduce boilerplate.
type UserRepository struct {
	base *BaseRepository
}

// NewUserRepository creates a new user repository with a DB handle.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{base: NewBaseRepository(db)}
}

// Create inserts a new user and returns the new ID.
func (r *UserRepository) Create(ctx context.Context, u *model.User) (int64, error) {
	now := time.Now().UTC()
	var id int64
	err := r.base.ScanRow(ctx,
		`INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3) RETURNING id`,
		func(row *sql.Row) error {
			return row.Scan(&id)
		},
		u.Name, u.Email, now,
	)
	return id, err
}

// GetByID returns a single user by ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := r.base.ScanRow(ctx,
		`SELECT id, name, email, created_at FROM users WHERE id = $1`,
		func(row *sql.Row) error {
			return row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt)
		},
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// Update modifies name and email of an existing user.
func (r *UserRepository) Update(ctx context.Context, u *model.User) error {
	return r.base.ExecUpdate(ctx,
		`UPDATE users SET name=$1, email=$2 WHERE id=$3`,
		u.Name, u.Email, u.ID,
	)
}

// Delete removes a user by ID.
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.base.ExecUpdate(ctx,
		`DELETE FROM users WHERE id=$1`,
		id,
	)
}

// List returns all users (small dataset for this example).
func (r *UserRepository) List(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	err := r.base.ScanRows(ctx,
		`SELECT id, name, email, created_at FROM users ORDER BY id`,
		func(rows *sql.Rows) error {
			for rows.Next() {
				var u model.User
				if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
					return err
				}
				users = append(users, &u)
			}
			return rows.Err()
		},
	)
	return users, err
}
