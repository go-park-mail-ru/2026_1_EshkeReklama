package repository

import "sync"

type User struct {
	ID       int
	Email    string
	Phone    string
	Password string
}

type UserRepository struct {
	mu           sync.RWMutex
	usersByEmail map[string]*User
	usersByPhone map[string]*User
	lastID       int
}

func InitUserRepository() *UserRepository {
	userRepo := &UserRepository{
		mu:           sync.RWMutex{},
		usersByEmail: make(map[string]*User),
		usersByPhone: make(map[string]*User),
		lastID:       0,
	}

	userRepo.lastID++

	firstUser := &User{
		ID:       userRepo.lastID,
		Email:    "test@mail.com",
		Phone:    "+79991234567",
		Password: "123123",
	}
	userRepo.usersByEmail[firstUser.Email] = firstUser
	userRepo.usersByPhone[firstUser.Phone] = firstUser

	return userRepo
}

func (r *UserRepository) GetByEmail(email string) (*User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByEmail[email]
	return user, exists
}

func (r *UserRepository) GetByPhone(phone string) (*User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.usersByPhone[phone]
	return user, exists
}

func (r *UserRepository) Create(email, phone, password string) *User {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastID++
	newUser := &User{
		ID:       r.lastID,
		Email:    email,
		Phone:    phone,
		Password: password,
	}
	r.usersByEmail[newUser.Email] = newUser
	r.usersByPhone[newUser.Phone] = newUser
	return newUser
}
