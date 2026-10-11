package editor

import (
	"slices"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func twoPages(t *testing.T) (*Editor, design.NodeID, design.NodeID) {
	t.Helper()
	ed := New(standard.Default(), design.NewDocument("App", 80, 24))
	home, _ := ed.Add("lipgloss.label", 2, 2)
	if err := ed.AddPage("Settings"); err != nil {
		t.Fatal(err)
	}
	box, _ := ed.Add("lipgloss.box", 4, 4)
	return ed, home, box
}

func TestAddingAPageSplitsTheDesignAndShowsTheNewOne(t *testing.T) {
	ed, home, box := twoPages(t)
	if got := ed.PageNames(); !slices.Equal(got, []string{"Page 1", "Settings"}) {
		t.Fatalf("pages = %v", got)
	}
	if ed.ActivePage() != 1 {
		t.Fatalf("active = %d", ed.ActivePage())
	}
	doc := ed.Document()
	if _, ok := doc.Get(box); !ok {
		t.Fatal("the new page shows its own component")
	}
	if _, ok := doc.Get(home); ok {
		t.Fatal("the first page's component is not on this page")
	}
	if len(doc.Pages[0].Nodes) != 1 {
		t.Fatalf("page 1 keeps its components: %+v", doc.Pages[0])
	}
}

func TestSwitchingPagesKeepsEachPagesComponentsAndLeavesNoUndoStep(t *testing.T) {
	ed, home, box := twoPages(t)
	steps := ed.CanUndo()
	if !ed.SwitchPage(0) {
		t.Fatal("no switch")
	}
	doc := ed.Document()
	if _, ok := doc.Get(home); !ok {
		t.Fatal("page 1 shows its component")
	}
	if len(doc.Pages[1].Nodes) != 1 || doc.Pages[1].Nodes[0].ID != box {
		t.Fatalf("page 2 keeps %+v", doc.Pages[1].Nodes)
	}
	if ed.CanUndo() != steps {
		t.Error("switching is not an undo step")
	}
	if len(ed.Selected()) != 0 {
		t.Error("switching clears the selection")
	}
}

func TestPagesRenameMoveAndDeleteAndGotoFollows(t *testing.T) {
	ed, home, _ := twoPages(t)
	ed.SwitchPage(0)
	if err := ed.SetGoto(home, "Settings"); err != nil {
		t.Fatal(err)
	}
	if err := ed.SetGoto(home, "Nowhere"); err == nil {
		t.Error("a page that does not exist is refused")
	}
	if err := ed.RenamePage(1, "Options"); err != nil {
		t.Fatal(err)
	}
	if n, _ := ed.Document().Get(home); n.Goto != "Options" {
		t.Fatalf("goto = %q", n.Goto)
	}
	if err := ed.RenamePage(0, "Options"); err == nil {
		t.Error("a taken name is refused")
	}
	if !ed.MovePage(0, 1) {
		t.Fatal("no move")
	}
	if got := ed.PageNames(); !slices.Equal(got, []string{"Options", "Page 1"}) {
		t.Fatalf("pages = %v", got)
	}
	if n, _ := ed.Document().Get(home); n.ID != home || ed.PageNames()[ed.ActivePage()] != "Page 1" {
		t.Fatal("the page being edited stays the one being edited")
	}
	if err := ed.DeletePage(0); err != nil {
		t.Fatal(err)
	}
	doc := ed.Document()
	if doc.IsPaged() {
		t.Fatalf("one page left is no split: %+v", doc.Pages)
	}
	if n, _ := doc.Get(home); n.Goto != "" {
		t.Errorf("goto to a deleted page = %q", n.Goto)
	}
	if err := ed.DeletePage(0); err == nil {
		t.Error("a design keeps one page")
	}
}

func TestEachPageIsOneScreenOfTheDesign(t *testing.T) {
	ed, _, _ := twoPages(t)
	screens := ed.Document().Screens()
	if len(screens) != 2 || screens[0].Name != "Page 1" || screens[1].Name != "Settings" {
		t.Fatalf("screens = %v", len(screens))
	}
	if len(screens[0].Nodes) != 1 || len(screens[1].Nodes) != 1 || screens[0].IsPaged() {
		t.Fatalf("nodes %d/%d", len(screens[0].Nodes), len(screens[1].Nodes))
	}
}

func TestUndoRestoresThePageStructure(t *testing.T) {
	ed, _, _ := twoPages(t)
	ed.Undo() // the box
	ed.Undo() // the page
	if ed.Document().IsPaged() {
		t.Fatal("undoing the page leaves one page")
	}
}

func TestEachPageKeepsItsOwnSamplesAndFolder(t *testing.T) {
	ed, _, _ := twoPages(t)
	if err := ed.SetSample("Rows", "9"); err != nil {
		t.Fatal(err)
	}
	if !ed.SwitchPage(0) {
		t.Fatal("no switch")
	}
	if got := ed.Samples()["Rows"]; got != "" {
		t.Fatalf("the first page took the sample of the second: %q", got)
	}
	if err := ed.SetSample("Rows", "3"); err != nil {
		t.Fatal(err)
	}
	ed.SwitchPage(1)
	if got := ed.Samples()["Rows"]; got != "9" {
		t.Fatalf("second page sample = %q", got)
	}
	screens := ed.Document().Screens()
	if screens[0].Samples["Rows"] != "3" || screens[1].Samples["Rows"] != "9" {
		t.Fatalf("screens: %v %v", screens[0].Samples, screens[1].Samples)
	}
	if err := ed.SetPageFolder(1, "setup/prefs"); err != nil {
		t.Fatal(err)
	}
	if got := ed.Document().Screens()[1].Folder; got != "setup/prefs" {
		t.Fatalf("folder = %q", got)
	}
	if err := ed.SetPageFolder(1, "../x"); err == nil {
		t.Fatal("a folder outside the export was accepted")
	}
}
