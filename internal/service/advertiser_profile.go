package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	errs "eshkere/internal/errors"
	serviceinput "eshkere/internal/service/input"

	"eshkere/internal/models"
)

func (s *Service) UpdateAdvertiserProfile(ctx context.Context, in *serviceinput.UpdateAdvertiserProfile) (*models.Advertiser, error) {
	if in == nil {
		return nil, fmt.Errorf("%w: nil update profile input", errs.ErrInvalidAdvertiserArg)
	}

	if in.AdvertiserID <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}

	adv, err := s.advertiserRepo.GetByID(ctx, in.AdvertiserID)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		adv.Name = *in.Name
	}
	if in.Surname != nil {
		adv.Surname = sql.NullString{String: *in.Surname, Valid: true}
	}
	if in.Company != nil {
		adv.Company = sql.NullString{String: *in.Company, Valid: true}
	}
	if in.City != nil {
		adv.City = sql.NullString{String: *in.City, Valid: true}
	}
	// Тариф меняется только через ActivatePro / DeactivatePro (после оплаты)

	if err = s.advertiserRepo.Update(ctx, adv); err != nil {
		return nil, err
	}

	s.decorateAdvertiserAvatarURL(adv)

	return adv, nil
}

func (s *Service) UpdateAdvertiserAvatar(
	ctx context.Context,
	advertiserID int,
	avatar []byte,
	avatarExt string,
	avatarContentType string,
) (*models.Advertiser, error) {
	if advertiserID <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}

	if len(avatar) == 0 {
		return nil, fmt.Errorf("%w: empty avatar", errs.ErrInvalidAdvertiserArg)
	}

	if s.avatarStorage == nil {
		return nil, fmt.Errorf("avatar storage is not configured")
	}

	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	prevAvatarKey := ""
	if adv.AvatarURL.Valid {
		prevAvatarKey = adv.AvatarURL.String
	}

	avatarKey, err := s.avatarStorage.UploadAvatar(ctx, advertiserID, avatar, avatarExt, avatarContentType)
	if err != nil {
		return nil, err
	}

	adv.AvatarURL = sql.NullString{String: avatarKey, Valid: true}

	if err = s.advertiserRepo.Update(ctx, adv); err != nil {
		_ = s.avatarStorage.DeleteAvatar(ctx, advertiserID, avatarKey)
		return nil, err
	}

	if prevAvatarKey != "" && prevAvatarKey != avatarKey {
		_ = s.avatarStorage.DeleteAvatar(ctx, advertiserID, prevAvatarKey)
	}

	s.decorateAdvertiserAvatarURL(adv)

	return adv, nil
}

func (s *Service) decorateAdvertiserAvatarURL(adv *models.Advertiser) {
	if s == nil || s.avatarStorage == nil || adv == nil {
		return
	}

	avatarKey := ""
	if adv.AvatarURL.Valid {
		avatarKey = adv.AvatarURL.String
	}

	avatarURL := strings.TrimSpace(s.avatarStorage.GetAvatarURL(avatarKey))
	if avatarURL == "" {
		return
	}

	adv.AvatarURL = sql.NullString{String: avatarURL, Valid: true}
}
