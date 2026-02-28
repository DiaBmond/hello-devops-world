package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-prod-app/internal/domain"
	"go-prod-app/internal/repository"
)

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrUserNotFound   = repository.ErrUserNotFound
	ErrDuplicateEmail = repository.ErrDuplicateEmail
	ErrConflict       = repository.ErrVersionConflict
)

type UserService struct {
	repo   repository.UserRepository
	health repository.HealthChecker
}

func NewUserService(
	repo repository.UserRepository,
	health repository.HealthChecker,
) *UserService {
	return &UserService{
		repo:   repo,
		health: health,
	}
}

//
// =========================
// CreateUser
// =========================
//

func (s *UserService) CreateUser(
	ctx context.Context,
	name string,
	email string,
) (*domain.User, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()

	user, err := domain.NewUser(name, email, now)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

//
// =========================
// UpdateUser
// =========================
//

func (s *UserService) UpdateUser(
	ctx context.Context,
	id domain.UserID,
	name string,
	email string,
) (*domain.User, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.IsDeleted() {
		return nil, ErrUserNotFound
	}

	now := time.Now().UTC()

	if err := user.ChangeName(name, now); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := user.ChangeEmail(email, now); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

//
// =========================
// DeleteUser (Soft Delete)
// =========================
//

func (s *UserService) DeleteUser(
	ctx context.Context,
	id domain.UserID,
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if user.IsDeleted() {
		return ErrUserNotFound
	}

	now := time.Now().UTC()

	if err := user.Delete(now); err != nil {
		return err
	}

	return s.repo.Update(ctx, user)
}

//
// =========================
// GetUser
// =========================
//

func (s *UserService) GetUser(
	ctx context.Context,
	id domain.UserID,
) (*domain.User, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.IsDeleted() {
		return nil, ErrUserNotFound
	}

	return user, nil
}

//
// =========================
// GetByEmail
// =========================
//

func (s *UserService) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.repo.GetByEmail(ctx, email)
}

//
// =========================
// ListUsers
// =========================
//

func (s *UserService) ListUsers(
	ctx context.Context,
	filter repository.UserFilter,
	cursor *repository.Cursor,
	limit int,
) ([]*domain.User, *repository.Cursor, error) {

	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	return s.repo.List(ctx, filter, cursor, limit)
}

//
// =========================
// CountUsers
// =========================
//

func (s *UserService) CountUsers(
	ctx context.Context,
	filter repository.UserFilter,
) (int64, error) {

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	return s.repo.Count(ctx, filter)
}

func (s *UserService) Ping(ctx context.Context) error {
	return s.health.Ping(ctx)
}
