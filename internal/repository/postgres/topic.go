package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	errs "eshkere/internal/errors"
)

const selectTopicIDByName = `SELECT id FROM eshkere.topic WHERE name = $1`

type TopicRepository struct {
	db *sql.DB
}

func NewTopicRepository(db *sql.DB) *TopicRepository {
	return &TopicRepository{db: db}
}

func (r *TopicRepository) GetIDByName(ctx context.Context, name string) (int, error) {
	var id int
	if err := r.db.QueryRowContext(ctx, selectTopicIDByName, name).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: topic %q not found", errs.NotFoundError, name)
		}
		return 0, fmt.Errorf("get topic id by name: %w", err)
	}
	return id, nil
}
