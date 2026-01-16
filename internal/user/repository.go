package user

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository interface {
	Create(ctx context.Context, u *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, u *User) error {
	query := `INSERT INTO users (email, password, created_at) VALUES ($1, $2, $3) RETURNING id`

	err := r.db.QueryRowContext(ctx, query, u.Email, u.Password, u.CreatedAt).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `SELECT id, email, password, created_at FROM users WHERE email = $1`

	u := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}
