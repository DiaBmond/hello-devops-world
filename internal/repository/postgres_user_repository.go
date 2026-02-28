package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go-prod-app/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const maxListLimit = 1000

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

//
// =========================
// Create
// =========================
// Repository owns ID + persistence metadata
//

func (r *PostgresUserRepository) Create(
	ctx context.Context,
	user *domain.User,
) error {

	now := time.Now().UTC()

	// UUID v7 → sortable by time
	id := uuid.Must(uuid.NewV7())

	query := `
		INSERT INTO users (
			id, name, email, version,
			created_at, updated_at, deleted_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id
	`

	var returnedID string

	err := r.db.QueryRowContext(
		ctx,
		query,
		id.String(),
		user.Name(),
		user.Email(),
		1,   // initial version
		now, // created_at
		now, // updated_at
		nil,
	).Scan(&returnedID)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateEmail
		}
		return err
	}

	// Set ID only AFTER successful insert
	if err := user.SetID(domain.UserID(returnedID)); err != nil {
		return err
	}

	return nil
}

//
// =========================
// Update (Optimistic Lock)
// Repository owns version + updated_at
//

func (r *PostgresUserRepository) Update(
	ctx context.Context,
	user *domain.User,
) error {

	now := time.Now().UTC()
	newVersion := user.Version() + 1

	query := `
		UPDATE users
		SET name = $1,
			email = $2,
			version = $3,
			updated_at = $4,
			deleted_at = $5
		WHERE id = $6
		  AND version = $7
	`

	res, err := r.db.ExecContext(
		ctx,
		query,
		user.Name(),
		user.Email(),
		newVersion,
		now,
		user.DeletedAt(),
		user.ID(),
		user.Version(),
	)

	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateEmail
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrVersionConflict
	}

	user.IncreaseVersion()
	return nil
}

//
// =========================
// GetByID
// Returns user even if soft-deleted
//

func (r *PostgresUserRepository) GetByID(
	ctx context.Context,
	id domain.UserID,
) (*domain.User, error) {

	query := `
		SELECT id, name, email, version,
		       created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

//
// =========================
// GetByEmail
// Returns only active users
//

func (r *PostgresUserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (*domain.User, error) {

	query := `
		SELECT id, name, email, version,
		       created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1
		  AND deleted_at IS NULL
	`

	row := r.db.QueryRowContext(ctx, query, strings.ToLower(email))

	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

//
// =========================
// List (Keyset Pagination)
// Ordered by id ASC (UUID v7 → time-ordered)
//

func (r *PostgresUserRepository) List(
	ctx context.Context,
	filter UserFilter,
	cursor *Cursor,
	limit int,
) ([]*domain.User, *Cursor, error) {

	if limit <= 0 {
		return nil, nil, fmt.Errorf("limit must be > 0")
	}
	if limit > maxListLimit {
		limit = maxListLimit
	}

	var (
		args       []interface{}
		conditions []string
	)

	if !filter.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	if filter.Email != nil {
		args = append(args, strings.ToLower(*filter.Email))
		conditions = append(conditions,
			fmt.Sprintf("email = $%d", len(args)))
	}

	if filter.CreatedAfter != nil {
		args = append(args, *filter.CreatedAfter)
		conditions = append(conditions,
			fmt.Sprintf("created_at > $%d", len(args)))
	}

	if filter.CreatedBefore != nil {
		args = append(args, *filter.CreatedBefore)
		conditions = append(conditions,
			fmt.Sprintf("created_at < $%d", len(args)))
	}

	if cursor != nil {
		args = append(args, cursor.AfterID)
		conditions = append(conditions,
			fmt.Sprintf("id > $%d", len(args)))
	}

	// limit parameterized
	args = append(args, limit)
	limitParam := len(args)

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, name, email, version,
		       created_at, updated_at, deleted_at
		FROM users
		%s
		ORDER BY id ASC
		LIMIT $%d
	`, where, limitParam)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var users []*domain.User
	var lastID domain.UserID

	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, nil, err
		}
		users = append(users, u)
		lastID = u.ID()
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var nextCursor *Cursor
	if len(users) == limit {
		nextCursor = &Cursor{AfterID: lastID}
	}

	return users, nextCursor, nil
}

//
// =========================
// Count
// =========================
//

func (r *PostgresUserRepository) Count(
	ctx context.Context,
	filter UserFilter,
) (int64, error) {

	var (
		args       []interface{}
		conditions []string
	)

	if !filter.IncludeDeleted {
		conditions = append(conditions, "deleted_at IS NULL")
	}

	if filter.Email != nil {
		args = append(args, strings.ToLower(*filter.Email))
		conditions = append(conditions,
			fmt.Sprintf("email = $%d", len(args)))
	}

	if filter.CreatedAfter != nil {
		args = append(args, *filter.CreatedAfter)
		conditions = append(conditions,
			fmt.Sprintf("created_at > $%d", len(args)))
	}

	if filter.CreatedBefore != nil {
		args = append(args, *filter.CreatedBefore)
		conditions = append(conditions,
			fmt.Sprintf("created_at < $%d", len(args)))
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM users %s`, where)

	var count int64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

//
// =========================
// Helpers
// =========================
//

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanUser(s scanner) (*domain.User, error) {
	var (
		id        string
		name      string
		email     string
		version   int
		createdAt time.Time
		updatedAt time.Time
		deletedAt *time.Time
	)

	if err := s.Scan(
		&id,
		&name,
		&email,
		&version,
		&createdAt,
		&updatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	return domain.RehydrateUser(
		domain.UserID(id),
		name,
		email,
		version,
		createdAt,
		updatedAt,
		deletedAt,
	), nil
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func (r *PostgresUserRepository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
