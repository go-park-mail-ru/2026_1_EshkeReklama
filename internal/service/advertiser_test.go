package service

import (
	"context"
	"database/sql"
	"errors"
	errs "eshkere/internal/errors"
	"testing"

	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

func TestNormalizeAdvertiserPhone(t *testing.T) {
	got, err := normalizeAdvertiserPhone("+7 (900) 123-45-67")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "9001234567" {
		t.Fatalf("expected 9001234567 got %q", got)
	}
}

func TestRegisterAdvertiser_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().GetByEmail(gomock.Any(), "a@a.test").Return(nil, sql.ErrNoRows)
	advRepo.EXPECT().GetByPhone(gomock.Any(), "9001234567").Return(nil, sql.ErrNoRows)
	advRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, a *models.Advertiser) (int, error) {
			if a.Email != "a@a.test" || a.Phone != "9001234567" || a.PasswordHash == "" {
				t.Fatalf("unexpected advertiser: %+v", a)
			}
			return 77, nil
		})

	adv, err := svc.RegisterAdvertiser(context.Background(), "", "A@A.TEST", "+7 900 123-45-67", "secret1")
	if err != nil {
		t.Fatalf("RegisterAdvertiser: %v", err)
	}
	if adv.ID != 77 {
		t.Fatalf("expected id 77 got %d", adv.ID)
	}
	if adv.Name == "" {
		t.Fatalf("expected generated name")
	}
}

func TestRegisterAdvertiser_EmailTaken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{AdvertiserRepo: advRepo})

	advRepo.EXPECT().GetByEmail(gomock.Any(), "a@a.test").Return(&models.Advertiser{ID: 1}, nil)

	_, err := svc.RegisterAdvertiser(context.Background(), "n", "a@a.test", "+7 900 123-45-67", "secret1")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, errs.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken got %v", err)
	}
}

func TestAuthenticateAdvertiser_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{AdvertiserRepo: advRepo})

	advRepo.EXPECT().GetByEmail(gomock.Any(), "a@a.test").Return(&models.Advertiser{
		ID:           1,
		Email:        "a@a.test",
		PasswordSalt: bcryptSaltMarker,
		PasswordHash: "$2a$10$5qPj/8rGPt8xUj5mI9wKQeB3o0jvCz9bBDmLr3nNzFQH0UoqTj/0y", // not matching "secret"
	}, nil)

	_, err := svc.AuthenticateAdvertiser(context.Background(), "a@a.test", "secret")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, errs.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials got %v", err)
	}
}
