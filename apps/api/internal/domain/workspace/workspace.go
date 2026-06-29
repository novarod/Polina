package workspace

import (
	"unicode/utf8"

	"github.com/novarod/polina/apps/api/pkg/apierr"
)

const (
	nameMin = 2
	nameMax = 255
	descMax = 1000
)

func ValidateName(name string) error {
	n := utf8.RuneCountInString(name)
	if n < nameMin || n > nameMax {
		return apierr.Validation("name", "name must be between 2 and 255 characters")
	}
	return nil
}

func ValidateDescription(description string) error {
	if utf8.RuneCountInString(description) > descMax {
		return apierr.Validation("description", "description must be at most 1000 characters")
	}
	return nil
}
