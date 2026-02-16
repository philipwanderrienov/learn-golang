package model

import "time"

// User represents the users table in the database.
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name" validate:"required,min=1,max=255"`
	Email     string    `json:"email" validate:"required,email"`
	CreatedAt time.Time `json:"created_at"`
}
