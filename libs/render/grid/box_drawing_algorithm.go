package grid

import "github.com/meta-ui-labs/cuppa/libs/document/design"

// Border is the set of characters a box is drawn with.
type Border struct {
	TL, TR, BL, BR, H, V rune
}

// BorderNamed returns the border with the given name; unknown names give the
// rounded border. "hidden" draws spaces, keeping the layout of a border.
func BorderNamed(name string) Border {
	switch name {
	case "normal":
		return Border{'┌', '┐', '└', '┘', '─', '│'}
	case "thick":
		return Border{'┏', '┓', '┗', '┛', '━', '┃'}
	case "double":
		return Border{'╔', '╗', '╚', '╝', '═', '║'}
	case "ascii":
		return Border{'+', '+', '+', '+', '-', '|'}
	case "hidden":
		return Border{' ', ' ', ' ', ' ', ' ', ' '}
	}
	return Border{'╭', '╮', '╰', '╯', '─', '│'}
}

// Box draws a border around r and clears its inside with spaces. Rectangles
// smaller than 2x2 are filled instead.
func (g *Grid) Box(r design.Rect, b Border, s Style) {
	g.Fill(r, ' ', s)
	if r.W < 2 || r.H < 2 {
		return
	}
	for x := r.X + 1; x < r.Right()-1; x++ {
		g.Set(x, r.Y, Cell{Ch: b.H, Style: s})
		g.Set(x, r.Bottom()-1, Cell{Ch: b.H, Style: s})
	}
	for y := r.Y + 1; y < r.Bottom()-1; y++ {
		g.Set(r.X, y, Cell{Ch: b.V, Style: s})
		g.Set(r.Right()-1, y, Cell{Ch: b.V, Style: s})
	}
	g.Set(r.X, r.Y, Cell{Ch: b.TL, Style: s})
	g.Set(r.Right()-1, r.Y, Cell{Ch: b.TR, Style: s})
	g.Set(r.X, r.Bottom()-1, Cell{Ch: b.BL, Style: s})
	g.Set(r.Right()-1, r.Bottom()-1, Cell{Ch: b.BR, Style: s})
}
