package stage

import (
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/container"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func stripANSI(s string) string { return ansi.Strip(s) }

func setup(t *testing.T) (*Model, *editor.Editor, design.NodeID) {
	t.Helper()
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 100, 40))
	id, err := ed.Add("lipgloss.box", 10, 5) // 24x6 at (10,5)
	if err != nil {
		t.Fatal(err)
	}
	ed.Clear()
	m := New(ed, cat)
	m.SetSize(60, 20)
	return m, ed, id
}

func down(m *Model, x, y int) { m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Down, Left: true}) }
func drag(m *Model, x, y int) { m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Move, Held: true}) }
func up(m *Model, x, y int)   { m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Up, Left: true}) }
func rect(ed *editor.Editor, id design.NodeID) design.Rect {
	n, _ := ed.Document().Get(id)
	return n.Rect
}

func TestClickSelectsAndEmptyClickClears(t *testing.T) {
	m, ed, id := setup(t)
	down(m, 15, 8)
	up(m, 15, 8)
	if sel := ed.Selected(); len(sel) != 1 || sel[0] != id {
		t.Fatalf("selection = %v", sel)
	}
	down(m, 50, 18)
	up(m, 50, 18)
	if len(ed.Selected()) != 0 {
		t.Fatal("clicking empty canvas should deselect")
	}
}

func TestDragMovesByThePointerDelta(t *testing.T) {
	m, ed, id := setup(t)
	down(m, 15, 8) // inner cell, grabbed 5,3 from the top-left
	drag(m, 20, 10)
	drag(m, 25, 12)
	up(m, 25, 12)
	if got := rect(ed, id); got.X != 20 || got.Y != 9 {
		t.Fatalf("rect = %+v", got)
	}
	ed.Undo()
	if got := rect(ed, id); got.X != 10 || got.Y != 5 {
		t.Fatalf("drag should undo as one step: %+v", got)
	}
}

func TestPlainClickLeavesNoUndoStep(t *testing.T) {
	m, ed, _ := setup(t)
	ed.MarkSaved()
	down(m, 15, 8)
	up(m, 15, 8)
	if ed.Dirty() {
		t.Fatal("clicking changed the document")
	}
}

func TestDragCornerResizes(t *testing.T) {
	m, ed, id := setup(t)
	down(m, 15, 8)
	up(m, 15, 8)    // select
	down(m, 33, 10) // bottom-right corner cell of (10,5,24,6)
	drag(m, 38, 13)
	up(m, 38, 13)
	if got := rect(ed, id); got.W != 29 || got.H != 9 || got.X != 10 || got.Y != 5 {
		t.Fatalf("rect = %+v", got)
	}
}

func TestMarqueeSelectsIntersectingNodes(t *testing.T) {
	m, ed, a := setup(t)
	b, _ := ed.Add("lipgloss.label", 40, 15)
	ed.Clear()
	down(m, 5, 2)
	drag(m, 45, 16)
	up(m, 45, 16)
	sel := ed.Selected()
	if len(sel) != 2 {
		t.Fatalf("selection = %v (want %v and %v)", sel, a, b)
	}
}

func TestShiftClickTogglesSelection(t *testing.T) {
	m, ed, a := setup(t)
	b, _ := ed.Add("lipgloss.label", 40, 15)
	ed.Select(a)
	m.Handle(pointer.Event{X: 41, Y: 15, Phase: pointer.Down, Left: true, Shift: true})
	m.Handle(pointer.Event{X: 41, Y: 15, Phase: pointer.Up, Left: true})
	if sel := ed.Selected(); len(sel) != 2 {
		t.Fatalf("selection = %v", sel)
	}
	m.Handle(pointer.Event{X: 41, Y: 15, Phase: pointer.Down, Left: true, Shift: true})
	m.Handle(pointer.Event{X: 41, Y: 15, Phase: pointer.Up, Left: true})
	if ed.IsSelected(b) {
		t.Fatal("second shift-click should deselect")
	}
}

func TestWheelScrollsTheViewport(t *testing.T) {
	m, _, _ := setup(t)
	m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: 1})
	if m.offY != 2 {
		t.Fatalf("offY = %d", m.offY)
	}
	for i := 0; i < 50; i++ {
		m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: 1})
	}
	if m.offY != 20 { // doc height 40 - pane height 20
		t.Fatalf("offY not clamped: %d", m.offY)
	}
	if x, y := m.Canvas(0, 0); x != 0 || y != 20 {
		t.Fatalf("Canvas(0,0) = %d,%d", x, y)
	}
}

func TestRenderIsExactlyPaneSized(t *testing.T) {
	m, ed, id := setup(t)
	ed.Select(id)
	g := design.Rect{X: 3, Y: 3, W: 5, H: 2}
	m.SetGhost(&g)
	lines := m.Lines()
	if len(lines) != 20 {
		t.Fatalf("lines = %d", len(lines))
	}
	if !strings.Contains(strings.Join(lines, ""), "■") {
		t.Fatal("selection handles not drawn")
	}
}

func TestDraggingSnapsToANeighbourAndShowsAGuide(t *testing.T) {
	m, ed, id := setup(t) // box at (10,5) 24x6
	other, _ := ed.Add("lipgloss.label", 60, 20)
	ed.Clear()
	// Move the box so its left edge lands one cell off the label's left edge (60).
	down(m, 15, 8)
	drag(m, 15+49, 8) // box.X -> 59
	if got := rect(ed, id); got.X != 60 {
		t.Fatalf("X = %d, want 60 (snapped to the label at %v)", got.X, other)
	}
	if len(m.guides) == 0 {
		t.Fatal("no guide recorded while snapped")
	}
	up(m, 15+49, 8)
	if len(m.guides) != 0 {
		t.Fatal("guides should clear on release")
	}
}

func TestSnapCanBeSwitchedOff(t *testing.T) {
	m, ed, id := setup(t)
	if _, err := ed.Add("lipgloss.label", 60, 20); err != nil {
		t.Fatal(err)
	}
	ed.Clear()
	m.SetSnap(false)
	down(m, 15, 8)
	drag(m, 15+49, 8)
	up(m, 15+49, 8)
	if got := rect(ed, id); got.X != 59 {
		t.Fatalf("X = %d, want the unsnapped 59", got.X)
	}
}

func TestMultipleSelectionIsMarkedEvenWithoutBorders(t *testing.T) {
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 100, 40))
	a, _ := ed.Add("lipgloss.list", 5, 3)
	b, _ := ed.Add("lipgloss.list", 30, 12)
	ed.Select(a, b)
	m := New(ed, cat)
	m.SetSize(60, 25)
	screen := stripANSI(strings.Join(m.Lines(), "\n"))
	if got := strings.Count(screen, "□"); got != 8 {
		t.Fatalf("expected 4 hollow corner marks per selected component, got %d:\n%s", got, screen)
	}
	if strings.Contains(screen, "■") {
		t.Fatal("solid squares are resize handles and belong to a single selection only")
	}
	if !strings.Contains(screen, "╌") || !strings.Contains(screen, "╎") {
		t.Fatalf("the group needs a frame:\n%s", screen)
	}
}

func TestSingleSelectionKeepsItsResizeHandles(t *testing.T) {
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 100, 40))
	a, _ := ed.Add("lipgloss.list", 5, 3)
	ed.Select(a)
	m := New(ed, cat)
	m.SetSize(60, 25)
	screen := stripANSI(strings.Join(m.Lines(), "\n"))
	if strings.Count(screen, "■") != 4 || strings.Contains(screen, "□") || strings.Contains(screen, "╌") {
		t.Fatalf("single selection should show four solid handles only:\n%s", screen)
	}
}

func TestALightTerminalPreviewPaintsPaleCanvasAndDarkDefaultText(t *testing.T) {
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 60, 12))
	id, _ := ed.Add("lipgloss.label", 2, 2)
	ed.Clear()
	_ = id
	m := New(ed, cat)
	m.SetSize(60, 12)
	dark := m.render()
	ed.SetLight(true)
	light := m.render()
	if dark.At(30, 8).Bg == light.At(30, 8).Bg {
		t.Fatal("the empty canvas changes colour")
	}
	if light.At(30, 8).Bg != lightCanvas {
		t.Fatalf("empty canvas = %q", light.At(30, 8).Bg)
	}
	c := light.At(2, 2)
	if c.Fg == "" || c.Bg == "" {
		t.Fatalf("text drawn in the terminal's own colours needs real ones on a light canvas: %+v", c)
	}
}

// While editing, a press on a tab selects the container and leaves the page
// alone: the person is editing a component, not using it. Pages change in the
// preview (or from the Properties list).
func TestPressingATabWhileEditingDoesNotChangeThePage(t *testing.T) {
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 80, 24))
	id, err := ed.Add("cuppa.tabs", 5, 3)
	if err != nil {
		t.Fatal(err)
	}
	ed.Clear()
	m := New(ed, cat)
	m.SetSize(80, 24)
	n, _ := ed.Document().Get(id)
	tab := container.Layout(n).Tabs[1]
	x, y := tab.X0+1-m.offX, n.Rect.Y-m.offY
	down(m, x, y)
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Up, Left: true})
	got, _ := ed.Document().Get(id)
	if container.Active(got) != 0 {
		t.Fatalf("active = %d: editing must not change the page", container.Active(got))
	}
	if !ed.IsSelected(id) {
		t.Error("the press should select the container")
	}
}
