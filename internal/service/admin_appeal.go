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

func (s *Service) AdminListAppeals(ctx context.Context, filter *serviceinput.AdminListAppealsFilter) ([]*models.Appeal, error) {
	if filter == nil {
		return nil, errors.New("filter is nil")
	}
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	appeals, err := s.appealRepo.AdminList(ctx, filter)
	if err != nil {
		return nil, err
	}

	for _, appeal := range appeals {
		s.decorateAppealImageURL(appeal)
	}

	return appeals, nil
}

func (s *Service) AdminGetAppealWithHistory(ctx context.Context, appealID int) (*models.Appeal, []*models.AppealMessage, []*models.AppealStatusHistory, error) {
	appeal, err := s.appealRepo.GetByID(ctx, appealID)
	if err != nil {
		return nil, nil, nil, err
	}
	s.decorateAppealImageURL(appeal)

	msgs, err := s.appealRepo.ListMessages(ctx, appealID)
	if err != nil {
		return nil, nil, nil, err
	}
	history, err := s.appealRepo.AdminListStatusHistory(ctx, appealID)
	if err != nil {
		return nil, nil, nil, err
	}

	return appeal, msgs, history, nil
}

func (s *Service) AdminPatchAppealStatus(ctx context.Context, in *serviceinput.AdminPatchAppealStatus) error {
	if in == nil {
		return errors.New("input is nil")
	}
	switch in.Status {
	case models.AppealStatusOpen, models.AppealStatusInProgress, models.AppealStatusClosed:
	default:
		return errs.BadRequestError
	}

	if err := s.appealRepo.AdminUpdateStatus(ctx, in.AppealID, in.Status); err != nil {
		return err
	}
	return nil
}

func (s *Service) AdminPostAppealMessage(ctx context.Context, in *serviceinput.AdminPostAppealMessage) (*models.AppealMessage, error) {
	if in == nil {
		return nil, errors.New("input is nil")
	}
	appeal, err := s.appealRepo.GetByID(ctx, in.AppealID)
	if err != nil {
		return nil, err
	}
	if !appeal.AdvertiserID.Valid {
		return nil, fmt.Errorf("%w: guest appeals do not support chat", errs.BusinessLogicError)
	}

	text := strings.TrimSpace(in.Text)
	if text == "" {
		return nil, errs.BadRequestError
	}

	msg := &models.AppealMessage{
		AppealID: in.AppealID,
		Author:   models.AppealMessageAuthorAdmin,
		Text:     text,
	}
	if err := s.appealRepo.AddMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
