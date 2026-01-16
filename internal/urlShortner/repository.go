package urlShortener

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Repository interface {
	Add(ctx context.Context, UrlData UrlDbModel) error
	GetVal(ctx context.Context, id string) (*UrlDbModel, error)
	GetByUser(ctx context.Context, userId int) ([]UrlDbModel, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Add(ctx context.Context, urlData UrlDbModel) error {
	query := `INSERT INTO urls (url,shortCode,createdAt,user_id) VALUES ($1,$2,$3,$4)`
	_, err := r.db.ExecContext(ctx, query, urlData.Url, urlData.ShortCode, urlData.CreatedAt, urlData.UserID)
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

func (r *PostgresRepository) GetByUser(ctx context.Context, userID int) ([]UrlDbModel, error) {
	query := `SELECT id, url, shortCode, createdAt FROM urls WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []UrlDbModel
	for rows.Next() {
		var u UrlDbModel
		if err := rows.Scan(&u.Id, &u.Url, &u.ShortCode, &u.CreatedAt); err != nil {
			return nil, err
		}
		urls = append(urls, u)
	}
	return urls, nil
}
