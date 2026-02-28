package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

//
// =========
// Identity
// =========
//

type UserID string

//
// =========
// Errors
// =========
//

var (
	ErrInvalidName        = errors.New("invalid name")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrUserAlreadyDeleted = errors.New("user already deleted")
	ErrUserNotDeleted     = errors.New("user is not deleted")
	ErrIDAlreadySet       = errors.New("id already set")
)

//
// =========
// Entity
// =========
//

type User struct {
	id      UserID
	name    string
	email   string
	version int

	createdAt time.Time
	updatedAt time.Time
	deletedAt *time.Time
}

var emailRegex = regexp.MustCompile(
	`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`,
)

//
// =========
// Constructors
// =========
//

// NewUser is used when creating new entity (business flow)
func NewUser(name, email string, now time.Time) (*User, error) {
	name = strings.TrimSpace(name)
	email = normalizeEmail(email)

	if err := validateName(name); err != nil {
		return nil, err
	}

	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}

	return &User{
		name:      name,
		email:     email,
		version:   1,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// RehydrateUser is used by repository to rebuild entity from storage
func RehydrateUser(
	id UserID,
	name string,
	email string,
	version int,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *User {
	return &User{
		id:        id,
		name:      name,
		email:     email,
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
		deletedAt: deletedAt,
	}
}

//
// =========
// Behaviors
// =========
//

func (u *User) Delete(now time.Time) error {
	if u.deletedAt != nil {
		return ErrUserAlreadyDeleted
	}

	u.deletedAt = &now
	u.updatedAt = now
	return nil
}

func (u *User) Restore(now time.Time) error {
	if u.deletedAt == nil {
		return ErrUserNotDeleted
	}

	u.deletedAt = nil
	u.updatedAt = now
	return nil
}

func (u *User) ChangeEmail(newEmail string, now time.Time) error {
	if u.deletedAt != nil {
		return ErrUserAlreadyDeleted
	}

	newEmail = normalizeEmail(newEmail)

	if !emailRegex.MatchString(newEmail) {
		return ErrInvalidEmail
	}

	if u.email == newEmail {
		return nil
	}

	u.email = newEmail
	u.updatedAt = now
	return nil
}

func (u *User) ChangeName(newName string, now time.Time) error {
	if u.deletedAt != nil {
		return ErrUserAlreadyDeleted
	}

	newName = strings.TrimSpace(newName)

	if err := validateName(newName); err != nil {
		return err
	}

	if u.name == newName {
		return nil
	}

	u.name = newName
	u.updatedAt = now
	return nil
}

//
// =========
// Version Control
// =========
//

// IncreaseVersion should be called by repository AFTER successful update
func (u *User) IncreaseVersion() {
	u.version++
}

//
// =========
// Repository Control
// =========
//

func (u *User) SetID(id UserID) error {
	if u.id != "" {
		return ErrIDAlreadySet
	}
	u.id = id
	return nil
}

//
// =========
// Getters
// =========
//

func (u *User) ID() UserID            { return u.id }
func (u *User) Name() string          { return u.name }
func (u *User) Email() string         { return u.email }
func (u *User) Version() int          { return u.version }
func (u *User) CreatedAt() time.Time  { return u.createdAt }
func (u *User) UpdatedAt() time.Time  { return u.updatedAt }
func (u *User) DeletedAt() *time.Time { return u.deletedAt }
func (u *User) IsDeleted() bool       { return u.deletedAt != nil }

//
// =========
// Validation
// =========ex``
//

func validateName(name string) error {
	if len(name) < 2 || len(name) > 100 {
		return ErrInvalidName
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
