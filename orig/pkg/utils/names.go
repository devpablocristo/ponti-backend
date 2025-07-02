package pkgutils

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateName checks if the name meets length and content requirements.
func ValidateName(name string, minLen, maxLen int) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	if len(name) < minLen || len(name) > maxLen {
		return fmt.Errorf("name length must be between %d and %d characters", minLen, maxLen)
	}
	if strings.Contains(name, "  ") {
		return errors.New("name cannot contain consecutive spaces")
	}
	return nil
}
