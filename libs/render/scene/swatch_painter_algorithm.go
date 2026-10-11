package scene

import (
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// paintSwatch draws a colour as a block with its value and a label, the way a
// list of colours is shown.
func paintSwatch(g *grid.Grid, p Props) {
	text := grid.Style{Fg: p.Str("textColor")}
	x := 0
	if label := p.Str("label"); label != "" {
		w := p.Int("labelWidth", 0)
		if w <= 0 {
			w = len([]rune(label)) + 2
		}
		g.Text(0, 0, label, text, min(w, g.W))
		x = w
	}
	width := max(p.Int("swatch", 2), 1)
	g.Fill(design.Rect{X: x, Y: 0, W: min(width, max(g.W-x, 0)), H: g.H}, ' ', grid.Style{Bg: p.Str("color")})
	if p.Bool("showValue") {
		g.Text(x+width+1, 0, p.Str("color"), text, max(g.W-x-width-1, 0))
	}
}
