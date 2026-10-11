package format

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestAPagedDesignSurvivesTheFile(t *testing.T) {
	doc := design.NewDocument("App", 80, 24)
	doc.Add(design.Node{Component: "lipgloss.label", Name: "Home label", Rect: design.Rect{W: 10, H: 1}, Goto: "Settings"})
	doc.Pages = []design.Page{
		{Name: "Home"},
		{Name: "Settings", Nodes: []design.Node{{ID: "n9", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 20, H: 5}}}},
	}
	data, err := Encode(doc)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Pages) != 2 || got.Pages[1].Name != "Settings" || len(got.Pages[1].Nodes) != 1 || got.Page != 0 {
		t.Fatalf("pages = %+v page %d", got.Pages, got.Page)
	}
	if got.Nodes[0].Goto != "Settings" {
		t.Errorf("goto = %q", got.Nodes[0].Goto)
	}
}

func TestBadPagesAreRepairedOnOpening(t *testing.T) {
	doc := design.NewDocument("App", 80, 24)
	doc.Pages = []design.Page{{Name: "Home"}, {Name: "Home"}, {Name: "9 bad"}, {Name: "Ok", Nodes: []design.Node{{Component: "lipgloss.box", Rect: design.Rect{W: 1, H: 1}}}}}
	doc.Page = 7
	got := normalise(doc)
	if len(got.Pages) != 2 || got.Pages[0].Name != "Home" || got.Pages[1].Name != "Ok" {
		t.Fatalf("pages = %+v", got.Pages)
	}
	if got.Page < 0 || got.Page >= len(got.Pages) {
		t.Errorf("page = %d", got.Page)
	}
	if len(got.Pages[1].Nodes) != 0 {
		t.Error("a node without an id is dropped")
	}
}
