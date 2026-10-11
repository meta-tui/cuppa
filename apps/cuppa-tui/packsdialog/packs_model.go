// Package packsdialog is the Packs box: every component pack with a checkbox
// that switches it on or off in the palette, a Remove button on installed
// packs and an Add button for new .cupp files.
package packsdialog

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const (
	width = 60
	// removeW is the width of the "[Remove]" button, with its leading space.
	removeW = 9
	// AddButton and RemoveButton are the Outcome.Button values for those actions;
	// a removal carries the pack id in Outcome.Value.
	AddButton    = "Add"
	RemoveButton = "Remove"
	doneButton   = "Done"
)

// Entry is one pack with how many components it holds.
type Entry struct {
	Pack  definition.Pack
	Count int
}

// Model is the open dialog. Switching a pack calls toggle at once, so the
// palette behind it follows; the box ends with Done, Add or Remove.
type Model struct {
	entries []Entry
	enabled func(definition.Family) bool
	toggle  func(definition.Family)

	// cursor is the pack the keyboard is on, shown once it has moved.
	cursor  int
	moved   bool
	rect    design.Rect
	hoverY  int
	hotBtn  string
	outcome modal.Outcome
	done    bool
}

// New returns the dialog. enabled reads the current state of a pack.
func New(entries []Entry, enabled func(definition.Family) bool, toggle func(definition.Family)) *Model {
	return &Model{entries: entries, enabled: enabled, toggle: toggle, hoverY: -1}
}

// Place implements modal.Modal.
func (m *Model) Place(sw, sh int) {
	w := min(width, sw)
	h := len(m.entries)*3 + 5
	m.rect = design.Rect{X: max((sw-w)/2, 0), Y: max((sh-h)/3, 0), W: w, H: h}
}

// Rect implements modal.Modal.
func (m *Model) Rect() design.Rect { return m.rect }

// rowTop is the y of an entry's first line, inside the box.
func rowTop(i int) int { return 2 + i*3 }

// Lines implements modal.Modal.
func (m *Model) Lines() []string {
	inner := m.rect.W - 4
	content := []string{""}
	for i, e := range m.entries {
		mark := "[ ]"
		if m.enabled(e.Pack.ID) {
			mark = "[x]"
		}
		tail := theme.Faded(countText(e.Count))
		if e.Pack.Source != "" {
			tail += " " + theme.Button("[Remove]", true)
		}
		head := theme.Fit(fmt.Sprintf(" %s %s", mark, e.Pack.Name), max(inner-ansi.StringWidth(tail)-1, 1))
		head += " " + tail
		if m.hoverY == rowTop(i) || m.moved && m.hoverY < 0 && i == m.cursor {
			head = theme.Hovered(ansi.Strip(head))
		}
		content = append(content, " "+head, "      "+theme.Faded(theme.Fit(e.Pack.Description, inner-6)), "")
	}
	add, done := m.button(AddButton, "[ Add pack… ]"), m.button(doneButton, "[ Done ]")
	row := add + " " + done
	pad := max(m.rect.W-2-ansi.StringWidth(row)-1, 0)
	content = append(content, spaces(pad)+row)
	return theme.Panel("Component packs", content, m.rect.W)
}

func (m *Model) button(id, label string) string {
	if m.hotBtn == id {
		return theme.Selected(label)
	}
	return theme.Button(label, true)
}

func countText(n int) string {
	if n == 1 {
		return "1 component"
	}
	return fmt.Sprintf("%d components", n)
}

func spaces(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = ' '
	}
	return string(out)
}

// buttonAt says which bottom button the box column x is over.
func (m *Model) buttonAt(x int) string {
	doneW, addW := ansi.StringWidth("[ Done ]"), ansi.StringWidth("[ Add pack… ]")
	doneX := m.rect.W - 2 - doneW
	addX := doneX - 1 - addW
	switch {
	case x >= doneX && x < doneX+doneW:
		return doneButton
	case x >= addX && x < addX+addW:
		return AddButton
	}
	return ""
}

// Handle implements modal.Modal.
func (m *Model) Handle(e pointer.Event) {
	if m.done {
		return
	}
	x, y := e.X-m.rect.X, e.Y-m.rect.Y
	m.hoverY, m.hotBtn = -1, ""
	press := e.Phase == pointer.Down && e.Left
	if i, ok := m.entryAt(y); ok && x > 0 && x < m.rect.W-1 {
		pack := m.entries[i].Pack
		if pack.Source != "" && y == rowTop(i) && x >= m.rect.W-1-removeW {
			if press {
				m.outcome, m.done = modal.Outcome{Button: RemoveButton, Value: string(pack.ID)}, true
			}
			return
		}
		m.hoverY = rowTop(i)
		if press {
			m.toggle(pack.ID)
		}
		return
	}
	if y == m.rect.H-2 {
		m.hotBtn = m.buttonAt(x)
		if press && m.hotBtn != "" {
			m.outcome, m.done = modal.Outcome{Button: m.hotBtn}, true
		}
	}
}

func (m *Model) entryAt(y int) (int, bool) {
	for i := range m.entries {
		if y >= rowTop(i) && y < rowTop(i)+2 {
			return i, true
		}
	}
	return 0, false
}

// Key implements modal.Modal.
func (m *Model) Key(_ string, _, enter, esc bool) {
	if enter || esc {
		m.outcome, m.done = modal.Outcome{Button: doneButton, Canceled: esc}, true
	}
}

// Nav implements modal.Navigator: Up and Down move through the packs, Space
// switches the pack on or off, Delete or r removes an installed pack, a adds
// one. Enter is Done.
func (m *Model) Nav(name string) bool {
	if len(m.entries) == 0 && name != "a" {
		return false
	}
	switch name {
	case "up":
		m.cursor = (m.cursor + len(m.entries) - 1) % len(m.entries)
	case "down":
		m.cursor = (m.cursor + 1) % len(m.entries)
	case "space":
		m.toggle(m.entries[m.cursor].Pack.ID)
	case "delete", "r":
		pack := m.entries[m.cursor].Pack
		if pack.Source == "" {
			return true
		}
		m.outcome, m.done = modal.Outcome{Button: RemoveButton, Value: string(pack.ID)}, true
	case "a":
		m.outcome, m.done = modal.Outcome{Button: AddButton}, true
	default:
		return false
	}
	m.moved = true
	return true
}

// Outcome implements modal.Modal.
func (m *Model) Outcome() (modal.Outcome, bool) { return m.outcome, m.done }

// Describe reads the dialog out: each pack, whether it is on, and the buttons.
func (m *Model) Describe() []a11y.Node {
	var items []a11y.Node
	for _, e := range m.entries {
		state := "off"
		if m.enabled(e.Pack.ID) {
			state = "on"
		}
		label := fmt.Sprintf("%s, %s, %s, %s", e.Pack.Name, state, countText(e.Count), e.Pack.Description)
		if e.Pack.Source != "" {
			label += ", installed, can be removed"
		}
		items = append(items, a11y.Item(label, false))
	}
	return []a11y.Node{a11y.Heading("Component packs"), a11y.List("Packs", items...), a11y.Button("Add pack"), a11y.Button("Done")}
}
