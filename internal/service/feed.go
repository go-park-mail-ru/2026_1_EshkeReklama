package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"

	"eshkere/internal/models"
)

func (s *Service) GenerateFeedLink(ctx context.Context, advertiserID int) (string, error) {
	if advertiserID <= 0 {
		return "", fmt.Errorf("%w: invalid advertiser id", ErrInvalidAdvertiserArg)
	}
	if s.feedLinkRepo == nil {
		return "", fmt.Errorf("feed link repository is not configured")
	}

	token, err := generateToken(16)
	if err != nil {
		return "", fmt.Errorf("generate feed token: %w", err)
	}

	if err = s.feedLinkRepo.UpsertByAdvertiserID(ctx, advertiserID, token); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) GetAdsByFeedToken(ctx context.Context, token string) ([]*models.Ad, error) {
	if token == "" {
		return nil, fmt.Errorf("%w: empty token", ErrInvalidAdvertiserArg)
	}
	if s.feedLinkRepo == nil {
		return nil, fmt.Errorf("feed link repository is not configured")
	}

	advertiserID, err := s.feedLinkRepo.GetAdvertiserIDByToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}

	ads, err := s.adRepo.ListByAdvertiserID(ctx, advertiserID)
	if err != nil {
		return nil, err
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
