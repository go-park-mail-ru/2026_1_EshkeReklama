package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	authrepo "eshkere/internal/auth/repository/postgres"

	"golang.org/x/crypto/bcrypt"
)

type CredentialsRepo interface {
	Create(ctx context.Context, email, phone, hash string) (int64, error)
	GetByEmail(ctx context.Context, email string) (*authrepo.Credential, error)
	GetByPhone(ctx context.Context, phone string) (*authrepo.Credential, error)
	GetByID(ctx context.Context, id int64) (*authrepo.Credential, error)
	Update(ctx context.Context, id int64, email, phone string) error
	Delete(ctx context.Context, id int64) error
}

type CredentialsService struct {
	repo CredentialsRepo
}

func NewCredentialsService(repo CredentialsRepo) *CredentialsService {
	return &CredentialsService{repo: repo}
}

var (
	ErrEmailTaken         = errors.New("email already registered")
	ErrPhoneTaken         = errors.New("phone already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidArg         = errors.New("invalid argument")
)

func (s *CredentialsService) Register(ctx context.Context, email, phone, password string) (id int64, err error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return 0, fmt.Errorf("%w: invalid email", ErrInvalidArg)
	}

	phone, err = normalizePhone(phone)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidArg, err)
	}

	if len(password) < 6 {
		return 0, fmt.Errorf("%w: password too short", ErrInvalidArg)
	}

	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return 0, ErrEmailTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if _, err := s.repo.GetByPhone(ctx, phone); err == nil {
		return 0, ErrPhoneTaken
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	return s.repo.Create(ctx, email, phone, string(hash))
}

func (s *CredentialsService) Authenticate(ctx context.Context, identifier, password string) (int64, error) {
	identifier = strings.TrimSpace(identifier)
	password = strings.TrimSpace(password)
	if identifier == "" || password == "" {
		return 0, ErrInvalidCredentials
	}

	var (
		cred *authrepo.Credential
		err  error
	)

	if strings.Contains(identifier, "@") {
		cred, err = s.repo.GetByEmail(ctx, strings.ToLower(identifier))
	} else {
		phone, nerr := normalizePhone(identifier)
		if nerr != nil {
			return 0, ErrInvalidCredentials
		}
		cred, err = s.repo.GetByPhone(ctx, phone)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(password)); err != nil {
		return 0, ErrInvalidCredentials
	}

	return cred.ID, nil
}

func (s *CredentialsService) GetByID(ctx context.Context, id int64) (*authrepo.Credential, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", ErrInvalidArg)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *CredentialsService) Update(ctx context.Context, id int64, email, phone string) (*authrepo.Credential, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", ErrInvalidArg)
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: invalid email", ErrInvalidArg)
	}

	var err error
	phone, err = normalizePhone(phone)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidArg, err)
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if current.Email != email {
		if existing, getErr := s.repo.GetByEmail(ctx, email); getErr == nil && existing.ID != id {
			return nil, ErrEmailTaken
		} else if getErr != nil && !errors.Is(getErr, sql.ErrNoRows) {
			return nil, getErr
		}
	}

	if current.Phone != phone {
		if existing, getErr := s.repo.GetByPhone(ctx, phone); getErr == nil && existing.ID != id {
			return nil, ErrPhoneTaken
		} else if getErr != nil && !errors.Is(getErr, sql.ErrNoRows) {
			return nil, getErr
		}
	}

	if err := s.repo.Update(ctx, id, email, phone); err != nil {
		return nil, err
	}

	current.Email = email
	current.Phone = phone

	return current, nil
}

func normalizePhone(raw string) (string, error) {
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
