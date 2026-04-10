package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"fmt"
)

type AdvertiserRepository struct {
	db *sql.DB
}

func NewAdvertiserRepository(db *sql.DB) *AdvertiserRepository {
	return &AdvertiserRepository{db: db}
}

const (
	insertAdvertiser = `INSERT INTO eshkere.advertiser (
		name, email, phone_number, password_hash, password_salt, balance) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	selectAdvertiserByID = `SELECT
        id, name, email, phone_number, password_hash, password_salt, balance, created_at, updated_at
    FROM eshkere.advertiser
    WHERE id = $1`

	selectAdvertiserByEmail = `SELECT
	id, name, email, phone_number, password_hash, password_salt, balance, created_at, updated_at
	FROM eshkere.advertiser
	WHERE email = $1`

	updateAdvertiser = `UPDATE eshkere.advertiser SET 
    name = $1, email = $2, phone_number = $3, password_hash = $4, password_salt = $5, balance = $6
    WHERE id = $7`

	deleteAdvertiser = `DELETE FROM eshkere.advertiser WHERE id = $1`
)

func (r *AdvertiserRepository) Create(ctx context.Context, a *models.Advertiser) (int, error) {
	if a == nil {
		return 0, fmt.Errorf("advertiser cannot be nil")
	}

	err := r.db.QueryRowContext(ctx, insertAdvertiser,
		a.Name, a.Email, a.Phone, a.PasswordHash, a.PasswordSalt, a.Balance,
	).Scan(&a.ID)
	if err != nil {
		return 0, fmt.Errorf("insert advertiser: %w", err)
	}

	return a.ID, nil
}

func (r *AdvertiserRepository) GetByID(ctx context.Context, id int) (*models.Advertiser, error) {
	var a models.Advertiser

	err := r.db.QueryRowContext(ctx, selectAdvertiserByID, id).Scan(
		&a.ID,
		&a.Name,
		&a.Email,
		&a.Phone,
		&a.PasswordHash,
		&a.PasswordSalt,
		&a.Balance,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("advertiser not found: %w", err)
		}
		return nil, fmt.Errorf("get advertiser by id: %w", err)
	}

	return &a, nil
}

func (r *AdvertiserRepository) GetByEmail(ctx context.Context, email string) (*models.Advertiser, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var a models.Advertiser
	err := r.db.QueryRowContext(ctx, selectAdvertiserByEmail, email).Scan(
		&a.ID,
		&a.Name,
		&a.Email,
		&a.Phone,
		&a.PasswordHash,
		&a.PasswordSalt,
		&a.Balance,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("advertiser not found: %w", err)
		}
		return nil, fmt.Errorf("get advertiser by email: %w", err)
	}

	return &a, nil
}

func (r *AdvertiserRepository) Update(ctx context.Context, a *models.Advertiser) error {
	if a == nil {
		return fmt.Errorf("advertiser cannot be nil")
	}

	_, err := r.db.ExecContext(ctx, updateAdvertiser, a.Name, a.Email, a.Phone, a.PasswordHash, a.PasswordSalt, a.Balance, a.ID)
	if err != nil {
		return fmt.Errorf("update advertiser: %w", err)
	}

	return nil
}

func (r *AdvertiserRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.ExecContext(ctx, deleteAdvertiser, id)
	if err != nil {
		return fmt.Errorf("delete advertiser: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("advertiser not found")
	}

	return nil
}
