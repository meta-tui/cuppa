package scene

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func themed() (design.Document, design.NodeID, design.NodeID) {
	doc := design.NewDocument("t", 60, 20)
	doc.Theme = design.Theme{Text: "#112233", Muted: "#445566", Border: "#336699", Secondary: "#aa5500"}
	box := doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 10, H: 3}})
	own := doc.Add(design.Node{Component: "lipgloss.box", Name: "Own", Rect: design.Rect{X: 20, W: 10, H: 3},
		Props: map[string]string{"color": "212"}})
	return doc, box.ID, own.ID
}

func TestComponentsFollowTheThemeAndOverridesStay(t *testing.T) {
	doc, _, _ := themed()
	g := Render(doc, standard.Default())
	if got := g.At(0, 0).Fg; got != "#336699" {
		t.Errorf("a box with no colour of its own follows the theme's border: %q", got)
	}
	if got := g.At(20, 0).Fg; got != "212" {
		t.Errorf("a box that sets its colour keeps it: %q", got)
	}
}

func TestNoThemeShowsTheComponentDefaults(t *testing.T) {
	doc, _, _ := themed()
	doc.Theme = design.Theme{}
	g := Render(doc, standard.Default())
	if got := g.At(0, 0).Fg; got != "212" {
		t.Errorf("without a theme the default shows: %q", got)
	}
}

func TestMutedColourIsUsedForQuietText(t *testing.T) {
	doc := design.NewDocument("t", 60, 20)
	doc.Theme = design.Theme{Muted: "#445566"}
	doc.Add(design.Node{Component: "bubbles.textinput", Name: "Input", Rect: design.Rect{W: 20, H: 1}})
	g := Render(doc, standard.Default())
	// "> " is the prompt; the placeholder starts at column 2.
	if got := g.At(2, 0).Fg; got != "#445566" {
		t.Errorf("the placeholder should be muted: %q", got)
	}
}

func TestThemeTextColourReachesLabelsAndPackParts(t *testing.T) {
	doc := design.NewDocument("t", 60, 20)
	doc.Theme = design.Theme{Text: "#112233"}
	doc.Add(design.Node{Component: "lipgloss.label", Name: "L", Rect: design.Rect{W: 10, H: 1}})
	g := Render(doc, standard.Default())
	if got := g.At(0, 0).Fg; got != "#112233" {
		t.Errorf("label text follows the theme: %q", got)
	}
}
