package service

import (
	"context"
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"fmt"
	"strings"
)

func (s *Service) CreateAppeal(ctx context.Context, in *serviceinput.CreateAppeal) (*models.Appeal, error) {
	if in == nil {
		return nil, errors.New("create appeal input is nil")
	}

	appeal := &models.Appeal{
		AdvertiserID: models.NullInt64FromPtr(in.AdvertiserID),
		Status:       models.AppealStatusOpen,
		Category:     in.Category,
		Title:        in.Title,
		Description:  in.Description,
		Name:         in.Name,
		Email:        in.Email,
	}

	if err := s.appealRepo.Create(ctx, appeal); err != nil {
		return nil, err
	}

	if len(in.Image) > 0 {
		if s.appealStorage == nil {
			return nil, fmt.Errorf("%w: appeal storage is not configured", errs.InternalServiceError)
		}

		imageKey, err := s.appealStorage.UploadAppealImage(ctx, appeal.ID, in.Image, in.ImageExt, in.ImageType)
		if err != nil {
			return nil, err
		}

		if err = s.appealRepo.UpdateImage(ctx, appeal.ID, imageKey); err != nil {
			_ = s.appealStorage.DeleteAppealImage(ctx, appeal.ID, imageKey)
			return nil, err
		}

		appeal.ImageURL = imageKey
	}

	s.decorateAppealImageURL(appeal)

	return appeal, nil
}

func (s *Service) ListAppeals(ctx context.Context, advertiserID int) ([]*models.Appeal, error) {
	appeals, err := s.appealRepo.ListByAdvertiserID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	for _, appeal := range appeals {
		s.decorateAppealImageURL(appeal)
	}

	return appeals, nil
}

func (s *Service) GetAppealByID(ctx context.Context, appealID int) (*models.Appeal, error) {
	appeal, err := s.appealRepo.GetByID(ctx, appealID)
	if err != nil {
		return nil, err
	}

	s.decorateAppealImageURL(appeal)

	return appeal, nil
}

func (s *Service) GetAppealMessages(ctx context.Context, advertiserID, appealID int) ([]*models.AppealMessage, error) {
	if _, err := s.getOwnedChatAppeal(ctx, advertiserID, appealID); err != nil {
		return nil, err
	}

	return s.appealRepo.ListMessages(ctx, appealID)
}

func (s *Service) PostAppealMessage(ctx context.Context, in *serviceinput.PostAppealMessage) (*models.AppealMessage, error) {
	if in == nil {
		return nil, errors.New("post appeal message input is nil")
	}

	if _, err := s.getOwnedChatAppeal(ctx, in.AdvertiserID, in.AppealID); err != nil {
		return nil, err
	}

	text := strings.TrimSpace(in.Text)
	if text == "" {
		return nil, errs.BadRequestError
	}

	msg := &models.AppealMessage{
		AppealID: in.AppealID,
		Author:   models.AppealMessageAuthorUser,
		Text:     text,
	}
	if err := s.appealRepo.AddMessage(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) decorateAppealImageURL(appeal *models.Appeal) {
	if s == nil || s.appealStorage == nil || appeal == nil {
		return
	}

	imageURL := s.appealStorage.GetAppealImageURL(appeal.ImageURL)
	if imageURL == "" {
		return
	}

	appeal.ImageURL = imageURL
}

func (s *Service) getOwnedChatAppeal(ctx context.Context, advertiserID, appealID int) (*models.Appeal, error) {
	appeal, err := s.appealRepo.GetByID(ctx, appealID)
	if err != nil {
		return nil, err
	}

	if !appeal.AdvertiserID.Valid || appeal.AdvertiserID.Int64 != int64(advertiserID) {
		return nil, errs.NotFoundError
	}

	return appeal, nil
}
