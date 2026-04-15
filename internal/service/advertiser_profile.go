package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"eshkere/internal/models"
)

func (s *Service) UpdateAdvertiserProfile(ctx context.Context, advertiserID int, name, email, phone string) (*models.Advertiser, error) {
	if advertiserID <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", ErrInvalidAdvertiserArg)
	}

	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name != "" {
		adv.Name = name
	}

	email = strings.TrimSpace(strings.ToLower(email))
	if email != "" {
		adv.Email = email
	}

	phone = strings.TrimSpace(phone)
	if phone != "" {
		normalizedPhone, normalizeErr := normalizeAdvertiserPhone(phone)
		if normalizeErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidAdvertiserArg, normalizeErr)
		}
		adv.Phone = normalizedPhone
	}

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
		return nil, fmt.Errorf("%w: invalid advertiser id", ErrInvalidAdvertiserArg)
	}

	if len(avatar) == 0 {
		return nil, fmt.Errorf("%w: empty avatar", ErrInvalidAdvertiserArg)
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
