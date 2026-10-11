package editor

import (
	"testing"

	"github.com/meta-tui/cuppa/libs/catalog/standard"
	"github.com/meta-tui/cuppa/libs/document/container"
	"github.com/meta-tui/cuppa/libs/document/design"
)

// tabsWithTwoLabels is a tabs container at (10, 4) with a label on each of two
// pages.
func tabsWithTwoLabels(t *testing.T) (*Editor, design.NodeID, design.NodeID, design.NodeID) {
	t.Helper()
	ed := New(standard.Default(), design.NewDocument("t", 100, 40))
	tabs, err := ed.Add("cuppa.tabs", 10, 4)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := ed.Add("lipgloss.label", 12, 7) // inside the shown page
	ed.PressContainerHeader(tabs, 1)
	b, _ := ed.Add("lipgloss.label", 14, 8)
	ed.PressContainerHeader(tabs, 0)
	return ed, tabs, a, b
}

func TestAComponentDroppedOnAPageBelongsToIt(t *testing.T) {
	ed, tabs, a, b := tabsWithTwoLabels(t)
	na, _ := ed.Document().Get(a)
	nb, _ := ed.Document().Get(b)
	if na.In != tabs || na.Page != 0 || nb.In != tabs || nb.Page != 1 {
		t.Fatalf("a %+v b %+v", na, nb)
	}
	doc := ed.Document()
	if !container.Visible(doc, na) || container.Visible(doc, nb) {
		t.Fatal("only the label of the shown page is visible")
	}
	free, _ := ed.Add("lipgloss.label", 70, 30)
	if n, _ := ed.Document().Get(free); n.In != "" {
		t.Errorf("outside the container: %+v", n)
	}
}

func TestPressingATabChangesWhatIsVisibleAndIsOneUndoStep(t *testing.T) {
	ed, tabs, a, b := tabsWithTwoLabels(t)
	ed.Select(a)
	if !ed.PressContainerHeader(tabs, 1) {
		t.Fatal("nothing changed")
	}
	doc := ed.Document()
	na, _ := doc.Get(a)
	nb, _ := doc.Get(b)
	if container.Visible(doc, na) || !container.Visible(doc, nb) {
		t.Fatal("page 1 should be the shown one")
	}
	if len(ed.Selected()) != 0 {
		t.Error("a member that went out of sight is deselected")
	}
	ed.Undo()
	if n, _ := ed.Document().Get(tabs); container.Active(n) != 0 {
		t.Fatal("undo did not restore the page")
	}
}

func TestSelectingAMemberShowsItsPage(t *testing.T) {
	ed, tabs, _, b := tabsWithTwoLabels(t)
	ed.Select(b)
	if n, _ := ed.Document().Get(tabs); container.Active(n) != 1 {
		t.Fatalf("active = %d", container.Active(n))
	}
}

func TestMovingAContainerMovesItsMembers(t *testing.T) {
	ed, tabs, a, b := tabsWithTwoLabels(t)
	before, _ := ed.Document().Get(a)
	beforeB, _ := ed.Document().Get(b)
	ed.Select(tabs)
	ed.Nudge(5, 3)
	after, _ := ed.Document().Get(a)
	afterB, _ := ed.Document().Get(b)
	if after.Rect.X != before.Rect.X+5 || after.Rect.Y != before.Rect.Y+3 || afterB.Rect.X != beforeB.Rect.X+5 {
		t.Fatalf("a %+v -> %+v, b %+v -> %+v", before.Rect, after.Rect, beforeB.Rect, afterB.Rect)
	}
	// Moving the container and a member together moves the member once.
	ed.Select(tabs, a)
	ed.Nudge(2, 0)
	again, _ := ed.Document().Get(a)
	if again.Rect.X != after.Rect.X+2 {
		t.Fatalf("a moved to %d, want %d", again.Rect.X, after.Rect.X+2)
	}
}

func TestAMemberStaysInsideItsPage(t *testing.T) {
	ed, tabs, a, _ := tabsWithTwoLabels(t)
	ed.MoveTo(a, 90, 38)
	n, _ := ed.Document().Get(a)
	c, _ := ed.Document().Get(tabs)
	body := container.Layout(c).Bodies[0]
	if !body.Contains(n.Rect.X, n.Rect.Y) {
		t.Fatalf("label at %+v left the body %+v", n.Rect, body)
	}
}

func TestOpeningAPanelSectionMovesTheMembersBelowIt(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("t", 100, 40))
	p, _ := ed.Add("cuppa.panels", 0, 0)
	for key, value := range map[string]string{"pages": "A,B,C", "open": "0,2"} {
		if err := ed.SetProp(p, key, value); err != nil {
			t.Fatal(err)
		}
	}
	ed.SetRect(p, design.Rect{W: 30, H: 13}, true)
	c, _ := ed.Document().Get(p)
	third := container.Layout(c).Bodies[2]
	m, _ := ed.Add("lipgloss.label", third.X+1, third.Y+1)
	n, _ := ed.Document().Get(m)
	if n.In != p || n.Page != 2 {
		t.Fatalf("member = %+v", n)
	}
	off := n.Rect.Y - third.Y
	ed.PressContainerHeader(p, 1) // opens B: C's body moves
	c, _ = ed.Document().Get(p)
	now := container.Layout(c).Bodies[2]
	n, _ = ed.Document().Get(m)
	if now.Y == third.Y {
		t.Fatal("the section did not move")
	}
	if n.Rect.Y-now.Y != off {
		t.Fatalf("the label kept %d rows under its section's top, had %d", n.Rect.Y-now.Y, off)
	}
}

func TestDeletingAContainerDeletesItsMembersAndDuplicatingCopiesThem(t *testing.T) {
	ed, tabs, _, _ := tabsWithTwoLabels(t)
	ed.Select(tabs)
	ed.Duplicate()
	if got := len(ed.Document().Nodes); got != 6 {
		t.Fatalf("after duplicate: %d nodes", got)
	}
	copyID := ed.Selected()[0]
	if got := len(ed.Members(copyID)); got != 2 {
		t.Fatalf("the copy has %d members", got)
	}
	ed.Select(tabs)
	ed.Delete()
	if got := len(ed.Document().Nodes); got != 3 {
		t.Fatalf("after delete: %d nodes", got)
	}
}

func TestPagesCanBeAddedAndRemoved(t *testing.T) {
	ed, tabs, a, b := tabsWithTwoLabels(t)
	if err := ed.AddPage(tabs, "Keys"); err != nil {
		t.Fatal(err)
	}
	if n, _ := ed.Document().Get(tabs); len(container.Pages(n)) != 3 {
		t.Fatalf("pages = %v", container.Pages(n))
	}
	if err := ed.RemovePage(tabs, 0); err != nil {
		t.Fatal(err)
	}
	doc := ed.Document()
	if _, ok := doc.Get(a); ok {
		t.Error("the components of a removed page are removed")
	}
	if nb, _ := doc.Get(b); nb.Page != 0 {
		t.Errorf("later pages move up: %+v", nb)
	}
	if err := ed.RemovePage(tabs, 0); err != nil {
		t.Fatal(err)
	}
	if err := ed.RemovePage(tabs, 0); err == nil {
		t.Error("the last page cannot be removed")
	}
}

func TestAContainerAndItsMembersCannotBeGrouped(t *testing.T) {
	ed, tabs, a, _ := tabsWithTwoLabels(t)
	ed.Select(tabs, a)
	if ed.CanGroup() {
		t.Fatal("grouping would take members out of their pages")
	}
}
