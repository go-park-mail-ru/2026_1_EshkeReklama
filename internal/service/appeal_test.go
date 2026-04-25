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
	listMessagesFn     func(ctx context.Context, appealID int) ([]*models.AppealMessage, error)
	addMessageFn       func(ctx context.Context, msg *models.AppealMessage) error
	updateImageFn      func(ctx context.Context, appealID int, imageKey string) error

	adminListFn         func(ctx context.Context, filter *serviceinput.AdminListAppealsFilter) ([]*models.Appeal, error)
	adminListMessagesFn func(ctx context.Context, appealID int) ([]*models.AppealMessage, error)
	adminListHistoryFn  func(ctx context.Context, appealID int) ([]*models.AppealStatusHistory, error)
	adminAddMessageFn   func(ctx context.Context, msg *models.AppealMessage) error
	adminUpdateStatusFn func(ctx context.Context, appealID int, status models.AppealStatus) error
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

func (r *testAppealRepo) ListMessages(ctx context.Context, appealID int) ([]*models.AppealMessage, error) {
	if r.listMessagesFn == nil {
		return nil, nil
	}
	return r.listMessagesFn(ctx, appealID)
}

func (r *testAppealRepo) AddMessage(ctx context.Context, msg *models.AppealMessage) error {
	if r.addMessageFn == nil {
		return nil
	}
	return r.addMessageFn(ctx, msg)
}

func (r *testAppealRepo) AdminList(ctx context.Context, filter *serviceinput.AdminListAppealsFilter) ([]*models.Appeal, error) {
	if r.adminListFn == nil {
		return nil, nil
	}
	return r.adminListFn(ctx, filter)
}

func (r *testAppealRepo) AdminListStatusHistory(ctx context.Context, appealID int) ([]*models.AppealStatusHistory, error) {
	if r.adminListHistoryFn == nil {
		return nil, nil
	}
	return r.adminListHistoryFn(ctx, appealID)
}

func (r *testAppealRepo) AdminUpdateStatus(ctx context.Context, appealID int, status models.AppealStatus) error {
	if r.adminUpdateStatusFn == nil {
		return nil
	}
	return r.adminUpdateStatusFn(ctx, appealID, status)
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

func TestPostAppealMessage_OwnedAppeal(t *testing.T) {
	repo := &testAppealRepo{
		getByIDFn: func(_ context.Context, appealID int) (*models.Appeal, error) {
			return &models.Appeal{ID: appealID, AdvertiserID: models.NullInt64FromPtr(ptrInt(7))}, nil
		},
		addMessageFn: func(_ context.Context, msg *models.AppealMessage) error {
			if msg.AppealID != 11 || msg.Author != models.AppealMessageAuthorUser || msg.Text != "hello" {
				t.Fatalf("unexpected message: %+v", msg)
			}
			msg.ID = 3
			return nil
		},
	}

	svc, err := NewService(&Config{AppealRepo: repo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	msg, err := svc.PostAppealMessage(context.Background(), &serviceinput.PostAppealMessage{
		AppealID:     11,
		AdvertiserID: 7,
		Text:         " hello ",
	})
	if err != nil {
		t.Fatalf("PostAppealMessage: %v", err)
	}
	if msg.ID != 3 {
		t.Fatalf("expected message id 3, got %d", msg.ID)
	}
}

func TestAdminPostAppealMessage_GuestAppealRejected(t *testing.T) {
	repo := &testAppealRepo{
		getByIDFn: func(_ context.Context, appealID int) (*models.Appeal, error) {
			return &models.Appeal{ID: appealID}, nil
		},
	}

	svc, err := NewService(&Config{AppealRepo: repo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	_, err = svc.AdminPostAppealMessage(context.Background(), &serviceinput.AdminPostAppealMessage{
		AppealID: 5,
		Text:     "reply",
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func ptrInt(v int) *int {
	return &v
}
