package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	errs "eshkere/internal/errors"
	"fmt"
	"math/big"

	"eshkere/internal/models"
)

const BaseFeedURL = "https://eshkereklama.ru/feed/"

func (s *Service) GenerateFeedLink(ctx context.Context, campaignID int) (string, error) {
	campaign, err := s.adCampaignRepo.GetByID(ctx, campaignID)
	if err != nil || campaign == nil {
		return "", fmt.Errorf("%w: invalid campaign id", errs.BadRequestError)
	}

	token, err := generateToken(16)
	if err != nil {
		return "", fmt.Errorf("generate feed token: %w", err)
	}

	if err = s.feedLinkRepo.Create(ctx, campaignID, token); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", BaseFeedURL, token), nil
}

func (s *Service) GetAdByFeedToken(ctx context.Context, token string) (*models.Ad, error) {
	if token == "" {
		return nil, fmt.Errorf("%w: empty token", errs.ErrInvalidAdvertiserArg)
	}

	campaignID, err := s.feedLinkRepo.GetCampaignIDByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: feed with this token not found", errs.NotFoundError)
		}
		return nil, err
	}

	ads, err := s.adRepo.ListByAdCampaignID(ctx, campaignID)
	if err != nil {
		return &models.Ad{}, nil
	}

	if ads == nil {
		return &models.Ad{}, nil
	}

	randAdInd, err := rand.Int(rand.Reader, big.NewInt(int64(len(ads))))
	if err != nil {
		randAdInd = big.NewInt(0)
	}

	return ads[randAdInd.Int64()], nil
}

func (s *Service) GetAdsByFeedToken(ctx context.Context, token string) ([]*models.Ad, error) {
	if token == "" {
		return nil, fmt.Errorf("%w: empty token", errs.ErrInvalidAdvertiserArg)
	}

	campaignID, err := s.feedLinkRepo.GetCampaignIDByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: feed with this token not found", errs.NotFoundError)
		}
		return nil, err
	}

	ads, err := s.adRepo.ListByAdCampaignID(ctx, campaignID)
	if err != nil {
		return []*models.Ad{}, nil
	}
	if ads == nil {
		return []*models.Ad{}, nil
	}
	return ads, nil
}

func generateToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
