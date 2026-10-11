// Package inspector is the right bar: details of the selection (position,
// size, layer order, properties) and the layer list.
package inspector

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/meta-tui/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-tui/cuppa/libs/canvas/editor"
	"github.com/meta-tui/cuppa/libs/catalog/definition"
	"github.com/meta-tui/cuppa/libs/document/design"
)

// Catalog is what the inspector needs to know about components.
type Catalog interface {
	Get(id string) (definition.Definition, bool)
}

// region is a clickable span of one content line.
type region struct {
	y, x0, x1 int
	act       func()
	// step changes the value by one in a direction (-1 or +1), for Left and
	// Right on the keyboard stop; mouseOnly spans are not keyboard stops.
	step      func(delta int)
	mouseOnly bool
	// label is the stop as it is read out: the text before it on its line and
	// its own text.
	label string
	// layer is set on a layer name: pressing it can start a drag of that layer.
	layer design.NodeID
	// header is set on a panel's header: a click opens or closes the panel and
	// a drag moves the edge above it.
	header panelID
}

// Model is the inspector pane.
type Model struct {
	ed      *editor.Editor
	cat     Catalog
	w, h    int
	focused bool
	regions []region
	// stop is the keyboard stop that has the focus and stops how many there
	// were at the last render.
	stop, stops int

	editing string // field being typed into: "x", "y", "w", "h", "name" or "prop:<key>"
	buf     string
	message string

	// snap reads and changes the stage setting this pane exposes as a checkbox.
	snapGet func() bool
	snapSet func(bool)

	// pickColor opens the colour dialog; nil means colours are edited as text.
	pickColor ColorPicker
	// pickTheme opens the dialog that fills the theme from a colour scheme.
	pickTheme ThemePicker
	// ask opens a question with buttons; nil means the answer is always the
	// first button.
	ask Asker

	// panels is what each panel remembers, views where they were drawn at the
	// last render, press the header being held, and lastSel the layer that was
	// selected then (a new selection is scrolled into view in the layer list).
	panels  map[panelID]*panelState
	views   []panelView
	press   *headerPress
	lastSel design.NodeID
	// menu is the dropdown that is open, if any.
	menu *dropdown

	// The layer list: a drag in progress, how long the list is (set at render),
	// and where the trash button is (columns; its row is the Layers header).
	drag             *layerDrag
	layerCount       int
	trashX0, trashX1 int
}

// New returns an inspector editing through ed.
func New(ed *editor.Editor, cat Catalog) *Model { return &Model{ed: ed, cat: cat} }

// BindSnap connects the "snap to guides" checkbox to the stage setting.
func (m *Model) BindSnap(get func() bool, set func(bool)) { m.snapGet, m.snapSet = get, set }

// SetSize sets the pane size in cells.
func (m *Model) SetSize(w, h int) { m.w, m.h = w, h }

// ColorPicker opens a colour dialog for a property: it shows title with the
// current colour selected, and calls apply with the chosen one ("" for none).
// The shell provides it.
type ColorPicker func(title, current string, apply func(color string))

// ThemePicker opens the dialog that fills the theme from a colour scheme and
// calls apply with the background and the theme colours it chose. The shell
// provides it.
type ThemePicker func(apply func(background string, t design.Theme))

// Asker opens a question box with buttons and calls then with the label of
// the one pressed. The shell provides it.
type Asker func(title, body string, buttons []string, then func(button string))

// BindAsker sets how the inspector asks the person a question, such as
// whether renaming a variable renames it everywhere.
func (m *Model) BindAsker(a Asker) { m.ask = a }

// BindThemePicker sets how the Theme section's scheme button opens the dialog.
// Without it the button is not shown.
func (m *Model) BindThemePicker(p ThemePicker) { m.pickTheme = p }

// BindColorPicker sets how clicking a colour opens the picker. Without it a
// click edits the colour as text.
func (m *Model) BindColorPicker(p ColorPicker) { m.pickColor = p }

// Editing reports whether a field is taking keyboard input.
// An open dropdown takes the keys too: Up and Down move through its options,
// Enter chooses and Esc closes it.
func (m *Model) Editing() bool { return m.editing != "" || m.menu != nil }

// Handle processes a pointer event in pane coordinates.
func (m *Model) Handle(e pointer.Event) {
	if m.menu != nil {
		m.handleMenu(e)
		return
	}
	switch e.Phase {
	case pointer.Wheel:
		// The wheel scrolls the panel under the pointer, not the whole bar.
		if v, ok := m.viewAt(e.Y); ok {
			m.scrollPanel(v, e.WheelY*2)
		}
	case pointer.Down:
		if !e.Left {
			return
		}
		cy, ok := m.contentY(e.Y)
		if ok {
			for _, r := range m.regions {
				if r.y == cy && e.X >= r.x0 && e.X < r.x1 {
					m.cancelEdit()
					switch {
					case r.layer != "":
						// Selecting changes what the bar shows above the list, which
						// would slide the list out from under a drag; so a layer is
						// selected on release, when it was not dragged.
						m.pressLayer(r.layer, e.Y, r.act)
					case r.header != "":
						m.pressHeader(r.header, e.Y, r.act)
					default:
						r.act()
					}
					return
				}
			}
		}
		m.cancelEdit()
	case pointer.Move:
		switch {
		case m.press != nil && e.Held:
			m.dragHeader(e)
		case m.drag != nil && e.Held:
			m.dragTo(e)
		}
	case pointer.Up:
		if m.press != nil {
			m.releaseHeader()
		}
		if m.drag != nil {
			m.drop(e)
		}
	}
}

// Key handles typing into the field being edited.
func (m *Model) Key(text string, backspace, enter, escape bool) {
	if m.menu != nil {
		switch {
		case escape:
			m.menu = nil
		case enter || text == " ":
			m.chooseMenu(m.menu.at)
		}
		return
	}
	switch {
	case escape:
		m.cancelEdit()
	case enter:
		m.commit()
	case backspace:
		if r := []rune(m.buf); len(r) > 0 {
			m.buf = string(r[:len(r)-1])
		}
	default:
		m.buf += text
	}
}

func (m *Model) startEdit(field, initial string) {
	m.editing, m.buf, m.message = field, initial, ""
}

func (m *Model) cancelEdit() { m.editing, m.buf = "", "" }

func (m *Model) commit() {
	field, value := m.editing, m.buf
	m.cancelEdit()
	m.apply(field, value)
}

// Step adds delta to the number being edited and applies it at once, so the
// design follows the arrow keys; Enter then ends the edit. It reports whether
// the field holds a number.
func (m *Model) Step(delta int) bool {
	if m.menu != nil {
		// Up (a positive step) goes to the option above.
		if delta > 0 {
			m.menu.move(-1)
		} else if delta < 0 {
			m.menu.move(1)
		}
		return true
	}
	if m.editing == "" || m.editing == "background" || m.editing == "name" {
		return false
	}
	n, err := strconv.Atoi(strings.TrimSpace(m.buf))
	if err != nil {
		return false
	}
	m.buf = strconv.Itoa(n + delta)
	m.apply(m.editing, m.buf)
	return true
}

// apply sets the field to value. The field is "background", "canvas-…",
// "name", "prop:key" or one of the geometry fields.
func (m *Model) apply(field, value string) {
	switch {
	case field == "background":
		m.report(m.ed.SetBackground(value))
		return
	case field == "swatch-new":
		m.report(m.ed.SetSwatch(value, "#ffffff"))
		return
	case strings.HasPrefix(field, "swatch:"):
		m.report(m.ed.SetSwatch(strings.TrimPrefix(field, "swatch:"), value))
		return
	case field == "keys":
		m.report(m.ed.SetKeys(value))
		return
	case strings.HasPrefix(field, "theme-"):
		m.report(m.ed.SetThemeColor(strings.TrimPrefix(field, "theme-"), value))
		return
	case strings.HasPrefix(field, "canvas-"):
		m.commitCanvas(field, value)
		return
	}
	n, ok := m.ed.Primary()
	if !ok {
		return
	}
	switch {
	case field == "name":
		m.ed.Rename(n.ID, value)
	case field == "showif":
		m.changeVariable(editor.VarInput, n.ShowIf, value, editor.Use{Node: n.ID, Role: editor.UseShowIf})
	case field == "event":
		m.changeVariable(editor.VarEvent, n.Event, value, editor.Use{Node: n.ID, Role: editor.UseEvent})
	case strings.HasPrefix(field, "bind:"):
		key := strings.TrimPrefix(field, "bind:")
		m.changeVariable(editor.VarInput, n.Bind[key], value, editor.Use{Node: n.ID, Prop: key, Role: editor.UseProp})
	case strings.HasPrefix(field, "swatchname:"):
		key := strings.TrimPrefix(field, "swatchname:")
		current, _ := strings.CutPrefix(n.Props[key], design.TokenPrefix)
		m.changeVariable(editor.VarColour, current, value, editor.Use{Node: n.ID, Prop: key, Role: editor.UseProp})
	case strings.HasPrefix(field, "prop:"):
		key := strings.TrimPrefix(field, "prop:")
		if def, ok := m.cat.Get(n.Component); ok {
			if spec, ok := def.Prop(key); ok && spec.Kind == definition.PropColor && !strings.HasPrefix(strings.TrimSpace(value), design.TokenPrefix) {
				m.report(m.ed.SetColour(n.ID, key, value))
				return
			}
		}
		m.report(m.ed.SetProp(n.ID, key, value))
	default:
		value = strings.TrimSpace(value)
		v, err := strconv.Atoi(value)
		switch {
		case err != nil:
			// Not a plain number: an expression such as 50% or 100% - 10.
			m.report(m.ed.SetLayout(n.ID, field, value))
		case layoutOf(m.ed.Layout(n.ID), field) != "":
			// A number typed over an expression makes the axis fixed.
			m.report(m.ed.SetLayout(n.ID, field, value))
		default:
			m.setGeometry(n, field, v)
		}
	}
}

func (m *Model) report(err error) {
	m.message = ""
	if err != nil {
		m.message = strings.TrimPrefix(err.Error(), "editor: ")
	}
}

func (m *Model) setGeometry(n design.Node, field string, v int) {
	r := n.Rect
	switch field {
	case "x":
		r.X = v
	case "y":
		r.Y = v
	case "w":
		r.W = v
	case "h":
		r.H = v
	}
	m.ed.SetRect(n.ID, r, true)
}

func (m *Model) nudge(field string, delta int) {
	n, ok := m.ed.Primary()
	if !ok {
		return
	}
	switch field {
	case "x":
		m.setGeometry(n, field, n.Rect.X+delta)
	case "y":
		m.setGeometry(n, field, n.Rect.Y+delta)
	case "w":
		m.setGeometry(n, field, n.Rect.W+delta)
	case "h":
		m.setGeometry(n, field, n.Rect.H+delta)
	}
}

func layoutOf(l design.Layout, field string) string {
	switch field {
	case "x":
		return l.X
	case "y":
		return l.Y
	case "w":
		return l.W
	case "h":
		return l.H
	}
	return ""
}

// changeVariable sets what a component's use of a variable is called. When the
// variable is used elsewhere too, the person is asked whether the new name
// renames the variable everywhere or only makes this use point to another
// variable.
func (m *Model) changeVariable(kind editor.VariableKind, current, to string, use editor.Use) {
	to = design.CleanInputName(to)
	relink := func() {
		if kind == editor.VarColour && use.Role == editor.UseProp {
			if _, ok := m.ed.Document().Theme.Colour(to); !ok && to != "" {
				// Another variable that does not exist yet: it starts with the
				// colour this one has.
				value, _ := m.ed.Document().Theme.Colour(current)
				if err := m.ed.SetSwatch(to, value); err != nil {
					m.report(err)
					return
				}
			}
		}
		m.report(m.ed.Relink(use, kind, to))
	}
	if to == "" || to == current {
		if to == "" && current != "" && use.Role == editor.UseProp && kind == editor.VarInput {
			m.report(m.ed.SetBinding(use.Node, use.Prop, ""))
		} else if to == "" && use.Role == editor.UseShowIf {
			m.report(m.ed.SetShowIf(use.Node, ""))
		} else if to == "" && use.Role == editor.UseEvent {
			m.report(m.ed.SetEvent(use.Node, ""))
		}
		return
	}
	var uses int
	exists := false
	for _, v := range m.ed.Variables() {
		if v.Kind != kind {
			continue
		}
		if v.Name == current {
			uses = len(v.Uses)
		}
		if v.Name == to {
			exists = true
		}
	}
	switch {
	case current == "" || exists:
		relink()
	case uses <= 1:
		m.report(m.ed.RenameVariable(kind, current, to))
	case m.ask == nil:
		relink()
	default:
		body := fmt.Sprintf("%q is used in %d places.\nRename it to %q everywhere, or make only this one use %q?", current, uses, to, to)
		m.ask("Rename variable", body, []string{"Rename everywhere", "Only this one", "Cancel"}, func(button string) {
			switch button {
			case "Rename everywhere":
				m.report(m.ed.RenameVariable(kind, current, to))
			case "Only this one":
				relink()
			}
		})
	}
}
