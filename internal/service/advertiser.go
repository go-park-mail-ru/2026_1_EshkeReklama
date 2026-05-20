package service

import (
	"context"
	"fmt"
	"strings"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
)

func displayNameFromEmail(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}

	return "Advertiser"
}

func (s *Service) CreateAdvertiserProfile(ctx context.Context, id int64, name, email string) error {
	if name == "" {
		name = displayNameFromEmail(email)
	}

	return s.advertiserRepo.CreateProfile(ctx, id, name)
}

func (s *Service) GetAdvertiserByID(ctx context.Context, id int) (*models.Advertiser, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", errs.ErrInvalidAdvertiserArg)
	}

	adv, err := s.advertiserRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.decorateAdvertiserAvatarURL(adv)

	return adv, nil
}
