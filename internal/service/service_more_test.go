package service

import (
	"context"
	"database/sql"
	"testing"

	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticateAdvertiser_OK_WithEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{AdvertiserRepo: advRepo})

	hash, err := bcrypt.GenerateFromPassword([]byte("secret1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}

	advRepo.EXPECT().GetByEmail(gomock.Any(), "a@a.test").Return(&models.Advertiser{
		ID:           10,
		Email:        "a@a.test",
		PasswordSalt: bcryptSaltMarker,
		PasswordHash: string(hash),
	}, nil)

	adv, err := svc.AuthenticateAdvertiser(context.Background(), "A@A.TEST", "secret1")
	if err != nil {
		t.Fatalf("AuthenticateAdvertiser: %v", err)
	}
	if adv.ID != 10 {
		t.Fatalf("expected id 10 got %d", adv.ID)
	}
}

func TestTopUpAdvertiserBalance_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	svc, _ := NewService(&Config{AdvertiserRepo: advRepo})

	adv := &models.Advertiser{ID: 1, Balance: 100}
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(adv, nil)
	advRepo.EXPECT().Update(gomock.Any(), adv).Return(nil)

	got, err := svc.TopUpAdvertiserBalance(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("TopUpAdvertiserBalance: %v", err)
	}
	if got != 150 {
		t.Fatalf("expected 150 got %d", got)
	}
}

func TestUpdateAdvertiserAvatar_UploadsAvatar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	st := NewMockAvatarStorage(ctrl)

	svc, _ := NewService(&Config{AdvertiserRepo: advRepo, AvatarStorage: st})

	adv := &models.Advertiser{ID: 1}
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(adv, nil)
	st.EXPECT().
		UploadAvatar(gomock.Any(), 1, []byte("img"), "a.png", "image/png").
		Return("https://cdn/avatar.png", nil)
	advRepo.EXPECT().Update(gomock.Any(), adv).Return(nil)
	st.EXPECT().
		GetAvatarURL("https://cdn/avatar.png").
		Return("https://cdn/avatar.png")

	updated, err := svc.UpdateAdvertiserAvatar(context.Background(), 1, []byte("img"), "a.png", "image/png")
	if err != nil {
		t.Fatalf("UpdateAdvertiserAvatar: %v", err)
	}
	if updated.AvatarURL.Valid != true {
		t.Fatalf("expected avatar url to be set")
	}
}

func TestUpdateAdvertiserAvatar_DeletesPrevious(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	advRepo := NewMockAdvertiserRepository(ctrl)
	st := NewMockAvatarStorage(ctrl)

	svc, _ := NewService(&Config{AdvertiserRepo: advRepo, AvatarStorage: st})

	adv := &models.Advertiser{
		ID:        1,
		AvatarURL: sql.NullString{String: "https://cdn/avatars/1/old.png", Valid: true},
	}
	advRepo.EXPECT().GetByID(gomock.Any(), 1).Return(adv, nil)
	st.EXPECT().
		UploadAvatar(gomock.Any(), 1, []byte("img"), "a.png", "image/png").
		Return("https://cdn/avatars/1/new.png", nil)
	advRepo.EXPECT().Update(gomock.Any(), adv).Return(nil)
	st.EXPECT().
		DeleteAvatar(gomock.Any(), 1, "https://cdn/avatars/1/old.png").
		Return(nil)
	st.EXPECT().
		GetAvatarURL("https://cdn/avatars/1/new.png").
		Return("https://cdn/avatars/1/new.png")

	_, err := svc.UpdateAdvertiserAvatar(context.Background(), 1, []byte("img"), "a.png", "image/png")
	if err != nil {
		t.Fatalf("UpdateAdvertiserAvatar: %v", err)
	}
}

func TestGenerateFeedLink_And_GetAdsByFeedToken_OK(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedRepo := NewMockFeedLinkRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)

	svc, _ := NewService(&Config{FeedLinkRepo: feedRepo, AdRepo: adRepo})

	feedRepo.EXPECT().
		UpsertByAdvertiserID(gomock.Any(), 1, gomock.Any()).
		Return(nil)

	token, err := svc.GenerateFeedLink(context.Background(), 1)
	if err != nil {
		t.Fatalf("GenerateFeedLink: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token")
	}

	feedRepo.EXPECT().GetAdvertiserIDByToken(gomock.Any(), "t").Return(1, nil)
	adRepo.EXPECT().ListByAdvertiserID(gomock.Any(), 1).Return([]*models.Ad{}, nil)

	ads, err := svc.GetAdsByFeedToken(context.Background(), "t")
	if err != nil {
		t.Fatalf("GetAdsByFeedToken: %v", err)
	}
	if len(ads) != 0 {
		t.Fatalf("expected empty ads")
	}
}

func TestGetAdsByFeedToken_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	feedRepo := NewMockFeedLinkRepository(ctrl)
	svc, _ := NewService(&Config{FeedLinkRepo: feedRepo, AdRepo: NewMockAdRepository(ctrl)})

	feedRepo.EXPECT().GetAdvertiserIDByToken(gomock.Any(), "t").Return(0, sql.ErrNoRows)

	_, err := svc.GetAdsByFeedToken(context.Background(), "t")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows got %v", err)
	}
}
