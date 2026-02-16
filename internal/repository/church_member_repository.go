package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/example/golang-project/internal/model"
)

// ChurchMemberRepository provides CRUD access to church members in Postgres.
type ChurchMemberRepository struct {
	base *BaseRepository
}

// NewChurchMemberRepository creates a new church member repository with a DB handle.
func NewChurchMemberRepository(db *sql.DB) *ChurchMemberRepository {
	return &ChurchMemberRepository{base: NewBaseRepository(db)}
}

// Create inserts a new church member and returns the new ID.
func (r *ChurchMemberRepository) Create(ctx context.Context, m *model.ChurchMember) (int64, error) {
	now := time.Now().UTC()
	var id int64
	err := r.base.ScanRow(ctx,
		`INSERT INTO church_members (name, email, phone, address, biography, joined_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		func(row *sql.Row) error {
			return row.Scan(&id)
		},
		m.Name, m.Email, m.Phone, m.Address, m.Biography, m.JoinedAt, now, now,
	)
	return id, err
}

// GetByID returns a single church member by ID.
func (r *ChurchMemberRepository) GetByID(ctx context.Context, id int64) (*model.ChurchMember, error) {
	var m model.ChurchMember
	err := r.base.ScanRow(ctx,
		`SELECT id, name, email, phone, address, biography, joined_at, created_at, updated_at
		 FROM church_members WHERE id = $1`,
		func(row *sql.Row) error {
			return row.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Address, &m.Biography, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt)
		},
		id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// GetByEmail returns a church member by email.
func (r *ChurchMemberRepository) GetByEmail(ctx context.Context, email string) (*model.ChurchMember, error) {
	var m model.ChurchMember
	err := r.base.ScanRow(ctx,
		`SELECT id, name, email, phone, address, biography, joined_at, created_at, updated_at
		 FROM church_members WHERE email = $1`,
		func(row *sql.Row) error {
			return row.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Address, &m.Biography, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt)
		},
		email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// Update modifies an existing church member's information.
func (r *ChurchMemberRepository) Update(ctx context.Context, m *model.ChurchMember) error {
	now := time.Now().UTC()
	return r.base.ExecUpdate(ctx,
		`UPDATE church_members SET name=$1, email=$2, phone=$3, address=$4, biography=$5, updated_at=$6
		 WHERE id=$7`,
		m.Name, m.Email, m.Phone, m.Address, m.Biography, now, m.ID,
	)
}

// Delete removes a church member by ID.
func (r *ChurchMemberRepository) Delete(ctx context.Context, id int64) error {
	return r.base.ExecUpdate(ctx,
		`DELETE FROM church_members WHERE id=$1`,
		id,
	)
}

// List returns all church members, ordered by joined_at (newest first).
func (r *ChurchMemberRepository) List(ctx context.Context) ([]*model.ChurchMember, error) {
	var members []*model.ChurchMember
	err := r.base.ScanRows(ctx,
		`SELECT id, name, email, phone, address, biography, joined_at, created_at, updated_at
		 FROM church_members ORDER BY joined_at DESC`,
		func(rows *sql.Rows) error {
			for rows.Next() {
				var m model.ChurchMember
				if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Address, &m.Biography, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
					return err
				}
				members = append(members, &m)
			}
			return rows.Err()
		},
	)
	return members, err
}

// ListByJoinedDateRange returns church members joined within a date range.
func (r *ChurchMemberRepository) ListByJoinedDateRange(ctx context.Context, startDate, endDate time.Time) ([]*model.ChurchMember, error) {
	var members []*model.ChurchMember
	err := r.base.ScanRows(ctx,
		`SELECT id, name, email, phone, address, biography, joined_at, created_at, updated_at
		 FROM church_members WHERE joined_at >= $1 AND joined_at <= $2 ORDER BY joined_at DESC`,
		func(rows *sql.Rows) error {
			for rows.Next() {
				var m model.ChurchMember
				if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.Address, &m.Biography, &m.JoinedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
					return err
				}
				members = append(members, &m)
			}
			return rows.Err()
		},
		startDate, endDate,
	)
	return members, err
}
