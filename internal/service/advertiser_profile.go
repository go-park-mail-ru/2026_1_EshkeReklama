package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"eshkere/internal/models"
)

func (s *Service) UpdateAdvertiserProfile(ctx context.Context, advertiserID int, name, email, phone string, avatar []byte, avatarFilename, avatarContentType string) (*models.Advertiser, error) {
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

	if len(avatar) > 0 {
		if s.avatarStorage == nil {
			return nil, fmt.Errorf("avatar storage is not configured")
		}

		avatarURL, uploadErr := s.avatarStorage.UploadAvatar(ctx, advertiserID, avatar, avatarFilename, avatarContentType)
		if uploadErr != nil {
			return nil, uploadErr
		}
		adv.AvatarURL = sql.NullString{String: avatarURL, Valid: true}
	}

	if err = s.advertiserRepo.Update(ctx, adv); err != nil {
		return nil, err
	}

	return adv, nil
}
