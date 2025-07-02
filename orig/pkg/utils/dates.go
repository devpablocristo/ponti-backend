package pkgutils

import (
	"fmt"
	"time"
)

// ValidateBirthDate checks if the birthDate is consistent with the provided age and not in the future.
func ValidateBirthDate(birthDate time.Time, expectedAge int) error {
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	if age != expectedAge {
		return fmt.Errorf("birth date does not match the provided age")
	}
	if birthDate.After(now) {
		return fmt.Errorf("birth date cannot be in the future")
	}
	return nil
}
