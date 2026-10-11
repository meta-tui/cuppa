package cupp

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func usesTheCard(extra ...design.Node) design.Document {
	doc := design.NewDocument("t", 60, 20)
	doc.Add(design.Node{Component: "tea-shop.card", Name: "Card", Rect: design.Rect{W: 20, H: 5}})
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 5, H: 3}})
	for _, n := range extra {
		doc.Add(n)
	}
	return doc
}

func TestEmbedCopiesOnlyTheCustomComponentsInUse(t *testing.T) {
	reg, _ := Extend(standard.Default(), []Pack{samplePack()})
	doc := Embed(usesTheCard(), reg)
	if len(doc.Embedded) != 1 || doc.Embedded[0].ID != "tea-shop.card" || doc.Embedded[0].Composite.W != 20 {
		t.Fatalf("embedded = %+v", doc.Embedded)
	}
	plain := design.NewDocument("t", 10, 10)
	plain.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 5, H: 3}})
	if got := Embed(plain, reg); len(got.Embedded) != 0 {
		t.Fatal("built-ins are never embedded")
	}
}

func TestEmbedFindsComponentsInsideGroupsAndOtherComponents(t *testing.T) {
	outer := Pack{ID: "outer", Name: "Outer", Components: []design.Composite{{
		ID: "wrap", Name: "Wrap", W: 30, H: 8,
		Nodes: []design.Node{{ID: "x", Component: "tea-shop.card", Name: "Inner card", Rect: design.Rect{W: 30, H: 8}}},
	}}}
	reg, _ := Extend(standard.Default(), []Pack{samplePack(), outer})
	doc := design.NewDocument("t", 60, 20)
	doc.Add(design.Node{Component: design.GroupComponent, Name: "G", Rect: design.Rect{W: 30, H: 8}, BaseW: 30, BaseH: 8,
		Children: []design.Node{{ID: "c", Component: "outer.wrap", Name: "Wrap", Rect: design.Rect{W: 30, H: 8}}}})
	got := Embed(doc, reg)
	if len(got.Embedded) != 2 {
		t.Fatalf("both the wrapper and what it uses travel with the file: %+v", got.Embedded)
	}
}

func TestAdoptMakesAFileDrawWithoutItsPack(t *testing.T) {
	withPack, _ := Extend(standard.Default(), []Pack{samplePack()})
	doc := Embed(usesTheCard(), withPack)

	bare := standard.Default()
	if _, ok := bare.Get("tea-shop.card"); ok {
		t.Fatal("the bare catalog does not know the card")
	}
	adopted := Adopt(bare, doc.Embedded)
	def, ok := adopted.Get("tea-shop.card")
	if !ok || def.Inner == nil || adopted.Title(EmbeddedPackID) != EmbeddedPackName {
		t.Fatalf("the embedded copy is adopted: %+v", def)
	}
	if _, ok := bare.Get("tea-shop.card"); ok {
		t.Fatal("the base is untouched")
	}
	if Adopt(withPack, doc.Embedded) != withPack {
		t.Fatal("an installed pack wins; nothing to adopt")
	}
	bad := []design.Embedded{{ID: "x.y", Composite: design.Composite{ID: "y"}}}
	if Adopt(bare, bad) != bare {
		t.Fatal("an invalid copy is ignored")
	}
}
