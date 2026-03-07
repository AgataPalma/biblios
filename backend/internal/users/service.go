package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type userRepository interface {
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUsername(ctx context.Context, username string) (bool, error)
	CreateUser(ctx context.Context, email, username, passwordHash string) (User, error)
	UpdateUser(ctx context.Context, userID, email, username string) (User, error)
	UpdatePassword(ctx context.Context, userID string, newPasswordHash string) error
	SoftDeleteWithCascade(ctx context.Context, userID string) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id string) (User, error)
	UpdateTheme(ctx context.Context, userID string, theme string) error
}

type Service struct {
	repo userRepository
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

type UpdatePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

type UpdateUserInput struct {
	UserID   string
	Email    *string
	Username *string
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func isUniqueViolation(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
	}
	return false
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (User, error) {
	// Check if email already exists
	var existsEmail bool
	var err error

	existsEmail, err = s.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return User{}, fmt.Errorf("failed to check email: %w", err)
	}
	if existsEmail {
		return User{}, fmt.Errorf("email already registered")
	}

	//Check if username already exists
	var existsUsername bool

	existsUsername, err = s.repo.ExistsByUsername(ctx, input.Username)
	if err != nil {
		return User{}, fmt.Errorf("failed to check username: %w", err)
	}
	if existsUsername {
		return User{}, fmt.Errorf("username already taken")
	}

	// Hash password
	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	var user User
	user, err = s.repo.CreateUser(ctx, input.Email, input.Username, string(hash))
	if err != nil {
		return User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (User, error) {
	var user User
	var err error

	user, err = s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		// Don't reveal whether email exists or not
		return User{}, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return User{}, fmt.Errorf("invalid credentials")
	}

	return user, nil
}

func (s *Service) UpdateTheme(ctx context.Context, userID string, theme string) error {
	return s.repo.UpdateTheme(ctx, userID, theme)
}

func (s *Service) GetByID(ctx context.Context, id string) (User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, input UpdateUserInput) (User, error) {
	var user User
	var err error

	user, err = s.repo.UpdateUser(ctx, input.UserID, input.Email, input.Username)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return User{}, fmt.Errorf("email already in use")
		}
		if isUniqueViolation(err, "users_username_key") {
			return User{}, fmt.Errorf("username already in use")
		}
		return User{}, fmt.Errorf("failed to update profile: %w", err)
	}

	return user, nil
}

func (s *Service) UpdatePassword(ctx context.Context, input UpdatePasswordInput) error {
	var user User
	var err error

	user, err = s.repo.FindByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword))
	if err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, input.UserID, string(hash))
}

func (s *Service) DeleteUser(ctx context.Context, userID string) error {
	return s.repo.SoftDeleteWithCascade(ctx, userID)
}
