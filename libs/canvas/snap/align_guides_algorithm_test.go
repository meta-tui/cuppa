package snap

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

var canvas = design.Rect{W: 100, H: 40}

func TestSnapsLeftEdgeToNeighbourWithinThreshold(t *testing.T) {
	other := design.Rect{X: 30, Y: 10, W: 10, H: 4}
	moving := design.Rect{X: 29, Y: 0, W: 8, H: 3} // left edge one cell short
	r := Snap(moving, []design.Rect{other}, canvas, 2)
	if r.DX != 1 {
		t.Fatalf("DX = %d", r.DX)
	}
	found := false
	for _, g := range r.Guides {
		if g.Vertical && g.Pos == 30 {
			found = true
		}
	}
	if !found {
		t.Fatalf("no vertical guide at 30: %+v", r.Guides)
	}
}

func TestOutsideThresholdDoesNothing(t *testing.T) {
	r := Snap(design.Rect{X: 20, Y: 20, W: 7, H: 5}, []design.Rect{{X: 50, Y: 30, W: 9, H: 9}}, design.Rect{X: 1000, Y: 1000, W: 1, H: 1}, 2)
	if r.DX != 0 || r.DY != 0 || len(r.Guides) != 0 {
		t.Fatalf("unexpected snap: %+v", r)
	}
}

func TestSnapsToCanvasEdgeAndCentre(t *testing.T) {
	r := Snap(design.Rect{X: 1, Y: 18, W: 6, H: 3}, nil, canvas, 2)
	if r.DX != -1 { // left edge to canvas column 0
		t.Fatalf("DX = %d", r.DX)
	}
	// Vertical centre of the moving rect (row 19) is 1 from the canvas centre (row 19).
	if r.DY != 0 || len(r.Guides) < 2 {
		t.Fatalf("res = %+v", r)
	}
}

func TestAxesSnapIndependentlyAndPreferTheNearest(t *testing.T) {
	far := design.Rect{X: 1000, Y: 1000, W: 1, H: 1}
	others := []design.Rect{{X: 50, Y: 5, W: 10, H: 10}}
	// X lines are 51,51,52: one cell from 50, two from 54 -> the nearest wins.
	// Y lines are 5,6,8: the top already matches the neighbour's top.
	r := Snap(design.Rect{X: 51, Y: 5, W: 2, H: 4}, others, far, 2)
	if r.DX != -1 || r.DY != 0 {
		t.Fatalf("res = %+v", r)
	}
}
