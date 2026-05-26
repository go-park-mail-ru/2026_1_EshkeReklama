package service

import (
	"context"
	"testing"

	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
)

type testAppealRepo struct {
	createFn           func(ctx context.Context, appeal *models.Appeal) error
	getByIDFn          func(ctx context.Context, appealID int) (*models.Appeal, error)
	listByAdvertiserFn func(ctx context.Context, advertiserID int) ([]*models.Appeal, error)
	updateImageFn      func(ctx context.Context, appealID int, imageKey string) error
}

func (r *testAppealRepo) Create(ctx context.Context, appeal *models.Appeal) error {
	return r.createFn(ctx, appeal)
}

func (r *testAppealRepo) GetByID(ctx context.Context, appealID int) (*models.Appeal, error) {
	return r.getByIDFn(ctx, appealID)
}

func (r *testAppealRepo) ListByAdvertiserID(ctx context.Context, advertiserID int) ([]*models.Appeal, error) {
	return r.listByAdvertiserFn(ctx, advertiserID)
}

func (r *testAppealRepo) UpdateImage(ctx context.Context, appealID int, imageKey string) error {
	return r.updateImageFn(ctx, appealID, imageKey)
}

type testAppealStorage struct {
	uploadFn func(ctx context.Context, appealID int, data []byte, ext string, contentType string) (string, error)
	deleteFn func(ctx context.Context, appealID int, imageKey string) error
	urlFn    func(imageKey string) string
}

func (s *testAppealStorage) UploadAppealImage(ctx context.Context, appealID int, data []byte, ext string, contentType string) (string, error) {
	return s.uploadFn(ctx, appealID, data, ext, contentType)
}

func (s *testAppealStorage) DeleteAppealImage(ctx context.Context, appealID int, imageKey string) error {
	if s.deleteFn == nil {
		return nil
	}
	return s.deleteFn(ctx, appealID, imageKey)
}

func (s *testAppealStorage) GetAppealImageURL(imageKey string) string {
	if s.urlFn == nil {
		return imageKey
	}
	return s.urlFn(imageKey)
}

func TestCreateAppeal_UploadsImageAndDecoratesURL(t *testing.T) {
	repo := &testAppealRepo{
		createFn: func(_ context.Context, appeal *models.Appeal) error {
			appeal.ID = 17
			return nil
		},
		updateImageFn: func(_ context.Context, appealID int, imageKey string) error {
			if appealID != 17 || imageKey != "appeals/17/attachments/file.png" {
				t.Fatalf("unexpected update image args: id=%d key=%s", appealID, imageKey)
			}
			return nil
		},
	}
	storage := &testAppealStorage{
		uploadFn: func(_ context.Context, appealID int, data []byte, ext string, contentType string) (string, error) {
			if appealID != 17 {
				t.Fatalf("unexpected appeal id: %d", appealID)
			}
			if string(data) != "png-bytes" || ext != ".png" || contentType != "image/png" {
				t.Fatalf("unexpected image payload: %q %s %s", string(data), ext, contentType)
			}
			return "appeals/17/attachments/file.png", nil
		},
		urlFn: func(imageKey string) string {
			return "https://cdn.example.com/" + imageKey
		},
	}

	svc, err := NewService(&Config{
		AppealRepo:    repo,
		AppealStorage: storage,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	appeal, err := svc.CreateAppeal(context.Background(), &serviceinput.CreateAppeal{
		Category:    models.AppealCategoryBug,
		Title:       "Crash",
		Description: "Steps to reproduce",
		Name:        "Ivan",
		Email:       "ivan@example.com",
		Image:       []byte("png-bytes"),
		ImageExt:    ".png",
		ImageType:   "image/png",
	})
	if err != nil {
		t.Fatalf("CreateAppeal: %v", err)
	}
	if appeal.ImageURL != "https://cdn.example.com/appeals/17/attachments/file.png" {
		t.Fatalf("unexpected image url: %s", appeal.ImageURL)
	}
}

func TestListAppeals_DecoratesImageURLs(t *testing.T) {
	repo := &testAppealRepo{
		listByAdvertiserFn: func(_ context.Context, advertiserID int) ([]*models.Appeal, error) {
			if advertiserID != 5 {
				t.Fatalf("unexpected advertiser id: %d", advertiserID)
			}
			return []*models.Appeal{{ID: 1, ImageURL: "appeals/1/attachments/file.png"}}, nil
		},
	}
	storage := &testAppealStorage{
		urlFn: func(imageKey string) string {
			return "https://cdn.example.com/" + imageKey
		},
	}

	svc, err := NewService(&Config{
		AppealRepo:    repo,
		AppealStorage: storage,
	})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	appeals, err := svc.ListAppeals(context.Background(), 5)
	if err != nil {
		t.Fatalf("ListAppeals: %v", err)
	}
	if len(appeals) != 1 || appeals[0].ImageURL != "https://cdn.example.com/appeals/1/attachments/file.png" {
		t.Fatalf("unexpected appeals: %+v", appeals)
	}
}
