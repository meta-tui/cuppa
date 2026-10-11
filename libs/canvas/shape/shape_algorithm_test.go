package shape

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

func at(cells []drawlayer.Cell, x, y int) (drawlayer.Cell, bool) {
	l := drawlayer.New()
	l.Paint(cells...)
	return l.Get(x, y)
}

func TestRectangleCornersFollowTheStyle(t *testing.T) {
	cases := map[string][4]rune{
		CornersSquare: {'┌', '┐', '└', '┘'},
		CornersRound:  {'╭', '╮', '╰', '╯'},
		CornersAngled: {'╱', '╲', '╲', '╱'},
	}
	for style, want := range cases {
		cells := Rectangle(5, 2, 1, 6, style, "212", "")
		for i, p := range [][2]int{{1, 2}, {5, 2}, {1, 6}, {5, 6}} {
			if c, ok := at(cells, p[0], p[1]); !ok || c.Ch != want[i] {
				t.Errorf("%s corner %d = %q, want %q", style, i, c.Ch, want[i])
			}
		}
	}
}

func TestRectangleFillIsBackgroundOnly(t *testing.T) {
	cells := Rectangle(0, 0, 4, 4, CornersSquare, "1", "#102030")
	if c, ok := at(cells, 2, 2); !ok || c.Ch != ' ' || c.Bg != "#102030" {
		t.Fatalf("inside = %+v", c)
	}
	if c, _ := at(cells, 2, 0); c.Ch != '─' || c.Fg != "1" {
		t.Fatalf("top edge = %+v", c)
	}
}

func TestLineAutoCharsAndArrowEnd(t *testing.T) {
	cells := Line(0, 0, 4, 0, LineOptions{Char: AutoChar, Thickness: 1, Color: "5", End: EndArrow})
	if c, _ := at(cells, 2, 0); c.Ch != '─' {
		t.Fatalf("middle = %q, want ─", c.Ch)
	}
	if c, _ := at(cells, 4, 0); c.Ch != '▶' {
		t.Fatalf("end = %q, want ▶", c.Ch)
	}
	down := Line(0, 0, 0, 3, LineOptions{Char: AutoChar, Thickness: 1})
	if c, _ := at(down, 0, 1); c.Ch != '│' {
		t.Fatalf("vertical = %q", c.Ch)
	}
}

func TestStrokeJoinsPointsAndThicknessWidensIt(t *testing.T) {
	cells := Stroke([][2]int{{0, 0}, {6, 0}}, BrushOptions{Char: "#", Thickness: 1})
	for x := 0; x <= 6; x++ {
		if _, ok := at(cells, x, 0); !ok {
			t.Fatalf("gap at x=%d", x)
		}
	}
	thick := Stroke([][2]int{{5, 5}}, BrushOptions{Char: "#", Thickness: 4})
	l := drawlayer.New()
	l.Paint(thick...)
	if l.Len() != 8 { // 4 wide, 2 tall
		t.Fatalf("a thickness 4 dab has %d cells, want 8", l.Len())
	}
}

func TestTextSkipsSpaces(t *testing.T) {
	cells := Text(2, 1, "a b", "3")
	if len(cells) != 2 {
		t.Fatalf("%d cells, want 2", len(cells))
	}
	if c, _ := at(cells, 4, 1); c.Ch != 'b' {
		t.Fatalf("got %q", c.Ch)
	}
}
