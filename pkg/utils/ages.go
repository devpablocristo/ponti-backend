package pkgutils

import "fmt"

// ValidateAge checks if age is within the allowed bounds.
func ValidateAge(age, minAge, maxAge int) error {
	if age < minAge || age > maxAge {
		return fmt.Errorf("age must be between %d and %d", minAge, maxAge)
	}
	return nil
}
