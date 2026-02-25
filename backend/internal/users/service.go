package users

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

type RegisterInput struct {
	Email    string
	Username string
	Password string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (User, error) {
	// Check if email already exists
	var exists bool
	var err error

	exists, err = s.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return User{}, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return User{}, fmt.Errorf("email already registered")
	}

	// Hash password
	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// Insert user
	var user User
	user, err = s.repo.Insert(ctx, input.Email, input.Username, string(hash))
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
