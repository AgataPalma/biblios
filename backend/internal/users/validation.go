package users

import (
	"fmt"
	"strings"
)

func ValidateRegisterInput(input RegisterInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if !strings.Contains(input.Email, "@") {
		return fmt.Errorf("email is invalid")
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
