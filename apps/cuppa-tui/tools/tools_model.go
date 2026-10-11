// Package tools is the drawing toolbox: the list of tools in the left bar and
// the contextual bar above the canvas that shows the options of the chosen
// one, as in an image editor. Choosing a tool does not place anything; with a
// tool chosen, dragging on the canvas draws.
package tools

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/canvas/shape"
)

// Tool is one of the drawing tools.
type Tool int

// The tools, in the order of the list.
const (
	Select Tool = iota
	Rectangle
	Path
	Brush
	Erase
	Text
)

// Rows is how many lines the tool list takes in the left bar: a heading and
// one per tool.
const Rows = 7

type entry struct {
	tool Tool
	key  rune
	name string
}

var list = []entry{
	{Select, 'v', "Select"},
	{Rectangle, 'u', "Rectangle"},
	{Path, 'p', "Path"},
	{Brush, 'b', "Brush"},
	{Erase, 'e', "Erase"},
	{Text, 't', "Text"},
}

// Options are what the contextual bar edits. Each tool uses the ones it shows.
type Options struct {
	// Fill and Stroke colour a rectangle: its inside and its border.
	Fill, Stroke string
	// Corners is shape.CornersSquare, CornersRound or CornersAngled.
	Corners string
	// LineChar is what a path draws with (shape.AutoChar picks the line
	// character for each step) and BrushChar what the brush draws with.
	LineChar, BrushChar string
	Thickness           int
	// Color is the colour of paths, brush strokes and text.
	Color string
	// End is how a path finishes: shape.EndNone, EndArrow or EndCircle.
	End string
}

// ColorPicker opens the colour dialog for one option; the shell provides it.
type ColorPicker func(title, current string, apply func(color string))

// Model is the toolbox.
type Model struct {
	tool Tool
	// Opts are the options of every tool; the bar edits them in place.
	Opts      Options
	pickColor ColorPicker

	// bar holds the clickable spans of the contextual bar, in bar columns.
	bar []span
}

type span struct {
	x0, x1 int
	act    func()
}

// New returns the toolbox with the select tool chosen.
func New() *Model {
	return &Model{Opts: Options{
		Stroke: "212", Corners: shape.CornersSquare,
		LineChar: shape.AutoChar, BrushChar: "█", Thickness: 1, Color: "212", End: shape.EndNone,
	}}
}

// BindColorPicker sets how a colour swatch of the bar opens the colour dialog.
func (m *Model) BindColorPicker(p ColorPicker) { m.pickColor = p }

// Tool is the chosen tool.
func (m *Model) Tool() Tool { return m.tool }

// Drawing reports whether a drawing tool (anything but select) is chosen.
func (m *Model) Drawing() bool { return m.tool != Select }

// Choose picks a tool.
func (m *Model) Choose(t Tool) { m.tool = t }

// ChooseKey picks the tool whose shortcut is r (V, U, P, B, E, T, either
// case) and reports whether there is one.
func (m *Model) ChooseKey(r rune) bool {
	r = []rune(strings.ToLower(string(r)))[0]
	for _, e := range list {
		if e.key == r {
			m.tool = e.tool
			return true
		}
	}
	return false
}

// Name is the tool's name.
func (t Tool) Name() string {
	for _, e := range list {
		if e.tool == t {
			return e.name
		}
	}
	return ""
}

// Key is the tool's shortcut letter, in capitals.
func (t Tool) Key() string {
	for _, e := range list {
		if e.tool == t {
			return strings.ToUpper(string(e.key))
		}
	}
	return ""
}

// Lines renders the tool list, Rows lines of w cells.
func (m *Model) Lines(w int, focused bool) []string {
	title := theme.Title(" TOOLS")
	if focused {
		title = theme.Selected(" TOOLS ")
	}
	lines := []string{title}
	for _, e := range list {
		text := fmt.Sprintf(" %s  %s", strings.ToUpper(string(e.key)), e.name)
		text = theme.Fit(text, w)
		if e.tool == m.tool {
			text = theme.Selected(text)
		}
		lines = append(lines, text)
	}
	return theme.Block(lines, w, Rows)
}

// HandleList takes a pointer event in the coordinates of the tool list. It
// reports whether a tool was chosen.
func (m *Model) HandleList(e pointer.Event) bool {
	if e.Phase != pointer.Down || !e.Left || e.Y < 1 || e.Y > len(list) {
		return false
	}
	m.tool = list[e.Y-1].tool
	return true
}

// ---- the contextual bar ---------------------------------------------------

// Instructions is the line the select tool shows.
const Instructions = "Click a component to select it, drag to move it, drag a corner to resize it. " +
	"Choose a drawing tool (U P B E T) to draw; V comes back here."

// EraseNote is the line the erase tool shows.
const EraseNote = "Erase only works on the drawing. Components are not touched."

// barBuilder collects the bar's text and its clickable spans.
type barBuilder struct {
	b     strings.Builder
	x     int
	spans []span
}

func (b *barBuilder) text(s string) {
	b.b.WriteString(s)
	b.x += ansi.StringWidth(s)
}

func (b *barBuilder) button(s string, act func()) {
	w := ansi.StringWidth(s)
	b.spans = append(b.spans, span{b.x, b.x + w, act})
	b.text(s)
}

// colour is a swatch and its value; clicking it opens the colour dialog.
func (m *Model) colour(b *barBuilder, label, value string, set func(string)) {
	b.text(theme.Dim(label + " "))
	sw := theme.Faded("--")
	if value != "" {
		sw = lipgloss.NewStyle().Background(lipgloss.Color(value)).Render("  ")
	}
	shown := value
	if shown == "" {
		shown = "none"
	}
	b.button(sw+" "+shown, func() {
		if m.pickColor != nil {
			m.pickColor(label, value, set)
		}
	})
	b.text("  ")
}

// stepper is a label with ◂ value ▸ over a list of values.
func (m *Model) stepper(b *barBuilder, label string, values []string, current string, set func(string)) {
	at := 0
	for i, v := range values {
		if v == current {
			at = i
		}
	}
	step := func(d int) func() {
		return func() { set(values[(at+d+len(values))%len(values)]) }
	}
	b.text(theme.Dim(label + " "))
	b.button(theme.Button("◂", true), step(-1))
	b.text(" " + theme.Bold(current) + " ")
	b.button(theme.Button("▸", true), step(1))
	b.text("  ")
}

// thickness is a [-] n [+] counter from 1 to 8.
func (m *Model) thickness(b *barBuilder) {
	b.text(theme.Dim("Thickness "))
	b.button(theme.Button("[-]", true), func() { m.Opts.Thickness = max(m.Opts.Thickness-1, 1) })
	b.text(" " + theme.Bold(fmt.Sprint(m.Opts.Thickness)) + " ")
	b.button(theme.Button("[+]", true), func() { m.Opts.Thickness = min(m.Opts.Thickness+1, 8) })
	b.text("  ")
}

// BarLine renders the contextual bar as one line of w cells and records what
// is clickable.
func (m *Model) BarLine(w int) string {
	b := &barBuilder{}
	b.text(theme.Title(" " + m.tool.Name() + " "))
	switch m.tool {
	case Select:
		b.text(theme.Dim(Instructions))
	case Rectangle:
		m.colour(b, "Fill", m.Opts.Fill, func(v string) { m.Opts.Fill = v })
		m.colour(b, "Border", m.Opts.Stroke, func(v string) { m.Opts.Stroke = v })
		b.text(theme.Dim("Corners "))
		glyphs := map[string]string{shape.CornersRound: " ╭ ", shape.CornersSquare: " ┌ ", shape.CornersAngled: " ╱ "}
		for _, c := range []string{shape.CornersRound, shape.CornersSquare, shape.CornersAngled} {
			c := c
			styled := theme.Button(glyphs[c], true)
			if m.Opts.Corners == c {
				styled = theme.Selected(glyphs[c])
			}
			b.button(styled, func() { m.Opts.Corners = c })
		}
		b.text(" " + theme.Dim(m.Opts.Corners))
	case Path:
		m.stepper(b, "Line", lineChars, m.Opts.LineChar, func(v string) { m.Opts.LineChar = v })
		m.stepper(b, "End", []string{shape.EndNone, shape.EndArrow, shape.EndCircle}, m.Opts.End, func(v string) { m.Opts.End = v })
		m.thickness(b)
		m.colour(b, "Colour", m.Opts.Color, func(v string) { m.Opts.Color = v })
	case Brush:
		m.stepper(b, "Char", brushChars, m.Opts.BrushChar, func(v string) { m.Opts.BrushChar = v })
		m.thickness(b)
		m.colour(b, "Colour", m.Opts.Color, func(v string) { m.Opts.Color = v })
	case Erase:
		m.thickness(b)
		b.text(theme.Dim(EraseNote))
	case Text:
		m.colour(b, "Colour", m.Opts.Color, func(v string) { m.Opts.Color = v })
		b.text(theme.Dim("Click the canvas, then type."))
	}
	m.bar = b.spans
	return theme.Fit(b.b.String(), w)
}

// Characters a path and a brush can draw with.
var (
	lineChars  = []string{shape.AutoChar, "─", "═", "━", "·", "*", "#", "█", "░"}
	brushChars = []string{"█", "▓", "▒", "░", "#", "*", "+", "·", "●", "~"}
)

// HandleBar takes a pointer event in the coordinates of the bar and reports
// whether it pressed something.
func (m *Model) HandleBar(e pointer.Event) bool {
	if e.Phase != pointer.Down || !e.Left || e.Y != 0 {
		return false
	}
	for _, s := range m.bar {
		if e.X >= s.x0 && e.X < s.x1 {
			s.act()
			return true
		}
	}
	return false
}

// LineOptions are the path options.
func (m *Model) LineOptions() shape.LineOptions {
	return shape.LineOptions{Char: m.Opts.LineChar, Thickness: m.Opts.Thickness, Color: m.Opts.Color, End: m.Opts.End}
}

// BrushOptions are the brush options.
func (m *Model) BrushOptions() shape.BrushOptions {
	return shape.BrushOptions{Char: m.Opts.BrushChar, Thickness: m.Opts.Thickness, Color: m.Opts.Color}
}

// EraseOptions are the eraser's: a solid block of the thickness.
func (m *Model) EraseOptions() shape.BrushOptions {
	return shape.BrushOptions{Char: "█", Thickness: m.Opts.Thickness}
}
