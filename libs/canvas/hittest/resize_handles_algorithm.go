package hittest

import "github.com/meta-ui-labs/cuppa/libs/document/design"

// Handle names the part of a selected node's outline that resizes it.
type Handle int

// Resize handles. Corners resize both axes, edges one.
const (
	HandleNone Handle = iota
	HandleLeft
	HandleRight
	HandleTop
	HandleBottom
	HandleTopLeft
	HandleTopRight
	HandleBottomLeft
	HandleBottomRight
)

// HandleAt returns the handle under (x, y) for a node occupying r. The outline
// cells are the handles; a one-cell-high node only has left and right handles
// and a one-cell-wide node only top and bottom, so the rest stays draggable.
func HandleAt(r design.Rect, x, y int) Handle {
	if !r.Contains(x, y) {
		return HandleNone
	}
	l, rt := x == r.X, x == r.Right()-1
	t, b := y == r.Y, y == r.Bottom()-1
	if r.H == 1 {
		t, b = false, false
	}
	if r.W == 1 {
		l, rt = false, false
	}
	switch {
	case t && l:
		return HandleTopLeft
	case t && rt:
		return HandleTopRight
	case b && l:
		return HandleBottomLeft
	case b && rt:
		return HandleBottomRight
	case l:
		return HandleLeft
	case rt:
		return HandleRight
	case t:
		return HandleTop
	case b:
		return HandleBottom
	}
	return HandleNone
}

// Resize returns r after dragging handle h by (dx, dy) cells from where the
// drag started, never smaller than minW x minH. The edge opposite the handle
// stays put.
func Resize(r design.Rect, h Handle, dx, dy, minW, minH int) design.Rect {
	left := h == HandleLeft || h == HandleTopLeft || h == HandleBottomLeft
	right := h == HandleRight || h == HandleTopRight || h == HandleBottomRight
	top := h == HandleTop || h == HandleTopLeft || h == HandleTopRight
	bottom := h == HandleBottom || h == HandleBottomLeft || h == HandleBottomRight
	out := r
	if left {
		out.W = max(r.W-dx, minW)
		out.X = r.Right() - out.W
	}
	if right {
		out.W = max(r.W+dx, minW)
	}
	if top {
		out.H = max(r.H-dy, minH)
		out.Y = r.Bottom() - out.H
	}
	if bottom {
		out.H = max(r.H+dy, minH)
	}
	return out
}
