package colorpicker

import (
	"fmt"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/color/scheme"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
)

// letters are the jump targets above the swatches: "#" is every name that
// starts with a digit.
const letters = "#ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// swatchOfRGB draws a true-colour block w cells wide, with a check mark when it
// is the current colour.
func swatchOfRGB(c space.RGB, w int, marked bool) string {
	text := strings.Repeat(" ", w)
	if marked {
		pad := (w - 1) / 2
		text = strings.Repeat(" ", pad) + "✓" + strings.Repeat(" ", w-pad-1)
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(c.Hex())).Foreground(contrast(c)).Render(text)
}

// stepScheme moves the shown scheme by n, wrapping around the list.
func (m *Model) stepScheme(n int) {
	count := len(scheme.All())
	m.scheme = ((m.scheme+n)%count + count) % count
}

// jumpScheme shows the first scheme whose name starts with the letter ("#" for a digit).
func (m *Model) jumpScheme(letter rune) {
	for i, s := range scheme.All() {
		first := []rune(s.Name)[0]
		if letter == '#' && unicode.IsDigit(first) || letter != '#' && unicode.ToUpper(first) == letter {
			m.scheme = i
			return
		}
	}
}

// themesLines draws the Themes tab: the scheme's name with step buttons, a row
// of letters to jump by, its 16 colours, and its text, page, cursor and
// selection colours when it has them.
func (m *Model) themesLines(y0 int) []string {
	all := scheme.All()
	s := all[m.scheme]
	var lines []string
	next := func() *row { return m.newRow(y0 + len(lines)).text(" ") }
	step := func(label string, n int) func(*row) {
		return func(r *row) { r.span(label, func(int) { m.stepScheme(n) }, nil) }
	}

	r := next()
	for _, b := range []func(*row){step(" « ", -10), step(" ◂ ", -1)} {
		b(r)
	}
	kind := "light"
	if s.Dark {
		kind = "dark"
	}
	r.text(theme.Selected(" " + s.Name + " "))
	step(" ▸ ", 1)(r)
	step(" » ", 10)(r)
	lines = append(lines, r.String())
	lines = append(lines, theme.Dim(fmt.Sprintf("   %s · %d of %d", kind, m.scheme+1, len(all))))

	r = next()
	for _, l := range letters {
		l := l
		r.span(theme.Dim(string(l)), func(int) { m.jumpScheme(l) }, nil)
	}
	lines = append(lines, r.String(), "")

	marked := func(c space.RGB) bool { return !m.empty && m.index < 0 && m.rgb == c }
	for half := 0; half < 2; half++ {
		for h := 0; h < 2; h++ { // each swatch is two terminal rows tall
			r := next()
			for col := 0; col < 8; col++ {
				c := s.ANSI[half*8+col]
				r.span(swatchOfRGB(c, 5, marked(c) && h == 0), func(int) { m.setRGB(c) }, nil).text(" ")
			}
			lines = append(lines, r.String())
		}
	}
	lines = append(lines, "")

	r = next()
	extra := func(label string, c space.RGB) {
		r.text(theme.Dim(label + " ")).span(swatchOfRGB(c, 4, marked(c)), func(int) { m.setRGB(c) }, nil).text("  ")
	}
	extra("Text", s.Foreground)
	extra("Page", s.Background)
	if s.HasCursor {
		extra("Cursor", s.Cursor)
	}
	if s.HasSelection {
		extra("Select", s.Selection)
	}
	return append(lines, r.String())
}

// startScheme is the scheme the Themes tab opens on: Dracula when the list has
// it, otherwise the first.
func startScheme() int {
	for i, s := range scheme.All() {
		if s.Name == "Dracula" || strings.EqualFold(s.Name, "dracula") {
			return i
		}
	}
	return 0
}
