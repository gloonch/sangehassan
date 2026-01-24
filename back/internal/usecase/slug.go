package usecase

import (
	"regexp"
	"strings"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = slugRegex.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	return value
}
