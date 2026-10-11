// Package stage is the centre pane: it draws the document and turns pointer
// gestures into editor commands (select, move, resize).
package stage

import (
	"github.com/meta-tui/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-tui/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-tui/cuppa/libs/canvas/editor"
	"github.com/meta-tui/cuppa/libs/canvas/hittest"
	"github.com/meta-tui/cuppa/libs/canvas/snap"
	"github.com/meta-tui/cuppa/libs/catalog/definition"
	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/document/drawlayer"
	"github.com/meta-tui/cuppa/libs/render/grid"
)

// Catalog is what the stage needs to know about components.
type Catalog interface {
	Get(id string) (definition.Definition, bool)
}

type gesture int

const (
	idle gesture = iota
	moving
	resizing
	marquee
)

// Model is the stage pane.
type Model struct {
	ed   *editor.Editor
	cat  Catalog
	w, h int
	// renderer, when set, draws the canvas instead of the designer (the preview).
	renderer func() *grid.Grid
	// offX, offY is the canvas cell shown at the pane's top-left corner.
	offX, offY int

	mode gesture
	// moving: where inside the primary node the pointer grabbed it.
	grabDX, grabDY int
	// resizing: the node, its starting rectangle, the handle and the press point.
	target         design.NodeID
	origin         design.Rect
	handle         hittest.Handle
	startX, startY int
	// marquee: the press point and current point, in canvas cells.
	marqueeFrom, marqueeTo [2]int
	marqueeAdditive        bool

	ghost *design.Rect
	// cursor is the canvas cell a keyboard user places components at; it is
	// drawn only while showCursor is set.
	cursor     [2]int
	showCursor bool
	// preview is what the drawing tool in use would draw (or erase), shown
	// over the canvas until the drag ends.
	preview      []drawlayer.Cell
	previewErase bool
	// snapOn turns alignment snapping on; guides are the lines of the last snap.
	snapOn bool
	guides []snap.Guide
}

// New returns a stage editing through ed.
func New(ed *editor.Editor, cat Catalog) *Model {
	return &Model{ed: ed, cat: cat, snapOn: true}
}

// SetRenderer replaces how the canvas is drawn (nil for the designer's own
// drawing). The preview uses it to show running components; while it is set
// the selection and drag feedback are not drawn.
func (m *Model) SetRenderer(draw func() *grid.Grid) { m.renderer = draw }

// SetSize sets the pane size in cells.
func (m *Model) SetSize(w, h int) {
	m.w, m.h = w, h
	m.clampOffset()
}

// Busy reports whether a pointer gesture started on the stage is in progress.
func (m *Model) Busy() bool { return m.mode != idle }

// Canvas converts pane coordinates to canvas cells.
func (m *Model) Canvas(x, y int) (int, int) { return x + m.offX, y + m.offY }

// Contains reports whether pane coordinates lie inside the pane.
func (m *Model) Contains(x, y int) bool { return x >= 0 && y >= 0 && x < m.w && y < m.h }

// Snap reports whether moving snaps to other components and the canvas edges.
func (m *Model) Snap() bool { return m.snapOn }

// SetSnap turns snapping on or off.
func (m *Model) SetSnap(on bool) {
	m.snapOn = on
	m.guides = nil
}

// SetGhost shows (or, with nil, hides) the outline of a component being
// dragged in from the palette. The rectangle is in canvas cells.
func (m *Model) SetGhost(r *design.Rect) { m.ghost = r }

func (m *Model) clampOffset() {
	doc := m.ed.Document()
	m.offX = min(max(m.offX, 0), max(doc.Width-m.w, 0))
	m.offY = min(max(m.offY, 0), max(doc.Height-m.h, 0))
}

// Handle processes a pointer event in pane coordinates.
func (m *Model) Handle(e pointer.Event) {
	cx, cy := m.Canvas(e.X, e.Y)
	switch e.Phase {
	case pointer.Wheel:
		if e.Shift {
			m.offX += e.WheelY * 3
		} else {
			m.offY += e.WheelY * 2
		}
		m.offX += e.WheelX * 3
		m.clampOffset()
	case pointer.Down:
		if e.Left {
			m.press(cx, cy, e.Shift)
		}
	case pointer.Move:
		if e.Held {
			m.drag(cx, cy, e.Shift, e.Alt)
		}
	case pointer.Up:
		m.release()
	}
}

func (m *Model) press(cx, cy int, additive bool) {
	doc := m.ed.Document()
	if sel := m.ed.Selected(); len(sel) == 1 {
		if n, ok := m.ed.Primary(); ok {
			if h := hittest.HandleAt(n.Rect, cx, cy); h != hittest.HandleNone {
				m.ed.Checkpoint()
				m.mode, m.target, m.origin, m.handle = resizing, n.ID, n.Rect, h
				m.startX, m.startY = cx, cy
				return
			}
		}
	}
	if id, ok := hittest.Node(doc, cx, cy); ok {
		switch {
		case additive:
			m.ed.Toggle(id)
		case !m.ed.IsSelected(id):
			m.ed.Select(id)
		}
		if !m.ed.IsSelected(id) {
			return
		}
		n, _ := m.ed.Primary()
		m.ed.Checkpoint()
		m.mode = moving
		m.grabDX, m.grabDY = cx-n.Rect.X, cy-n.Rect.Y
		m.startX, m.startY, m.origin = cx, cy, n.Rect
		return
	}
	if !additive {
		m.ed.Clear()
	}
	m.mode = marquee
	m.marqueeFrom, m.marqueeTo = [2]int{cx, cy}, [2]int{cx, cy}
	m.marqueeAdditive = additive
}

// drag follows the pointer. With shift a move keeps to one of the four
// directions, 0°, 45°, 90° and 135° as they look on screen, and a resize keeps
// the proportions; with alt a resize grows from the centre.
func (m *Model) drag(cx, cy int, shift, alt bool) {
	switch m.mode {
	case moving:
		n, ok := m.ed.Primary()
		if !ok {
			return
		}
		dx, dy := cx-m.grabDX-n.Rect.X, cy-m.grabDY-n.Rect.Y
		switch {
		case shift:
			// Measured from where the drag began, so the direction is the
			// one the whole drag points in, not the last step.
			lx, ly := snap.LockAngle(cx-m.startX, cy-m.startY)
			dx, dy = m.origin.X+lx-n.Rect.X, m.origin.Y+ly-n.Rect.Y
			m.guides = nil
		case m.snapOn:
			dx, dy = m.snapDelta(dx, dy)
		}
		m.ed.MoveSelectionBy(dx, dy)
	case resizing:
		minW, minH := 1, 1
		if n, ok := m.ed.Document().Get(m.target); ok {
			if def, known := m.cat.Get(n.Component); known {
				minW, minH = max(def.MinSize.W, 1), max(def.MinSize.H, 1)
			}
		}
		r := hittest.ResizeWith(m.origin, m.handle, cx-m.startX, cy-m.startY, minW, minH,
			hittest.ResizeModifiers{Proportional: shift, FromCenter: alt})
		m.ed.SetRect(m.target, r, false)
	case marquee:
		m.marqueeTo = [2]int{cx, cy}
	}
}

func (m *Model) release() {
	m.guides = nil
	switch m.mode {
	case moving, resizing:
		m.ed.EndGesture()
	case marquee:
		r := m.marqueeRect()
		ids := hittest.NodesIn(m.ed.Document(), r)
		if m.marqueeAdditive {
			ids = append(m.ed.Selected(), ids...)
		}
		m.ed.Select(ids...)
	}
	m.mode = idle
}

// marqueeRect is the rubber-band rectangle, normalised and at least one cell.
func (m *Model) marqueeRect() design.Rect {
	a, b := m.marqueeFrom, m.marqueeTo
	x, y := min(a[0], b[0]), min(a[1], b[1])
	return design.Rect{X: x, Y: y, W: max(a[0], b[0]) - x + 1, H: max(a[1], b[1]) - y + 1}
}

// Lines renders the pane as exactly h lines of w cells.
func (m *Model) Lines() []string {
	return theme.Block(m.render().Lines(), m.w, m.h)
}

// snapDelta adjusts a proposed move of the selection so it aligns with the
// other components or the canvas, remembering the guides to draw.
func (m *Model) snapDelta(dx, dy int) (int, int) {
	doc := m.ed.Document()
	var union design.Rect
	var others []design.Rect
	first := true
	for _, n := range doc.Nodes {
		if !m.ed.IsSelected(n.ID) {
			others = append(others, n.Rect)
			continue
		}
		if first {
			union, first = n.Rect, false
		} else {
			union = union.Union(n.Rect)
		}
	}
	proposed := union.Translate(dx, dy).MoveInto(doc.Bounds())
	res := snap.Snap(proposed, others, doc.Bounds(), snapThreshold)
	m.guides = res.Guides
	return dx + res.DX, dy + res.DY
}

// snapThreshold is how close, in cells, an edge must be to snap.
const snapThreshold = 2
