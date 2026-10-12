package runcmd

import "fmt"

// splitCommand is a command line as the program and its arguments. Words are
// separated by spaces; a word in double quotes keeps its spaces, and a
// backslash before a quote or a backslash keeps it. There is no shell: no
// variables, no pipes, no &&.
func splitCommand(line string) ([]string, error) {
	var words []string
	var cur []rune
	inWord, quoted := false, false
	rs := []rune(line)
	for i := 0; i < len(rs); i++ {
		r := rs[i]
		switch {
		case r == '\\' && i+1 < len(rs) && (rs[i+1] == '"' || rs[i+1] == '\\') && quoted:
			cur = append(cur, rs[i+1])
			i++
		case r == '"':
			quoted = !quoted
			inWord = true
		case (r == ' ' || r == '\t') && !quoted:
			if inWord {
				words = append(words, string(cur))
				cur, inWord = nil, false
			}
		default:
			cur = append(cur, r)
			inWord = true
		}
	}
	if quoted {
		return nil, fmt.Errorf("the quote in %q is not closed", line)
	}
	if inWord {
		words = append(words, string(cur))
	}
	return words, nil
}
