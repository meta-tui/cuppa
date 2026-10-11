package inspector

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
)

// menuRows is the most options a dropdown shows at once; more scroll.
const menuRows = 8

// menuItem is one option of a dropdown.
type menuItem struct {
	label string
	// current marks the option that is the value now.
	current bool
	act     func()
}

// dropdown is an open list of options under a control. It is drawn over the
// panels, so a long list is not cut by the panel the control is in.
type dropdown struct {
	title string
	// y is the content line of the control and x its column.
	y, x  int
	items []menuItem
	at    int
	// scroll is the first option shown.
	scroll int
	// box is where it was drawn at the last render, in bar cells.
	boxX, boxY, boxW, boxH int
}

// here is the content line and column the next span will start at.
func (b *builder) here() (y, x int) { return len(b.lines), b.x }

// openMenu opens a dropdown for the control at content line y, column x.
func (m *Model) openMenu(title string, y, x int, items []menuItem) {
	if len(items) == 0 {
		return
	}
	d := &dropdown{title: title, y: y, x: x, items: items}
	for i, it := range items {
		if it.current {
			d.at = i
		}
	}
	d.keepInView()
	m.cancelEdit()
	m.menu = d
}

// selectRow adds a control that shows the current value and opens a dropdown
// of the options. step, when set, lets Left and Right on the keyboard go to
// the previous or next option without opening it.
func (m *Model) selectRow(b *builder, shown, title string, items []menuItem, step func(delta int)) {
	y, x := b.here()
	open := func() { m.openMenu(title, y, x, items) }
	label := theme.Button("["+shown+" ▾]", true)
	if step != nil {
		b.addStep(label, open, step)
		return
	}
	b.add(label, open)
}

func (d *dropdown) keepInView() {
	if d.at < d.scroll {
		d.scroll = d.at
	}
	if d.at >= d.scroll+menuRows {
		d.scroll = d.at - menuRows + 1
	}
	d.scroll = min(max(d.scroll, 0), max(len(d.items)-menuRows, 0))
}

func (d *dropdown) move(delta int) {
	d.at = min(max(d.at+delta, 0), len(d.items)-1)
	d.keepInView()
}

// choose closes the list and runs the option at index i.
func (m *Model) chooseMenu(i int) {
	d := m.menu
	m.menu = nil
	if d != nil && i >= 0 && i < len(d.items) && d.items[i].act != nil {
		d.items[i].act()
	}
}

// CloseMenu closes an open dropdown without choosing.
func (m *Model) CloseMenu() { m.menu = nil }

// handleMenu takes a pointer event while a dropdown is open. A click on an
// option chooses it; a click anywhere else closes the list and does nothing
// more.
func (m *Model) handleMenu(e pointer.Event) {
	d := m.menu
	inside := e.X >= d.boxX && e.X < d.boxX+d.boxW && e.Y > d.boxY && e.Y < d.boxY+d.boxH-1
	row := d.scroll + e.Y - d.boxY - 1
	switch e.Phase {
	case pointer.Wheel:
		d.scroll = min(max(d.scroll+e.WheelY, 0), max(len(d.items)-menuRows, 0))
	case pointer.Move:
		if inside && row < len(d.items) {
			d.at = row
		}
	case pointer.Down:
		switch {
		case !e.Left:
		case inside && row < len(d.items):
			m.chooseMenu(row)
		default:
			m.menu = nil
		}
	}
}

// drawMenu overlays the open dropdown on the bar's lines.
func (m *Model) drawMenu(out []string) {
	d := m.menu
	y, ok := m.paneY(d.y)
	if !ok {
		m.menu = nil
		return
	}
	rows := min(len(d.items), menuRows)
	width := ansi.StringWidth(d.title) + 4
	for _, it := range d.items {
		width = max(width, ansi.StringWidth(it.label)+6)
	}
	width = min(width, m.w)
	x := min(d.x, max(m.w-width, 0))
	top := y + 1
	if top+rows+2 > len(out) && y-rows-2 >= 0 {
		top = y - rows - 2
	}
	body := make([]string, rows)
	for r := range body {
		i := d.scroll + r
		it := d.items[i]
		mark := "  "
		if it.current {
			mark = "• "
		}
		line := theme.Fit(mark+it.label, width-2)
		if i == d.at {
			line = theme.Selected(ansi.Strip(line))
		}
		body[r] = line
	}
	d.boxX, d.boxY, d.boxW, d.boxH = x, top, width, rows+2
	for i, l := range theme.Panel(strings.TrimSpace(d.title), body, width) {
		if row := top + i; row >= 0 && row < len(out) {
			out[row] = theme.Overlay(out[row], l, x)
		}
	}
}
