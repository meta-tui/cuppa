package scene

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func boxDoc() design.Document {
	doc := design.NewDocument("t", 20, 8)
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{X: 2, Y: 2, W: 6, H: 3}})
	return doc
}

func TestBackgroundFillsEmptyCellsAndComponentSpaces(t *testing.T) {
	doc := boxDoc()
	doc.Background = "#102030"
	g := Render(doc, standard.Default())
	if c := g.At(15, 6); c.Ch != 0 || c.Bg != "#102030" {
		t.Fatalf("empty cell should stay unpainted but carry the background: %+v", c)
	}
	if c := g.At(4, 3); c.Bg != "#102030" {
		t.Fatalf("a component's own spaces take the background too: %+v", c)
	}
}

func TestNoBackgroundLeavesTerminalDefault(t *testing.T) {
	g := Render(boxDoc(), standard.Default())
	if c := g.At(15, 6); c.Ch != 0 || c.Bg != "" {
		t.Fatalf("no background means untouched cells: %+v", c)
	}
}

func TestShadowFallsBelowAndRightOfAComponentOnly(t *testing.T) {
	doc := boxDoc()
	doc.Effects.Shadow = true
	g := Render(doc, standard.Default())
	if g.At(8, 3).Ch != '▒' || g.At(4, 5).Ch != '▒' {
		t.Fatalf("shadow missing right or below: %q %q", g.At(8, 3).Ch, g.At(4, 5).Ch)
	}
	if g.At(1, 1).Ch != 0 || g.At(2, 2).Ch == '▒' {
		t.Fatal("shadow must not appear above-left or inside the component")
	}
}

func TestScanlinesDimEveryOtherRow(t *testing.T) {
	doc := boxDoc()
	doc.Effects.Scanlines = true
	g := Render(doc, standard.Default())
	if g.At(15, 1).Dim == g.At(15, 2).Dim {
		t.Fatal("neighbouring rows should differ")
	}
}

func TestVignetteDimsOnlyTheEdges(t *testing.T) {
	doc := boxDoc()
	doc.Effects.Vignette = true
	g := Render(doc, standard.Default())
	if !g.At(0, 4).Dim || !g.At(10, 0).Dim || g.At(10, 4).Dim {
		t.Fatal("edges dim, the middle does not")
	}
}
