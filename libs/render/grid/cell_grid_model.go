// Package grid is a small character-cell canvas that components are painted
// onto and that is turned into an ANSI string with Lip Gloss.
package grid

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Style is the look of one cell. Colors are Lip Gloss color specs such as
// "212" (ANSI 256) or "#ff5faf"; empty means the terminal default.
type Style struct {
	Fg, Bg    string
	Bold, Dim bool
	// Reverse swaps the foreground and the background.
	Reverse bool
}

// Cell is one character with its style. A zero Ch means "never painted".
type Cell struct {
	Ch rune
	Style
}

// Grid is a fixed-size rectangle of cells.
type Grid struct {
	W, H  int
	cells []Cell
}

// New returns a grid of unpainted cells.
func New(w, h int) *Grid {
	w, h = max(w, 0), max(h, 0)
	return &Grid{W: w, H: h, cells: make([]Cell, w*h)}
}

// In reports whether (x, y) is inside the grid.
func (g *Grid) In(x, y int) bool { return x >= 0 && y >= 0 && x < g.W && y < g.H }

// At returns the cell at (x, y), or the zero cell outside the grid.
func (g *Grid) At(x, y int) Cell {
	if !g.In(x, y) {
		return Cell{}
	}
	return g.cells[y*g.W+x]
}

// Set paints one cell; positions outside the grid are ignored.
func (g *Grid) Set(x, y int, c Cell) {
	if g.In(x, y) {
		g.cells[y*g.W+x] = c
	}
}

// Fill paints every cell of r (clipped to the grid) with the same character and style.
func (g *Grid) Fill(r design.Rect, ch rune, s Style) {
	for y := r.Y; y < r.Bottom(); y++ {
		for x := r.X; x < r.Right(); x++ {
			g.Set(x, y, Cell{Ch: ch, Style: s})
		}
	}
}

// wideTail is the second cell of a double-width character: the first cell
// holds the character, this one is drawn as nothing.
const wideTail rune = -1

// RuneWidth is how many terminal cells a rune takes: 0 for a combining mark,
// 2 for a wide character such as a Japanese one, 1 otherwise.
func RuneWidth(r rune) int {
	if r < 0x300 {
		return 1
	}
	return lipgloss.Width(string(r))
}

// Width is how many terminal cells a text takes.
func Width(s string) int { return lipgloss.Width(s) }

// Text paints text starting at (x, y), stopping after maxW cells (no limit if
// maxW <= 0) or at the end of the grid. A double-width character takes two
// cells and is not split. It returns the number of cells written.
func (g *Grid) Text(x, y int, text string, s Style, maxW int) int {
	n := 0
	for _, ch := range text {
		w := RuneWidth(ch)
		if w == 0 {
			continue
		}
		if maxW > 0 && n+w > maxW || w == 2 && x+n+w > g.W {
			break
		}
		g.Set(x+n, y, Cell{Ch: ch, Style: s})
		if w == 2 {
			g.Set(x+n+1, y, Cell{Ch: wideTail, Style: s})
		}
		n += w
	}
	return n
}

// Restyle changes the style of an existing cell without touching its character.
func (g *Grid) Restyle(x, y int, fn func(Style) Style) {
	if g.In(x, y) {
		c := &g.cells[y*g.W+x]
		c.Style = fn(c.Style)
	}
}

// Blit copies src onto g with its top-left at (ox, oy). Unpainted source cells
// are skipped so they do not erase what is below.
func (g *Grid) Blit(src *Grid, ox, oy int) {
	for y := 0; y < src.H; y++ {
		for x := 0; x < src.W; x++ {
			if c := src.At(x, y); c.Ch != 0 {
				g.Set(ox+x, oy+y, c)
			}
		}
	}
}

// Clone returns an independent copy.
func (g *Grid) Clone() *Grid {
	c := &Grid{W: g.W, H: g.H, cells: make([]Cell, len(g.cells))}
	copy(c.cells, g.cells)
	return c
}

// Lines renders each row to an ANSI string. Unpainted cells become spaces.
func (g *Grid) Lines() []string {
	lines := make([]string, g.H)
	for y := 0; y < g.H; y++ {
		lines[y] = g.row(y)
	}
	return lines
}

// String renders the whole grid, rows joined by newlines.
func (g *Grid) String() string { return strings.Join(g.Lines(), "\n") }

// row groups runs of identical style so each run is styled once.
func (g *Grid) row(y int) string {
	var out, run strings.Builder
	var cur Style
	flush := func() {
		if run.Len() > 0 {
			out.WriteString(render(cur, run.String()))
			run.Reset()
		}
	}
	for x := 0; x < g.W; x++ {
		c := g.cells[y*g.W+x]
		if c.Ch == wideTail {
			continue
		}
		if c.Style != cur {
			flush()
			cur = c.Style
		}
		ch := c.Ch
		if ch == 0 {
			ch = ' '
		}
		run.WriteRune(ch)
	}
	flush()
	return out.String()
}

func render(s Style, text string) string {
	if s == (Style{}) {
		return text
	}
	st := lipgloss.NewStyle().Bold(s.Bold).Faint(s.Dim).Reverse(s.Reverse)
	if s.Fg != "" {
		st = st.Foreground(lipgloss.Color(s.Fg))
	}
	if s.Bg != "" {
		st = st.Background(lipgloss.Color(s.Bg))
	}
	return st.Render(text)
}
