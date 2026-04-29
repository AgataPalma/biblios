package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
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

type LoginInput struct {
	Email    string
	Password string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (User, error) {
	exists, err := s.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return User{}, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return User{}, fmt.Errorf("email already registered")
	}

	exists, err = s.repo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return User{}, fmt.Errorf("check username: %w", err)
	}
	if exists {
		return User{}, fmt.Errorf("username already taken")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, input.Email, input.Username, string(hash))
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	// Create default library
	if err = s.repo.CreateDefaultLibrary(ctx, user.ID); err != nil {
		return User{}, fmt.Errorf("create default library: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (User, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return User{}, fmt.Errorf("invalid credentials")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return User{}, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, username, bio, avatarURL *string) (User, error) {
	user, err := s.repo.UpdateProfile(ctx, userID, username, bio, avatarURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if pgErr.ConstraintName == "users_username_key" {
				return User{}, fmt.Errorf("username already taken")
			}
		}
		return User{}, err
	}
	return user, nil
}

func (s *Service) UpdateEmail(ctx context.Context, userID, currentPassword, newEmail string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}
	return s.repo.UpdateEmail(ctx, userID, newEmail)
}

func (s *Service) UpdatePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return fmt.Errorf("current password is incorrect")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	return s.repo.UpdatePassword(ctx, userID, string(hash))
}

func (s *Service) UpdateTheme(ctx context.Context, userID, theme string) error {
	validThemes := map[string]bool{
		"default-light": true, "woody": true, "nordic": true, "metallic": true,
		"futuristic": true, "post-apocalyptic": true, "dark-academia": true,
		"ocean": true, "space": true,
	}
	if !validThemes[theme] {
		return fmt.Errorf("invalid theme")
	}
	return s.repo.UpdateTheme(ctx, userID, theme)
}

func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.SoftDelete(ctx, userID)
}
