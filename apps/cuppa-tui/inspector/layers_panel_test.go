package inspector

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// layersFixture is an inspector over three boxes named Box 1 (back), Box 2, Box 3 (front).
func layersFixture(t *testing.T) (*Model, *editor.Editor, []design.NodeID) {
	t.Helper()
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 100, 40))
	var ids []design.NodeID
	for i := 0; i < 3; i++ {
		id, err := ed.Add("lipgloss.box", 5+i*20, 5)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	ed.Clear()
	m := New(ed, cat)
	m.SetSize(34, 60)
	m.Lines()
	return m, ed, ids
}

// find returns the pane position of the first rendered cell of text.
func find(t *testing.T, m *Model, text string) (x, y int) {
	t.Helper()
	for row, l := range m.Lines() {
		plain := stripANSI(l)
		if i := strings.Index(plain, text); i >= 0 {
			return len([]rune(plain[:i])), row
		}
	}
	t.Fatalf("%q is not on screen:\n%s", text, strings.Join(m.Lines(), "\n"))
	return 0, 0
}

func press(m *Model, x, y int) {
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Down, Left: true})
	m.Lines()
}
func dragTo(m *Model, x, y int) {
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Move, Held: true})
	m.Lines()
}
func release(m *Model, x, y int) {
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Up, Left: true})
	m.Lines()
}

func names(ed *editor.Editor) []string {
	var out []string
	for _, n := range ed.Document().Nodes {
		out = append(out, n.Name)
	}
	return out
}

func TestEachLayerRowHasAnEyeAndAPadlock(t *testing.T) {
	m, _, _ := layersFixture(t)
	for _, l := range m.Lines() {
		plain := stripANSI(l)
		if strings.Contains(plain, "Box 3") && (!strings.Contains(plain, "◉") || !strings.Contains(plain, "▢")) {
			t.Fatalf("row lacks its toggles: %q", plain)
		}
	}
}

func TestClickingTheEyeHidesAndShowsALayer(t *testing.T) {
	m, ed, ids := layersFixture(t)
	x, y := find(t, m, "◉") // the first row's eye: the front layer
	press(m, x, y)
	release(m, x, y)
	front, _ := ed.Document().Get(ids[2])
	if !front.Hidden {
		t.Fatal("clicking the eye hides the layer")
	}
	if strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "○ ▢ Box 3") == false {
		t.Fatalf("a hidden layer shows a closed eye:\n%s", stripANSI(strings.Join(m.Lines(), "\n")))
	}
	x, y = find(t, m, "○")
	press(m, x, y)
	release(m, x, y)
	if n, _ := ed.Document().Get(ids[2]); n.Hidden {
		t.Fatal("clicking again shows it")
	}
}

func TestClickingThePadlockLocksALayer(t *testing.T) {
	m, ed, ids := layersFixture(t)
	x, y := find(t, m, "▢")
	press(m, x, y)
	release(m, x, y)
	if n, _ := ed.Document().Get(ids[2]); !n.Locked {
		t.Fatal("clicking the padlock locks the layer")
	}
	if !strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "▣") {
		t.Fatal("a locked layer shows a closed padlock")
	}
}

func TestAPressWithoutMovingOnlySelects(t *testing.T) {
	m, ed, ids := layersFixture(t)
	x, y := find(t, m, "Box 1")
	press(m, x, y)
	release(m, x, y)
	if !ed.IsSelected(ids[0]) {
		t.Fatal("pressing a layer selects it")
	}
	if got := names(ed); got[0] != "Box 1" || got[2] != "Box 3" {
		t.Fatalf("order changed: %v", got)
	}
}

func TestDraggingALayerUpTheListMovesItForward(t *testing.T) {
	m, ed, _ := layersFixture(t)
	x, y := find(t, m, "Box 1") // the back layer is the last row
	_, top := find(t, m, "Box 3")
	press(m, x, y)
	dragTo(m, x, y-1)
	if !m.Dragging() {
		t.Fatal("moving to another row starts a drag")
	}
	if !strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "▸") {
		t.Fatal("the row it would land on is marked")
	}
	dragTo(m, x, top)
	release(m, x, top)
	if got := names(ed); got[2] != "Box 1" || got[0] != "Box 2" || got[1] != "Box 3" {
		t.Fatalf("Box 1 should be in front now: %v", got)
	}
	if m.Dragging() {
		t.Fatal("the drag ends on release")
	}
}

func TestDraggingALayerDownTheListSendsItBack(t *testing.T) {
	m, ed, _ := layersFixture(t)
	x, y := find(t, m, "Box 3")
	_, bottom := find(t, m, "Box 1")
	press(m, x, y)
	dragTo(m, x, bottom)
	release(m, x, bottom)
	if got := names(ed); got[0] != "Box 3" {
		t.Fatalf("Box 3 should be at the back: %v", got)
	}
	ed.Undo()
	if got := names(ed); got[2] != "Box 3" {
		t.Fatalf("one undo step restores the order: %v", got)
	}
}

func TestDroppingALayerOnTheTrashDeletesIt(t *testing.T) {
	m, ed, ids := layersFixture(t)
	tx, ty := find(t, m, "[Trash]")
	x, y := find(t, m, "Box 2")
	press(m, x, y)
	dragTo(m, tx+2, ty)
	if !strings.Contains(strings.Join(m.Lines(), "\n"), "[Trash]") {
		t.Fatal("the trash button stays visible while dragging")
	}
	release(m, tx+2, ty)
	if got := names(ed); len(got) != 2 || got[0] != "Box 1" || got[1] != "Box 3" {
		t.Fatalf("Box 2 should be gone: %v", got)
	}
	if ed.IsSelected(ids[1]) {
		t.Fatal("the deleted layer cannot stay selected")
	}
}

func TestDroppingASelectedLayerOnTheTrashDeletesTheWholeSelection(t *testing.T) {
	m, ed, ids := layersFixture(t)
	ed.Select(ids[0], ids[1])
	m.Lines()
	tx, ty := find(t, m, "[Trash]")
	x, y := find(t, m, "Box 2")
	press(m, x, y)
	dragTo(m, tx+2, ty)
	release(m, tx+2, ty)
	if got := names(ed); len(got) != 1 || got[0] != "Box 3" {
		t.Fatalf("both selected layers go: %v", got)
	}
}

func TestClickingTheTrashButtonDeletesTheSelection(t *testing.T) {
	m, ed, ids := layersFixture(t)
	ed.Select(ids[1])
	m.Lines()
	tx, ty := find(t, m, "[Trash]")
	press(m, tx+2, ty)
	release(m, tx+2, ty)
	if got := names(ed); len(got) != 2 || got[0] != "Box 1" || got[1] != "Box 3" {
		t.Fatalf("the selected layer is deleted: %v", got)
	}
}

func TestEscCancelsALayerDragWithoutChangingAnything(t *testing.T) {
	m, ed, _ := layersFixture(t)
	x, y := find(t, m, "Box 1")
	_, top := find(t, m, "Box 3")
	press(m, x, y)
	dragTo(m, x, top)
	m.CancelDrag()
	release(m, x, top)
	if got := names(ed); got[0] != "Box 1" || got[2] != "Box 3" {
		t.Fatalf("a cancelled drag must not reorder: %v", got)
	}
}

func TestALockedSelectionSaysSoAndRefusesEdits(t *testing.T) {
	m, ed, ids := layersFixture(t)
	ed.SetLocked(ids[2], true)
	ed.Select(ids[2])
	screen := stripANSI(strings.Join(m.Lines(), "\n"))
	if !strings.Contains(screen, "Locked: unlock it in Layers") {
		t.Fatalf("the details bar should say why edits do nothing:\n%s", screen)
	}
	barClick(t, m, "X", "plus")
	if n, _ := ed.Document().Get(ids[2]); n.Rect.X != 45 {
		t.Fatalf("a locked layer must not move, X = %d", n.Rect.X)
	}
}
