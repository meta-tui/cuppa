package design

import (
	"strings"
	"unicode"
)

// ValidSwatchName reports whether s can name a colour of the palette: it starts
// with a letter and holds letters, digits, spaces, dashes and underscores.
func ValidSwatchName(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case unicode.IsLetter(r):
		case i > 0 && (unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_'):
		default:
			return false
		}
	}
	return true
}

// CleanSwatchName trims a name and collapses its inner spaces.
func CleanSwatchName(s string) string { return strings.Join(strings.Fields(s), " ") }
