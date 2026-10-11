package shell

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/prompt"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/tools"
	"github.com/meta-ui-labs/cuppa/libs/canvas/shape"
	"github.com/meta-ui-labs/cuppa/libs/canvas/snap"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

// gesture is a drawing drag in progress.
type gesture struct {
	active bool
	start  [2]int
	cur    [2]int
	// points are the cells the pointer passed over, for the brush and the eraser.
	points [][2]int
	shift  bool
}

// canvasCell is the canvas cell under a screen position, kept on the canvas.
func (m *Model) canvasCell(x, y int) (int, int) {
	r := m.layout.stage
	cx, cy := m.stg.Canvas(x-r.X, y-r.Y)
	doc := m.ed.Document()
	return min(max(cx, 0), max(doc.Width-1, 0)), min(max(cy, 0), max(doc.Height-1, 0))
}

// drawMouse handles a pointer event on the canvas while a drawing tool is
// chosen: a drag draws, a click with the text tool asks for the text.
func (m *Model) drawMouse(e pointer.Event) {
	cx, cy := m.canvasCell(e.X, e.Y)
	switch e.Phase {
	case pointer.Down:
		if !e.Left {
			return
		}
		if m.tb.Tool() == tools.Text {
			m.askText(cx, cy)
			return
		}
		m.gest = gesture{active: true, start: [2]int{cx, cy}, cur: [2]int{cx, cy}, points: [][2]int{{cx, cy}}, shift: e.Shift}
		m.showPreview()
	case pointer.Move:
		if !m.gest.active || !e.Held {
			return
		}
		m.gest.cur, m.gest.shift = [2]int{cx, cy}, e.Shift
		if last := m.gest.points[len(m.gest.points)-1]; last != m.gest.cur {
			m.gest.points = append(m.gest.points, m.gest.cur)
		}
		m.showPreview()
	case pointer.Up:
		if !m.gest.active {
			return
		}
		m.gest.cur, m.gest.shift = [2]int{cx, cy}, e.Shift
		m.commitGesture()
	}
}

// cancelGesture drops a drawing drag without drawing it.
func (m *Model) cancelGesture() {
	m.gest = gesture{}
	m.stg.SetPreview(nil, false)
}

// gestureCells are the cells the drag would draw, and whether they erase.
func (m *Model) gestureCells() (cells []drawlayer.Cell, erase bool) {
	g := m.gest
	o := m.tb.Opts
	switch m.tb.Tool() {
	case tools.Rectangle:
		x1, y1 := g.cur[0], g.cur[1]
		if g.shift {
			// A square as it looks on screen: twice as many columns as rows.
			dx, dy := x1-g.start[0], y1-g.start[1]
			side := max(abs(dx), 2*abs(dy))
			x1 = g.start[0] + sign(dx)*side
			y1 = g.start[1] + sign(dy)*side/2
			if dx == 0 {
				x1 = g.start[0] + side
			}
			if dy == 0 {
				y1 = g.start[1] + side/2
			}
		}
		return shape.Rectangle(g.start[0], g.start[1], x1, y1, o.Corners, o.Stroke, o.Fill), false
	case tools.Path:
		x1, y1 := g.cur[0], g.cur[1]
		if g.shift {
			dx, dy := snap.LockAngle(x1-g.start[0], y1-g.start[1])
			x1, y1 = g.start[0]+dx, g.start[1]+dy
		}
		return shape.Line(g.start[0], g.start[1], x1, y1, m.tb.LineOptions()), false
	case tools.Brush:
		return shape.Stroke(g.points, m.tb.BrushOptions()), false
	case tools.Erase:
		return shape.Stroke(g.points, m.tb.EraseOptions()), true
	}
	return nil, false
}

func (m *Model) showPreview() {
	cells, erase := m.gestureCells()
	m.stg.SetPreview(cells, erase)
}

// commitGesture paints (or erases) what the drag drew, as one undo step.
func (m *Model) commitGesture() {
	cells, erase := m.gestureCells()
	m.cancelGesture()
	m.paintCells(cells, erase)
}

// paintCells puts cells into the drawing, kept on the canvas, and says what
// happened.
func (m *Model) paintCells(cells []drawlayer.Cell, erase bool) {
	doc := m.ed.Document()
	var on []drawlayer.Cell
	for _, c := range cells {
		if c.X >= 0 && c.Y >= 0 && c.X < doc.Width && c.Y < doc.Height {
			on = append(on, c)
		}
	}
	if len(on) == 0 {
		return
	}
	changed := m.ed.ApplyDrawing(func(l *drawlayer.Layer) {
		if erase {
			l.Erase(on...)
		} else {
			l.Paint(on...)
		}
	})
	if !changed {
		if n, ok := m.drawingLocked(); ok && n {
			m.say("Cannot draw, the drawing is locked")
		}
		return
	}
	verb := "Drew"
	if erase {
		verb = "Erased"
	}
	m.say(fmt.Sprintf("%s %d cells", verb, len(on)))
}

// drawingLocked reports whether there is a drawing and it is locked.
func (m *Model) drawingLocked() (locked, has bool) {
	for _, n := range m.ed.Document().Nodes {
		if n.Component == drawlayer.Component {
			return n.Locked, true
		}
	}
	return false, false
}

// askText opens the prompt for the text tool and draws what is typed with its
// first letter at the clicked cell.
func (m *Model) askText(cx, cy int) {
	m.flow.Show(prompt.New("Text", "Text to draw", ""), func(o modal.Outcome) {
		if o.Canceled || o.Value == "" {
			return
		}
		m.paintCells(shape.Text(cx, cy, o.Value, m.tb.Opts.Color), false)
	})
}

// toolKey chooses a tool from its letter (V, U, P, B, E, T) when no other
// modifier is held, and reports whether it did.
func (m *Model) toolKey(k tea.Key) bool {
	if k.Mod&^tea.ModShift != 0 || k.Text == "" || len([]rune(k.Text)) != 1 {
		return false
	}
	if !m.tb.ChooseKey([]rune(k.Text)[0]) {
		return false
	}
	m.cancelGesture()
	m.say(m.tb.Tool().Name() + " tool")
	return true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}
