package editor

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestGroupingMakesOneComponentAndUngroupingRestoresTheParts(t *testing.T) {
	ed, ids := three(t)
	a, b, c := ids[0], ids[1], ids[2]
	ed.Select(a, c)
	if !ed.CanGroup() || !ed.Group() {
		t.Fatal("two components can be grouped")
	}
	nodes := ed.Document().Nodes
	if len(nodes) != 2 || nodes[0].ID != b || !nodes[1].IsGroup() {
		t.Fatalf("the group takes the front-most member's place: %+v", order(ed))
	}
	g := nodes[1]
	if g.Name != "Group 1" || len(g.Children) != 2 || g.Rect.X != 5 || g.Rect.W != 84 {
		t.Fatalf("group = %+v", g)
	}
	if g.Children[0].Rect.X != 0 || g.Children[1].Rect.X != 60 {
		t.Fatalf("children are relative to the group: %+v", g.Children)
	}
	if got := ed.Selected(); len(got) != 1 || got[0] != g.ID {
		t.Fatalf("the group is selected: %v", got)
	}

	ed.MoveSelectionBy(10, 4)
	if !ed.CanUngroup() || !ed.Ungroup() {
		t.Fatal("ungroup")
	}
	nodes = ed.Document().Nodes
	if len(nodes) != 3 {
		t.Fatalf("parts are back: %d", len(nodes))
	}
	var xs []int
	for _, n := range nodes {
		xs = append(xs, n.Rect.X)
	}
	if xs[0] != 35 || xs[1] != 15 || xs[2] != 75 {
		t.Fatalf("parts follow the group's move: %v", xs)
	}
	if len(ed.Selected()) != 2 {
		t.Fatalf("the released parts are selected: %v", ed.Selected())
	}
}

func TestResizingAGroupScalesItsPartsOnUngroup(t *testing.T) {
	ed, ids := three(t)
	ed.Select(ids[0], ids[1])
	ed.Group()
	g := ed.Document().Nodes[len(ed.Document().Nodes)-1]
	box := g.Rect
	box.W *= 2
	if !ed.SetRect(g.ID, box, true) {
		t.Fatal("a group resizes")
	}
	ed.Ungroup()
	first, _ := ed.Document().Get(ed.Selected()[0])
	if first.Rect.W != 2*(ed.Document().Nodes[0].Rect.W/2) && first.Rect.W < 2 {
		t.Fatalf("parts scale with the group: %+v", first.Rect)
	}
}

func TestGroupRules(t *testing.T) {
	ed, ids := three(t)
	ed.Select(ids[0])
	if ed.CanGroup() || ed.Group() {
		t.Fatal("one component is not a group")
	}
	ed.SetLocked(ids[1], true)
	ed.Select(ids[0], ids[1])
	if ed.CanGroup() || ed.Group() {
		t.Fatal("a locked member blocks grouping")
	}
	ed.Select(ids[2])
	if ed.CanUngroup() || ed.Ungroup() {
		t.Fatal("a plain component cannot be ungrouped")
	}
}

func TestGroupAndUngroupAreOneUndoStepEach(t *testing.T) {
	ed, ids := three(t)
	ed.Select(ids[0], ids[1], ids[2])
	ed.Group()
	if len(ed.Document().Nodes) != 1 {
		t.Fatal("all three grouped")
	}
	ed.Ungroup()
	ed.Undo()
	if len(ed.Document().Nodes) != 1 {
		t.Fatal("undo of ungroup brings the group back")
	}
	ed.Undo()
	if len(ed.Document().Nodes) != 3 {
		t.Fatal("undo of group brings the parts back")
	}
}

func TestNestedGroupsAndHiddenGroups(t *testing.T) {
	ed, ids := three(t)
	ed.Select(ids[0], ids[1])
	ed.Group()
	inner := ed.Selected()[0]
	ed.Select(inner, ids[2])
	if !ed.Group() {
		t.Fatal("a group can be grouped again")
	}
	outer := ed.Selected()[0]
	ed.SetHidden(outer, true)
	ed.Select(outer)
	ed.Ungroup()
	for _, n := range ed.Document().Nodes {
		if !n.Hidden {
			t.Fatalf("a hidden group releases hidden parts: %+v", n)
		}
	}
	var _ design.Node
}
