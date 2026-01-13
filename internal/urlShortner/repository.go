package urlShortener

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository interface {
	Add(ctx context.Context, UrlData UrlDbModel) error
	GetVal(ctx context.Context, id string) (*UrlDbModel, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Add(ctx context.Context, urlData UrlDbModel) error {
	query := `INSERT INTO urls (url,shortCode,createdAt) VALUES ($1,$2,$3)`
	_, err := r.db.ExecContext(ctx, query, urlData.Url, urlData.ShortCode, urlData.CreatedAt)
	return err
}

func (r *PostgresRepository) GetVal(ctx context.Context, code string) (*UrlDbModel, error) {
	query := `SELECT id,url,shortCode,CreatedAt FROM urls WHERE shortCode = $1`

	rows := r.db.QueryRowContext(ctx, query, code)
	var Entry UrlDbModel
	if err := rows.Scan(&Entry.Id, &Entry.Url, &Entry.ShortCode, &Entry.CreatedAt); err != nil {
		return nil, err
	}
	return &Entry, nil
}
