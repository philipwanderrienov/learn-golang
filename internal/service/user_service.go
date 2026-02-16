package service

import (
	"context"

	"github.com/example/golang-project/internal/model"
	"github.com/example/golang-project/internal/repository"
)

// UserService contains business logic for users. It delegates persistence to the repository.
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService constructs a new UserService.
func NewUserService(r *repository.UserRepository) *UserService {
	return &UserService{repo: r}
}

// CreateUser validates and creates a new user, returning the created ID.
func (s *UserService) CreateUser(ctx context.Context, u *model.User) (int64, error) {
	// In a .NET style you'd validate DTOs here; keep simple and delegate to repo.
	return s.repo.Create(ctx, u)
}

// GetUser returns a user by ID.
func (s *UserService) GetUser(ctx context.Context, id int64) (*model.User, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateUser updates an existing user.
func (s *UserService) UpdateUser(ctx context.Context, u *model.User) error {
	return s.repo.Update(ctx, u)
}

// DeleteUser removes a user by ID.
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// ListUsers returns all users.
func (s *UserService) ListUsers(ctx context.Context) ([]*model.User, error) {
	return s.repo.List(ctx)
}
