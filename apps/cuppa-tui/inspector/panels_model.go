package inspector

import "github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"

// panelID names a panel of the bar. The state of a panel (open or closed, how
// tall, how far it is scrolled) belongs to the name, not to the selection, so
// choosing another component does not move anything.
type panelID string

// The panels of the bar, in the order they can appear.
const (
	panelProperties panelID = "properties"
	panelBehaviour  panelID = "behaviour"
	panelCanvas     panelID = "canvas"
	panelTheme      panelID = "theme"
	panelColours    panelID = "colours"
	panelKeys       panelID = "keys"
	panelDisplay    panelID = "display"
	panelInputs     panelID = "inputs"
	panelPages      panelID = "pages"
	panelLayers     panelID = "layers"
)

// panelState is what a panel remembers.
type panelState struct {
	closed bool
	// want is the body height the person gave it (or the default).
	want   int
	scroll int
}

// panelDefault is how a panel starts: how tall, whether it is closed, and the
// fewest body rows it can be given.
type panelDefault struct {
	want, min int
	closed    bool
}

var panelDefaults = map[panelID]panelDefault{
	panelProperties: {want: 12, min: 3},
	panelBehaviour:  {want: 6, min: 2},
	panelCanvas:     {want: 5, min: 2},
	panelTheme:      {want: 8, min: 3},
	panelColours:    {want: 6, min: 2, closed: true},
	panelKeys:       {want: 3, min: 2, closed: true},
	panelDisplay:    {want: 8, min: 3},
	panelInputs:     {want: 4, min: 2},
	panelPages:      {want: 4, min: 2},
	panelLayers:     {want: 8, min: 3},
}

// panelSpec is one panel to show: its title and what fills its body.
type panelSpec struct {
	id    panelID
	title string
	fill  func(b *builder)
	// trash puts the delete button in the header (the layer list).
	trash bool
}

// panelView is where a panel was drawn at the last render.
type panelView struct {
	id panelID
	// start is the content line of the header; the body lines follow it.
	start int
	// total is how many body lines there are, shown how many rows in h.
	total, h int
	// top is the bar row of the header.
	top  int
	open bool
}

// fixedRows are the rows above the panels: the title and the history buttons.
const fixedRows = 2

func (m *Model) state(id panelID) *panelState {
	if m.panels == nil {
		m.panels = map[panelID]*panelState{}
	}
	st, ok := m.panels[id]
	if !ok {
		d := panelDefaults[id]
		st = &panelState{want: max(d.want, 1), closed: d.closed}
		m.panels[id] = st
	}
	return st
}

// content is the width panels draw into: the last column is the scroll bar.
func (m *Model) content() int { return max(m.w-1, 1) }

// reflow gives each view its header row and body height for the pane size.
func (m *Model) reflow() {
	n := len(m.views)
	wants, mins, open := make([]int, n), make([]int, n), make([]bool, n)
	for i, v := range m.views {
		wants[i] = m.state(v.id).want
		mins[i] = panelDefaults[v.id].min
		open[i] = v.open
	}
	h := panelHeights(max(m.h-fixedRows-n, 0), wants, mins, open)
	top := fixedRows
	for i := range m.views {
		m.views[i].top, m.views[i].h = top, h[i]
		top += 1 + h[i]
	}
}

// bodyRow is the content line shown on bar row y of a view's body, and whether
// there is one.
func (m *Model) bodyRow(v panelView, y int) (int, bool) {
	row := y - v.top - 1
	if row < 0 || row >= v.h {
		return 0, false
	}
	line := m.state(v.id).scroll + row
	if line >= v.total {
		return 0, false
	}
	return v.start + 1 + line, true
}

// contentY is the content line shown on bar row y: the header and history rows
// as they are, a panel's header, or the body line the panel has scrolled to.
func (m *Model) contentY(y int) (int, bool) {
	if y < 0 {
		return 0, false
	}
	if y < fixedRows {
		return y, true
	}
	for _, v := range m.views {
		switch {
		case y == v.top:
			return v.start, true
		case y > v.top && y <= v.top+v.h:
			return m.bodyRow(v, y)
		}
	}
	return 0, false
}

// paneY is the bar row on which content line cy is visible.
func (m *Model) paneY(cy int) (int, bool) {
	if cy < fixedRows {
		return cy, true
	}
	for _, v := range m.views {
		if cy == v.start {
			return v.top, true
		}
		if cy > v.start && cy <= v.start+v.total {
			row := cy - v.start - 1 - m.state(v.id).scroll
			if row >= 0 && row < v.h {
				return v.top + 1 + row, true
			}
			return 0, false
		}
	}
	return 0, false
}

// viewAt is the panel whose header or body is on bar row y.
func (m *Model) viewAt(y int) (panelView, bool) {
	for _, v := range m.views {
		if y >= v.top && y <= v.top+v.h {
			return v, true
		}
	}
	return panelView{}, false
}

// scrollPanel scrolls a panel's body by delta lines.
func (m *Model) scrollPanel(v panelView, delta int) {
	st := m.state(v.id)
	st.scroll = min(max(st.scroll+delta, 0), max(v.total-v.h, 0))
}

// toggle opens a closed panel and closes an open one.
func (m *Model) toggle(id panelID, closed bool) { m.state(id).closed = closed }

// headerPress is a panel header held down: letting go without moving opens or
// closes the panel, moving the pointer drags the edge between this panel and
// the one above it.
type headerPress struct {
	id     panelID
	startY int
	// lastY is the bar row the edge is at now.
	lastY int
	moved bool
	// flip is what a plain click does; it runs on release.
	flip func()
}

func (m *Model) pressHeader(id panelID, y int, flip func()) {
	m.press = &headerPress{id: id, startY: y, lastY: y, flip: flip}
}

// dragHeader moves the edge above the pressed header with the pointer.
func (m *Model) dragHeader(e pointer.Event) {
	p := m.press
	if e.Y != p.startY {
		p.moved = true
	}
	if !p.moved {
		return
	}
	if applied := m.resizeAbove(p.id, e.Y-p.lastY); applied != 0 {
		p.lastY += applied
	}
}

// releaseHeader ends the press: a click that did not move flips the panel.
func (m *Model) releaseHeader() {
	p := m.press
	m.press = nil
	if p != nil && !p.moved && p.flip != nil {
		p.flip()
	}
}

// resizeAbove moves the edge above panel id by delta rows (down is positive):
// the nearest open panel above it grows and the nearest open one at or below
// it shrinks by as much. It returns how far the edge moved.
func (m *Model) resizeAbove(id panelID, delta int) int {
	at := -1
	for i, v := range m.views {
		if v.id == id {
			at = i
		}
	}
	if at < 0 || delta == 0 {
		return 0
	}
	above, below := -1, -1
	for i := at - 1; i >= 0; i-- {
		if m.views[i].open {
			above = i
			break
		}
	}
	for i := at; i < len(m.views); i++ {
		if m.views[i].open {
			below = i
			break
		}
	}
	if above < 0 || below < 0 {
		return 0
	}
	room := func(i int) int { return m.views[i].h - panelDefaults[m.views[i].id].min }
	switch {
	case delta > 0:
		delta = min(delta, max(room(below), 0))
	default:
		delta = -min(-delta, max(room(above), 0))
	}
	if delta == 0 {
		return 0
	}
	// Every open panel keeps the height it has now, apart from the two at the
	// edge, so nothing else moves.
	for _, v := range m.views {
		if v.open {
			m.state(v.id).want = v.h
		}
	}
	m.state(m.views[above].id).want += delta
	m.state(m.views[below].id).want -= delta
	m.reflow()
	return delta
}

// reveal scrolls the panel that holds content line y so the line is in view.
func (m *Model) reveal(y int) {
	for _, v := range m.views {
		if y <= v.start || y > v.start+v.total {
			continue
		}
		st, row := m.state(v.id), y-v.start-1
		if row < st.scroll {
			st.scroll = row
		}
		if v.h > 0 && row >= st.scroll+v.h {
			st.scroll = row - v.h + 1
		}
		return
	}
}
