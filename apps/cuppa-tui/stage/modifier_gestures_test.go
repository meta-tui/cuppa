package stage

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
)

func dragWith(m *Model, x, y int, shift, alt bool) {
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Move, Held: true, Shift: shift, Alt: alt})
}

func TestShiftKeepsAMoveToAStraightLine(t *testing.T) {
	m, ed, id := setup(t)
	down(m, 15, 8)
	dragWith(m, 30, 9, true, false) // far right, one down: horizontal
	up(m, 30, 9)
	if got := rect(ed, id); got.X != 25 || got.Y != 5 {
		t.Fatalf("rect = %+v, want x 25 and y unchanged at 5", got)
	}
}

func TestShiftMovesOnTheDiagonalAsItLooksOnScreen(t *testing.T) {
	m, ed, id := setup(t)
	down(m, 15, 8)
	dragWith(m, 35, 18, true, false) // 20 across, 10 down: two columns to a row
	up(m, 35, 18)
	if got := rect(ed, id); got.X != 30 || got.Y != 15 {
		t.Fatalf("rect = %+v, want (30,15)", got)
	}
}

func TestShiftResizeKeepsTheProportions(t *testing.T) {
	m, ed, id := setup(t)
	down(m, 15, 8)
	up(m, 15, 8)
	down(m, 33, 10) // bottom-right corner of the 24x6 box
	dragWith(m, 45, 11, true, false)
	up(m, 45, 11)
	got := rect(ed, id)
	if got.W != 36 || got.H != 9 || got.X != 10 || got.Y != 5 {
		t.Fatalf("rect = %+v, want 36x9 at (10,5): the 4:1 shape kept", got)
	}
}

func TestAltResizeGrowsFromTheCentre(t *testing.T) {
	m, ed, id := setup(t)
	down(m, 15, 8)
	up(m, 15, 8)
	down(m, 33, 10)
	dragWith(m, 37, 10, false, true) // 4 to the right: both sides grow by 4
	up(m, 37, 10)
	got := rect(ed, id)
	if got.W != 32 || got.X != 6 || got.H != 6 {
		t.Fatalf("rect = %+v, want width 32 from x 6", got)
	}
}
