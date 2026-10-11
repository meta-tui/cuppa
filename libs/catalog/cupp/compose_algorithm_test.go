package cupp

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestSlugMakesAnIDFromAName(t *testing.T) {
	for in, want := range map[string]string{
		"Tea Card":         "tea-card",
		"  --Hello, World": "hello-world",
		"100% sure":        "100-sure",
		"???":              "component",
		"":                 "component",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
		if !design.ValidID(Slug(in)) {
			t.Errorf("Slug(%q) must be a valid id", in)
		}
	}
}

func TestAGroupBecomesAComponentThatExposesItsText(t *testing.T) {
	g := design.Node{
		ID: "g", Component: design.GroupComponent, Rect: design.Rect{W: 40, H: 10}, BaseW: 20, BaseH: 5,
		Children: []design.Node{
			{ID: "a", Component: "lipgloss.box", Name: "Frame", Rect: design.Rect{W: 20, H: 5}},
			{ID: "b", Component: "lipgloss.label", Name: "Title", Rect: design.Rect{X: 2, Y: 1, W: 10, H: 1}, Props: map[string]string{"text": "Matcha"}},
		},
	}
	c := FromGroup(g, "Tea Card", standard.Default())
	if err := c.Check(); err != nil {
		t.Fatal(err)
	}
	if c.ID != "tea-card" || c.W != 40 || c.H != 10 {
		t.Fatalf("component = %+v", c)
	}
	if c.Nodes[1].Rect.X != 4 || c.Nodes[1].Rect.W != 20 {
		t.Fatalf("parts are scaled to the group's size: %+v", c.Nodes[1].Rect)
	}
	var title *design.Exposed
	for i := range c.Props {
		if c.Props[i].Target == "b" && c.Props[i].TargetProp == "text" {
			title = &c.Props[i]
		}
	}
	if title == nil || title.Default != "Matcha" || title.Kind != "text" {
		t.Fatalf("the label's text is exposed with its current value: %+v", c.Props)
	}
	if g.Children[1].Rect.X != 2 {
		t.Fatal("the group itself is untouched")
	}
}

func TestAddComponentFindsAFreeID(t *testing.T) {
	p := Pack{ID: "mine", Name: "Mine"}
	c := design.Composite{ID: "card", Name: "Card", W: 5, H: 3, Nodes: []design.Node{{ID: "a", Component: "lipgloss.box", Rect: design.Rect{W: 5, H: 3}}}}
	if got := AddComponent(&p, c); got != "card" {
		t.Fatalf("first = %q", got)
	}
	if got := AddComponent(&p, c); got != "card-2" {
		t.Fatalf("second = %q", got)
	}
	if got := AddComponent(&p, c); got != "card-3" {
		t.Fatalf("third = %q", got)
	}
	if err := p.check(); err != nil {
		t.Fatalf("the pack stays valid: %v", err)
	}
}
