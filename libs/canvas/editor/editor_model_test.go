package editor

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func newEditor() *Editor {
	return New(standard.Default(), design.NewDocument("t", 80, 24))
}

func mustAdd(t *testing.T, e *Editor, comp string, x, y int) design.NodeID {
	t.Helper()
	id, err := e.Add(comp, x, y)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func rectOf(e *Editor, id design.NodeID) design.Rect {
	n, _ := e.Document().Get(id)
	return n.Rect
}

func TestAddUsesDefaultsClampsAndSelects(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.box", 70, 20) // 24x6 would overflow
	if got := rectOf(e, id); got != (design.Rect{X: 56, Y: 18, W: 24, H: 6}) {
		t.Fatalf("rect = %+v", got)
	}
	if sel := e.Selected(); len(sel) != 1 || sel[0] != id {
		t.Fatalf("selection = %v", sel)
	}
	n, _ := e.Primary()
	if n.Name != "Box 1" {
		t.Fatalf("name = %q", n.Name)
	}
	if _, err := e.Add("nope", 0, 0); err == nil {
		t.Fatal("unknown component accepted")
	}
	second := mustAdd(t, e, "lipgloss.box", 0, 0)
	if n2, _ := e.Document().Get(second); n2.Name != "Box 2" {
		t.Fatalf("second name = %q", n2.Name)
	}
}

func TestMoveSelectionIsClampedAndGrouped(t *testing.T) {
	e := newEditor()
	a := mustAdd(t, e, "lipgloss.label", 0, 0)
	b := mustAdd(t, e, "lipgloss.label", 10, 5)
	e.Select(a, b)
	e.Checkpoint()
	e.MoveSelectionBy(-5, 100)
	if rectOf(e, a).X != 0 || rectOf(e, b).X != 10 {
		t.Fatalf("group should not move horizontally past the edge: %+v %+v", rectOf(e, a), rectOf(e, b))
	}
	if rectOf(e, b).Bottom() != 24 || rectOf(e, a).Y != 24-1-5 {
		t.Fatalf("group not clamped vertically: a=%+v b=%+v", rectOf(e, a), rectOf(e, b))
	}
}

func TestResizeEnforcesMinimumSize(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.box", 0, 0)
	e.SetRect(id, design.Rect{W: 1, H: 1}, true)
	if got := rectOf(e, id); got.W != 3 || got.H != 3 {
		t.Fatalf("min size not enforced: %+v", got)
	}
}

func TestUndoRedoAndGestures(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.label", 0, 0)
	e.Checkpoint()
	e.MoveSelectionBy(3, 0)
	e.MoveSelectionBy(2, 0)
	e.EndGesture()
	if rectOf(e, id).X != 5 {
		t.Fatal("drag did not move")
	}
	if !e.Undo() || rectOf(e, id).X != 0 {
		t.Fatalf("undo should revert the whole drag in one step: %+v", rectOf(e, id))
	}
	if !e.Redo() || rectOf(e, id).X != 5 {
		t.Fatal("redo failed")
	}
	// A click that moves nothing must not create an undo step.
	before := len(e.undo)
	e.Checkpoint()
	e.EndGesture()
	if len(e.undo) != before {
		t.Fatal("empty gesture left an undo step")
	}
	e.Undo()
	e.Undo() // back before the Add
	if len(e.Document().Nodes) != 0 || e.CanUndo() {
		t.Fatal("undo past add should empty the document")
	}
}

func TestDeleteAndDuplicate(t *testing.T) {
	e := newEditor()
	a := mustAdd(t, e, "lipgloss.label", 0, 0)
	if !e.Duplicate() || len(e.Document().Nodes) != 2 {
		t.Fatal("duplicate failed")
	}
	sel := e.Selected()
	if len(sel) != 1 || sel[0] == a {
		t.Fatalf("copy should be selected: %v", sel)
	}
	n, _ := e.Primary()
	if n.Rect.X != 2 || n.Rect.Y != 1 || n.Name != "Label 2" {
		t.Fatalf("copy = %+v", n)
	}
	e.Delete()
	if len(e.Document().Nodes) != 1 || len(e.Selected()) != 0 {
		t.Fatal("delete failed")
	}
	if e.Delete() {
		t.Fatal("deleting an empty selection should report no change")
	}
	e.Undo()
	if len(e.Document().Nodes) != 2 {
		t.Fatal("undo delete failed")
	}
}

func TestReorder(t *testing.T) {
	e := newEditor()
	a := mustAdd(t, e, "lipgloss.label", 0, 0)
	b := mustAdd(t, e, "lipgloss.label", 0, 2)
	c := mustAdd(t, e, "lipgloss.label", 0, 4)
	order := func() []design.NodeID {
		var out []design.NodeID
		for _, n := range e.Document().Nodes {
			out = append(out, n.ID)
		}
		return out
	}
	e.Select(a)
	e.Reorder(BringToFront)
	if got := order(); got[2] != a {
		t.Fatalf("front: %v", got)
	}
	e.Reorder(SendToBack)
	if got := order(); got[0] != a {
		t.Fatalf("back: %v", got)
	}
	e.Select(a, b)
	e.Reorder(BringForward) // a,b,c -> c above? selected block moves up one
	if got := order(); got[0] != c || got[1] != a || got[2] != b {
		t.Fatalf("block forward: %v", got)
	}
	if e.Reorder(BringForward) {
		t.Fatal("block already at the front should not move")
	}
}

func TestSetPropValidates(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "bubbles.progress", 0, 0)
	if err := e.SetProp(id, "percent", "250"); err != nil {
		t.Fatal(err)
	}
	if n, _ := e.Primary(); n.Props["percent"] != "100" {
		t.Fatalf("percent not clamped: %v", n.Props)
	}
	if e.SetProp(id, "percent", "abc") == nil || e.SetProp(id, "nope", "1") == nil {
		t.Fatal("invalid values accepted")
	}
	if e.SetProp(id, "show_percentage", "maybe") == nil {
		t.Fatal("bad bool accepted")
	}
	steps := len(e.undo)
	_ = e.SetProp(id, "percent", "100") // unchanged
	if len(e.undo) != steps {
		t.Fatal("no-op SetProp created an undo step")
	}
	lg := mustAdd(t, e, "lipgloss.box", 0, 0)
	if e.SetProp(lg, "border", "wavy") == nil {
		t.Fatal("bad choice accepted")
	}
}

func TestDirtyTracking(t *testing.T) {
	e := newEditor()
	if e.Dirty() {
		t.Fatal("fresh editor is dirty")
	}
	mustAdd(t, e, "lipgloss.label", 0, 0)
	if !e.Dirty() {
		t.Fatal("edit did not mark dirty")
	}
	e.MarkSaved()
	if e.Dirty() {
		t.Fatal("saved editor is dirty")
	}
	e.Undo()
	if !e.Dirty() {
		t.Fatal("undo after save should be dirty")
	}
}

func TestNudgeIsOneUndoStepAndLeavesNoneWhenNothingMoved(t *testing.T) {
	e := newEditor()
	id, _ := e.Add("lipgloss.box", 0, 0)
	e.Select(id)
	if e.Nudge(-1, 0) {
		t.Fatal("already at the left edge: nothing should move")
	}
	if e.CanUndo() {
		// Adding the node is the only step.
		e.Undo()
		if e.CanUndo() {
			t.Fatal("a refused nudge must not leave an undo step")
		}
		e.Redo()
	}
	if !e.Nudge(3, 2) {
		t.Fatal("nudging into the canvas should move")
	}
	n, _ := e.Document().Get(id)
	if n.Rect.X != 3 || n.Rect.Y != 2 {
		t.Fatalf("at (%d,%d), want (3,2)", n.Rect.X, n.Rect.Y)
	}
	e.Undo()
	n, _ = e.Document().Get(id)
	if n.Rect.X != 0 || n.Rect.Y != 0 {
		t.Fatalf("undo: (%d,%d)", n.Rect.X, n.Rect.Y)
	}
}

func TestThemeColoursUndoAndOverridesStayWhenTheThemeChanges(t *testing.T) {
	e := newEditor()
	id, _ := e.Add("lipgloss.box", 0, 0)
	if err := e.SetThemeColor(ThemeBorder, "#336699"); err != nil {
		t.Fatal(err)
	}
	if got := e.Document().Theme.Border; got != "#336699" {
		t.Fatalf("border = %q", got)
	}
	n, _ := e.Document().Get(id)
	if len(n.Props) != 0 {
		t.Fatalf("a new component comes with no colours of its own: %v", n.Props)
	}
	// An override stays whatever the theme becomes, and can be removed.
	if err := e.SetProp(id, "color", "212"); err != nil {
		t.Fatal(err)
	}
	if err := e.SetThemeColor(ThemeBorder, "#aa0000"); err != nil {
		t.Fatal(err)
	}
	n, _ = e.Document().Get(id)
	if n.Props["color"] != "212" {
		t.Fatalf("the override should stay: %v", n.Props)
	}
	if !e.ClearProp(id, "color") {
		t.Fatal("there was an override to remove")
	}
	n, _ = e.Document().Get(id)
	if _, set := n.Props["color"]; set {
		t.Fatal("the override should be gone")
	}
	if e.ClearProp(id, "color") {
		t.Fatal("nothing left to remove")
	}
	e.Undo()
	n, _ = e.Document().Get(id)
	if n.Props["color"] != "212" {
		t.Fatal("removing an override is one undo step")
	}
	if err := e.SetThemeColor("nonsense", "1"); err == nil {
		t.Fatal("unknown role")
	}
	if err := e.SetThemeColor(ThemeText, "not a colour"); err == nil {
		t.Fatal("bad colour")
	}
}

func TestSetThemeIsOneUndoStep(t *testing.T) {
	e := newEditor()
	err := e.SetTheme("#101010", design.Theme{Text: "#eeeeee", Muted: "#888888", Border: "#3366aa", Secondary: "#aa66cc"})
	if err != nil {
		t.Fatal(err)
	}
	doc := e.Document()
	if doc.Background != "#101010" || doc.Theme.Border != "#3366aa" {
		t.Fatalf("theme = %q %+v", doc.Background, doc.Theme)
	}
	e.Undo()
	if doc = e.Document(); doc.Background != "" || !doc.Theme.IsZero() {
		t.Fatalf("one undo should take back the whole preset: %q %+v", doc.Background, doc.Theme)
	}
	if err := e.SetTheme("nope", design.Theme{}); err == nil {
		t.Fatal("a bad colour is refused")
	}
}
