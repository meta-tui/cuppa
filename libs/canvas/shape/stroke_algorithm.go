package shape

import "github.com/meta-ui-labs/cuppa/libs/document/drawlayer"

// Ends of a path.
const (
	EndNone   = "none"
	EndArrow  = "arrow"
	EndCircle = "circle"
)

// AutoChar asks Line to pick the line-drawing character for each step.
const AutoChar = "auto"

// LineOptions describe a path.
type LineOptions struct {
	// Char is AutoChar or the character to draw with.
	Char      string
	Thickness int
	Color     string
	End       string
}

// BrushOptions describe a brush or an eraser.
type BrushOptions struct {
	Char      string
	Thickness int
	Color     string
}

// Line draws a path from (x0, y0) to (x1, y1).
func Line(x0, y0, x1, y1 int, o LineOptions) []drawlayer.Cell {
	steps := bresenham(x0, y0, x1, y1)
	var out []drawlayer.Cell
	thick := max(o.Thickness, 1)
	for i, p := range steps {
		ch := []rune(o.Char)
		var r rune
		switch {
		case o.Char != "" && o.Char != AutoChar && len(ch) > 0:
			r = ch[0]
		case thick > 1:
			r = '█'
		default:
			r = autoChar(steps, i)
		}
		out = append(out, stamp(p[0], p[1], thick, r, o.Color)...)
	}
	if len(steps) > 1 {
		last := steps[len(steps)-1]
		dx, dy := last[0]-steps[len(steps)-2][0], last[1]-steps[len(steps)-2][1]
		switch o.End {
		case EndArrow:
			out = append(out, drawlayer.Cell{X: last[0], Y: last[1], Ch: arrowFor(dx, dy), Fg: o.Color})
		case EndCircle:
			out = append(out, drawlayer.Cell{X: last[0], Y: last[1], Ch: '●', Fg: o.Color})
		}
	} else if o.End == EndCircle && len(steps) == 1 {
		out = append(out, drawlayer.Cell{X: x0, Y: y0, Ch: '●', Fg: o.Color})
	}
	return out
}

// Stroke stamps a brush along the points of a drag, joining them.
func Stroke(points [][2]int, o BrushOptions) []drawlayer.Cell {
	ch := []rune(o.Char)
	r := '█'
	if len(ch) > 0 {
		r = ch[0]
	}
	var out []drawlayer.Cell
	for i, p := range points {
		if i == 0 {
			out = append(out, stamp(p[0], p[1], max(o.Thickness, 1), r, o.Color)...)
			continue
		}
		for _, q := range bresenham(points[i-1][0], points[i-1][1], p[0], p[1]) {
			out = append(out, stamp(q[0], q[1], max(o.Thickness, 1), r, o.Color)...)
		}
	}
	return out
}

// Text writes a string from (x, y) to the right.
func Text(x, y int, s, color string) []drawlayer.Cell {
	var out []drawlayer.Cell
	for i, r := range []rune(s) {
		if r == ' ' {
			continue
		}
		out = append(out, drawlayer.Cell{X: x + i, Y: y, Ch: r, Fg: color})
	}
	return out
}

// stamp is one brush dab: a block as wide as the thickness and half as tall
// (cells are about twice as tall as they are wide).
func stamp(x, y, thickness int, r rune, color string) []drawlayer.Cell {
	w, h := thickness, (thickness+1)/2
	var out []drawlayer.Cell
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			out = append(out, drawlayer.Cell{X: x + dx - w/2, Y: y + dy - h/2, Ch: r, Fg: color})
		}
	}
	return out
}

// bresenham lists the cells of the straight line, ends included.
func bresenham(x0, y0, x1, y1 int) [][2]int {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := sign(x1-x0), sign(y1-y0)
	err := dx + dy
	var out [][2]int
	for {
		out = append(out, [2]int{x0, y0})
		if x0 == x1 && y0 == y1 {
			return out
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

// autoChar picks the line character for step i from the way the path goes.
func autoChar(steps [][2]int, i int) rune {
	a, b := steps[max(i-1, 0)], steps[min(i+1, len(steps)-1)]
	dx, dy := b[0]-a[0], b[1]-a[1]
	switch {
	case dy == 0:
		return '─'
	case dx == 0:
		return '│'
	case dx*dy > 0:
		return '╲'
	}
	return '╱'
}

func arrowFor(dx, dy int) rune {
	switch {
	case dx > 0 && dy == 0:
		return '▶'
	case dx < 0 && dy == 0:
		return '◀'
	case dx == 0 && dy < 0:
		return '▲'
	case dx == 0 && dy > 0:
		return '▼'
	case dx > 0 && dy < 0:
		return '↗'
	case dx > 0 && dy > 0:
		return '↘'
	case dx < 0 && dy > 0:
		return '↙'
	}
	return '↖'
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}
