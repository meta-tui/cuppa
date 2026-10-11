package editor

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestCopyThenPasteAddsOffsetCopiesAndSelectsThem(t *testing.T) {
	ed, ids := three(t)
	if ed.Copy() || ed.CanPaste() {
		t.Fatal("nothing selected, nothing copied")
	}
	if ed.Paste() {
		t.Fatal("nothing to paste")
	}
	ed.Select(ids[0], ids[1])
	if !ed.Copy() || !ed.CanPaste() {
		t.Fatal("copy")
	}
	if !ed.Paste() {
		t.Fatal("paste")
	}
	nodes := ed.Document().Nodes
	if len(nodes) != 5 {
		t.Fatalf("two copies added: %d", len(nodes))
	}
	a, _ := ed.Document().Get(ids[0])
	first, second := nodes[3], nodes[4]
	if first.Rect.X != a.Rect.X+duplicateDX || first.Rect.Y != a.Rect.Y+duplicateDY {
		t.Fatalf("the copy is offset: %+v vs %+v", first.Rect, a.Rect)
	}
	if first.ID == ids[0] || first.Name == a.Name || second.Name == "" {
		t.Fatalf("copies get their own ids and names: %+v", first)
	}
	if got := ed.Selected(); len(got) != 2 || got[0] != first.ID || got[1] != second.ID {
		t.Fatalf("the pasted copies are selected: %v", got)
	}
	ed.Paste()
	third := ed.Document().Nodes[5]
	if third.Rect.X != a.Rect.X+2*duplicateDX {
		t.Fatalf("each paste lands further on: %+v", third.Rect)
	}
	ed.Undo()
	if len(ed.Document().Nodes) != 5 {
		t.Fatal("a paste is one undo step")
	}
}

func TestPasteDropsLockAndHiddenAndSurvivesNewDocuments(t *testing.T) {
	ed, ids := three(t)
	ed.SetLocked(ids[0], true)
	ed.Select(ids[0])
	if !ed.Copy() {
		t.Fatal("a locked layer can be copied")
	}
	ed.Load(design.NewDocument("other", 100, 40))
	if !ed.Paste() {
		t.Fatal("the clipboard outlives the document")
	}
	n := ed.Document().Nodes[0]
	if n.Locked || n.Hidden {
		t.Fatalf("a pasted copy is a normal layer: %+v", n)
	}
}

func TestCopiedGroupsAreIndependentOfTheOriginal(t *testing.T) {
	ed, ids := three(t)
	ed.Select(ids[0], ids[1])
	ed.Group()
	ed.Copy()
	ed.Paste()
	groups := 0
	for _, n := range ed.Document().Nodes {
		if n.IsGroup() {
			groups++
		}
	}
	if groups != 2 {
		t.Fatalf("the group was copied whole: %d", groups)
	}
	g := ed.Document().Nodes[len(ed.Document().Nodes)-1]
	g.Children[0].Name = "changed"
	orig, _ := ed.Document().Get(ed.Document().Nodes[1].ID)
	_ = orig
	if ed.Document().Nodes[len(ed.Document().Nodes)-1].Children[0].Name == "changed" {
		t.Fatal("Document returns copies")
	}
	ed.Select(ed.Document().Nodes[len(ed.Document().Nodes)-1].ID)
	ed.MoveSelectionBy(5, 0)
	if first := ed.Document().Nodes[1]; first.IsGroup() && first.Rect.X != 5 {
		t.Fatalf("moving the copy leaves the original: %+v", first.Rect)
	}
}
