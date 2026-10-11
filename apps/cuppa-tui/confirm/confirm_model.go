// Package confirm is a small dialog with a message and a row of buttons:
// "Save changes?", "Replace file?", error notices.
package confirm

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const (
	minW = 34
	maxW = 64
)

// Model is one open dialog. The first button is the default (Enter).
type Model struct {
	title   string
	body    []string
	buttons []string
	// art is a picture shown centred above the message (the logo), one string
	// per row, possibly with colour codes.
	art []string

	rect  design.Rect
	hover int
	// focus is the button Enter presses; it moves with Tab and the arrows and
	// is drawn once the keyboard has moved it.
	focus   int
	moved   bool
	spans   []span
	outcome modal.Outcome
	done    bool
}

type span struct{ x0, x1 int }

// New returns a dialog. With no buttons it gets a single "OK".
func New(title, body string, buttons ...string) *Model {
	if len(buttons) == 0 {
		buttons = []string{"OK"}
	}
	return &Model{title: title, body: strings.Split(body, "\n"), buttons: buttons, hover: -1}
}

// WithArt puts a picture above the message and returns the dialog.
func (m *Model) WithArt(art []string) *Model {
	m.art = art
	return m
}

func (m *Model) artHeight() int {
	if len(m.art) == 0 {
		return 0
	}
	return len(m.art) + 1
}

// Place implements modal.Modal.
func (m *Model) Place(sw, sh int) {
	w := minW
	for _, l := range m.body {
		w = max(w, ansi.StringWidth(l)+6)
	}
	for _, l := range m.art {
		w = max(w, ansi.StringWidth(l)+6)
	}
	w = min(w, min(maxW, sw))
	lines := m.wrapped(w - 4)
	h := len(lines) + 5 + m.artHeight()
	m.rect = design.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/3, 0), W: w, H: h}
}

// Rect implements modal.Modal.
func (m *Model) Rect() design.Rect { return m.rect }

func (m *Model) wrapped(w int) []string {
	var out []string
	for _, l := range m.body {
		out = append(out, strings.Split(ansi.Wrap(l, w, ""), "\n")...)
	}
	return out
}

// Lines implements modal.Modal.
func (m *Model) Lines() []string {
	body := m.wrapped(m.rect.W - 4)
	content := []string{""}
	for _, l := range m.art {
		pad := max((m.rect.W-2-ansi.StringWidth(l))/2, 0)
		content = append(content, strings.Repeat(" ", pad)+l)
	}
	if len(m.art) > 0 {
		content = append(content, "")
	}
	for _, l := range body {
		content = append(content, "  "+l)
	}
	row, spans := m.buttonRow()
	m.spans = spans
	content = append(content, "", row)
	return theme.Panel(m.title, content, m.rect.W)
}

// buttonRow lays the buttons out right-aligned inside the box and records
// their columns (relative to the box).
func (m *Model) buttonRow() (string, []span) {
	var parts []string
	var spans []span
	total := 0
	for _, b := range m.buttons {
		total += ansi.StringWidth(b) + 4 + 1
	}
	x := max(m.rect.W-2-total, 1)
	line := strings.Repeat(" ", x)
	col := 1 + x
	for i, b := range m.buttons {
		label := " " + b + " "
		style := theme.Button
		text := "[" + label + "]"
		if i == m.hover || m.hover < 0 && m.moved && i == m.focus {
			text = theme.Selected(text)
		} else {
			text = style(text, true)
		}
		parts = append(parts, text)
		spans = append(spans, span{col, col + ansi.StringWidth(text)})
		col += ansi.StringWidth(text) + 1
	}
	return line + strings.Join(parts, " "), spans
}

// Handle implements modal.Modal.
func (m *Model) Handle(e pointer.Event) {
	if m.done {
		return
	}
	x, y := e.X-m.rect.X, e.Y-m.rect.Y
	idx := -1
	if y == m.rect.H-2 {
		for i, s := range m.spans {
			if x >= s.x0 && x < s.x1 {
				idx = i
			}
		}
	}
	switch e.Phase {
	case pointer.Move:
		m.hover = idx
	case pointer.Down:
		if e.Left && idx >= 0 {
			m.finish(m.buttons[idx], false)
		}
	}
}

// Key implements modal.Modal.
func (m *Model) Key(_ string, _, enter, esc bool) {
	switch {
	case esc:
		m.finish("", true)
	case enter:
		m.finish(m.buttons[m.focus], false)
	}
}

// Nav implements modal.Navigator: Tab, Shift+Tab and the arrows choose the
// button that Enter presses.
func (m *Model) Nav(name string) bool {
	n := len(m.buttons)
	switch name {
	case "tab", "right", "down":
		m.focus = (m.focus + 1) % n
	case "shift+tab", "left", "up":
		m.focus = (m.focus + n - 1) % n
	default:
		return false
	}
	m.moved = true
	return true
}

func (m *Model) finish(button string, canceled bool) {
	m.done = true
	m.outcome = modal.Outcome{Button: button, Canceled: canceled}
}

// Outcome implements modal.Modal.
func (m *Model) Outcome() (modal.Outcome, bool) { return m.outcome, m.done }

// Describe reads the dialog out: its title, message and buttons.
func (m *Model) Describe() []a11y.Node {
	nodes := []a11y.Node{a11y.Heading(m.title)}
	for _, l := range m.body {
		if l != "" {
			nodes = append(nodes, a11y.Text(l))
		}
	}
	for i, b := range m.buttons {
		btn := a11y.Button(b)
		btn.Focused = i == m.focus // Enter presses this one
		nodes = append(nodes, btn)
	}
	return nodes
}
