package hittest

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestNodePicksTopmost(t *testing.T) {
	d := design.NewDocument("t", 40, 10)
	d.Add(design.Node{Rect: design.Rect{W: 10, H: 5}})
	top := d.Add(design.Node{Rect: design.Rect{X: 5, W: 10, H: 5}})
	if id, ok := Node(d, 6, 1); !ok || id != top.ID {
		t.Fatalf("got %v %v", id, ok)
	}
	if id, ok := Node(d, 1, 1); !ok || id != "n1" {
		t.Fatalf("got %v %v", id, ok)
	}
	if _, ok := Node(d, 30, 8); ok {
		t.Fatal("hit on empty canvas")
	}
	if got := NodesIn(d, design.Rect{X: 12, W: 3, H: 1}); len(got) != 1 || got[0] != top.ID {
		t.Fatalf("NodesIn = %v", got)
	}
}

func TestHandleAt(t *testing.T) {
	r := design.Rect{X: 2, Y: 2, W: 6, H: 4}
	cases := []struct {
		x, y int
		want Handle
	}{
		{2, 2, HandleTopLeft}, {7, 2, HandleTopRight}, {2, 5, HandleBottomLeft}, {7, 5, HandleBottomRight},
		{2, 3, HandleLeft}, {7, 3, HandleRight}, {4, 2, HandleTop}, {4, 5, HandleBottom},
		{4, 3, HandleNone}, {0, 0, HandleNone},
	}
	for _, c := range cases {
		if got := HandleAt(r, c.x, c.y); got != c.want {
			t.Errorf("(%d,%d) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
	line := design.Rect{X: 0, Y: 0, W: 8, H: 1}
	if HandleAt(line, 0, 0) != HandleLeft || HandleAt(line, 7, 0) != HandleRight || HandleAt(line, 3, 0) != HandleNone {
		t.Fatal("one-row node handles wrong")
	}
}

func TestResizeKeepsOppositeEdgeAndMinimum(t *testing.T) {
	r := design.Rect{X: 10, Y: 10, W: 8, H: 4}
	if got := Resize(r, HandleTopLeft, -2, -1, 3, 3); got != (design.Rect{X: 8, Y: 9, W: 10, H: 5}) {
		t.Fatalf("grow top-left: %+v", got)
	}
	if got := Resize(r, HandleLeft, 100, 0, 3, 3); got != (design.Rect{X: 15, Y: 10, W: 3, H: 4}) {
		t.Fatalf("shrink past min: %+v", got)
	}
	if got := Resize(r, HandleBottomRight, 2, 3, 3, 3); got != (design.Rect{X: 10, Y: 10, W: 10, H: 7}) {
		t.Fatalf("grow bottom-right: %+v", got)
	}
	if got := Resize(r, HandleBottom, 0, -50, 3, 3); got.H != 3 || got.Y != 10 {
		t.Fatalf("bottom min: %+v", got)
	}
}

func TestHiddenNodesCannotBeHit(t *testing.T) {
	doc := design.NewDocument("t", 50, 20)
	under := doc.Add(design.Node{Name: "under", Rect: design.Rect{X: 0, Y: 0, W: 10, H: 5}})
	over := doc.Add(design.Node{Name: "over", Rect: design.Rect{X: 0, Y: 0, W: 10, H: 5}, Hidden: true})
	if id, ok := Node(doc, 3, 3); !ok || id != under.ID {
		t.Fatalf("the hidden node on top must be skipped, got %v %v", id, ok)
	}
	for _, id := range NodesIn(doc, design.Rect{W: 20, H: 20}) {
		if id == over.ID {
			t.Fatal("a marquee must not select a hidden node")
		}
	}
}
