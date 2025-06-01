package pkgutils

import (
	"errors"
	"log"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt.
func HashPassword(password string, cost int) (string, error) {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return "", errors.New("error hashing password")
	}
	return string(hashedPassword), nil
}

// VerifyPassword checks if the password matches the hashed password.
func VerifyPassword(password, hashedPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil
		}
		log.Printf("error verifying password: %v", err)
		return false, errors.New("error verifying password")
	}
	return true, nil
}

// ValidatePasswordComplexity enforces strong password requirements.
func ValidatePasswordComplexity(password string) error {
	var hasMinLen, hasUpper, hasLower, hasNumber, hasSpecial bool
	const minLen = 8
	if len(password) >= minLen {
		hasMinLen = true
	}
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	if !hasMinLen {
		return errors.New("password must have at least 8 characters")
	}
	if !hasUpper {
		return errors.New("password must have at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must have at least one lowercase letter")
	}
	if !hasNumber {
		return errors.New("password must have at least one number")
	}
	if !hasSpecial {
		return errors.New("password must have at least one special character")
	}
	return nil
}
