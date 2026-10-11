package hittest

import (
	"math"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// ResizeModifiers are the keys that change how a drag resizes, as in an image
// editor.
type ResizeModifiers struct {
	// Proportional keeps the width to height ratio of the start.
	Proportional bool
	// FromCenter grows and shrinks the node around its centre, so both
	// opposite edges move.
	FromCenter bool
}

// ResizeWith is Resize with the modifier keys applied. A corner keeps the
// ratio by following whichever of the two sizes changed more; an edge keeps it
// by changing the other size to match and centring the node on the edge's
// axis. From the centre, the pointer's distance counts twice, since both edges
// move.
func ResizeWith(r design.Rect, h Handle, dx, dy, minW, minH int, mods ResizeModifiers) design.Rect {
	if !mods.Proportional && !mods.FromCenter {
		return Resize(r, h, dx, dy, minW, minH)
	}
	left := h == HandleLeft || h == HandleTopLeft || h == HandleBottomLeft
	right := h == HandleRight || h == HandleTopRight || h == HandleBottomRight
	top := h == HandleTop || h == HandleTopLeft || h == HandleTopRight
	bottom := h == HandleBottom || h == HandleBottomLeft || h == HandleBottomRight

	grow := func(on1, on2 bool, d int) int {
		switch {
		case on1:
			return -d
		case on2:
			return d
		}
		return 0
	}
	factor := 1
	if mods.FromCenter {
		factor = 2
	}
	w := r.W + factor*grow(left, right, dx)
	hh := r.H + factor*grow(top, bottom, dy)

	if mods.Proportional && r.W > 0 && r.H > 0 {
		sw, sh := float64(w)/float64(r.W), float64(hh)/float64(r.H)
		var s float64
		switch {
		case (left || right) && !top && !bottom:
			s = sw
		case (top || bottom) && !left && !right:
			s = sh
		case math.Abs(sw-1) >= math.Abs(sh-1):
			s = sw
		default:
			s = sh
		}
		s = math.Max(s, math.Max(float64(minW)/float64(r.W), float64(minH)/float64(r.H)))
		w = max(int(math.Round(float64(r.W)*s)), 1)
		hh = max(int(math.Round(float64(r.H)*s)), 1)
	}
	w, hh = max(w, minW), max(hh, minH)

	out := design.Rect{W: w, H: hh}
	centreX, centreY := 2*r.X+r.W, 2*r.Y+r.H
	switch {
	case mods.FromCenter, !left && !right:
		out.X = (centreX - w) / 2
	case left:
		out.X = r.Right() - w
	default:
		out.X = r.X
	}
	switch {
	case mods.FromCenter, !top && !bottom:
		out.Y = (centreY - hh) / 2
	case top:
		out.Y = r.Bottom() - hh
	default:
		out.Y = r.Y
	}
	return out
}
