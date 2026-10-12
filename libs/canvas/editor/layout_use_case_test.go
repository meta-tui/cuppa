package editor

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// topBar is a box that is the full width and as short as a box can be.
func topBar(t *testing.T, e *Editor) design.NodeID {
	t.Helper()
	id := mustAdd(t, e, "lipgloss.box", 0, 0)
	for axis, v := range map[string]string{AxisX: "0", AxisY: "0", AxisW: "100%", AxisH: "3"} {
		if err := e.SetLayout(id, axis, v); err != nil {
			t.Fatal(err)
		}
	}
	return id
}

func TestLayoutExpressionsFollowTheCanvas(t *testing.T) {
	e := newEditor()
	id := topBar(t, e)
	if got := rectOf(e, id); got != (design.Rect{X: 0, Y: 0, W: 80, H: 3}) {
		t.Fatalf("at 80x24: %+v", got)
	}
	if err := e.SetCanvasSize(120, 40); err != nil {
		t.Fatal(err)
	}
	if got := rectOf(e, id); got != (design.Rect{X: 0, Y: 0, W: 120, H: 3}) {
		t.Fatalf("at 120x40: %+v", got)
	}
	side := mustAdd(t, e, "lipgloss.box", 0, 2)
	must(t, e.SetLayout(side, AxisH, "100% - 3"))
	must(t, e.SetLayout(side, AxisW, "30"))
	if got := rectOf(e, side); got.H != 37 || got.W != 30 {
		t.Fatalf("sidebar: %+v", got)
	}
	must(t, e.SetCanvasSize(80, 24))
	if got := rectOf(e, side); got.H != 21 {
		t.Fatalf("sidebar at 80x24: %+v", got)
	}
}

func TestLayoutRefusesBadExpressionsAndLockedNodes(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.box", 0, 0)
	if err := e.SetLayout(id, AxisW, "100% -"); err == nil {
		t.Fatal("bad expression accepted")
	}
	if err := e.SetLayout(id, "z", "1"); err == nil {
		t.Fatal("bad axis accepted")
	}
	e.SetLocked(id, true)
	if err := e.SetLayout(id, AxisW, "50%"); err == nil {
		t.Fatal("locked node accepted a layout")
	}
	if e.Layout(id) != (design.Layout{}) {
		t.Fatal("layout changed")
	}
}

func TestLayoutIsOneUndoStepAndClearingFixesTheAxis(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.box", 0, 0)
	must(t, e.SetLayout(id, AxisW, "50%"))
	if rectOf(e, id).W != 40 {
		t.Fatalf("w = %d", rectOf(e, id).W)
	}
	e.Undo()
	if got := rectOf(e, id).W; got != 24 || e.Layout(id).W != "" {
		t.Fatalf("after undo: w=%d layout=%+v", got, e.Layout(id))
	}
	e.Redo()
	must(t, e.SetLayout(id, AxisW, ""))
	must(t, e.SetCanvasSize(120, 24))
	if got := rectOf(e, id).W; got != 40 {
		t.Fatalf("fixed axis moved: %d", got)
	}
}

func TestDraggingKeepsTheUnit(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.box", 0, 0)
	must(t, e.SetLayout(id, AxisW, "100% - 10"))
	must(t, e.SetLayout(id, AxisX, "0"))
	e.Checkpoint()
	e.Select(id)
	e.SetRect(id, design.Rect{X: 0, Y: 0, W: 60, H: 6}, false)
	if got := e.Layout(id).W; got != "100% - 20" {
		t.Fatalf("w = %q, want 100%% - 20", got)
	}
	must(t, e.SetCanvasSize(120, 24))
	if got := rectOf(e, id).W; got != 100 {
		t.Fatalf("after resize w = %d", got)
	}
	e.MoveSelectionBy(5, 0)
	if got := rectOf(e, id).X; got != 5 || e.Layout(id).X != "" {
		t.Fatalf("x = %d (%q): a fixed axis just moves", got, e.Layout(id).X)
	}
}

func TestLoadResolvesLayouts(t *testing.T) {
	doc := design.NewDocument("t", 100, 30)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box 1", Rect: design.Rect{X: 0, Y: 0, W: 5, H: 5}, Layout: design.Layout{W: "50%"}})
	e := New(newEditor().cat, doc)
	if got := e.Document().Nodes[0].Rect.W; got != 50 {
		t.Fatalf("w = %d", got)
	}
	if e.Dirty() {
		t.Fatal("resolving a loaded document is not an edit")
	}
}

func TestToggleLayoutUnitKeepsTheSize(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.box", 8, 0)
	if err := e.ToggleLayoutUnit(id, AxisX); err != nil {
		t.Fatal(err)
	}
	if got := e.Layout(id).X; got != "10%" {
		t.Fatalf("x = %q", got)
	}
	must(t, e.SetCanvasSize(160, 24))
	if got := rectOf(e, id).X; got != 16 {
		t.Fatalf("x at 160 = %d", got)
	}
	must(t, e.ToggleLayoutUnit(id, AxisX))
	if e.Layout(id).X != "" {
		t.Fatal("not fixed again")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
