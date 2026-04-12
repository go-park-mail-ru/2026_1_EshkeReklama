package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"eshkere/internal/models"

	"golang.org/x/crypto/bcrypt"
)

const bcryptSaltMarker = "bcrypt"

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailTaken           = errors.New("email already registered")
	ErrPhoneTaken           = errors.New("phone already registered")
	ErrInvalidAdvertiserArg = errors.New("invalid argument")
)

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

func (s *Service) RegisterAdvertiser(ctx context.Context, name, email, phone, password string) (*models.Advertiser, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	phone = strings.TrimSpace(phone)
	password = strings.TrimSpace(password)

	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: invalid email", ErrInvalidAdvertiserArg)
	}
	nphone, err := normalizeAdvertiserPhone(phone)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAdvertiserArg, err)
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("%w: password too short", ErrInvalidAdvertiserArg)
	}

	if name == "" {
		name = displayNameFromEmail(email)
	}

	_, errEmail := s.advertiserRepo.GetByEmail(ctx, email)
	if errEmail == nil {
		return nil, ErrEmailTaken
	}
	if !errors.Is(errEmail, sql.ErrNoRows) {
		return nil, errEmail
	}

	_, errPhone := s.advertiserRepo.GetByPhone(ctx, nphone)
	if errPhone == nil {
		return nil, ErrPhoneTaken
	}
	if !errors.Is(errPhone, sql.ErrNoRows) {
		return nil, errPhone
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	a := &models.Advertiser{
		Name:         name,
		Email:        email,
		Phone:        nphone,
		PasswordHash: string(hash),
		PasswordSalt: bcryptSaltMarker,
		Balance:      0,
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
		return nil, ErrInvalidCredentials
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
			return nil, ErrInvalidCredentials
		}
		adv, err = s.advertiserRepo.GetByPhone(ctx, phone)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if adv.PasswordSalt == bcryptSaltMarker || strings.HasPrefix(adv.PasswordHash, "$2") {
		if err := bcrypt.CompareHashAndPassword([]byte(adv.PasswordHash), []byte(password)); err != nil {
			return nil, ErrInvalidCredentials
		}
		return adv, nil
	}

	return nil, ErrInvalidCredentials
}

func (s *Service) GetAdvertiserByID(ctx context.Context, id int) (*models.Advertiser, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid id", ErrInvalidAdvertiserArg)
	}
	return s.advertiserRepo.GetByID(ctx, id)
}
