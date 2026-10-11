package scene

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestBoxGradientBlendsTheBorderLeftToRight(t *testing.T) {
	cat := standard.Default()
	n := design.Node{Component: "lipgloss.box", Rect: design.Rect{W: 11, H: 3},
		Props: map[string]string{"color": "#000000", "gradient": "#ffffff"}}
	g := RenderNode(n, cat)
	left, mid, right := g.At(0, 0).Fg, g.At(5, 0).Fg, g.At(10, 0).Fg
	if left != "#000000" || right != "#ffffff" {
		t.Fatalf("ends = %q and %q", left, right)
	}
	if mid != "#808080" {
		t.Fatalf("middle = %q, want the halfway grey", mid)
	}
	if g.At(0, 1).Fg != "#000000" || g.At(10, 2).Fg != "#ffffff" {
		t.Fatal("the side edges and the bottom row follow the same blend")
	}
}

func TestBoxWithoutGradientKeepsOneColour(t *testing.T) {
	n := design.Node{Component: "lipgloss.box", Rect: design.Rect{W: 8, H: 3}, Props: map[string]string{"color": "212"}}
	g := RenderNode(n, standard.Default())
	if g.At(0, 0).Fg != "212" || g.At(7, 0).Fg != "212" {
		t.Fatalf("border = %q .. %q", g.At(0, 0).Fg, g.At(7, 0).Fg)
	}
}
