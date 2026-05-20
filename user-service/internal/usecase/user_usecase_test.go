package usecase

import (
	"errors"
	"testing"

	"user-service/internal/auth"
	"user-service/internal/domain"
)

type fakeUserRepository struct {
	createCalled bool
	createError  error

	usersByID    map[int]*domain.User
	usersByEmail map[string]*domain.User
}

func (f *fakeUserRepository) Create(user *domain.User) error {
	f.createCalled = true

	if f.createError != nil {
		return f.createError
	}

	if f.usersByID == nil {
		f.usersByID = make(map[int]*domain.User)
	}

	if f.usersByEmail == nil {
		f.usersByEmail = make(map[string]*domain.User)
	}

	if user.ID == 0 {
		user.ID = len(f.usersByID) + 1
	}

	f.usersByID[user.ID] = user
	f.usersByEmail[user.Email] = user

	return nil
}

func (f *fakeUserRepository) GetByEmail(email string) (*domain.User, error) {
	if user, ok := f.usersByEmail[email]; ok {
		return user, nil
	}

	return nil, errors.New("user not found")
}

func (f *fakeUserRepository) GetByID(id int) (*domain.User, error) {
	if user, ok := f.usersByID[id]; ok {
		return user, nil
	}

	return nil, errors.New("user not found")
}

func TestRegisterUserSuccess(t *testing.T) {
	repo := &fakeUserRepository{}
	uc := NewUserUsecase(repo, "test-secret")

	err := uc.RegisterUser(
		"Aruzhan Toktarbekova",
		"aruzhan@example.com",
		"password123",
		"student",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repo.createCalled {
		t.Error("expected repository Create to be called")
	}

	user := repo.usersByEmail["aruzhan@example.com"]

	if user == nil {
		t.Fatal("expected user to be saved")
	}

	if user.FullName != "Aruzhan Toktarbekova" {
		t.Errorf("expected full name Aruzhan Toktarbekova, got %s", user.FullName)
	}

	if user.Email != "aruzhan@example.com" {
		t.Errorf("expected email aruzhan@example.com, got %s", user.Email)
	}

	if user.Role != "student" {
		t.Errorf("expected role student, got %s", user.Role)
	}

	if user.PasswordHash == "password123" {
		t.Error("expected password to be hashed, but got plain password")
	}

	if user.PasswordHash == "" {
		t.Error("expected password hash, got empty string")
	}
}

func TestRegisterUserAlreadyExists(t *testing.T) {
	passwordHash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &fakeUserRepository{
		usersByEmail: map[string]*domain.User{
			"aruzhan@example.com": {
				ID:           1,
				FullName:     "Aruzhan Toktarbekova",
				Email:        "aruzhan@example.com",
				PasswordHash: passwordHash,
				Role:         "student",
			},
		},
	}

	uc := NewUserUsecase(repo, "test-secret")

	err = uc.RegisterUser(
		"Aruzhan Toktarbekova",
		"aruzhan@example.com",
		"password123",
		"student",
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if repo.createCalled {
		t.Error("repository Create should not be called when user already exists")
	}
}

func TestLoginUserSuccess(t *testing.T) {
	passwordHash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &fakeUserRepository{
		usersByEmail: map[string]*domain.User{
			"aruzhan@example.com": {
				ID:           1,
				FullName:     "Aruzhan Toktarbekova",
				Email:        "aruzhan@example.com",
				PasswordHash: passwordHash,
				Role:         "student",
			},
		},
	}

	uc := NewUserUsecase(repo, "test-secret")

	token, err := uc.LoginUser("aruzhan@example.com", "password123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if token == "" {
		t.Fatal("expected JWT token, got empty string")
	}
}

func TestLoginUserInvalidPassword(t *testing.T) {
	passwordHash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &fakeUserRepository{
		usersByEmail: map[string]*domain.User{
			"aruzhan@example.com": {
				ID:           1,
				FullName:     "Aruzhan Toktarbekova",
				Email:        "aruzhan@example.com",
				PasswordHash: passwordHash,
				Role:         "student",
			},
		},
	}

	uc := NewUserUsecase(repo, "test-secret")

	token, err := uc.LoginUser("aruzhan@example.com", "wrong-password")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if token != "" {
		t.Errorf("expected empty token, got %s", token)
	}
}

func TestGetUserByIDSuccess(t *testing.T) {
	repo := &fakeUserRepository{
		usersByID: map[int]*domain.User{
			1: {
				ID:       1,
				FullName: "Aruzhan Toktarbekova",
				Email:    "aruzhan@example.com",
				Role:     "student",
			},
		},
	}

	uc := NewUserUsecase(repo, "test-secret")

	user, err := uc.GetUserByID(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("expected user, got nil")
	}

	if user.ID != 1 {
		t.Errorf("expected user ID 1, got %d", user.ID)
	}

	if user.Email != "aruzhan@example.com" {
		t.Errorf("expected email aruzhan@example.com, got %s", user.Email)
	}
}

func TestValidateTokenSuccess(t *testing.T) {
	repo := &fakeUserRepository{
		usersByID: map[int]*domain.User{
			1: {
				ID:       1,
				FullName: "Aruzhan Toktarbekova",
				Email:    "aruzhan@example.com",
				Role:     "student",
			},
		},
	}

	secret := "test-secret"

	token, err := auth.GenerateJWT(1, "aruzhan@example.com", "student", secret)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	uc := NewUserUsecase(repo, secret)

	user, err := uc.ValidateToken(token)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user == nil {
		t.Fatal("expected user, got nil")
	}

	if user.ID != 1 {
		t.Errorf("expected user ID 1, got %d", user.ID)
	}
}

func TestValidateTokenInvalid(t *testing.T) {
	repo := &fakeUserRepository{}
	uc := NewUserUsecase(repo, "test-secret")

	user, err := uc.ValidateToken("invalid-token")

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if user != nil {
		t.Fatal("expected nil user")
	}
}
