package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	authrepo "eshkere/internal/auth/repository/postgres"
	"eshkere/internal/auth/vkid"

	"golang.org/x/crypto/bcrypt"
)

type stubCredentialsRepo struct {
	createFunc     func(context.Context, string, string, string) (int64, error)
	createVKFunc   func(context.Context, string, string, int64) (int64, error)
	getByEmailFunc func(context.Context, string) (*authrepo.Credential, error)
	getByPhoneFunc func(context.Context, string) (*authrepo.Credential, error)
	getByVKFunc    func(context.Context, int64) (*authrepo.Credential, error)
	getByIDFunc    func(context.Context, int64) (*authrepo.Credential, error)
	linkVKFunc     func(context.Context, int64, int64) error
	updateFunc     func(context.Context, int64, string, string) error
	deleteFunc     func(context.Context, int64) error
}

func (s *stubCredentialsRepo) Create(ctx context.Context, email, phone, hash string) (int64, error) {
	return s.createFunc(ctx, email, phone, hash)
}
func (s *stubCredentialsRepo) CreateVK(ctx context.Context, email, phone string, vkUserID int64) (int64, error) {
	return s.createVKFunc(ctx, email, phone, vkUserID)
}
func (s *stubCredentialsRepo) GetByEmail(ctx context.Context, email string) (*authrepo.Credential, error) {
	return s.getByEmailFunc(ctx, email)
}
func (s *stubCredentialsRepo) GetByPhone(ctx context.Context, phone string) (*authrepo.Credential, error) {
	return s.getByPhoneFunc(ctx, phone)
}
func (s *stubCredentialsRepo) GetByVKUserID(ctx context.Context, vkUserID int64) (*authrepo.Credential, error) {
	return s.getByVKFunc(ctx, vkUserID)
}
func (s *stubCredentialsRepo) GetByID(ctx context.Context, id int64) (*authrepo.Credential, error) {
	return s.getByIDFunc(ctx, id)
}
func (s *stubCredentialsRepo) LinkVKUserID(ctx context.Context, id, vkUserID int64) error {
	return s.linkVKFunc(ctx, id, vkUserID)
}
func (s *stubCredentialsRepo) Update(ctx context.Context, id int64, email, phone string) error {
	return s.updateFunc(ctx, id, email, phone)
}
func (s *stubCredentialsRepo) Delete(ctx context.Context, id int64) error {
	return s.deleteFunc(ctx, id)
}

type stubVKIDAuth struct {
	resolveFunc func(context.Context, string) (*vkid.Identity, error)
}

func (s *stubVKIDAuth) ResolveUser(ctx context.Context, accessToken string) (*vkid.Identity, error) {
	return s.resolveFunc(ctx, accessToken)
}

func TestRegisterAndAuthenticate(t *testing.T) {
	var storedHash string
	repo := &stubCredentialsRepo{
		createFunc: func(_ context.Context, email, phone, hash string) (int64, error) {
			if email != "user@example.com" || phone != "9001234567" {
				t.Fatalf("unexpected create args: %q %q", email, phone)
			}
			storedHash = hash
			return 10, nil
		},
		createVKFunc: func(context.Context, string, string, int64) (int64, error) { return 0, nil },
		getByEmailFunc: func(_ context.Context, email string) (*authrepo.Credential, error) {
			if email == "user@example.com" && storedHash == "" {
				return nil, sql.ErrNoRows
			}
			if email == "user@example.com" && storedHash != "" {
				return &authrepo.Credential{ID: 10, Email: email, Phone: "9001234567", PasswordHash: storedHash}, nil
			}
			return nil, sql.ErrNoRows
		},
		getByPhoneFunc: func(_ context.Context, phone string) (*authrepo.Credential, error) {
			if phone == "9001234567" && storedHash == "" {
				return nil, sql.ErrNoRows
			}
			return &authrepo.Credential{ID: 10, Email: "user@example.com", Phone: phone, PasswordHash: storedHash}, nil
		},
		getByVKFunc: func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByIDFunc: func(context.Context, int64) (*authrepo.Credential, error) { return nil, nil },
		linkVKFunc:  func(context.Context, int64, int64) error { return nil },
		updateFunc:  func(context.Context, int64, string, string) error { return nil },
		deleteFunc:  func(context.Context, int64) error { return nil },
	}

	svc := NewCredentialsService(repo, nil)
	id, err := svc.Register(context.Background(), " User@Example.com ", "+7 (900) 123-45-67", "secret123")
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if id != 10 || storedHash == "" {
		t.Fatalf("unexpected register result: id=%d hash=%q", id, storedHash)
	}
	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("secret123")) != nil {
		t.Fatal("expected password to be hashed")
	}

	authID, err := svc.Authenticate(context.Background(), " user@example.com ", "secret123")
	if err != nil {
		t.Fatalf("authenticate by email: %v", err)
	}
	if authID != 10 {
		t.Fatalf("unexpected auth id: %d", authID)
	}

	authID, err = svc.Authenticate(context.Background(), "8 (900) 123-45-67", "secret123")
	if err != nil {
		t.Fatalf("authenticate by phone: %v", err)
	}
	if authID != 10 {
		t.Fatalf("unexpected auth id: %d", authID)
	}
}

func TestRegisterValidationAndConflictErrors(t *testing.T) {
	repo := &stubCredentialsRepo{
		createFunc:     func(context.Context, string, string, string) (int64, error) { return 0, nil },
		createVKFunc:   func(context.Context, string, string, int64) (int64, error) { return 0, nil },
		getByEmailFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByPhoneFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByVKFunc:    func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByIDFunc:    func(context.Context, int64) (*authrepo.Credential, error) { return nil, nil },
		linkVKFunc:     func(context.Context, int64, int64) error { return nil },
		updateFunc:     func(context.Context, int64, string, string) error { return nil },
		deleteFunc:     func(context.Context, int64) error { return nil },
	}
	svc := NewCredentialsService(repo, nil)

	cases := []struct {
		name string
		fn   func() error
		want error
	}{
		{"bad email", func() error {
			_, err := svc.Register(context.Background(), "bad", "9001234567", "secret123")
			return err
		}, ErrInvalidArg},
		{"bad phone", func() error {
			_, err := svc.Register(context.Background(), "user@example.com", "123", "secret123")
			return err
		}, ErrInvalidArg},
		{"short password", func() error {
			_, err := svc.Register(context.Background(), "user@example.com", "9001234567", "123")
			return err
		}, ErrInvalidArg},
		{"empty auth", func() error { _, err := svc.Authenticate(context.Background(), "", ""); return err }, ErrInvalidCredentials},
	}
	for _, tc := range cases {
		if err := tc.fn(); !errors.Is(err, tc.want) {
			t.Fatalf("%s: expected %v, got %v", tc.name, tc.want, err)
		}
	}

	repo.getByEmailFunc = func(context.Context, string) (*authrepo.Credential, error) {
		return &authrepo.Credential{ID: 1}, nil
	}
	if _, err := svc.Register(context.Background(), "user@example.com", "9001234567", "secret123"); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}

	repo.getByEmailFunc = func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows }
	repo.getByPhoneFunc = func(context.Context, string) (*authrepo.Credential, error) {
		return &authrepo.Credential{ID: 2}, nil
	}
	if _, err := svc.Register(context.Background(), "user@example.com", "9001234567", "secret123"); !errors.Is(err, ErrPhoneTaken) {
		t.Fatalf("expected ErrPhoneTaken, got %v", err)
	}
}

func TestAuthenticateVKIDAndUpdateAndHelpers(t *testing.T) {
	var linkedID, linkedVKID int64
	var updatedEmail, updatedPhone string
	repo := &stubCredentialsRepo{
		createFunc:   func(context.Context, string, string, string) (int64, error) { return 0, nil },
		createVKFunc: func(_ context.Context, email, phone string, vkUserID int64) (int64, error) { return 99, nil },
		getByEmailFunc: func(_ context.Context, email string) (*authrepo.Credential, error) {
			switch email {
			case "vk@example.com":
				return &authrepo.Credential{ID: 5}, nil
			case "new@example.com":
				return nil, sql.ErrNoRows
			case "taken@example.com":
				return &authrepo.Credential{ID: 77}, nil
			default:
				return nil, sql.ErrNoRows
			}
		},
		getByPhoneFunc: func(_ context.Context, phone string) (*authrepo.Credential, error) {
			switch phone {
			case "9000000000":
				return &authrepo.Credential{ID: 5}, nil
			case "9111111111":
				return nil, sql.ErrNoRows
			case "9222222222":
				return &authrepo.Credential{ID: 88}, nil
			default:
				return nil, sql.ErrNoRows
			}
		},
		getByVKFunc: func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByIDFunc: func(_ context.Context, id int64) (*authrepo.Credential, error) {
			return &authrepo.Credential{ID: id, Email: "old@example.com", Phone: "9001234567"}, nil
		},
		linkVKFunc: func(_ context.Context, id, vkUserID int64) error {
			linkedID, linkedVKID = id, vkUserID
			return nil
		},
		updateFunc: func(_ context.Context, id int64, email, phone string) error {
			updatedEmail, updatedPhone = email, phone
			return nil
		},
		deleteFunc: func(context.Context, int64) error { return nil },
	}
	vk := &stubVKIDAuth{
		resolveFunc: func(context.Context, string) (*vkid.Identity, error) {
			return &vkid.Identity{UserID: 123, Email: "vk@example.com", Phone: "+7 900 000 00 00"}, nil
		},
	}
	svc := NewCredentialsService(repo, vk)

	id, identity, err := svc.AuthenticateVKID(context.Background(), "vk-token", 123)
	if err != nil {
		t.Fatalf("authenticate vkid: %v", err)
	}
	if id != 5 || linkedID != 5 || linkedVKID != 123 {
		t.Fatalf("unexpected vk link result: id=%d linked=%d/%d", id, linkedID, linkedVKID)
	}
	if identity == nil || identity.Email != "vk@example.com" {
		t.Fatalf("unexpected vk identity: %#v", identity)
	}

	updated, err := svc.Update(context.Background(), 5, "new@example.com", "8 911 111 11 11")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updatedEmail != "new@example.com" || updatedPhone != "9111111111" {
		t.Fatalf("unexpected update args: %q %q", updatedEmail, updatedPhone)
	}
	if updated.Email != "new@example.com" || updated.Phone != "9111111111" {
		t.Fatalf("unexpected updated credential: %#v", updated)
	}

	got, err := svc.GetByID(context.Background(), 5)
	if err != nil || got.ID != 5 {
		t.Fatalf("get by id: %#v %v", got, err)
	}

	if _, err := svc.GetByID(context.Background(), 0); !errors.Is(err, ErrInvalidArg) {
		t.Fatalf("expected invalid arg, got %v", err)
	}
	if _, err := svc.Update(context.Background(), 0, "a@b.c", "9001234567"); !errors.Is(err, ErrInvalidArg) {
		t.Fatalf("expected invalid arg, got %v", err)
	}

	if _, err := svc.Update(context.Background(), 5, "taken@example.com", "9001234567"); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
	repo.getByEmailFunc = func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows }
	if _, err := svc.Update(context.Background(), 5, "new@example.com", "9222222222"); !errors.Is(err, ErrPhoneTaken) {
		t.Fatalf("expected ErrPhoneTaken, got %v", err)
	}

	if _, _, err := svc.AuthenticateVKID(context.Background(), "", 123); !errors.Is(err, ErrInvalidArg) {
		t.Fatalf("expected invalid arg, got %v", err)
	}
	if _, _, err := NewCredentialsService(repo, nil).AuthenticateVKID(context.Background(), "vk-token", 123); !errors.Is(err, ErrVKIDUnavailable) {
		t.Fatalf("expected unavailable, got %v", err)
	}
}

func TestAuthenticateVKIDSyncsMissingContactsForExistingVKUser(t *testing.T) {
	var updatedID int64
	var updatedEmail, updatedPhone string

	repo := &stubCredentialsRepo{
		createFunc:     func(context.Context, string, string, string) (int64, error) { return 0, nil },
		createVKFunc:   func(context.Context, string, string, int64) (int64, error) { return 0, nil },
		getByEmailFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByPhoneFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByVKFunc: func(context.Context, int64) (*authrepo.Credential, error) {
			return &authrepo.Credential{ID: 15, Email: "", Phone: "", VKUserID: sql.NullInt64{Int64: 123, Valid: true}}, nil
		},
		getByIDFunc: func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		linkVKFunc:  func(context.Context, int64, int64) error { return nil },
		updateFunc: func(_ context.Context, id int64, email, phone string) error {
			updatedID = id
			updatedEmail = email
			updatedPhone = phone
			return nil
		},
		deleteFunc: func(context.Context, int64) error { return nil },
	}

	vk := &stubVKIDAuth{
		resolveFunc: func(context.Context, string) (*vkid.Identity, error) {
			return &vkid.Identity{UserID: 123, Email: "vk@example.com", Phone: "+7 900 000 00 00"}, nil
		},
	}

	svc := NewCredentialsService(repo, vk)
	id, identity, err := svc.AuthenticateVKID(context.Background(), "vk-token", 123)
	if err != nil {
		t.Fatalf("authenticate vkid existing user: %v", err)
	}
	if id != 15 {
		t.Fatalf("unexpected auth id: %d", id)
	}
	if identity == nil || identity.Email != "vk@example.com" {
		t.Fatalf("unexpected vk identity: %#v", identity)
	}
	if updatedID != 15 || updatedEmail != "vk@example.com" || updatedPhone != "9000000000" {
		t.Fatalf("unexpected synced contacts: id=%d email=%q phone=%q", updatedID, updatedEmail, updatedPhone)
	}
}

func TestUpdateAllowsPartialVKContacts(t *testing.T) {
	var updatedEmail, updatedPhone string

	repo := &stubCredentialsRepo{
		createFunc:     func(context.Context, string, string, string) (int64, error) { return 0, nil },
		createVKFunc:   func(context.Context, string, string, int64) (int64, error) { return 0, nil },
		getByEmailFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByPhoneFunc: func(context.Context, string) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByVKFunc:    func(context.Context, int64) (*authrepo.Credential, error) { return nil, sql.ErrNoRows },
		getByIDFunc: func(_ context.Context, id int64) (*authrepo.Credential, error) {
			return &authrepo.Credential{ID: id, Email: "", Phone: ""}, nil
		},
		linkVKFunc: func(context.Context, int64, int64) error { return nil },
		updateFunc: func(_ context.Context, id int64, email, phone string) error {
			updatedEmail = email
			updatedPhone = phone
			return nil
		},
		deleteFunc: func(context.Context, int64) error { return nil },
	}

	svc := NewCredentialsService(repo, nil)
	updated, err := svc.Update(context.Background(), 42, "vk@example.com", "")
	if err != nil {
		t.Fatalf("partial update with empty phone: %v", err)
	}
	if updatedEmail != "vk@example.com" || updatedPhone != "" {
		t.Fatalf("unexpected update args: email=%q phone=%q", updatedEmail, updatedPhone)
	}
	if updated.Email != "vk@example.com" || updated.Phone != "" {
		t.Fatalf("unexpected updated credential: %#v", updated)
	}
}

func TestAuthenticateVKIDConflictAndHelpers(t *testing.T) {
	if phone, err := normalizePhone("+7 (900) 123-45-67"); err != nil || phone != "9001234567" {
		t.Fatalf("unexpected normalized phone: %q %v", phone, err)
	}
	if _, err := normalizePhone("123"); err == nil {
		t.Fatal("expected phone validation error")
	}

	emailCred := &authrepo.Credential{ID: 1}
	phoneCred := &authrepo.Credential{ID: 2}
	if _, err := resolveVKIDTarget(1, emailCred, phoneCred); !errors.Is(err, ErrVKIDConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}

	emailCred = &authrepo.Credential{ID: 1}
	phoneCred = &authrepo.Credential{ID: 1}
	target, err := resolveVKIDTarget(10, emailCred, phoneCred)
	if err != nil || target.ID != 1 {
		t.Fatalf("unexpected target: %#v %v", target, err)
	}

	emailCred.VKUserID.Valid = true
	emailCred.VKUserID.Int64 = 99
	if _, err := resolveVKIDTarget(10, emailCred, nil); !errors.Is(err, ErrVKIDConflict) {
		t.Fatalf("expected vk conflict, got %v", err)
	}
}
