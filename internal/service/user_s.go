package service

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"awesomeProject/internal/auth"
	"awesomeProject/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("неверный email или пароль")
	ErrUserNotFound       = repository.ErrUserNotFound
)

type UserRepository interface {
	CreateUser(user repository.User) (repository.User, error)
	GetUser(id string) (repository.User, error)
	GetUserByEmail(email string) (repository.User, error)
}

type UserService struct {
	repo UserRepository
	jwt  *auth.Manager
}

func NewUserService(repo UserRepository, jwt *auth.Manager) *UserService {
	return &UserService{repo: repo, jwt: jwt}
}

type AuthResult struct {
	User        repository.User
	AccessToken string
}

func (s *UserService) Register(email, password, name string) (AuthResult, error) {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)

	if email == "" || password == "" {
		return AuthResult{}, errors.New("email и password обязательны")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	role := "user"
	saved, err := s.repo.CreateUser(repository.User{
		Email:        email,
		PasswordHash: string(hash),
		Name:         name,
		Role:         role,
	})
	if err != nil {
		return AuthResult{}, err
	}

	token, err := s.jwt.Issue(saved.ID, saved.Email, saved.Role)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: saved, AccessToken: token}, nil
}

func (s *UserService) Login(email, password string) (AuthResult, error) {
	email = strings.TrimSpace(email)
	if email == "" || password == "" {
		return AuthResult{}, errors.New("email и password обязательны")
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	token, err := s.jwt.Issue(user.ID, user.Email, user.Role)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: user, AccessToken: token}, nil
}

func (s *UserService) GetUser(userID string) (repository.User, error) {
	if userID == "" {
		return repository.User{}, errors.New("user_id не может быть пустым")
	}
	return s.repo.GetUser(userID)
}
