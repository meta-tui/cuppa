package editor

import (
	"slices"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func names(nodes []design.Node) []string {
	var out []string
	for _, n := range nodes {
		out = append(out, n.Name)
	}
	return out
}

// A master's components are on every screen below the page's own, a page can
// hide one, and the master itself is not a screen.
func TestTheMasterIsShownOnEveryScreenUnlessAPageHidesIt(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("App", 80, 24))
	home, _ := ed.Add("lipgloss.label", 2, 2)
	_ = home
	if err := ed.AddPage("Settings"); err != nil {
		t.Fatal(err)
	}
	_, _ = ed.Add("lipgloss.box", 4, 4)
	if err := ed.AddMasterPage(); err != nil {
		t.Fatal(err)
	}
	if !ed.OnMaster() {
		t.Fatal("the master is not shown after adding it")
	}
	title, _ := ed.Add("lipgloss.label", 0, 0)
	if err := ed.AddMasterPage(); err == nil {
		t.Fatal("a second master was added")
	}
	doc := ed.Document()
	screens := doc.Screens()
	if len(screens) != 2 || doc.ScreenCount() != 2 {
		t.Fatalf("screens = %d", len(screens))
	}
	for _, s := range screens {
		if len(s.Nodes) != 2 || s.Nodes[0].ID != design.MasterPrefix+title {
			t.Fatalf("screen %s has %v", s.Name, names(s.Nodes))
		}
	}
	// Hide it on Settings.
	if err := ed.SetMasterHidden(1, title, true); err != nil {
		t.Fatal(err)
	}
	screens = ed.Document().Screens()
	if len(screens[0].Nodes) != 2 || len(screens[1].Nodes) != 1 {
		t.Fatalf("after hiding: %d and %d nodes", len(screens[0].Nodes), len(screens[1].Nodes))
	}
	if !ed.MasterHidden(1, title) || ed.MasterHidden(0, title) {
		t.Fatal("MasterHidden is wrong")
	}
	// Walking the pages keeps everything.
	ed.SwitchPage(0)
	if ed.OnMaster() || len(ed.RenderDocument().Nodes) != len(ed.Document().Nodes)+1 {
		t.Fatalf("the page does not draw the master: %d vs %d", len(ed.RenderDocument().Nodes), len(ed.Document().Nodes))
	}
	ed.SwitchPage(1)
	if len(ed.RenderDocument().Nodes) != len(ed.Document().Nodes) {
		t.Fatal("the page that hides it draws it")
	}
	if err := ed.SetGoto(ed.Document().Nodes[0].ID, "Master"); err == nil {
		t.Fatal("a click can go to the master")
	}
	if err := ed.RenamePage(2, "Chrome"); err == nil {
		t.Fatal("the master was renamed")
	}
}

func TestComponentsMoveToTheMasterFromAPage(t *testing.T) {
	ed := New(standard.Default(), design.NewDocument("App", 80, 24))
	id, _ := ed.Add("lipgloss.label", 0, 0)
	if _, err := ed.MoveToMaster(); err == nil {
		t.Fatal("moved with no master")
	}
	if err := ed.AddMasterPage(); err != nil {
		t.Fatal(err)
	}
	ed.SwitchPage(0)
	ed.Select(id)
	n, err := ed.MoveToMaster()
	if err != nil || n != 1 {
		t.Fatalf("moved %d: %v", n, err)
	}
	if got := names(ed.Document().Nodes); len(got) != 0 {
		t.Fatalf("the page kept %v", got)
	}
	if got := ed.MasterComponents(); len(got) != 1 {
		t.Fatalf("master has %v", names(got))
	}
	ed.Undo()
	if len(ed.MasterComponents()) != 0 || len(ed.Document().Nodes) != 1 {
		t.Fatal("undo did not bring it back")
	}
	if slices.Contains(ed.PageNames(), "") {
		t.Fatal("empty page name")
	}
}
