package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/apperror"
	"github.com/yourusername/ai-interview/services/job-candidate-service/internal/model"
)

// UserRepository handles persistence for organizations and users.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository constructs a UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// CreateOrgAndUser creates an organization and its first user in a single transaction.
// Returns ErrConflict if org_slug or email already exists.
func (r *UserRepository) CreateOrgAndUser(
	ctx context.Context,
	orgName, orgSlug, email, passwordHash, fullName string,
) (*model.Organization, *model.User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("UserRepository.CreateOrgAndUser: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	org := &model.Organization{}
	err = tx.QueryRow(ctx,
		`INSERT INTO organizations (name, slug)
		 VALUES ($1, $2)
		 RETURNING id, name, slug, created_at, updated_at`,
		orgName, orgSlug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, fmt.Errorf("UserRepository.CreateOrgAndUser: %w", apperror.ErrConflict)
		}
		return nil, nil, fmt.Errorf("UserRepository.CreateOrgAndUser: insert org: %w", err)
	}

	user := &model.User{}
	err = tx.QueryRow(ctx,
		`INSERT INTO users (org_id, email, password_hash, full_name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, org_id, email, password_hash, full_name, created_at, updated_at`,
		org.ID, email, passwordHash, fullName,
	).Scan(&user.ID, &user.OrgID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, nil, fmt.Errorf("UserRepository.CreateOrgAndUser: %w", apperror.ErrConflict)
		}
		return nil, nil, fmt.Errorf("UserRepository.CreateOrgAndUser: insert user: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("UserRepository.CreateOrgAndUser: commit: %w", err)
	}

	return org, user, nil
}

// GetByEmail fetches a user by email address.
// Returns ErrNotFound if no matching user exists.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRow(ctx,
		`SELECT id, org_id, email, password_hash, full_name, created_at, updated_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.OrgID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("UserRepository.GetByEmail: %w", apperror.ErrNotFound)
		}
		return nil, fmt.Errorf("UserRepository.GetByEmail: %w", err)
	}
	return user, nil
}
