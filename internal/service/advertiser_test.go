package service

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"
)

func TestCreateAdvertiserProfile_DefaultName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().
		CreateProfile(gomock.Any(), int64(77), "a").
		Return(nil)

	if err := svc.CreateAdvertiserProfile(context.Background(), 77, "", "a@a.test"); err != nil {
		t.Fatalf("CreateAdvertiserProfile: %v", err)
	}
}

func TestCreateAdvertiserProfile_UsesProvidedName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, err := NewService(&Config{AdvertiserRepo: advRepo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	advRepo.EXPECT().
		CreateProfile(gomock.Any(), int64(88), "brand").
		DoAndReturn(func(_ context.Context, id int64, name string) error {
			if id != 88 || name != "brand" {
				t.Fatalf("unexpected profile args: id=%d name=%s", id, name)
			}
			return nil
		})

	if err := svc.CreateAdvertiserProfile(context.Background(), 88, "brand", "a@a.test"); err != nil {
		t.Fatalf("CreateAdvertiserProfile: %v", err)
	}
}
