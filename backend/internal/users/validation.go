package users

import (
	"fmt"
	"regexp"
	"strings"
)

func isEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func ValidateRegisterInput(input RegisterInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email é obrigatório")
	}
	if !isEmailValid(input.Email) {
		return fmt.Errorf("email inválido")
	}

	if strings.TrimSpace(input.Username) == "" {
		return fmt.Errorf("utilizador obrigatório")
	}
	if len(input.Username) < 3 {
		return fmt.Errorf("utilizador tem de conter pelo menos 3 caracteres")
	}
	if len(input.Password) < 8 {
		return fmt.Errorf("password tem de ter pelo menos 8 caracteres")
	}
	return nil
}
