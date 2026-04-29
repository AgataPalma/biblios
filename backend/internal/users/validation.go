package users

import (
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)

func isEmailValid(e string) bool {
	return emailRegex.MatchString(strings.ToLower(e))
}

func ValidateRegisterInput(input RegisterInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if !isEmailValid(input.Email) {
		return fmt.Errorf("invalid email")
	}
	if strings.TrimSpace(input.Username) == "" {
		return fmt.Errorf("username is required")
	}
	if len(input.Username) < 3 {
		return fmt.Errorf("username must be at least 3 characters")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}
