package service

import (
	"context"
	"errors"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
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

	return appeal, nil
}
func (s *Service) ListAppeals(ctx context.Context, advertiserID int) ([]*models.Appeal, error) {
	return s.appealRepo.ListByAdvertiserID(ctx, advertiserID)
}

func (s *Service) GetAppealByID(ctx context.Context, appealID int) (*models.Appeal, error) {
	return s.appealRepo.GetByID(ctx, appealID)
}
