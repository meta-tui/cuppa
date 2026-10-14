// Package panels is a stack of collapsible sections, or a bar of tabs, for
// Bubble Tea v2 programs: the pages hold content, and pressing a header shows
// or hides it, with no code of your own. It is the panel stack and the tab
// pages of Cuppa (https://github.com/meta-ui-labs/cuppa), as a component you can
// drop into any program.
//
//	p := panels.New(
//	    panels.WithPages(
//	        panels.Page{Title: "Details", Body: details},
//	        panels.Page{Title: "Layers", Body: layers},
//	    ),
//	    panels.WithOpen(0),
//	    panels.WithSize(40, 16),
//	)
//
// A Page body is a function of the room it is given, so a page can hold any
// view: a Bubbles model's View, a list of lines.
package panels

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Mode is how the pages are shown.
type Mode int

const (
	// Sections shows one header row per page; the rows left after the headers
	// are shared by the open pages.
	Sections Mode = iota
	// Tabs shows a bar of the page names and the active page under it.
	Tabs
)

// TabBarRows is the height of the tab bar: the names and a rule under them.
const TabBarRows = 2

// Page is one tab or one section.
type Page struct {
	// Title is the name on the tab or the header.
	Title string
	// Body draws the page in w by h cells. It may be nil for an empty page.
	Body func(w, h int) string
}

// ChangedMsg is sent when a press or a key changes the page that is shown.
type ChangedMsg struct {
	// Active is the tab shown (Tabs).
	Active int
	// Open are the open sections, in order (Sections).
	Open []int
}

// Rect is a rectangle of the program's screen, in cells.
type Rect struct{ X, Y, W, H int }

// Model is the panels component.
type Model struct {
	mode      Mode
	pages     []Page
	active    int
	open      map[int]bool
	exclusive bool
	// cursor is the page the keyboard is on.
	cursor              int
	w, h                int
	originX, originY    int
	accent, muted, rule string
}

// Option configures New.
type Option func(*Model)

// New returns the component, in Sections mode with no page open.
func New(opts ...Option) Model {
	m := Model{open: map[int]bool{}, w: 40, h: 12, accent: "212", muted: "245", rule: "240"}
	for _, o := range opts {
		o(&m)
	}
	return m
}

// WithMode chooses Sections or Tabs.
func WithMode(mode Mode) Option { return func(m *Model) { m.mode = mode } }

// WithPages sets the pages.
func WithPages(pages ...Page) Option { return func(m *Model) { m.pages = pages } }

// WithActive sets the tab shown.
func WithActive(i int) Option { return func(m *Model) { m.active = i } }

// WithOpen sets the open sections.
func WithOpen(pages ...int) Option {
	return func(m *Model) {
		m.open = map[int]bool{}
		for _, i := range pages {
			m.open[i] = true
		}
	}
}

// WithExclusive makes the sections an accordion: opening one closes the others.
func WithExclusive(on bool) Option { return func(m *Model) { m.exclusive = on } }

// WithSize sets the room the component takes, in cells.
func WithSize(w, h int) Option { return func(m *Model) { m.w, m.h = w, h } }

// WithOrigin tells the component where on the screen it is drawn, so mouse
// clicks line up with its headers.
func WithOrigin(x, y int) Option { return func(m *Model) { m.originX, m.originY = x, y } }

// WithColours sets the accent (the shown tab, an open header), the muted colour
// (the others) and the rule colour. Empty keeps the current one.
func WithColours(accent, muted, rule string) Option {
	return func(m *Model) {
		if accent != "" {
			m.accent = accent
		}
		if muted != "" {
			m.muted = muted
		}
		if rule != "" {
			m.rule = rule
		}
	}
}

// WithModeName chooses the mode by name: "sections" or "tabs". A design file
// carries it as a choice.
func WithModeName(name string) Option {
	return func(m *Model) {
		if name == "tabs" {
			m.mode = Tabs
		} else {
			m.mode = Sections
		}
	}
}

// WithTitles sets the pages from their names, with no content: give each its
// body with SetBody. A design file carries the names as one comma separated
// text.
func WithTitles(titles ...string) Option {
	return func(m *Model) {
		m.pages = nil
		for _, t := range titles {
			if t = strings.TrimSpace(t); t != "" {
				m.pages = append(m.pages, Page{Title: t})
			}
		}
	}
}

// WithOpenText sets the open sections from numbers in a text ("0,2"), as a
// design file writes them.
func WithOpenText(text string) Option {
	return func(m *Model) {
		m.open = map[int]bool{}
		for _, f := range strings.Split(text, ",") {
			if i, err := strconv.Atoi(strings.TrimSpace(f)); err == nil {
				m.open[i] = true
			}
		}
	}
}

// WithAccent, WithMuted and WithRule set one colour each: the shown tab or an
// open header, the other tabs or closed headers, and the rules.
func WithAccent(c string) Option { return func(m *Model) { m.accent = c } }
func WithMuted(c string) Option  { return func(m *Model) { m.muted = c } }
func WithRule(c string) Option   { return func(m *Model) { m.rule = c } }

// SetAccent, SetMuted and SetRule change the colours later.
func (m *Model) SetAccent(c string) { m.accent = c }
func (m *Model) SetMuted(c string)  { m.muted = c }
func (m *Model) SetRule(c string)   { m.rule = c }

// SetBody gives page i its content.
func (m *Model) SetBody(i int, body func(w, h int) string) {
	if i >= 0 && i < len(m.pages) {
		pages := append([]Page(nil), m.pages...)
		pages[i].Body = body
		m.pages = pages
	}
}

// SetExclusive turns the accordion behaviour on or off.
func (m *Model) SetExclusive(on bool) { m.exclusive = on }

// Exclusive reports whether the sections are an accordion.
func (m Model) Exclusive() bool { return m.exclusive }

// Mode is how the pages are shown.
func (m Model) Mode() Mode { return m.mode }

// Pages are the pages.
func (m Model) Pages() []Page { return m.pages }

// Active is the tab shown, kept inside the pages.
func (m Model) Active() int {
	return min(max(m.active, 0), max(len(m.pages)-1, 0))
}

// Open reports whether section i is open.
func (m Model) Open(i int) bool { return m.open[i] }

// OpenSections are the open sections, in order.
func (m Model) OpenSections() []int {
	var out []int
	for i := range m.pages {
		if m.open[i] {
			out = append(out, i)
		}
	}
	return out
}

// Shown reports whether a page is shown: the active tab, or an open section.
func (m Model) Shown(i int) bool {
	if m.mode == Tabs {
		return i == m.Active()
	}
	return m.open[i]
}

// SetActive shows tab i.
func (m *Model) SetActive(i int) {
	m.active = min(max(i, 0), max(len(m.pages)-1, 0))
}

// SetOpen opens or closes section i; in an accordion opening closes the others.
func (m *Model) SetOpen(i int, open bool) {
	if i < 0 || i >= len(m.pages) {
		return
	}
	if open && m.exclusive {
		m.open = map[int]bool{}
	}
	if open {
		m.open[i] = true
	} else {
		delete(m.open, i)
	}
}

// SetSize sets the room the component takes.
func (m *Model) SetSize(w, h int) { m.w, m.h = w, h }

// SetOrigin tells the component where on the screen it is drawn.
func (m *Model) SetOrigin(x, y int) { m.originX, m.originY = x, y }

// Width and Height are the room the component takes.
func (m Model) Width() int  { return m.w }
func (m Model) Height() int { return m.h }

// Geometry is where the parts are, relative to the component's top-left.
type Geometry struct {
	// Tabs are the columns each tab takes (Tabs mode).
	Tabs []Span
	// Heads are the row of each section's header (Sections mode).
	Heads []int
	// Bodies is each page's body. A closed section's body has no height; its
	// top is where its content keeps its offset from.
	Bodies []Rect
}

// Span is a run of columns.
type Span struct{ X0, X1 int }

// Layout computes the geometry for the current state. Tabs: the names on the
// first row (each as " name ", a bar between them), a rule under them, and
// every page's body is the rest. Sections: one header row per page, the rows
// left after the headers shared by the open pages, the last open one taking
// the remainder.
func (m Model) Layout() Geometry {
	var g Geometry
	n := len(m.pages)
	g.Bodies = make([]Rect, n)
	if m.mode == Tabs {
		x := 0
		for _, p := range m.pages {
			w := lipgloss.Width(p.Title) + 2
			g.Tabs = append(g.Tabs, Span{x, min(x+w, m.w)})
			x += w + 1
		}
		for i := range g.Bodies {
			g.Bodies[i] = Rect{0, TabBarRows, m.w, max(m.h-TabBarRows, 0)}
		}
		return g
	}
	open := m.OpenSections()
	avail := max(m.h-n, 0)
	y, left, opened := 0, avail, 0
	for i := range m.pages {
		g.Heads = append(g.Heads, y)
		y++
		h := 0
		if m.open[i] && len(open) > 0 {
			opened++
			h = avail / len(open)
			if opened == len(open) {
				h = left
			}
			left -= h
		}
		g.Bodies[i] = Rect{0, y, m.w, h}
		y += h
	}
	return g
}

// HeaderAt is the page whose tab or header is at (x, y) inside the component.
func (m Model) HeaderAt(x, y int) (int, bool) {
	g := m.Layout()
	if m.mode == Tabs {
		if y != 0 {
			return 0, false
		}
		for i, t := range g.Tabs {
			if x >= t.X0 && x < t.X1 {
				return i, true
			}
		}
		return 0, false
	}
	if x < 0 || x >= m.w {
		return 0, false
	}
	for i, hy := range g.Heads {
		if hy == y {
			return i, true
		}
	}
	return 0, false
}

// Press does what pressing page i's tab or header does and reports whether
// the state changed.
func (m *Model) Press(i int) bool {
	if i < 0 || i >= len(m.pages) {
		return false
	}
	if m.mode == Tabs {
		if m.Active() == i {
			return false
		}
		m.SetActive(i)
		return true
	}
	m.cursor = i
	m.SetOpen(i, !m.open[i])
	return true
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles a click on a tab or header and the keyboard: Left and Right
// change tab, Up and Down move between sections and Enter or Space opens or
// closes the one the keyboard is on.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	changed := false
	switch msg := msg.(type) {
	case tea.MouseClickMsg:
		if msg.Button != tea.MouseLeft {
			return m, nil
		}
		if i, ok := m.HeaderAt(msg.X-m.originX, msg.Y-m.originY); ok {
			changed = m.clonePress(i)
		}
	case tea.KeyPressMsg:
		last := len(m.pages) - 1
		switch msg.String() {
		case "left", "shift+tab":
			if m.mode == Tabs {
				changed = m.clonePress(max(m.Active()-1, 0))
			}
		case "right", "tab":
			if m.mode == Tabs {
				changed = m.clonePress(min(m.Active()+1, last))
			}
		case "up":
			if m.mode == Sections {
				m.cursor = max(m.cursor-1, 0)
			}
		case "down":
			if m.mode == Sections {
				m.cursor = min(m.cursor+1, last)
			}
		case "enter", "space":
			if m.mode == Sections {
				changed = m.clonePress(m.cursor)
			}
		}
	}
	if !changed {
		return m, nil
	}
	out := ChangedMsg{Active: m.Active(), Open: m.OpenSections()}
	return m, func() tea.Msg { return out }
}

// clonePress presses a page on a copy of the open set, so a Model that was
// copied before does not share it.
func (m *Model) clonePress(i int) bool {
	open := make(map[int]bool, len(m.open))
	for k, v := range m.open {
		open[k] = v
	}
	m.open = open
	return m.Press(i)
}

// View draws the component: the tab bar or the headers, and the body of each
// shown page in its place.
func (m Model) View() string {
	rows := make([]string, m.h)
	for i := range rows {
		rows[i] = strings.Repeat(" ", m.w)
	}
	g := m.Layout()
	put := func(y int, line string) {
		if y >= 0 && y < len(rows) {
			rows[y] = fit(line, m.w)
		}
	}
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color(m.accent)).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(m.muted))
	rule := lipgloss.NewStyle().Foreground(lipgloss.Color(m.rule))
	if m.mode == Tabs {
		var bar strings.Builder
		for i, t := range g.Tabs {
			if i > 0 {
				bar.WriteString(rule.Render("│"))
			}
			label := " " + m.pages[i].Title + " "
			if i == m.Active() {
				bar.WriteString(accent.Render(label))
			} else {
				bar.WriteString(muted.Render(label))
			}
			_ = t
		}
		put(0, bar.String())
		line := rule.Render(strings.Repeat("─", m.w))
		if m.Active() < len(g.Tabs) {
			t := g.Tabs[m.Active()]
			line = rule.Render(strings.Repeat("─", t.X0)) + accent.Render(strings.Repeat("━", t.X1-t.X0)) + rule.Render(strings.Repeat("─", max(m.w-t.X1, 0)))
		}
		put(1, line)
	} else {
		for i, y := range g.Heads {
			arrow, style := "▸ ", muted
			if m.open[i] {
				arrow, style = "▾ ", accent
			}
			head := style.Render(arrow+m.pages[i].Title) + " "
			fill := max(m.w-lipgloss.Width(head), 0)
			put(y, head+rule.Render(strings.Repeat("─", fill)))
		}
	}
	for i, p := range m.pages {
		b := g.Bodies[i]
		if !m.Shown(i) || b.H <= 0 || p.Body == nil {
			continue
		}
		for r, line := range strings.Split(p.Body(b.W, b.H), "\n") {
			if r >= b.H {
				break
			}
			put(b.Y+r, line)
		}
	}
	return strings.Join(rows, "\n")
}

// fit cuts or pads a line to exactly w cells.
func fit(s string, w int) string {
	if ansi.StringWidth(s) > w {
		return ansi.Truncate(s, w, "")
	}
	return s + strings.Repeat(" ", w-ansi.StringWidth(s))
}

// String is the open set as the design file writes it, e.g. "0,2".
func (m Model) String() string {
	var parts []string
	for _, i := range m.OpenSections() {
		parts = append(parts, strconv.Itoa(i))
	}
	return strings.Join(parts, ",")
}
