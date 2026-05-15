package usecase

import (
	"errors"
	"user-service/internal/auth"
	"user-service/internal/domain"
	"user-service/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type UserUsecase interface {
	RegisterUser(fullName, email, password, role string) error
	LoginUser(email, password string) (string, error)
	GetUserByID(id int) (*domain.User, error)
	ValidateToken(token string) (*domain.User, error)
}

type userUsecase struct {
	repo      repository.UserRepository
	jwtSecret string
}

func NewUserUsecase(repo repository.UserRepository, jwtSecret string) UserUsecase {
	return &userUsecase{repo: repo, jwtSecret: jwtSecret}
}

func (u *userUsecase) RegisterUser(fullName, email, password, role string) error {
	// Check if user exists
	_, err := u.repo.GetByEmail(email)
	if err == nil {
		return errors.New("user already exists")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	user := &domain.User{
		FullName:     fullName,
		Email:        email,
		PasswordHash: hash,
		Role:         role,
	}
	return u.repo.Create(user)
}

func (u *userUsecase) LoginUser(email, password string) (string, error) {
	user, err := u.repo.GetByEmail(email)
	if err != nil {
		return "", ErrInvalidCredentials
	}
	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		return "", ErrInvalidCredentials
	}
	return auth.GenerateJWT(user.ID, user.Email, user.Role, u.jwtSecret)
}

func (u *userUsecase) GetUserByID(id int) (*domain.User, error) {
	return u.repo.GetByID(id)
}

func (u *userUsecase) ValidateToken(token string) (*domain.User, error) {
	claims, err := auth.ValidateJWT(token, u.jwtSecret)
	if err != nil {
		return nil, err
	}
	return u.repo.GetByID(claims.UserID)
}
