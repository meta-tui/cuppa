package gosource

import (
	"strconv"
	"strings"
)

// emptyDefault is the app default that says an input starts empty.
const emptyDefault = "(empty)"

// defaultLiteral is the Go value an app default gives an input of type typ, or
// false when the text is not one of that type.
func defaultLiteral(typ, text string) (string, bool) {
	text = strings.TrimSpace(text)
	if text == emptyDefault {
		switch {
		case typ == "bool":
			return "false", true
		case typ == "int", typ == "float64":
			return "0", true
		case typ == "string":
			return `""`, true
		case strings.HasPrefix(typ, "[]"):
			return "nil", true
		}
		return typ + "{}", true
	}
	switch typ {
	case "bool":
		if text == "true" || text == "false" {
			return text, true
		}
	case "int":
		if n, err := strconv.Atoi(text); err == nil {
			return strconv.Itoa(n), true
		}
	case "float64":
		if f, err := strconv.ParseFloat(text, 64); err == nil {
			return strconv.FormatFloat(f, 'f', -1, 64), true
		}
	case "string":
		return strconv.Quote(text), true
	}
	return "", false
}
