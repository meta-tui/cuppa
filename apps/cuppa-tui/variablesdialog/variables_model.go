// Package variablesdialog is the Variables box: every named colour, screen
// input and event of the design, what each is linked to, and a way to go to
// where it is used or to rename it.
package variablesdialog

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const (
	width = 72
	// GoToButton and RenameButton are the Outcome.Button values. A go-to
	// carries the component id in Outcome.Value ("" for a screen key); a
	// rename carries "kind:name".
	GoToButton   = "Go to"
	RenameButton = "Rename"
	doneButton   = "Done"
)

// Model is the open dialog.
type Model struct {
	vars []editor.Variable
	// cursor is the variable the keyboard is on; inUses is true while the
	// places a variable is used are listed for choosing one.
	cursor    int
	useCursor int
	inUses    bool
	top       int
	rect      design.Rect
	hoverY    int
	hotBtn    string
	outcome   modal.Outcome
	done      bool
	// rowY is the y of the first listed row inside the box, for the pointer.
	rowY, rows int
}

// New returns the dialog for the variables of a design.
func New(vars []editor.Variable) *Model {
	return &Model{vars: vars, hoverY: -1}
}

// Place implements modal.Modal.
func (m *Model) Place(sw, sh int) {
	w := min(width, sw)
	most := len(m.vars)
	for _, v := range m.vars {
		most = max(most, len(v.Uses))
	}
	m.rows = min(max(most, 3), max(sh-10, 3))
	h := m.rows + 6
	m.rect = design.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/3, 0), W: w, H: h}
}

// Rect implements modal.Modal.
func (m *Model) Rect() design.Rect { return m.rect }

// Lines implements modal.Modal.
func (m *Model) Lines() []string {
	inner := m.rect.W - 4
	var content []string
	if m.inUses {
		v := m.vars[m.cursor]
		content = append(content, " "+theme.Bold(v.Name)+" "+theme.Faded("is used in "+plural(len(v.Uses), "place")+": choose one"))
		m.scrollTo(m.useCursor, len(v.Uses))
		m.rowY = len(content) + 1
		for i := m.top; i < min(m.top+m.rows, len(v.Uses)); i++ {
			u := v.Uses[i]
			line := fmt.Sprintf(" %s  %s", theme.Fit(u.NodeName, inner/2-2), theme.Faded(u.Where))
			content = append(content, m.row(line, i == m.useCursor, i))
		}
	} else {
		content = append(content, " "+theme.Faded("Name, kind and where it is linked. Enter goes to a use."))
		m.scrollTo(m.cursor, len(m.vars))
		m.rowY = len(content) + 1
		if len(m.vars) == 0 {
			content = append(content, " "+theme.Faded("No named colours, inputs or events yet."))
		}
		for i := m.top; i < min(m.top+m.rows, len(m.vars)); i++ {
			content = append(content, m.row(m.variableLine(m.vars[i], inner), i == m.cursor, i))
		}
	}
	for len(content) < m.rows+1 {
		content = append(content, "")
	}
	content = append(content, "")
	buttons := []string{m.button(GoToButton, "[ Go to ]"), m.button(RenameButton, "[ Rename… ]"), m.button(doneButton, "[ Done ]")}
	row := strings.Join(buttons, " ")
	pad := max(m.rect.W-2-ansi.StringWidth(row)-1, 0)
	content = append(content, strings.Repeat(" ", pad)+row)
	return theme.Panel("Variables", content, m.rect.W)
}

func (m *Model) variableLine(v editor.Variable, inner int) string {
	value := v.Value
	swatch := "  "
	if v.Kind == editor.VarColour {
		swatch = lipgloss.NewStyle().Background(lipgloss.Color(v.Value)).Render("  ")
	}
	uses := theme.Faded(plural(len(v.Uses), "use"))
	if len(v.Uses) == 0 {
		uses = theme.Faded("not used")
	}
	name := theme.Fit(v.Name, 22)
	kind := theme.Faded(theme.Fit(v.Type, 8))
	return " " + swatch + " " + name + " " + kind + " " + theme.Fit(theme.Faded(value), max(inner-46, 4)) + " " + uses
}

func (m *Model) row(text string, selected bool, i int) string {
	if selected {
		return theme.Selected(ansi.Strip(text))
	}
	if m.hoverY == i {
		return theme.Hovered(ansi.Strip(text))
	}
	return text
}

func (m *Model) button(id, label string) string {
	if m.hotBtn == id {
		return theme.Selected(label)
	}
	return theme.Button(label, true)
}

// scrollTo keeps the cursor row in the visible window.
func (m *Model) scrollTo(cursor, total int) {
	if cursor < m.top {
		m.top = cursor
	}
	if cursor >= m.top+m.rows {
		m.top = cursor - m.rows + 1
	}
	m.top = min(max(m.top, 0), max(total-m.rows, 0))
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// Handle implements modal.Modal.
func (m *Model) Handle(e pointer.Event) {
	if m.done {
		return
	}
	x, y := e.X-m.rect.X, e.Y-m.rect.Y
	m.hoverY, m.hotBtn = -1, ""
	buttonRow := m.rect.H - 2
	if y == buttonRow {
		m.hotBtn = m.buttonAt(x)
	}
	if y >= m.rowY && y < m.rowY+m.rows {
		m.hoverY = m.top + y - m.rowY
	}
	if e.Phase != pointer.Down || !e.Left {
		return
	}
	switch {
	case y == buttonRow:
		switch m.buttonAt(x) {
		case GoToButton:
			m.goTo()
		case RenameButton:
			m.rename()
		case doneButton:
			m.finish(modal.Outcome{Button: doneButton})
		}
	case m.hoverY >= 0:
		if m.inUses {
			if m.hoverY < len(m.vars[m.cursor].Uses) {
				m.useCursor = m.hoverY
				m.goTo()
			}
		} else if m.hoverY < len(m.vars) {
			m.cursor = m.hoverY
		}
	}
}

func (m *Model) buttonAt(x int) string {
	labels := []struct{ id, text string }{{GoToButton, "[ Go to ]"}, {RenameButton, "[ Rename… ]"}, {doneButton, "[ Done ]"}}
	total := 0
	for _, l := range labels {
		total += ansi.StringWidth(l.text) + 1
	}
	at := m.rect.W - 2 - total
	for _, l := range labels {
		w := ansi.StringWidth(l.text)
		if x >= at && x < at+w {
			return l.id
		}
		at += w + 1
	}
	return ""
}

// goTo ends the dialog on the variable's use, or lists the uses first when
// there are several.
func (m *Model) goTo() {
	if len(m.vars) == 0 {
		return
	}
	v := m.vars[m.cursor]
	switch {
	case len(v.Uses) == 0:
		return
	case m.inUses:
		m.finish(modal.Outcome{Button: GoToButton, Value: string(v.Uses[m.useCursor].Node)})
	case len(v.Uses) == 1:
		m.finish(modal.Outcome{Button: GoToButton, Value: string(v.Uses[0].Node)})
	default:
		m.inUses, m.useCursor, m.top = true, 0, 0
	}
}

func (m *Model) rename() {
	if len(m.vars) == 0 {
		return
	}
	v := m.vars[m.cursor]
	m.finish(modal.Outcome{Button: RenameButton, Value: string(v.Kind) + ":" + v.Name})
}

func (m *Model) finish(o modal.Outcome) {
	m.outcome, m.done = o, true
}

// Key implements modal.Modal.
func (m *Model) Key(_ string, _, enter, esc bool) {
	switch {
	case esc && m.inUses:
		m.inUses, m.top = false, 0
	case esc:
		m.finish(modal.Outcome{Button: doneButton, Canceled: true})
	case enter:
		m.goTo()
	}
}

// Nav implements modal.Navigator.
func (m *Model) Nav(name string) bool {
	count, cursor := len(m.vars), &m.cursor
	if m.inUses {
		count, cursor = len(m.vars[m.cursor].Uses), &m.useCursor
	}
	if count == 0 {
		return false
	}
	switch name {
	case "up":
		*cursor = max(*cursor-1, 0)
	case "down":
		*cursor = min(*cursor+1, count-1)
	case "home":
		*cursor = 0
	case "end":
		*cursor = count - 1
	case "r", "F2":
		if !m.inUses {
			m.rename()
		}
	default:
		return false
	}
	return true
}

// Outcome implements modal.Modal.
func (m *Model) Outcome() (modal.Outcome, bool) { return m.outcome, m.done }
