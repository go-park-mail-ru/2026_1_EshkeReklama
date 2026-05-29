package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
)

const defaultSupportMessagesLimit = 50

func (s *Service) GetSupportThreadByCampaign(ctx context.Context, advertiserID, campaignID int) (*models.SupportThread, error) {
	if campaignID <= 0 {
		return nil, fmt.Errorf("%w: invalid campaign id", errs.BadRequestError)
	}

	campaign, err := s.ownedCampaign(ctx, advertiserID, campaignID)
	if err != nil {
		return nil, err
	}

	return s.getOrCreateSupportThread(ctx, campaign.ID, campaign.AdvertiserID)
}

func (s *Service) GetSupportThread(ctx context.Context, threadID int) (*models.SupportThread, error) {
	if threadID <= 0 {
		return nil, fmt.Errorf("%w: invalid thread id", errs.BadRequestError)
	}
	if s.supportRepo == nil {
		return nil, fmt.Errorf("%w: support repository is not configured", errs.InternalServiceError)
	}

	return s.supportRepo.GetThreadByID(ctx, threadID)
}

func (s *Service) ListSupportMessages(ctx context.Context, advertiserID, threadID, limit int, beforeID *int) ([]*models.SupportMessage, error) {
	thread, err := s.GetSupportThread(ctx, threadID)
	if err != nil {
		return nil, err
	}
	if thread.AdvertiserID != advertiserID {
		return nil, ownershipNotFound("support thread")
	}

	return s.supportRepo.ListMessages(ctx, threadID, normalizeSupportLimit(limit), beforeID)
}

func (s *Service) CreateSupportMessage(ctx context.Context, advertiserID, threadID int, text string) (*models.SupportMessage, error) {
	thread, err := s.GetSupportThread(ctx, threadID)
	if err != nil {
		return nil, err
	}
	if thread.AdvertiserID != advertiserID {
		return nil, ownershipNotFound("support thread")
	}

	return s.createSupportMessage(ctx, threadID, models.SupportMessageAuthorAdvertiser, sql.NullInt64{Int64: int64(advertiserID), Valid: true}, text)
}

func (s *Service) ListAdminSupportThreads(ctx context.Context) ([]*models.SupportThreadSummary, error) {
	if s.supportRepo == nil {
		return nil, fmt.Errorf("%w: support repository is not configured", errs.InternalServiceError)
	}

	return s.supportRepo.ListThreads(ctx)
}

func (s *Service) ListAdminSupportMessages(ctx context.Context, threadID, limit int, beforeID *int) ([]*models.SupportMessage, error) {
	if _, err := s.GetSupportThread(ctx, threadID); err != nil {
		return nil, err
	}

	return s.supportRepo.ListMessages(ctx, threadID, normalizeSupportLimit(limit), beforeID)
}

func (s *Service) CreateAdminSupportMessage(ctx context.Context, adminID, threadID int, text string) (*models.SupportMessage, error) {
	if _, err := s.GetSupportThread(ctx, threadID); err != nil {
		return nil, err
	}

	return s.createSupportMessage(ctx, threadID, models.SupportMessageAuthorModerator, sql.NullInt64{}, text)
}

func (s *Service) createModeratorMessageForCampaign(ctx context.Context, campaignID, advertiserID int, text string) error {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	thread, err := s.getOrCreateSupportThread(ctx, campaignID, advertiserID)
	if err != nil {
		return err
	}

	_, err = s.createSupportMessage(ctx, thread.ID, models.SupportMessageAuthorModerator, sql.NullInt64{}, text)
	return err
}

func (s *Service) getOrCreateSupportThread(ctx context.Context, campaignID, advertiserID int) (*models.SupportThread, error) {
	if s.supportRepo == nil {
		return nil, fmt.Errorf("%w: support repository is not configured", errs.InternalServiceError)
	}

	thread, err := s.supportRepo.GetThreadByCampaignID(ctx, campaignID)
	if err == nil {
		return thread, nil
	}

	thread = &models.SupportThread{
		CampaignID:   campaignID,
		AdvertiserID: advertiserID,
	}
	if createErr := s.supportRepo.CreateThread(ctx, thread); createErr == nil {
		return thread, nil
	}

	// If another request created the thread first, a re-read gives us the canonical row.
	thread, getErr := s.supportRepo.GetThreadByCampaignID(ctx, campaignID)
	if getErr == nil {
		return thread, nil
	}

	return nil, err
}

func (s *Service) createSupportMessage(ctx context.Context, threadID int, authorType models.SupportMessageAuthor, authorID sql.NullInt64, text string) (*models.SupportMessage, error) {
	if s.supportRepo == nil {
		return nil, fmt.Errorf("%w: support repository is not configured", errs.InternalServiceError)
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("%w: text is required", errs.BadRequestError)
	}

	msg := &models.SupportMessage{
		ThreadID:   threadID,
		AuthorType: authorType,
		AuthorID:   authorID,
		Text:       text,
	}

	if err := s.supportRepo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}
	if s.supportBroadcaster != nil {
		if err := s.supportBroadcaster.Broadcast(ctx, threadID, msg); err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func normalizeSupportLimit(limit int) int {
	if limit <= 0 {
		return defaultSupportMessagesLimit
	}
	if limit > 100 {
		return 100
	}
	return limit
}
