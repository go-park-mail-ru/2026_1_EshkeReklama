package postgres

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCredentialsRepositoryCRUD(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewCredentialsRepository(db)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO auth.credentials (email, phone, password_hash) VALUES ($1, $2, $3) RETURNING id`)).
		WithArgs("user@example.com", "9001234567", "hash").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	id, err := repo.Create(context.Background(), "user@example.com", "9001234567", "hash")
	if err != nil || id != 1 {
		t.Fatalf("create: id=%d err=%v", id, err)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO auth.credentials (email, phone, password_hash, vk_user_id)
		 VALUES ($1, $2, NULL, $3)
		 RETURNING id`)).
		WithArgs(nil, "9001234567", int64(77)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	id, err = repo.CreateVK(context.Background(), "", "9001234567", 77)
	if err != nil || id != 2 {
		t.Fatalf("create vk: id=%d err=%v", id, err)
	}

	rows := sqlmock.NewRows([]string{"id", "coalesce", "coalesce", "coalesce", "vk_user_id", "created_at"}).
		AddRow(3, "user@example.com", "9001234567", "hash", sql.NullInt64{Int64: 88, Valid: true}, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE email = $1`)).
		WithArgs("user@example.com").
		WillReturnRows(rows)
	cred, err := repo.GetByEmail(context.Background(), "user@example.com")
	if err != nil || cred.ID != 3 || cred.Email != "user@example.com" {
		t.Fatalf("get by email: %#v %v", cred, err)
	}

	rows = sqlmock.NewRows([]string{"id", "coalesce", "coalesce", "coalesce", "vk_user_id", "created_at"}).
		AddRow(4, "user@example.com", "9001234567", "hash", sql.NullInt64{}, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE phone = $1`)).
		WithArgs("9001234567").
		WillReturnRows(rows)
	if cred, err = repo.GetByPhone(context.Background(), "9001234567"); err != nil || cred.ID != 4 {
		t.Fatalf("get by phone: %#v %v", cred, err)
	}

	rows = sqlmock.NewRows([]string{"id", "coalesce", "coalesce", "coalesce", "vk_user_id", "created_at"}).
		AddRow(5, "vk@example.com", "", "", sql.NullInt64{Int64: 99, Valid: true}, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE vk_user_id = $1`)).
		WithArgs(int64(99)).
		WillReturnRows(rows)
	if cred, err = repo.GetByVKUserID(context.Background(), 99); err != nil || cred.ID != 5 {
		t.Fatalf("get by vk user id: %#v %v", cred, err)
	}

	rows = sqlmock.NewRows([]string{"id", "coalesce", "coalesce", "coalesce", "vk_user_id", "created_at"}).
		AddRow(6, "user@example.com", "9001234567", "hash", sql.NullInt64{}, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, COALESCE(email, ''), COALESCE(phone, ''), COALESCE(password_hash, ''), vk_user_id, created_at
		 FROM auth.credentials
		 WHERE id = $1`)).
		WithArgs(int64(6)).
		WillReturnRows(rows)
	if cred, err = repo.GetByID(context.Background(), 6); err != nil || cred.ID != 6 {
		t.Fatalf("get by id: %#v %v", cred, err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE auth.credentials
		 SET vk_user_id = $2, updated_at = NOW()
		 WHERE id = $1`)).
		WithArgs(int64(6), int64(100)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.LinkVKUserID(context.Background(), 6, 100); err != nil {
		t.Fatalf("link vk user id: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE auth.credentials SET email = $2, phone = $3, updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(6), "new@example.com", "9111111111").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Update(context.Background(), 6, "new@example.com", "9111111111"); err != nil {
		t.Fatalf("update: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE auth.credentials SET password_hash = $2, updated_at = NOW() WHERE id = $1`)).
		WithArgs(int64(6), "new-hash").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.UpdatePasswordHash(context.Background(), 6, "new-hash"); err != nil {
		t.Fatalf("update password hash: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM auth.credentials WHERE id = $1`)).
		WithArgs(int64(6)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Delete(context.Background(), 6); err != nil {
		t.Fatalf("delete: %v", err)
	}

	if nullableText("") != nil || nullableText("x") != "x" {
		t.Fatal("unexpected nullableText behavior")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
