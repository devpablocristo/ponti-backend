package pkgutils

import (
	"fmt"
	"strings"
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

// ISODate para bind “YYYY-MM-DD”
type ISODate time.Time

func (d *ISODate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	*d = ISODate(t)
	return nil
}

func (d ISODate) MarshalJSON() ([]byte, error) {
	return []byte(`"` + time.Time(d).Format("2006-01-02") + `"`), nil
}
