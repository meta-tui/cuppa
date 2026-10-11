package scene

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/cupp"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func teaPack() cupp.Pack {
	return cupp.Pack{
		ID: "tea-shop", Name: "Tea shop",
		Components: []design.Composite{{
			ID: "card", Name: "Card", W: 20, H: 5,
			Nodes: []design.Node{
				{ID: "a", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 20, H: 5}},
				{ID: "b", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{X: 2, Y: 1, W: 10, H: 1}},
			},
			Props: []design.Exposed{{Key: "title", Label: "Title", Kind: "text", Default: "Tea", Target: "b", TargetProp: "text"}},
		}},
	}
}

func rows(doc design.Document, cat Catalog) string {
	g := Render(doc, cat)
	var b strings.Builder
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			ch := g.At(x, y).Ch
			if ch == 0 {
				ch = ' '
			}
			b.WriteRune(ch)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestACompositeDrawsItsPartsWithItsOwnProperties(t *testing.T) {
	cat, _ := cupp.Extend(standard.Default(), []cupp.Pack{teaPack()})
	doc := design.NewDocument("t", 30, 8)
	doc.Add(design.Node{Component: "tea-shop.card", Name: "Card", Rect: design.Rect{X: 1, Y: 1, W: 20, H: 5}})
	if out := rows(doc, cat); !strings.Contains(out, "Tea") || !strings.ContainsAny(out, "╭┌") {
		t.Fatalf("default title and frame expected:\n%s", out)
	}
	doc.Nodes[0].Props = map[string]string{"title": "Matcha"}
	if out := rows(doc, cat); !strings.Contains(out, "Matcha") {
		t.Fatalf("the exposed property drives the inner label:\n%s", out)
	}
}

func TestACompositeScalesWithItsNode(t *testing.T) {
	cat, _ := cupp.Extend(standard.Default(), []cupp.Pack{teaPack()})
	doc := design.NewDocument("t", 60, 12)
	doc.Add(design.Node{Component: "tea-shop.card", Name: "Card", Rect: design.Rect{X: 0, Y: 0, W: 40, H: 10}})
	g := Render(doc, cat)
	// The frame spans the whole node, so its right edge is at column 39.
	if g.At(39, 0).Ch == 0 || g.At(39, 9).Ch == 0 || g.At(40, 0).Ch != 0 {
		t.Fatal("the frame should stretch to the placed size")
	}
}

func TestAComponentThatContainsItselfStopsInsteadOfLooping(t *testing.T) {
	loop := cupp.Pack{ID: "loop", Name: "Loop", Components: []design.Composite{{
		ID: "again", Name: "Again", W: 10, H: 4,
		Nodes: []design.Node{{ID: "x", Component: "loop.again", Name: "Self", Rect: design.Rect{W: 10, H: 4}}},
	}}}
	cat, _ := cupp.Extend(standard.Default(), []cupp.Pack{loop})
	doc := design.NewDocument("t", 20, 6)
	doc.Add(design.Node{Component: "loop.again", Name: "Again", Rect: design.Rect{W: 10, H: 4}})
	Render(doc, cat) // must return
}

func TestAMissingPackShowsAPlaceholder(t *testing.T) {
	doc := design.NewDocument("t", 30, 8)
	doc.Add(design.Node{Component: "gone.widget", Name: "Widget", Rect: design.Rect{W: 20, H: 4}})
	if out := rows(doc, standard.Default()); !strings.Contains(out, "Widget") {
		t.Fatalf("the placeholder carries the node's name:\n%s", out)
	}
}
