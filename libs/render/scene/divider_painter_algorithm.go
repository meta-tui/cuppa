package scene

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// dividerLines are the characters of each line style: the horizontal and the
// vertical one.
var dividerLines = map[string][2]rune{
	"single": {'─', '│'},
	"double": {'═', '║'},
	"heavy":  {'━', '┃'},
	"dashed": {'╌', '╎'},
	"dotted": {'┈', '┊'},
}

// paintDivider draws a rule across the middle of its box: a horizontal one on
// the middle row, a vertical one down the middle column. A horizontal rule may
// carry a label, which the line breaks around.
//
//	──── Settings ──────────────
func paintDivider(g *grid.Grid, p Props) {
	chars, ok := dividerLines[p.Str("style")]
	if !ok {
		chars = dividerLines["single"]
	}
	style := grid.Style{Fg: p.Str("color")}
	if p.Str("orientation") == "vertical" {
		x := g.W / 2
		for y := 0; y < g.H; y++ {
			g.Set(x, y, grid.Cell{Ch: chars[1], Style: style})
		}
		return
	}
	y := g.H / 2
	line := func(from, to int) {
		for x := from; x < to; x++ {
			g.Set(x, y, grid.Cell{Ch: chars[0], Style: style})
		}
	}
	label := strings.TrimSpace(p.Str("label"))
	if label == "" || g.W < grid.Width(label)+6 {
		line(0, g.W)
		return
	}
	text := " " + label + " "
	w := grid.Width(text)
	at := (g.W - w) / 2
	switch p.Str("align") {
	case "left":
		at = 2
	case "right":
		at = g.W - w - 2
	}
	line(0, at)
	g.Text(at, y, text, grid.Style{Fg: p.Str("color"), Bold: true}, w)
	line(at+w, g.W)
}
