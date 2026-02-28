package repository

import (
	"context"
	"errors"
	"time"

	"go-prod-app/internal/domain"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrDuplicateEmail  = errors.New("duplicate email")
	ErrVersionConflict = errors.New("version conflict")
)

//
// =========
// Filtering
// =========
//

// UserFilter defines query constraints for listing users.
//
// If IncludeDeleted is false, soft-deleted users MUST be excluded.
type UserFilter struct {
	IncludeDeleted bool

	Email         *string
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// =========
// Cursor (Keyset Pagination)
// =========
//
// List is ordered by (id ASC) for stable pagination.
// Cursor represents the last seen record.
type Cursor struct {
	AfterID domain.UserID
}

//
// =========
// Repository Contract
// =========
//

type UserRepository interface {
	// =====================
	// Write Operations
	// =====================

	// Create persists a new user.
	// Must return ErrDuplicateEmail if unique constraint violated.
	Create(ctx context.Context, user *domain.User) error

	// Update persists changes using optimistic locking.
	// Must return ErrVersionConflict if version mismatch.
	Update(ctx context.Context, user *domain.User) error

	// =====================
	// Read Operations
	// =====================

	// GetByID returns a user by ID.
	// Must return ErrUserNotFound if not found.
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)

	// GetByEmail returns a user by normalized email.
	// Must return ErrUserNotFound if not found.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// List returns users ordered by ID ASC.
	//
	// - limit must be > 0
	// - keyset pagination via cursor
	// - if filter.IncludeDeleted is false, deleted users are excluded
	List(
		ctx context.Context,
		filter UserFilter,
		cursor *Cursor,
		limit int,
	) ([]*domain.User, *Cursor, error)

	// Count returns total number of users matching filter.
	Count(ctx context.Context, filter UserFilter) (int64, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}
