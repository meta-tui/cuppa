package container

import (
	"slices"
	"testing"

	"github.com/meta-tui/cuppa/libs/document/design"
)

func tabs(active string) design.Node {
	return design.Node{ID: "t", Component: design.TabsComponent, Rect: design.Rect{X: 5, Y: 2, W: 40, H: 10},
		Props: map[string]string{PropPages: "Overview, Colours ,Keys", PropActive: active}}
}

func panels(open string, exclusive bool) design.Node {
	n := design.Node{ID: "p", Component: design.PanelsComponent, Rect: design.Rect{X: 0, Y: 1, W: 30, H: 13},
		Props: map[string]string{PropPages: "A,B,C", PropOpen: open}}
	if exclusive {
		n.Props[PropExclusive] = "true"
	}
	return n
}

func TestTabsHaveABarAndABodyUnderIt(t *testing.T) {
	g := Layout(tabs("1"))
	if len(g.Tabs) != 3 || g.Tabs[0].X0 != 5 || g.Tabs[0].X1 != 5+len("Overview")+2 || g.Tabs[1].X0 != g.Tabs[0].X1+1 {
		t.Fatalf("tabs = %+v", g.Tabs)
	}
	if want := (design.Rect{X: 5, Y: 4, W: 40, H: 8}); g.Bodies[0] != want || g.Bodies[2] != want {
		t.Fatalf("bodies = %+v", g.Bodies)
	}
	if !Shown(tabs("1"), 1) || Shown(tabs("1"), 0) {
		t.Error("only the active tab is shown")
	}
	if Active(tabs("9")) != 2 || Active(tabs("-3")) != 0 {
		t.Error("the active page is kept inside the pages")
	}
}

func TestPanelsShareTheRowsAfterTheHeadersBetweenTheOpenSections(t *testing.T) {
	g := Layout(panels("0,2", false))
	// 13 rows, 3 headers: 10 rows for 2 open sections.
	if g.Heads[0].Y != 1 || g.Bodies[0] != (design.Rect{X: 0, Y: 2, W: 30, H: 5}) {
		t.Fatalf("section 0: %+v %+v", g.Heads[0], g.Bodies[0])
	}
	if g.Heads[1].Y != 7 || g.Bodies[1].H != 0 || g.Bodies[1].Y != 8 {
		t.Fatalf("closed section 1: %+v %+v", g.Heads[1], g.Bodies[1])
	}
	if g.Heads[2].Y != 8 || g.Bodies[2] != (design.Rect{X: 0, Y: 9, W: 30, H: 5}) {
		t.Fatalf("section 2: %+v %+v", g.Heads[2], g.Bodies[2])
	}
}

func TestPressingAHeaderChangesTheState(t *testing.T) {
	n := tabs("0")
	page, ok := HeaderAt(n, 5+len("Overview")+2+2, 2)
	if !ok || page != 1 {
		t.Fatalf("header at = %d, %v", page, ok)
	}
	if _, ok := HeaderAt(n, 5, 3); ok {
		t.Error("the rule is not a tab")
	}
	if got := Pressed(n, 2); got[PropActive] != "2" {
		t.Errorf("tab press = %v", got)
	}
	p := panels("0", false)
	if got := Pressed(p, 1); got[PropOpen] != "0,1" {
		t.Errorf("open a section = %v", got)
	}
	if got := Pressed(p, 0); got[PropOpen] != "" {
		t.Errorf("close a section = %v", got)
	}
	if got := Pressed(panels("0", true), 2); got[PropOpen] != "2" {
		t.Errorf("accordion keeps one open = %v", got)
	}
	if page, ok := HeaderAt(p, 3, 1); !ok || page != 0 {
		t.Errorf("section header = %d %v", page, ok)
	}
}

func TestAMemberIsVisibleOnlyWhileItsPageIsShown(t *testing.T) {
	doc := design.Document{Width: 80, Height: 24, Nodes: []design.Node{
		tabs("0"),
		{ID: "a", Component: "lipgloss.label", In: "t", Page: 0},
		{ID: "b", Component: "lipgloss.label", In: "t", Page: 1},
		{ID: "c", Component: "lipgloss.label"},
	}}
	got := []bool{Visible(doc, doc.Nodes[1]), Visible(doc, doc.Nodes[2]), Visible(doc, doc.Nodes[3])}
	if !slices.Equal(got, []bool{true, false, true}) {
		t.Fatalf("visible = %v", got)
	}
	doc.Nodes[0].Hidden = true
	if Visible(doc, doc.Nodes[1]) {
		t.Error("members go with a hidden container")
	}
}

func TestMembersAreDrawnRightAfterTheirContainer(t *testing.T) {
	doc := design.Document{Nodes: []design.Node{
		{ID: "a", Component: "lipgloss.label", In: "t"},
		{ID: "t", Component: design.TabsComponent},
		{ID: "x", Component: "lipgloss.label"},
		{ID: "b", Component: "lipgloss.label", In: "t"},
	}}
	if got := Order(doc); !slices.Equal(got, []int{1, 0, 3, 2}) {
		t.Fatalf("order = %v", got)
	}
}

func TestMovesSaysHowFarEachBodyMoved(t *testing.T) {
	a := panels("0", false)
	b := panels("0,1", false)
	m := Moves(a, b)
	// Opening B halves A: B's body starts higher, C's stays where it was.
	if m[0] != ([2]int{}) || m[1] != ([2]int{0, -5}) || m[2] != ([2]int{}) {
		t.Fatalf("moves = %v", m)
	}
	c := a
	c.Rect.X += 4
	if got := Moves(a, c); got[0] != ([2]int{4, 0}) {
		t.Fatalf("moved container = %v", got)
	}
}
