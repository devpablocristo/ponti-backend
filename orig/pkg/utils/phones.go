package pkgutils

import (
	"fmt"
	"strings"
	"unicode"
)

// ValidatePhone checks if the phone number has at least the required number of digits.
func ValidatePhone(phone string, minDigits int) error {
	digits := strings.Map(func(r rune) rune {
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, phone)

	if len(digits) < minDigits {
		return fmt.Errorf("phone number must have at least %d digits", minDigits)
	}
	return nil
}
