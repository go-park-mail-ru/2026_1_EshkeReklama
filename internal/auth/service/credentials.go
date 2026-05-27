package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	authrepo "eshkere/internal/auth/repository/postgres"
	"eshkere/internal/auth/vkid"

	"golang.org/x/crypto/bcrypt"
)

type CredentialsRepo interface {
	Create(ctx context.Context, email, phone, hash string) (int64, error)
	CreateVK(ctx context.Context, email, phone string, vkUserID int64) (int64, error)
	GetByEmail(ctx context.Context, email string) (*authrepo.Credential, error)
	GetByPhone(ctx context.Context, phone string) (*authrepo.Credential, error)
	GetByVKUserID(ctx context.Context, vkUserID int64) (*authrepo.Credential, error)
	GetByID(ctx context.Context, id int64) (*authrepo.Credential, error)
	LinkVKUserID(ctx context.Context, id, vkUserID int64) error
	Update(ctx context.Context, id int64, email, phone string) error
	UpdatePasswordHash(ctx context.Context, id int64, passwordHash string) error
	Delete(ctx context.Context, id int64) error
}

type VKIDAuthenticator interface {
	ResolveUser(ctx context.Context, accessToken string) (*vkid.Identity, error)
}

type CredentialsService struct {
	repo     CredentialsRepo
	vkidAuth VKIDAuthenticator
}

func NewCredentialsService(repo CredentialsRepo, vkidAuth VKIDAuthenticator) *CredentialsService {
	return &CredentialsService{repo: repo, vkidAuth: vkidAuth}
}

var (
	ErrEmailTaken          = errors.New("email already registered")
	ErrPhoneTaken          = errors.New("phone already registered")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidArg          = errors.New("invalid argument")
	ErrVKIDConflict        = errors.New("vk credentials conflict")
	ErrVKIDUnavailable     = errors.New("vk id auth is not configured")
	ErrPasswordUnavailable = errors.New("password change is unavailable for vk id accounts")
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

func (s *CredentialsService) FindByIdentifier(ctx context.Context, identifier string) (*authrepo.Credential, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return nil, fmt.Errorf("%w: identifier is required", ErrInvalidArg)
	}

	if strings.Contains(identifier, "@") {
		return s.repo.GetByEmail(ctx, strings.ToLower(identifier))
	}

	phone, err := normalizePhone(identifier)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidArg, err)
	}
	return s.repo.GetByPhone(ctx, phone)
}

func (s *CredentialsService) AuthenticateVKID(ctx context.Context, accessToken string, expectedUserID int64) (int64, *vkid.Identity, error) {
	if strings.TrimSpace(accessToken) == "" || expectedUserID <= 0 {
		return 0, nil, fmt.Errorf("%w: access_token and user_id are required", ErrInvalidArg)
	}
	if s.vkidAuth == nil {
		return 0, nil, ErrVKIDUnavailable
	}

	identity, err := s.vkidAuth.ResolveUser(ctx, strings.TrimSpace(accessToken))
	if err != nil {
		if errors.Is(err, vkid.ErrUnauthorized) {
			return 0, nil, ErrInvalidCredentials
		}
		return 0, nil, err
	}
	if identity.UserID != expectedUserID {
		return 0, nil, ErrInvalidCredentials
	}

	email := strings.ToLower(strings.TrimSpace(identity.Email))
	phone := ""
	if strings.TrimSpace(identity.Phone) != "" {
		phone, err = normalizePhone(identity.Phone)
		if err != nil {
			return 0, nil, fmt.Errorf("%w: %v", ErrInvalidArg, err)
		}
	}

	if cred, getErr := s.repo.GetByVKUserID(ctx, identity.UserID); getErr == nil {
		if err := s.syncVKIdentityContacts(ctx, cred, email, phone); err != nil {
			return 0, nil, err
		}
		return cred.ID, identity, nil
	} else if !errors.Is(getErr, sql.ErrNoRows) {
		return 0, nil, getErr
	}

	emailCred, err := s.lookupByEmail(ctx, email)
	if err != nil {
		return 0, nil, err
	}
	phoneCred, err := s.lookupByPhone(ctx, phone)
	if err != nil {
		return 0, nil, err
	}

	targetCred, err := resolveVKIDTarget(identity.UserID, emailCred, phoneCred)
	if err != nil {
		return 0, nil, err
	}

	if targetCred != nil {
		if !targetCred.VKUserID.Valid {
			if err := s.repo.LinkVKUserID(ctx, targetCred.ID, identity.UserID); err != nil {
				return 0, nil, err
			}
		}
		if err := s.syncVKIdentityContacts(ctx, targetCred, email, phone); err != nil {
			return 0, nil, err
		}
		return targetCred.ID, identity, nil
	}

	id, err := s.repo.CreateVK(ctx, email, phone, identity.UserID)
	if err != nil {
		return 0, nil, err
	}
	return id, identity, nil
}

func (s *CredentialsService) GetByID(ctx context.Context, id int64) (*authrepo.Credential, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", ErrInvalidArg)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *CredentialsService) ChangePassword(ctx context.Context, id int64, currentPassword, newPassword string) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid advertiser id", ErrInvalidArg)
	}

	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return fmt.Errorf("%w: current and new password are required", ErrInvalidArg)
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("%w: password too short", ErrInvalidArg)
	}

	cred, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if cred.PasswordHash == "" {
		return ErrPasswordUnavailable
	}
	if err := bcrypt.CompareHashAndPassword([]byte(cred.PasswordHash), []byte(currentPassword)); err != nil {
		return ErrInvalidCredentials
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.repo.UpdatePasswordHash(ctx, id, string(hash))
}

func (s *CredentialsService) SetPassword(ctx context.Context, id int64, newPassword string) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid advertiser id", ErrInvalidArg)
	}

	newPassword = strings.TrimSpace(newPassword)
	if newPassword == "" {
		return fmt.Errorf("%w: new password is required", ErrInvalidArg)
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("%w: password too short", ErrInvalidArg)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	return s.repo.UpdatePasswordHash(ctx, id, string(hash))
}

func (s *CredentialsService) Update(ctx context.Context, id int64, email, phone string) (*authrepo.Credential, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", ErrInvalidArg)
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email != "" && !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: invalid email", ErrInvalidArg)
	}

	var err error
	if strings.TrimSpace(phone) != "" {
		phone, err = normalizePhone(phone)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidArg, err)
		}
	} else {
		phone = ""
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if current.Email != email && email != "" {
		if existing, getErr := s.repo.GetByEmail(ctx, email); getErr == nil && existing.ID != id {
			return nil, ErrEmailTaken
		} else if getErr != nil && !errors.Is(getErr, sql.ErrNoRows) {
			return nil, getErr
		}
	}

	if current.Phone != phone && phone != "" {
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

func (s *CredentialsService) syncVKIdentityContacts(ctx context.Context, cred *authrepo.Credential, email, phone string) error {
	if cred == nil {
		return nil
	}

	nextEmail := cred.Email
	if nextEmail == "" && email != "" {
		nextEmail = email
	}

	nextPhone := cred.Phone
	if nextPhone == "" && phone != "" {
		nextPhone = phone
	}

	if nextEmail == cred.Email && nextPhone == cred.Phone {
		return nil
	}

	if err := s.repo.Update(ctx, cred.ID, nextEmail, nextPhone); err != nil {
		return err
	}

	cred.Email = nextEmail
	cred.Phone = nextPhone
	return nil
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

func (s *CredentialsService) lookupByEmail(ctx context.Context, email string) (*authrepo.Credential, error) {
	if email == "" {
		return nil, nil
	}

	cred, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return cred, nil
}

func (s *CredentialsService) lookupByPhone(ctx context.Context, phone string) (*authrepo.Credential, error) {
	if phone == "" {
		return nil, nil
	}

	cred, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return cred, nil
}

func resolveVKIDTarget(vkUserID int64, emailCred, phoneCred *authrepo.Credential) (*authrepo.Credential, error) {
	if emailCred != nil && phoneCred != nil && emailCred.ID != phoneCred.ID {
		return nil, ErrVKIDConflict
	}

	target := emailCred
	if target == nil {
		target = phoneCred
	}
	if target == nil {
		return nil, nil
	}

	if target.VKUserID.Valid && target.VKUserID.Int64 != vkUserID {
		return nil, ErrVKIDConflict
	}

	return target, nil
}
