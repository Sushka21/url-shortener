package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sushka21/url-shortener/internal/entity"
	sqlcShortenet "github.com/Sushka21/url-shortener/internal/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
)

type (
	DB interface {
		sqlcShortenet.DBTX
	}
)

type postgresRepository struct {
	queries sqlcShortenet.Querier
	db      DB
}

func NewPostgresRepository(qdb DB) *postgresRepository {
	return &postgresRepository{
		queries: sqlcShortenet.New(qdb),
		db:      qdb,
	}
}

func (r *postgresRepository) Save(ctx context.Context, shortKey string, longURL string) error {
	res, err := r.queries.SaveShortKey(ctx, sqlcShortenet.SaveShortKeyParams{
		ShortKey: shortKey,
		LongUrl:  longURL,
	})

	if err != nil {
		return fmt.Errorf("failed to execute save query: %w", err)
	}

	if res.RowsAffected() == 0 {
		return entity.ErrConflictURL
	}

	return nil
}

func (r *postgresRepository) GetByShortKey(ctx context.Context, shortKey string) (string, error) {
	longURL, err := r.queries.GetByShortKey(ctx, shortKey)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrURLNotFound
		}
		return "", fmt.Errorf("failed to get long url by short key: %w", err)
	}

	return longURL, nil
}
