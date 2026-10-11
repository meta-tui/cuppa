// Package pagesbar is the bar of pages under the canvas, as the artboard
// navigator of an illustration program: the page being edited, arrows to the
// neighbouring pages, a list of all of them and a button that adds one
// (ADR 0014).
package pagesbar

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
)

// Asker asks for a line of text (a page's new name) and calls then with it.
// The shell provides it.
type Asker func(title, initial string, then func(value string))

// Reporter tells the person something went wrong (a name that is taken).
type Reporter func(title, body string)

// item is one row of the list.
type item struct {
	label   string
	current bool
	// separator draws a rule and does nothing; with a heading it draws the
	// heading of a group of pages instead.
	separator bool
	heading   string
	disabled  bool
	act       func()
}

type span struct {
	x0, x1 int
	act    func()
}

// Model is the pages bar.
type Model struct {
	ed     *editor.Editor
	ask    Asker
	report Reporter

	// open says the list is shown; at is the row the keyboard is on.
	open  bool
	at    int
	items []item
	// spans are the clickable parts of the bar, in bar columns.
	spans []span
	// box is where the list was drawn, in screen cells.
	boxX, boxY, boxW, boxH int
}

// New returns the bar for an editor.
func New(ed *editor.Editor) *Model { return &Model{ed: ed} }

// Bind sets how the bar asks for a name and reports a refusal.
func (m *Model) Bind(ask Asker, report Reporter) { m.ask, m.report = ask, report }

// Open reports whether the list is shown.
func (m *Model) Open() bool { return m.open }

// Close hides the list.
func (m *Model) Close() { m.open = false }

// Line renders the bar as one line of w cells and records what is clickable.
// While the design is previewing the bar is shown but idle.
func (m *Model) Line(w int, idle bool) string {
	m.spans = m.spans[:0]
	names := m.ed.PageNames()
	at := m.ed.ActivePage()
	var b strings.Builder
	x := 0
	text := func(s string) {
		b.WriteString(s)
		x += ansi.StringWidth(s)
	}
	button := func(s string, enabled bool, act func()) {
		width := ansi.StringWidth(s)
		if enabled && !idle {
			m.spans = append(m.spans, span{x, x + width, act})
		}
		b.WriteString(theme.Button(s, enabled && !idle))
		x += width
	}
	text(" ")
	button("◂", at > 0, func() { m.ed.SwitchPage(at - 1) })
	text(" ")
	name := names[at]
	label := fmt.Sprintf("%s  %d/%d ▾", ansi.Truncate(name, 24, "…"), at+1, len(names))
	if idle {
		text(theme.Dim(label))
	} else {
		width := ansi.StringWidth(label)
		m.spans = append(m.spans, span{x, x + width, m.toggle})
		text(theme.Bold(label))
	}
	text(" ")
	button("▸", at < len(names)-1, func() { m.ed.SwitchPage(at + 1) })
	text("  ")
	button("[+ Page]", true, m.add)
	return theme.Fit(b.String(), w)
}

func (m *Model) add() {
	m.open = false
	if err := m.ed.AddPage(""); err != nil && m.report != nil {
		m.report("Cannot add page", err.Error())
	}
}

func (m *Model) toggle() {
	if m.open {
		m.open = false
		return
	}
	m.build()
	m.open = true
}

// build makes the rows of the list: the pages, then what can be done to the
// one being edited.
func (m *Model) build() {
	names := m.ed.PageNames()
	cur := m.ed.ActivePage()
	m.items = m.items[:0]
	m.at = 0
	add := func(i int) {
		page := i
		if i == cur {
			m.at = len(m.items)
		}
		m.items = append(m.items, item{label: names[i], current: i == cur, act: func() { m.ed.SwitchPage(page) }})
	}
	// Pages that have a folder are listed under it, so a design of many pages
	// is as easy to read as the folders its export makes.
	grouped := false
	for i := range names {
		if m.ed.PageFolder(i) != "" {
			grouped = true
		}
	}
	if !grouped {
		for i := range names {
			add(i)
		}
	} else {
		var order []string
		members := map[string][]int{}
		for i := range names {
			g := "pages"
			if f := m.ed.PageFolder(i); f != "" {
				g, _, _ = strings.Cut(f, "/")
			}
			if _, ok := members[g]; !ok {
				order = append(order, g)
			}
			members[g] = append(members[g], i)
		}
		for _, g := range order {
			m.items = append(m.items, item{separator: true, heading: g + "/"})
			for _, i := range members[g] {
				add(i)
			}
		}
	}
	m.items = append(m.items,
		item{separator: true},
		item{label: "Add page", act: m.add},
		item{label: "Rename…", act: func() { m.rename(cur) }},
		item{label: "Move left", disabled: cur == 0, act: func() { m.ed.MovePage(cur, cur-1) }},
		item{label: "Move right", disabled: cur >= len(names)-1, act: func() { m.ed.MovePage(cur, cur+1) }},
		item{label: "Delete page", disabled: len(names) < 2, act: func() {
			if err := m.ed.DeletePage(cur); err != nil && m.report != nil {
				m.report("Cannot delete page", err.Error())
			}
		}},
		item{label: "Folder…", act: func() { m.folder(cur) }},
	)
}

func (m *Model) rename(i int) {
	if m.ask == nil {
		return
	}
	names := m.ed.PageNames()
	if i < 0 || i >= len(names) {
		return
	}
	m.ask("Rename page", names[i], func(value string) {
		if err := m.ed.RenamePage(i, value); err != nil && m.report != nil {
			m.report("Cannot rename page", strings.TrimPrefix(err.Error(), "editor: "))
		}
	})
}

// folder asks for the folder an export keeps the page in.
func (m *Model) folder(i int) {
	if m.ask == nil {
		return
	}
	title := "Folder of the page (empty: pages/<page>)"
	seen := map[string]bool{}
	var known []string
	for j := range m.ed.PageNames() {
		if f := m.ed.PageFolder(j); f != "" && !seen[f] {
			seen[f] = true
			known = append(known, f)
		}
	}
	if len(known) > 0 {
		title += ". In use: " + strings.Join(known, ", ")
	}
	m.ask(title, m.ed.PageFolder(i), func(value string) {
		if err := m.ed.SetPageFolder(i, value); err != nil && m.report != nil {
			m.report("Cannot set the folder", strings.TrimPrefix(err.Error(), "editor: "))
		}
	})
}

// HandleBar takes a pointer event in the bar's own coordinates and reports
// whether it pressed something.
func (m *Model) HandleBar(e pointer.Event) bool {
	if e.Phase != pointer.Down || !e.Left || e.Y != 0 {
		return false
	}
	for _, s := range m.spans {
		if e.X >= s.x0 && e.X < s.x1 {
			s.act()
			return true
		}
	}
	return false
}

// Overlay is the list as lines to draw over the screen, and where: its
// bottom edge is the row above the bar (barY), as the list of an artboard
// navigator opens upward. It is nil while the list is closed.
func (m *Model) Overlay(barX, barY int) (x, y int, lines []string) {
	if !m.open {
		return 0, 0, nil
	}
	width := 12
	for _, it := range m.items {
		width = max(width, ansi.StringWidth(it.label)+6, ansi.StringWidth(it.heading)+2)
	}
	body := make([]string, len(m.items))
	for i, it := range m.items {
		switch {
		case it.separator:
			body[i] = theme.Faded(strings.Repeat("─", width))
		default:
			mark := "  "
			if it.current {
				mark = "• "
			}
			line := theme.Fit(mark+it.label, width)
			switch {
			case it.disabled:
				line = theme.Faded(ansi.Strip(line))
			case i == m.at:
				line = theme.Selected(ansi.Strip(line))
			}
			body[i] = line
		}
	}
	lines = theme.Panel("", body, width+2)
	x, y = barX+1, max(barY-len(lines), 1)
	m.boxX, m.boxY, m.boxW, m.boxH = x, y, width+2, len(lines)
	return x, y, lines
}

// HandleMenu takes a pointer event in screen coordinates while the list is
// open. A press on a row does it; a press anywhere else closes the list. It
// reports whether the event was used.
func (m *Model) HandleMenu(e pointer.Event) bool {
	if !m.open {
		return false
	}
	inside := e.X >= m.boxX && e.X < m.boxX+m.boxW && e.Y > m.boxY && e.Y < m.boxY+m.boxH-1
	row := e.Y - m.boxY - 1
	switch e.Phase {
	case pointer.Move:
		if inside && row < len(m.items) && !m.items[row].separator {
			m.at = row
		}
	case pointer.Down:
		switch {
		case !e.Left:
		case inside && row >= 0 && row < len(m.items):
			m.choose(row)
		default:
			m.open = false
		}
	}
	return true
}

func (m *Model) choose(i int) {
	it := m.items[i]
	if it.separator || it.disabled || it.act == nil {
		return
	}
	m.open = false
	it.act()
}

// Key takes a key while the list is open: Up and Down move through the rows,
// Enter does one, Esc closes. It reports whether the key was used.
func (m *Model) Key(name string) bool {
	if !m.open {
		return false
	}
	step := func(d int) {
		for i := 0; i < len(m.items); i++ {
			m.at = ((m.at+d)%len(m.items) + len(m.items)) % len(m.items)
			if !m.items[m.at].separator {
				return
			}
		}
	}
	switch name {
	case "up":
		step(-1)
	case "down":
		step(1)
	case "enter", "space":
		m.choose(m.at)
	case "esc", "escape":
		m.open = false
	}
	return true
}
