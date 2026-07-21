package utils

import "github.com/microcosm-cc/bluemonday"

var htmlSanitizePolicy = bluemonday.StrictPolicy()

func SanitizeHTML(value string) string {
	return htmlSanitizePolicy.Sanitize(value)
}
