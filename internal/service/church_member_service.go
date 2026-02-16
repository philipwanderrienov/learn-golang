package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/example/golang-project/internal/model"
	"github.com/example/golang-project/internal/repository"
)

// ChurchMemberService contains business logic for church members.
type ChurchMemberService struct {
	repo *repository.ChurchMemberRepository
}

// NewChurchMemberService constructs a new ChurchMemberService.
func NewChurchMemberService(r *repository.ChurchMemberRepository) *ChurchMemberService {
	return &ChurchMemberService{repo: r}
}

// CreateMember validates and creates a new church member, returning the created ID.
func (s *ChurchMemberService) CreateMember(ctx context.Context, m *model.ChurchMember) (int64, error) {
	// Validate input
	if err := s.validateMember(m); err != nil {
		return 0, err
	}

	// Check if email already exists
	existing, err := s.repo.GetByEmail(ctx, m.Email)
	if err != nil {
		return 0, err
	}
	if existing != nil {
		return 0, errors.New("email already exists")
	}

	// Set default joined_at to now if not provided
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now().UTC()
	}

	return s.repo.Create(ctx, m)
}

// GetMember returns a church member by ID.
func (s *ChurchMemberService) GetMember(ctx context.Context, id int64) (*model.ChurchMember, error) {
	if id <= 0 {
		return nil, errors.New("invalid member id")
	}
	return s.repo.GetByID(ctx, id)
}

// UpdateMember updates an existing church member's information.
func (s *ChurchMemberService) UpdateMember(ctx context.Context, m *model.ChurchMember) error {
	if m.ID <= 0 {
		return errors.New("invalid member id")
	}

	// Validate input
	if err := s.validateMember(m); err != nil {
		return err
	}

	// Check if member exists
	existing, err := s.repo.GetByID(ctx, m.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("member not found")
	}

	// Check if new email is already taken by another member
	if m.Email != existing.Email {
		emailExists, err := s.repo.GetByEmail(ctx, m.Email)
		if err != nil {
			return err
		}
		if emailExists != nil {
			return errors.New("email already exists")
		}
	}

	return s.repo.Update(ctx, m)
}

// DeleteMember removes a church member by ID.
func (s *ChurchMemberService) DeleteMember(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid member id")
	}
	return s.repo.Delete(ctx, id)
}

// ListMembers returns all church members.
func (s *ChurchMemberService) ListMembers(ctx context.Context) ([]*model.ChurchMember, error) {
	return s.repo.List(ctx)
}

// ListMembersByJoinedDate returns members joined within a date range.
func (s *ChurchMemberService) ListMembersByJoinedDate(ctx context.Context, startDate, endDate time.Time) ([]*model.ChurchMember, error) {
	if startDate.After(endDate) {
		return nil, errors.New("start date must be before end date")
	}
	return s.repo.ListByJoinedDateRange(ctx, startDate, endDate)
}

// validateMember checks if the member data is valid.
func (s *ChurchMemberService) validateMember(m *model.ChurchMember) error {
	// Validate name
	name := strings.TrimSpace(m.Name)
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) < 2 || len(name) > 255 {
		return errors.New("name must be between 2 and 255 characters")
	}

	// Validate email
	email := strings.TrimSpace(m.Email)
	if email == "" {
		return errors.New("email is required")
	}
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}

	// Validate phone (optional, but if provided must be reasonable)
	if m.Phone != "" && len(m.Phone) > 20 {
		return errors.New("phone must not exceed 20 characters")
	}

	// Validate address (optional)
	if len(m.Address) > 500 {
		return errors.New("address must not exceed 500 characters")
	}

	// Validate biography (optional)
	if len(m.Biography) > 5000 {
		return errors.New("biography must not exceed 5000 characters")
	}

	return nil
}

// isValidEmail checks if an email format is valid using regex.
func isValidEmail(email string) bool {
	// Simple email regex validation
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}
