package organization

import (
	"regexp"
	"unicode/utf8"

	"github.com/novarod/polina/apps/api/pkg/apierr"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

const (
	nameMin = 2
	nameMax = 255
	slugMin = 3
	slugMax = 63
)

func ValidateName(name string) error {
	n := utf8.RuneCountInString(name)
	if n < nameMin || n > nameMax {
		return apierr.Validation("name", "name must be between 2 and 255 characters")
	}
	return nil
}

func ValidateSlug(slug string) error {
	if len(slug) < slugMin || len(slug) > slugMax || !slugPattern.MatchString(slug) {
		return apierr.Validation("slug", "slug must be 3-63 characters, lowercase alphanumeric and hyphens")
	}
	return nil
}
