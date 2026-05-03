package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"fmt"
)

type AdRequestResult struct {
	RequestID string
	Ad        *models.Ad
}

func (s *Service) RequestAd(ctx context.Context, embedToken string) (*AdRequestResult, error) {
	if embedToken == "" {
		return nil, fmt.Errorf("%w: empty embed token", errs.BadRequestError)
	}

	if _, err := s.partnerBlockRepo.GetByEmbedToken(ctx, embedToken); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: partner block not found", errs.NotFoundError)
		}
		return nil, err
	}

	ad, err := s.adRepo.GetRandomWorking(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: working ad not found", errs.NotFoundError)
		}
		return nil, err
	}

	requestID, err := newRequestID()
	if err != nil {
		return nil, fmt.Errorf("generate request id: %w", err)
	}

	return &AdRequestResult{
		RequestID: requestID,
		Ad:        ad,
	}, nil
}

func newRequestID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
