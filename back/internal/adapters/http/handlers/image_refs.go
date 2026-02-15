package handlers

import (
	"net/url"
	"strings"
)

// We only store and accept images that are served from our own /images/ path.
// This keeps the system "upload-only" and prevents embedding arbitrary external URLs.
func isAllowedImageRef(value string) bool {
	if value == "" {
		return true
	}
	if strings.HasPrefix(value, "/images/") {
		return true
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	return strings.HasPrefix(parsed.Path, "/images/")
}

func normalizeImageRef(value string) string {
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "/images/") {
		return value
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return value
	}
	if strings.HasPrefix(parsed.Path, "/images/") {
		return parsed.Path
	}
	return value
}

func normalizeImageRefs(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		out = append(out, normalizeImageRef(v))
	}
	return out
}

func allAllowedImageRefs(values []string) bool {
	for _, v := range values {
		if !isAllowedImageRef(v) {
			return false
		}
	}
	return true
}
