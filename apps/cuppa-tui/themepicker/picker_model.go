// Package themepicker is the dialog that fills a design's theme from a terminal
// colour scheme: step through the schemes, see the five colours it would set
// and a small sample, and apply them. It is a modal box, operated by mouse.
package themepicker

import (
	"fmt"
	"strings"
	"unicode"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/color/scheme"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const (
	innerW = 46
	boxW   = innerW + 2
	// letters are the jump targets under the name: "#" is every name that
	// starts with a digit.
	letters = "#ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Rows of the box, counted from its top border (row 0). Between the letters and
// the buttons are the five colours (rows 4 to 8) and the sample (rows 10 to 12).
const (
	rowName    = 1
	rowLetters = 2
	rowButtons = 14
	boxH       = 16
)

type region struct {
	y, x0, x1 int
	act       func()
}

// Model is one open dialog.
type Model struct {
	title string
	index int // into scheme.All

	sw, sh  int
	rect    design.Rect
	regions []region

	outcome modal.Outcome
	done    bool
}

// New opens the dialog on the scheme named start (the first when it is not
// there).
func New(title, start string) *Model {
	m := &Model{title: title}
	for i, s := range scheme.All() {
		if strings.EqualFold(s.Name, start) {
			m.index = i
		}
	}
	return m
}

// Parse reads the value an OK outcome carries: the background and the four
// theme colours, as hex.
func Parse(value string) (background string, t design.Theme, ok bool) {
	parts := strings.Split(value, "|")
	if len(parts) != 5 {
		return "", design.Theme{}, false
	}
	return parts[0], design.Theme{Text: parts[1], Muted: parts[2], Border: parts[3], Secondary: parts[4]}, true
}

func (m *Model) current() scheme.Scheme { return scheme.All()[m.index] }

func (m *Model) value() string {
	t := m.current().Theme()
	return strings.Join([]string{t.Background.Hex(), t.Text.Hex(), t.Muted.Hex(), t.Border.Hex(), t.Secondary.Hex()}, "|")
}

func (m *Model) step(n int) {
	count := len(scheme.All())
	m.index = ((m.index+n)%count + count) % count
}

func (m *Model) jump(letter rune) {
	for i, s := range scheme.All() {
		first := []rune(s.Name)[0]
		if letter == '#' && unicode.IsDigit(first) || letter != '#' && unicode.ToUpper(first) == letter {
			m.index = i
			return
		}
	}
}

// Place implements modal.Modal.
func (m *Model) Place(sw, sh int) { m.sw, m.sh = sw, sh; m.rect = m.layout() }

func (m *Model) layout() design.Rect {
	w := min(boxW, max(m.sw, 20))
	return design.Rect{X: max((m.sw-w)/2, 0), Y: max((m.sh-boxH)/4, 0), W: w, H: boxH}
}

// Rect implements modal.Modal.
func (m *Model) Rect() design.Rect { m.rect = m.layout(); return m.rect }

// Outcome implements modal.Modal.
func (m *Model) Outcome() (modal.Outcome, bool) { return m.outcome, m.done }

// Key implements modal.Modal: Enter applies, Esc cancels.
func (m *Model) Key(_ string, _, enter, esc bool) {
	switch {
	case esc:
		m.done, m.outcome = true, modal.Outcome{Canceled: true}
	case enter:
		m.apply()
	}
}

// Nav implements modal.Navigator: Left and Right step through the schemes
// (Shift: ten), Up and Down by ten, Home and End jump, a letter or digit
// jumps to the first scheme that starts with it.
func (m *Model) Nav(name string) bool {
	switch name {
	case "left":
		m.step(-1)
	case "right":
		m.step(1)
	case "shift+left", "up", "pgup":
		m.step(-10)
	case "shift+right", "down", "pgdown":
		m.step(10)
	case "home":
		m.index = 0
	case "end":
		m.index = len(scheme.All()) - 1
	default:
		r := []rune(name)
		if len(r) != 1 || !unicode.IsLetter(r[0]) && !unicode.IsDigit(r[0]) {
			return false
		}
		if unicode.IsDigit(r[0]) {
			m.jump('#')
		} else {
			m.jump(unicode.ToUpper(r[0]))
		}
	}
	return true
}

func (m *Model) apply() {
	m.done, m.outcome = true, modal.Outcome{Button: "Apply", Value: m.value()}
}

// Handle implements modal.Modal.
func (m *Model) Handle(e pointer.Event) {
	if m.done {
		return
	}
	x, y := e.X-m.rect.X, e.Y-m.rect.Y
	switch e.Phase {
	case pointer.Down:
		if !e.Left {
			return
		}
		for _, r := range m.regions {
			if r.y == y && x >= r.x0 && x < r.x1 {
				r.act()
				return
			}
		}
	case pointer.Wheel:
		if e.WheelY != 0 {
			m.step(e.WheelY)
		}
	}
}

// Lines implements modal.Modal.
func (m *Model) Lines() []string {
	m.rect = m.layout()
	m.regions = m.regions[:0]
	s := m.current()
	t := s.Theme()
	var content []string

	// Name row: « ◂ Name ▸ » and what kind of scheme it is.
	r := m.row(rowName)
	for _, b := range []struct {
		label string
		n     int
	}{{" « ", -10}, {" ◂ ", -1}} {
		n := b.n
		r.span(b.label, func() { m.step(n) })
	}
	r.text(theme.Selected(" " + s.Name + " "))
	r.span(" ▸ ", func() { m.step(1) })
	r.span(" » ", func() { m.step(10) })
	kind := "light"
	if s.Dark {
		kind = "dark"
	}
	r.text(theme.Dim(fmt.Sprintf(" %s · %d of %d", kind, m.index+1, len(scheme.All()))))
	content = append(content, r.String())

	// Letters.
	r = m.row(rowLetters)
	for _, l := range letters {
		l := l
		r.span(theme.Dim(string(l)), func() { m.jump(l) })
	}
	content = append(content, r.String(), "")

	// The five colours.
	for _, c := range []struct {
		label string
		c     space.RGB
	}{{"Background", t.Background}, {"Text", t.Text}, {"Muted", t.Muted}, {"Border", t.Border}, {"Secondary", t.Secondary}} {
		swatch := lipgloss.NewStyle().Background(lipgloss.Color(c.c.Hex())).Render("      ")
		content = append(content, fmt.Sprintf(" %s %s %s", theme.Dim(fmt.Sprintf("%-11s", c.label)), swatch, theme.Dim(c.c.Hex())))
	}
	content = append(content, "")

	// A small sample in the theme's colours.
	content = append(content, m.sample(t)...)
	content = append(content, "")

	// Buttons.
	r = m.row(rowButtons)
	apply, cancel := "[ Apply ]", "[ Cancel ]"
	gap := max(innerW-1-ansi.StringWidth(apply)-ansi.StringWidth(cancel)-1, 1)
	r.text(strings.Repeat(" ", gap)).span(theme.Button(apply, true), m.apply)
	r.text(" ").span(theme.Button(cancel, true), func() { m.done, m.outcome = true, modal.Outcome{Canceled: true} })
	content = append(content, r.String())

	return theme.Panel(m.title, content, m.rect.W)
}

// sample draws three lines on the scheme's page colour: a frame in the border
// colour, text, muted text and an accent.
func (m *Model) sample(t scheme.ThemeColours) []string {
	page := lipgloss.NewStyle().Background(lipgloss.Color(t.Background.Hex()))
	on := func(c space.RGB) lipgloss.Style { return page.Foreground(lipgloss.Color(c.Hex())) }
	pad := func(line string, width int) string {
		return line + page.Render(strings.Repeat(" ", max(width-ansi.StringWidth(line), 0)))
	}
	const w = 40
	b := on(t.Border)
	top := b.Render("╭─ ") + on(t.Secondary).Bold(true).Render("Sample") + b.Render(" "+strings.Repeat("─", w-11)+"╮")
	mid := b.Render("│ ") + on(t.Text).Render("Some text, ") + on(t.Muted).Render("some quiet text ") + on(t.Secondary).Render("› accent")
	mid = pad(mid, w-1) + b.Render("│")
	bottom := b.Render("╰" + strings.Repeat("─", w-2) + "╯")
	return []string{" " + top, " " + mid, " " + bottom}
}

// Describe reads the dialog out for a screen reader.
func (m *Model) Describe() []a11y.Node {
	s := m.current()
	t := s.Theme()
	return []a11y.Node{
		a11y.Heading(m.title),
		a11y.Text(fmt.Sprintf("Scheme %s, number %d of %d. Background %s, text %s, muted %s, border %s, secondary %s.",
			s.Name, m.index+1, len(scheme.All()), t.Background.Hex(), t.Text.Hex(), t.Muted.Hex(), t.Border.Hex(), t.Secondary.Hex())),
		a11y.Button("Apply"), a11y.Button("Cancel"),
	}
}

// row builds one content line and records its clickable spans; coordinates are
// relative to the box (its left border is column 0).
type row struct {
	m *Model
	y int
	x int
	b strings.Builder
}

func (m *Model) row(y int) *row { return &row{m: m, y: y, x: 1} }

func (r *row) text(s string) *row {
	r.b.WriteString(s)
	r.x += ansi.StringWidth(s)
	return r
}

func (r *row) span(s string, act func()) *row {
	w := ansi.StringWidth(s)
	r.m.regions = append(r.m.regions, region{y: r.y, x0: r.x, x1: r.x + w, act: act})
	return r.text(s)
}

func (r *row) String() string { return r.b.String() }
