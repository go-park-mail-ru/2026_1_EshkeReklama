package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	errs "eshkere/internal/errors"
)

type TopicRepository struct {
	db *sql.DB
}

func NewTopicRepository(db *sql.DB) *TopicRepository {
	return &TopicRepository{db: db}
}

func (r *TopicRepository) GetIDByName(ctx context.Context, name string) (int, error) {
	return dictionaryIDByName(ctx, r.db, "topic", name)
}

type RegionRepository struct {
	db *sql.DB
}

func NewRegionRepository(db *sql.DB) *RegionRepository {
	return &RegionRepository{db: db}
}

func (r *RegionRepository) GetIDByName(ctx context.Context, name string) (int, error) {
	return dictionaryIDByName(ctx, r.db, "region", name)
}

func dictionaryIDByName(ctx context.Context, db *sql.DB, table string, name string) (int, error) {
	query := fmt.Sprintf("SELECT id FROM eshkere.%s WHERE name = $1", table)

	var id int
	if err := db.QueryRowContext(ctx, query, name).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: %s %q not found", errs.NotFoundError, table, name)
		}
		return 0, fmt.Errorf("get %s id by name: %w", table, err)
	}
	return id, nil
}
