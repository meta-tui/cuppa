// Package colorpicker is the colour dialog: swatches for the 16 and 256 colour
// palettes, and RGB and HSL sliders for any colour a terminal can show in true
// colour. It is a modal box, operated by mouse; a value field takes typing.
package colorpicker

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

type tab int

const (
	tab16 tab = iota
	tab256
	tabRGB
	tabHSL
	tabThemes
)

var tabNames = []string{"16", "256", "RGB", "HSL", "Themes"}

const (
	innerW = 46
	boxW   = innerW + 2
	// bodyTop is the box row of the first line of the active tab: below the
	// border, the tab bar and a blank line.
	bodyTop = 3
)

// region is a clickable (and for sliders, draggable) area in box coordinates.
type region struct {
	y, x0, x1 int
	down      func(x int)
	drag      func(x int)
}

// Model is one open colour dialog.
type Model struct {
	title string

	tab   tab
	// slide is the slider the keyboard is on (RGB and HSL tabs).
	slide int
	// scheme is the colour scheme the Themes tab shows (an index into scheme.All).
	scheme int
	empty bool // no colour
	index int  // palette index, or -1 for a colour made with the sliders
	rgb   space.RGB
	hsl   space.HSL

	text   []rune // the value field
	// replace is true while the field still shows a value nobody has edited:
	// the first keystroke replaces it, like a selected text field.
	replace bool
	errMsg string

	sw, sh  int
	rect    design.Rect
	regions []region
	grab    *region

	outcome modal.Outcome
	done    bool
}

// New opens the dialog on the colour currently stored (a palette number, hex,
// or "" for none).
func New(title, current string) *Model {
	m := &Model{title: title, index: -1}
	m.scheme = startScheme()
	if c, idx, isIndex, ok := space.Parse(current); ok && strings.TrimSpace(current) != "" {
		if isIndex {
			m.setIndex(idx)
		} else {
			m.setRGB(c)
		}
		if isIndex && idx < 16 {
			m.tab = tab16
		} else if isIndex {
			m.tab = tab256
		} else {
			m.tab = tabRGB
		}
	} else {
		m.empty = true
		m.rgb = space.RGB{R: 255, G: 135, B: 215}
		m.hsl = space.ToHSL(m.rgb)
	}
	m.syncText()
	return m
}

// Value is the colour as a design stores it: "", "212" or "#ff87d7".
func (m *Model) Value() string {
	switch {
	case m.empty:
		return ""
	case m.index >= 0:
		return strconv.Itoa(m.index)
	}
	return m.rgb.Hex()
}

func (m *Model) setIndex(i int) {
	m.empty, m.index = false, i
	m.rgb = space.ANSI(i)
	m.hsl = space.ToHSL(m.rgb)
	m.syncText()
}

func (m *Model) setRGB(c space.RGB) {
	m.empty, m.index = false, -1
	m.rgb = c
	m.hsl = space.ToHSL(c)
	m.syncText()
}

func (m *Model) setHSL(h space.HSL) {
	m.empty, m.index = false, -1
	m.hsl = h
	m.rgb = h.RGB()
	m.syncText()
}

func (m *Model) syncText() {
	m.text = []rune(m.Value())
	m.errMsg = ""
	m.replace = true
}

// Place implements modal.Modal.
func (m *Model) Place(sw, sh int) { m.sw, m.sh = sw, sh; m.rect = m.layout() }

func (m *Model) layout() design.Rect {
	h := m.bodyHeight() + 9
	w := min(boxW, max(m.sw, 20))
	return design.Rect{X: max((m.sw-w)/2, 0), Y: max((m.sh-h)/4, 0), W: w, H: h}
}

// Rect implements modal.Modal.
func (m *Model) Rect() design.Rect { m.rect = m.layout(); return m.rect }

// Outcome implements modal.Modal.
func (m *Model) Outcome() (modal.Outcome, bool) { return m.outcome, m.done }

func (m *Model) finish(button string, canceled bool) {
	m.done = true
	m.outcome = modal.Outcome{Button: button, Value: m.Value(), Canceled: canceled}
}

// Lines implements modal.Modal.
func (m *Model) Lines() []string {
	m.rect = m.layout()
	m.regions = m.regions[:0]
	var content []string
	content = append(content, m.tabBar())
	content = append(content, "")
	content = append(content, m.body()...)
	content = append(content, "")
	content = append(content, m.footer()...)
	return theme.Panel(m.title, content, m.rect.W)
}

// row builds one content line and records its clickable spans. Coordinates are
// relative to the box: the left border is column 0 and the top border row 0.
type row struct {
	m *Model
	y int
	x int
	b strings.Builder
}

func (m *Model) newRow(y int) *row { return &row{m: m, y: y, x: 1} }

func (r *row) text(s string) *row {
	r.b.WriteString(s)
	r.x += ansi.StringWidth(s)
	return r
}

// span adds styled text and makes it clickable (down) and draggable (drag).
func (r *row) span(s string, down, drag func(x int)) *row {
	w := ansi.StringWidth(s)
	if down != nil || drag != nil {
		r.m.regions = append(r.m.regions, region{y: r.y, x0: r.x, x1: r.x + w, down: down, drag: drag})
	}
	return r.text(s)
}

func (r *row) String() string { return r.b.String() }

// tabBar draws the tabs.
func (m *Model) tabBar() string {
	r := m.newRow(1)
	r.text(" ")
	for i, name := range tabNames {
		t := tab(i)
		label := " " + name + " "
		if t == m.tab {
			label = theme.Selected(label)
		} else {
			label = theme.Dim(label)
		}
		r.span(label, func(int) { m.tab = t; m.rect = m.layout() }, nil).text(" ")
	}
	return r.String()
}

// bodyHeight is the number of rows the active tab needs.
func (m *Model) bodyHeight() int {
	switch m.tab {
	case tab16:
		return 6
	case tab256:
		return 18
	}
	return 12
}

func (m *Model) body() []string {
	var lines []string
	switch m.tab {
	case tab16:
		lines = m.system16Lines(bodyTop)
	case tab256:
		lines = m.palette256Lines(bodyTop)
	case tabRGB:
		lines = m.rgbLines(bodyTop)
	case tabThemes:
		lines = m.themesLines(bodyTop)
	default:
		lines = m.hslLines(bodyTop)
	}
	for len(lines) < m.bodyHeight() {
		lines = append(lines, "")
	}
	return lines
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
		for i := range m.regions {
			g := m.regions[i]
			if g.y == y && x >= g.x0 && x < g.x1 {
				if g.down != nil {
					g.down(x - g.x0)
				}
				if g.drag != nil {
					m.grab = &m.regions[i]
				}
				return
			}
		}
	case pointer.Move:
		if m.grab != nil && e.Held {
			m.grab.drag(x - m.grab.x0)
		}
	case pointer.Up:
		m.grab = nil
	case pointer.Wheel:
		if m.tab == tabThemes && e.WheelY != 0 {
			m.stepScheme(e.WheelY)
		}
	}
}

// Key implements modal.Modal: typing edits the value field.
func (m *Model) Key(text string, back, enter, esc bool) {
	switch {
	case esc:
		m.finish("", true)
	case enter:
		m.acceptText()
	case back:
		if m.replace {
			m.text, m.replace = nil, false
		} else if len(m.text) > 0 {
			m.text = m.text[:len(m.text)-1]
		}
		m.errMsg = ""
	case text != "":
		if m.replace {
			m.text, m.replace = nil, false
		}
		m.text = append(m.text, []rune(text)...)
		m.errMsg = ""
	}
}

// acceptText applies the value field; a valid one also closes the dialog.
func (m *Model) acceptText() {
	v, err := space.Normalise(string(m.text))
	if err != nil {
		m.errMsg = err.Error()
		return
	}
	if v == "" {
		m.empty = true
		m.finish("None", false)
		return
	}
	c, idx, isIndex, _ := space.Parse(v)
	if isIndex {
		m.setIndex(idx)
	} else {
		m.setRGB(c)
	}
	m.finish("OK", false)
}

// Describe reads the dialog out: the current colour, the tab and the buttons.
func (m *Model) Describe() []a11y.Node {
	value := string(m.text)
	if m.empty {
		value = "none"
	}
	field := a11y.Field("Colour value", value)
	field.Focused = true
	nodes := []a11y.Node{
		a11y.Heading(m.title),
		a11y.Text("Tab " + tabNames[m.tab] + ". Type a number from 0 to 255 or a hex colour like #ff5fd7."),
		field,
	}
	if m.errMsg != "" {
		nodes = append(nodes, a11y.Text(m.errMsg))
	}
	return append(nodes, a11y.Button("None"), a11y.Button("OK"), a11y.Button("Cancel"))
}
