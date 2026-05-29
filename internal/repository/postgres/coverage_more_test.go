package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
)

func newSQLMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

func TestRegionAndTopicRepository(t *testing.T) {
	db, mock := newSQLMockDB(t)

	regionRepo := NewRegionRepository(db)
	topicRepo := NewTopicRepository(db)

	mock.ExpectQuery("SELECT id FROM eshkere.region WHERE name = \\$1").
		WithArgs("Moscow").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
	id, err := regionRepo.GetIDByName(context.Background(), "Moscow")
	if err != nil || id != 7 {
		t.Fatalf("expected region id 7, got %d err=%v", id, err)
	}

	mock.ExpectQuery("SELECT id FROM eshkere.region WHERE name = \\$1").
		WithArgs("Unknown").
		WillReturnError(sql.ErrNoRows)
	if _, err := regionRepo.GetIDByName(context.Background(), "Unknown"); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found error, got %v", err)
	}

	mock.ExpectQuery("SELECT id FROM eshkere.topic WHERE name = \\$1").
		WithArgs("Finance").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
	id, err = topicRepo.GetIDByName(context.Background(), "Finance")
	if err != nil || id != 3 {
		t.Fatalf("expected topic id 3, got %d err=%v", id, err)
	}

	mock.ExpectQuery("SELECT id FROM eshkere.topic WHERE name = \\$1").
		WithArgs("Broken").
		WillReturnError(errors.New("db down"))
	if _, err := topicRepo.GetIDByName(context.Background(), "Broken"); err == nil {
		t.Fatal("expected db error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdvertiserAutopaySettingsRepository(t *testing.T) {
	db, mock := newSQLMockDB(t)
	repo := NewAdvertiserAutopaySettingsRepository(db)
	now := time.Now()

	mock.ExpectQuery("SELECT\\s+advertiser_id, enabled, threshold_amount, top_up_amount, updated_at").
		WithArgs(11).
		WillReturnRows(sqlmock.NewRows([]string{"advertiser_id", "enabled", "threshold_amount", "top_up_amount", "updated_at"}).
			AddRow(11, true, int64(1000), int64(7000), now))
	settings, err := repo.GetByAdvertiserID(context.Background(), 11)
	if err != nil || settings.AdvertiserID != 11 || !settings.Enabled {
		t.Fatalf("unexpected autopay settings: %+v err=%v", settings, err)
	}

	mock.ExpectQuery("SELECT\\s+advertiser_id, enabled, threshold_amount, top_up_amount, updated_at").
		WithArgs(12).
		WillReturnError(sql.ErrNoRows)
	settings, err = repo.GetByAdvertiserID(context.Background(), 12)
	if err != nil || settings.AdvertiserID != 12 || settings.Enabled || settings.ThresholdAmount != 5000 || settings.TopUpAmount != 30000 {
		t.Fatalf("unexpected default autopay settings: %+v err=%v", settings, err)
	}

	mock.ExpectExec("INSERT INTO eshkere.advertiser_autopay_settings").
		WithArgs(11, true, int64(1000), int64(7000)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Upsert(context.Background(), &models.AdvertiserAutopaySettings{
		AdvertiserID:    11,
		Enabled:         true,
		ThresholdAmount: 1000,
		TopUpAmount:     7000,
	}); err != nil {
		t.Fatalf("upsert autopay: %v", err)
	}

	if err := repo.Upsert(context.Background(), nil); err == nil {
		t.Fatal("expected nil settings error")
	}

	mock.ExpectQuery("SELECT\\s+aas.advertiser_id,").
		WillReturnRows(sqlmock.NewRows([]string{"advertiser_id", "balance", "saved_payment_method_id", "threshold_amount", "top_up_amount"}).
			AddRow(11, int64(200), "pm_1", int64(1000), int64(7000)))
	candidates, err := repo.ListEligible(context.Background())
	if err != nil || len(candidates) != 1 || candidates[0].AdvertiserID != 11 {
		t.Fatalf("unexpected autopay candidates: %+v err=%v", candidates, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestAdvertiserNotificationSettingsRepository(t *testing.T) {
	db, mock := newSQLMockDB(t)
	repo := NewAdvertiserNotificationSettingsRepository(db)
	now := time.Now()

	mock.ExpectQuery("SELECT\\s+advertiser_id, email_enabled, telegram_enabled, telegram_chat_id, warning_threshold, critical_threshold, updated_at").
		WithArgs(21).
		WillReturnRows(sqlmock.NewRows([]string{"advertiser_id", "email_enabled", "telegram_enabled", "telegram_chat_id", "warning_threshold", "critical_threshold", "updated_at"}).
			AddRow(21, true, false, nil, int64(900), int64(300), now))
	settings, err := repo.GetByAdvertiserID(context.Background(), 21)
	if err != nil || settings.AdvertiserID != 21 || !settings.EmailEnabled {
		t.Fatalf("unexpected notification settings: %+v err=%v", settings, err)
	}

	mock.ExpectQuery("SELECT\\s+advertiser_id, email_enabled, telegram_enabled, telegram_chat_id, warning_threshold, critical_threshold, updated_at").
		WithArgs(22).
		WillReturnError(sql.ErrNoRows)
	settings, err = repo.GetByAdvertiserID(context.Background(), 22)
	if err != nil || settings.AdvertiserID != 22 || settings.WarningThreshold != 500 || settings.CriticalThreshold != 100 {
		t.Fatalf("unexpected default notification settings: %+v err=%v", settings, err)
	}

	if err := repo.Upsert(context.Background(), nil); err == nil {
		t.Fatal("expected nil settings error")
	}

	err = repo.Upsert(context.Background(), &models.AdvertiserNotificationSettings{
		AdvertiserID:      21,
		EmailEnabled:      true,
		WarningThreshold:  100,
		CriticalThreshold: 100,
	})
	if !errors.Is(err, errs.ErrInvalidAdvertiserArg) {
		t.Fatalf("expected invalid advertiser arg, got %v", err)
	}

	mock.ExpectExec("INSERT INTO eshkere.advertiser_notification_settings").
		WithArgs(21, true, false, nil, int64(900), int64(300)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Upsert(context.Background(), &models.AdvertiserNotificationSettings{
		AdvertiserID:      21,
		EmailEnabled:      true,
		WarningThreshold:  900,
		CriticalThreshold: 300,
	}); err != nil {
		t.Fatalf("upsert notification settings: %v", err)
	}

	mock.ExpectQuery("SELECT\\s+ans.advertiser_id,").
		WillReturnRows(sqlmock.NewRows([]string{"advertiser_id", "email_enabled", "telegram_enabled", "telegram_chat_id", "warning_threshold", "critical_threshold", "updated_at"}).
			AddRow(21, true, false, nil, int64(900), int64(300), now))
	enabled, err := repo.ListEnabled(context.Background())
	if err != nil || len(enabled) != 1 || enabled[0].AdvertiserID != 21 {
		t.Fatalf("unexpected enabled notification settings: %+v err=%v", enabled, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPaymentTransactionRepository_CreateGetByIDAndNullableString(t *testing.T) {
	db, mock := newSQLMockDB(t)
	repo := NewPaymentTransactionRepository(db)
	now := time.Now()
	ctx := logger.CtxWithLogger(context.Background(), zap.NewNop().Sugar())

	if got := nullableString(""); got != nil {
		t.Fatalf("expected nil for empty string, got %#v", got)
	}
	if got := nullableString("card"); got != "card" {
		t.Fatalf("expected raw string, got %#v", got)
	}

	if err := repo.Create(context.Background(), nil); err == nil {
		t.Fatal("expected nil tx error")
	}

	mock.ExpectExec("INSERT INTO eshkere.payment_transaction").
		WithArgs("pay_1", 15, int64(1200), models.PaymentTransactionStatusPending, models.PaymentTypeBalance).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Create(context.Background(), &models.PaymentTransaction{
		ID:           "pay_1",
		AdvertiserID: 15,
		Amount:       1200,
		Status:       models.PaymentTransactionStatusPending,
		PaymentType:  models.PaymentTypeBalance,
	}); err != nil {
		t.Fatalf("create payment tx: %v", err)
	}

	mock.ExpectQuery("SELECT\\s+id, advertiser_id, amount, status, payment_type, created_at, updated_at").
		WithArgs("pay_1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "advertiser_id", "amount", "status", "payment_type", "created_at", "updated_at"}).
			AddRow("pay_1", 15, int64(1200), models.PaymentTransactionStatusPending, models.PaymentTypeBalance, now, now))
	tx, err := repo.GetByID(ctx, "pay_1")
	if err != nil || tx.ID != "pay_1" || tx.AdvertiserID != 15 {
		t.Fatalf("unexpected payment tx: %+v err=%v", tx, err)
	}

	mock.ExpectQuery("SELECT\\s+id, advertiser_id, amount, status, payment_type, created_at, updated_at").
		WithArgs("missing").
		WillReturnError(sql.ErrNoRows)
	if _, err := repo.GetByID(ctx, "missing"); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPaymentTransactionRepository_CompleteSucceeded(t *testing.T) {
	db, mock := newSQLMockDB(t)
	repo := NewPaymentTransactionRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT advertiser_id, amount, status, payment_type\\s+FROM eshkere.payment_transaction").
		WithArgs("pay_2").
		WillReturnRows(sqlmock.NewRows([]string{"advertiser_id", "amount", "status", "payment_type"}).
			AddRow(19, int64(2500), models.PaymentTransactionStatusPending, models.PaymentTypeBalance))
	mock.ExpectExec("UPDATE eshkere.payment_transaction").
		WithArgs("pay_2", models.PaymentTransactionStatusSucceeded, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE eshkere.advertiser").
		WithArgs("pm_22", "Visa", sqlmock.AnyArg(), 19).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("UPDATE eshkere.advertiser\\s+SET balance = balance \\+ \\$1, updated_at = \\$3\\s+WHERE id = \\$2\\s+RETURNING balance").
		WithArgs(int64(2500), 19, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(int64(9100)))
	mock.ExpectCommit()

	result, err := repo.Complete(context.Background(), "pay_2", models.PaymentTransactionStatusSucceeded, "pm_22", "Visa")
	if err != nil {
		t.Fatalf("complete succeeded payment: %v", err)
	}
	if result.AdvertiserID != 19 || result.Balance != 9100 || result.AlreadyFinal {
		t.Fatalf("unexpected payment completion result: %+v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestPaymentTransactionRepository_CompleteAlreadyFinal(t *testing.T) {
	db, mock := newSQLMockDB(t)
	repo := NewPaymentTransactionRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT advertiser_id, amount, status, payment_type\\s+FROM eshkere.payment_transaction").
		WithArgs("pay_3").
		WillReturnRows(sqlmock.NewRows([]string{"advertiser_id", "amount", "status", "payment_type"}).
			AddRow(23, int64(1000), models.PaymentTransactionStatusSucceeded, models.PaymentTypeBalance))
	mock.ExpectQuery("SELECT balance FROM eshkere.advertiser WHERE id = \\$1").
		WithArgs(23).
		WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(int64(5000)))
	mock.ExpectCommit()

	result, err := repo.Complete(context.Background(), "pay_3", models.PaymentTransactionStatusSucceeded, "", "")
	if err != nil {
		t.Fatalf("complete already-final payment: %v", err)
	}
	if !result.AlreadyFinal || result.AdvertiserID != 23 || result.Balance != 5000 {
		t.Fatalf("unexpected already-final result: %+v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
