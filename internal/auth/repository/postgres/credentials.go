package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Credential struct {
	ID           int64
	Email        string
	Phone        string
	PasswordHash string
	VKUserID     sql.NullInt64
	CreatedAt    time.Time
}

type CredentialsRepository struct {
	db *sql.DB
}

func NewCredentialsRepository(db *sql.DB) *CredentialsRepository {
	return &CredentialsRepository{db: db}
}

func (r *CredentialsRepository) Create(ctx context.Context, email, phone, hash string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO auth.credentials (email, phone, password_hash) VALUES ($1, $2, $3) RETURNING id`,
		email, phone, hash,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create credentials: %w", err)
	}
	return id, nil
}

func (r *CredentialsRepository) CreateVK(ctx context.Context, email, phone string, vkUserID int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO auth.credentials (email, phone, password_hash, vk_user_id)
		 VALUES ($1, $2, NULL, $3)
		 RETURNING id`,
		nullableText(email), nullableText(phone), vkUserID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create vk credentials: %w", err)
	}
	return id, nil
}

func (r *CredentialsRepository) GetByEmail(ctx context.Context, email string) (*Credential, error) {
	c := &Credential{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE email = $1`,
		email,
	).Scan(&c.ID, &c.Email, &c.Phone, &c.PasswordHash, &c.VKUserID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get by email: %w", err)
	}
	return c, nil
}

func (r *CredentialsRepository) GetByPhone(ctx context.Context, phone string) (*Credential, error) {
	c := &Credential{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE phone = $1`,
		phone,
	).Scan(&c.ID, &c.Email, &c.Phone, &c.PasswordHash, &c.VKUserID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get by phone: %w", err)
	}
	return c, nil
}

func (r *CredentialsRepository) GetByVKUserID(ctx context.Context, vkUserID int64) (*Credential, error) {
	c := &Credential{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE vk_user_id = $1`,
		vkUserID,
	).Scan(&c.ID, &c.Email, &c.Phone, &c.PasswordHash, &c.VKUserID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get by vk user id: %w", err)
	}
	return c, nil
}

func (r *CredentialsRepository) GetByID(ctx context.Context, id int64) (*Credential, error) {
	c := &Credential{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Email, &c.Phone, &c.PasswordHash, &c.VKUserID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get by id: %w", err)
	}
	return c, nil
}

func (r *CredentialsRepository) LinkVKUserID(ctx context.Context, id, vkUserID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE auth.credentials
		 SET vk_user_id = $2, updated_at = NOW()
		 WHERE id = $1`,
		id, vkUserID,
	)
	if err != nil {
		return fmt.Errorf("link vk user id: %w", err)
	}
	return nil
}

func (r *CredentialsRepository) Update(ctx context.Context, id int64, email, phone string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE auth.credentials SET email = $2, phone = $3, updated_at = NOW() WHERE id = $1`,
		id, email, phone,
	)
	if err != nil {
		return fmt.Errorf("update credentials: %w", err)
	}
	return nil
}

func (r *CredentialsRepository) UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE auth.credentials SET password_hash = $2, updated_at = NOW() WHERE id = $1`,
		id, passwordHash,
	)
	if err != nil {
		return fmt.Errorf("update password hash: %w", err)
	}
	return nil
}

func (r *CredentialsRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM auth.credentials WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete credentials: %w", err)
	}
	return nil
}

func nullableText(v string) any {
	if v == "" {
		return nil
	}
	return v
}
