package editor

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// three returns an editor with three boxes a, b, c (back to front).
func three(t *testing.T) (*Editor, [3]design.NodeID) {
	t.Helper()
	ed := New(standard.Default(), design.NewDocument("t", 100, 40))
	var ids [3]design.NodeID
	for i := range ids {
		id, err := ed.Add("lipgloss.box", 5+i*30, 5)
		if err != nil {
			t.Fatal(err)
		}
		ids[i] = id
	}
	ed.Clear()
	return ed, ids
}

func order(ed *Editor) []design.NodeID {
	var out []design.NodeID
	for _, n := range ed.Document().Nodes {
		out = append(out, n.ID)
	}
	return out
}

func TestLockedLayersResistEveryEdit(t *testing.T) {
	ed, ids := three(t)
	a := ids[0]
	if !ed.SetLocked(a, true) {
		t.Fatal("locking should report a change")
	}
	ed.Select(a)
	before, _ := ed.Document().Get(a)

	if ed.MoveSelectionBy(5, 5) {
		t.Error("a locked layer must not move")
	}
	if ed.MoveTo(a, 50, 20) {
		t.Error("MoveTo must respect the lock")
	}
	if ed.SetRect(a, design.Rect{X: 5, Y: 5, W: 40, H: 10}, true) {
		t.Error("resize must respect the lock")
	}
	if err := ed.SetProp(a, "title", "x"); err == nil || !strings.Contains(err.Error(), "locked") {
		t.Errorf("properties must be refused, got %v", err)
	}
	if ed.Delete() || ed.DeleteLayer(a) {
		t.Error("a locked layer must not be deleted")
	}
	after, _ := ed.Document().Get(a)
	if after.Rect != before.Rect || after.Props["title"] != "" {
		t.Errorf("the locked layer changed: %+v", after)
	}
	// Reordering and renaming stay allowed, as in Photoshop.
	if !ed.MoveLayer(a, 2) {
		t.Error("a locked layer can still change z-order")
	}
	if !ed.Rename(a, "Background art") {
		t.Error("a locked layer can still be renamed")
	}
}

func TestLockOnlyHoldsTheLockedMembersOfASelection(t *testing.T) {
	ed, ids := three(t)
	ed.SetLocked(ids[0], true)
	ed.Select(ids[0], ids[1])
	if !ed.MoveSelectionBy(3, 0) {
		t.Fatal("the unlocked member should still move")
	}
	a, _ := ed.Document().Get(ids[0])
	b, _ := ed.Document().Get(ids[1])
	if a.Rect.X != 5 || b.Rect.X != 38 {
		t.Fatalf("locked stayed, unlocked moved: a=%d b=%d", a.Rect.X, b.Rect.X)
	}
	if !ed.Delete() {
		t.Fatal("delete should remove the unlocked one")
	}
	if len(ed.Document().Nodes) != 2 {
		t.Fatalf("only the unlocked layer goes: %d nodes", len(ed.Document().Nodes))
	}
}

func TestHidingALayerLeavesTheSelectionAndUndoes(t *testing.T) {
	ed, ids := three(t)
	ed.Select(ids[0], ids[1])
	if !ed.SetHidden(ids[1], true) {
		t.Fatal("hiding should report a change")
	}
	if ed.IsSelected(ids[1]) || !ed.IsSelected(ids[0]) {
		t.Fatalf("selection = %v", ed.Selected())
	}
	if ed.SetHidden(ids[1], true) {
		t.Error("hiding twice is not a change")
	}
	ed.Undo()
	if n, _ := ed.Document().Get(ids[1]); n.Hidden {
		t.Fatal("undo should show it again")
	}
}

func TestLockAndHideAreUndoableSteps(t *testing.T) {
	ed, ids := three(t)
	ed.SetLocked(ids[2], true)
	ed.Undo()
	if n, _ := ed.Document().Get(ids[2]); n.Locked {
		t.Fatal("undo should unlock")
	}
	ed.Redo()
	if n, _ := ed.Document().Get(ids[2]); !n.Locked {
		t.Fatal("redo should lock again")
	}
}

func TestDuplicatingGivesAnUnlockedVisibleCopy(t *testing.T) {
	ed, ids := three(t)
	ed.SetLocked(ids[0], true)
	ed.SetHidden(ids[0], true)
	ed.Select(ids[0])
	// A hidden layer is selectable from the list (Select), so duplicate works.
	if !ed.Duplicate() {
		t.Fatal("duplicate")
	}
	nodes := ed.Document().Nodes
	copy := nodes[len(nodes)-1]
	if copy.Locked || copy.Hidden {
		t.Fatalf("the copy should be a normal layer: %+v", copy)
	}
}

func TestMoveLayerToAnExactPosition(t *testing.T) {
	ed, ids := three(t)
	a, b, c := ids[0], ids[1], ids[2]
	if !ed.MoveLayer(a, 2) { // to the front
		t.Fatal("move")
	}
	if got := order(ed); got[0] != b || got[1] != c || got[2] != a {
		t.Fatalf("order = %v", got)
	}
	if !ed.MoveLayer(a, 0) {
		t.Fatal("move back")
	}
	if got := order(ed); got[0] != a || got[1] != b || got[2] != c {
		t.Fatalf("order = %v", got)
	}
	if ed.MoveLayer(a, 0) {
		t.Error("moving to where it already is changes nothing")
	}
	if !ed.MoveLayer(a, 99) || order(ed)[2] != a {
		t.Error("an out-of-range position clamps to the front")
	}
	ed.Undo()
	if order(ed)[0] != a {
		t.Error("each move is one undo step")
	}
}

func TestMovingASelectedLayerMovesTheWholeSelectionTogether(t *testing.T) {
	ed, ids := three(t)
	a, b, c := ids[0], ids[1], ids[2]
	ed.Select(a, b)
	if !ed.MoveLayer(a, 1) { // a and b go after c
		t.Fatal("move")
	}
	if got := order(ed); got[0] != c || got[1] != a || got[2] != b {
		t.Fatalf("order = %v, want [c a b]", got)
	}
}

func TestDeleteLayerWorksWithoutSelectingIt(t *testing.T) {
	ed, ids := three(t)
	ed.Select(ids[0])
	if !ed.DeleteLayer(ids[1]) {
		t.Fatal("delete the layer dropped on the trash")
	}
	if len(ed.Document().Nodes) != 2 || !ed.IsSelected(ids[0]) {
		t.Fatalf("only that layer goes and the selection stays: %v", ed.Selected())
	}
	ed.Undo()
	if len(ed.Document().Nodes) != 3 {
		t.Fatal("undo brings it back")
	}
	if ed.DeleteLayer("nope") {
		t.Error("unknown layer")
	}
}
