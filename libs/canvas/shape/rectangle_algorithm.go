// Package shape turns a drawing gesture (corners of a rectangle, the ends of a
// line, the points of a brush stroke, a word) into painted cells.
package shape

import "github.com/meta-ui-labs/cuppa/libs/document/drawlayer"

// Corner styles of a rectangle.
const (
	CornersSquare = "square"
	CornersRound  = "round"
	CornersAngled = "angled"
)

// Rectangle outlines the box between two corners (either order) in the corner
// style and, when fill is set, paints its inside with that background.
func Rectangle(x0, y0, x1, y1 int, corners, stroke, fill string) []drawlayer.Cell {
	if x1 < x0 {
		x0, x1 = x1, x0
	}
	if y1 < y0 {
		y0, y1 = y1, y0
	}
	tl, tr, bl, br := '┌', '┐', '└', '┘'
	switch corners {
	case CornersRound:
		tl, tr, bl, br = '╭', '╮', '╰', '╯'
	case CornersAngled:
		tl, tr, bl, br = '╱', '╲', '╲', '╱'
	}
	var out []drawlayer.Cell
	put := func(x, y int, ch rune) {
		out = append(out, drawlayer.Cell{X: x, Y: y, Ch: ch, Fg: stroke, Bg: fill})
	}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			onX := x == x0 || x == x1
			onY := y == y0 || y == y1
			switch {
			case onX && onY && x0 == x1 && y0 == y1:
				put(x, y, '▪')
			case onX && onY:
				ch := tl
				switch {
				case x == x1 && y == y0:
					ch = tr
				case x == x0 && y == y1:
					ch = bl
				case x == x1 && y == y1:
					ch = br
				}
				if x0 == x1 {
					ch = '│'
				} else if y0 == y1 {
					ch = '─'
				}
				put(x, y, ch)
			case onY:
				put(x, y, '─')
			case onX:
				put(x, y, '│')
			case fill != "":
				out = append(out, drawlayer.Cell{X: x, Y: y, Ch: ' ', Bg: fill})
			}
		}
	}
	return out
}
