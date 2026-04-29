package service

import (
	"context"
	"database/sql"
	"errors"
	errs "eshkere/internal/errors"
	serviceinput "eshkere/internal/service/input"
	"fmt"
	"strings"
	"time"

	"eshkere/internal/models"

	"golang.org/x/crypto/bcrypt"
)

const bcryptSaltMarker = "bcrypt"

func normalizeAdvertiserPhone(raw string) (string, error) {
	var digits strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	d := digits.String()
	switch len(d) {
	case 10:
		return d, nil
	case 11:
		if d[0] == '7' || d[0] == '8' {
			return d[1:], nil
		}
	}
	return "", fmt.Errorf("phone must be 10 digits (e.g. 9001234567 or +7 900 123-45-67)")
}

func displayNameFromEmail(email string) string {
	if i := strings.IndexByte(email, '@'); i > 0 {
		return email[:i]
	}
	return "Advertiser"
}

func (s *Service) RegisterAdvertiser(ctx context.Context, in *serviceinput.RegisterAdvertiser) (*models.Advertiser, error) {
	if in.Email == "" || !strings.Contains(in.Email, "@") {
		return nil, fmt.Errorf("%w: invalid email", errs.ErrInvalidAdvertiserArg)
	}
	nphone, err := normalizeAdvertiserPhone(in.Phone)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrInvalidAdvertiserArg, err)
	}
	if len(in.Password) < 6 {
		return nil, fmt.Errorf("%w: password too short", errs.ErrInvalidAdvertiserArg)
	}

	if in.Name == "" {
		in.Name = displayNameFromEmail(in.Email)
	}

	_, errEmail := s.advertiserRepo.GetByEmail(ctx, in.Email)
	if errEmail == nil {
		return nil, errs.ErrEmailTaken
	}
	if !errors.Is(errEmail, sql.ErrNoRows) {
		return nil, errEmail
	}

	_, errPhone := s.advertiserRepo.GetByPhone(ctx, nphone)
	if errPhone == nil {
		return nil, errs.ErrPhoneTaken
	}
	if !errors.Is(errPhone, sql.ErrNoRows) {
		return nil, errPhone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	a := &models.Advertiser{
		Name:         in.Name,
		Email:        in.Email,
		Phone:        nphone,
		PasswordHash: string(hash),
		PasswordSalt: bcryptSaltMarker,
		Balance:      0,
		Tariff:       models.TariffTypeNoob,
		CreatedAt:    time.Now(),
	}

	id, err := s.advertiserRepo.Create(ctx, a)
	if err != nil {
		return nil, err
	}
	a.ID = id
	return a, nil
}

func (s *Service) AuthenticateAdvertiser(ctx context.Context, identifier, password string) (*models.Advertiser, error) {
	identifier = strings.TrimSpace(identifier)
	password = strings.TrimSpace(password)
	if identifier == "" || password == "" {
		return nil, errs.ErrInvalidCredentials
	}

	var (
		adv *models.Advertiser
		err error
	)

	if strings.Contains(identifier, "@") {
		email := strings.ToLower(identifier)
		adv, err = s.advertiserRepo.GetByEmail(ctx, email)
	} else {
		phone, nerr := normalizeAdvertiserPhone(identifier)
		if nerr != nil {
			return nil, errs.ErrInvalidCredentials
		}
		adv, err = s.advertiserRepo.GetByPhone(ctx, phone)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrInvalidCredentials
		}
		return nil, err
	}

	if adv.PasswordSalt == bcryptSaltMarker || strings.HasPrefix(adv.PasswordHash, "$2") {
		if err := bcrypt.CompareHashAndPassword([]byte(adv.PasswordHash), []byte(password)); err != nil {
			return nil, errs.ErrInvalidCredentials
		}
		return adv, nil
	}

	return nil, errs.ErrInvalidCredentials
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
