package hittest

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestFromCenterMovesBothEdges(t *testing.T) {
	r := design.Rect{X: 10, Y: 10, W: 20, H: 10}
	got := ResizeWith(r, HandleRight, 4, 0, 1, 1, ResizeModifiers{FromCenter: true})
	want := design.Rect{X: 6, Y: 10, W: 28, H: 10}
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestProportionalCornerKeepsTheRatioAndAnchorsTheFarCorner(t *testing.T) {
	r := design.Rect{X: 10, Y: 10, W: 20, H: 10}
	got := ResizeWith(r, HandleBottomRight, 10, 1, 1, 1, ResizeModifiers{Proportional: true})
	if got.W != 30 || got.H != 15 || got.X != 10 || got.Y != 10 {
		t.Fatalf("got %+v, want 30x15 at 10,10", got)
	}
}

func TestProportionalEdgeChangesTheOtherSizeAndCentresOnTheAxis(t *testing.T) {
	r := design.Rect{X: 10, Y: 10, W: 20, H: 10}
	got := ResizeWith(r, HandleRight, 10, 0, 1, 1, ResizeModifiers{Proportional: true})
	if got.W != 30 || got.H != 15 || got.X != 10 || got.Y != 7 {
		t.Fatalf("got %+v, want 30x15 at 10,7 (centred vertically)", got)
	}
}

func TestBothModifiersScaleAroundTheCentre(t *testing.T) {
	r := design.Rect{X: 10, Y: 10, W: 20, H: 10}
	got := ResizeWith(r, HandleBottomRight, 5, 0, 1, 1, ResizeModifiers{Proportional: true, FromCenter: true})
	if got.W != 30 || got.H != 15 {
		t.Fatalf("got %+v, want 30x15", got)
	}
	if got.X != 5 || got.Y != 7 {
		t.Fatalf("got origin %d,%d, want 5,7 (centre kept)", got.X, got.Y)
	}
}

func TestMinimumSizeIsKeptWithTheRatio(t *testing.T) {
	r := design.Rect{X: 0, Y: 0, W: 20, H: 10}
	got := ResizeWith(r, HandleBottomRight, -100, -100, 10, 6, ResizeModifiers{Proportional: true})
	if got.W < 10 || got.H < 6 || got.W != got.H*2 {
		t.Fatalf("got %+v, want at least 10x6 in a 2:1 ratio", got)
	}
}

func TestNoModifiersIsPlainResize(t *testing.T) {
	r := design.Rect{X: 10, Y: 10, W: 20, H: 10}
	if got, want := ResizeWith(r, HandleLeft, -3, 0, 1, 1, ResizeModifiers{}), Resize(r, HandleLeft, -3, 0, 1, 1); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
