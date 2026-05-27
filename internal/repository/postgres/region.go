package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	errs "eshkere/internal/errors"
)

const selectRegionIDByName = `SELECT id FROM eshkere.region WHERE name = $1`

type RegionRepository struct {
	db *sql.DB
}

func NewRegionRepository(db *sql.DB) *RegionRepository {
	return &RegionRepository{db: db}
}

func (r *RegionRepository) GetIDByName(ctx context.Context, name string) (int, error) {
	var id int
	if err := r.db.QueryRowContext(ctx, selectRegionIDByName, name).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: region %q not found", errs.NotFoundError, name)
		}
		return 0, fmt.Errorf("get region id by name: %w", err)
	}
	return id, nil
}
