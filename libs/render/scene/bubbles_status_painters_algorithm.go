package scene

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// designRow is the full-width one-row rectangle at y.
func designRow(g *grid.Grid, y int) design.Rect { return design.Rect{X: 0, Y: y, W: g.W, H: 1} }

func paintTimer(g *grid.Grid, p Props) {
	paintClock(g, p, p.Str("label"))
}

func paintStopwatch(g *grid.Grid, p Props) {
	paintClock(g, p, "")
}

// paintClock shows an optional label and a time, bright while running.
func paintClock(g *grid.Grid, p Props, label string) {
	x := 0
	if label != "" {
		x = g.Text(0, 0, label+" ", grid.Style{}, g.W)
	}
	style := grid.Style{Fg: p.Str("color"), Bold: true}
	if !p.Bool("running") {
		style = p.Dim()
	}
	g.Text(x, 0, p.Str("value"), style, g.W-x)
}

func paintHelp(g *grid.Grid, p Props) {
	key := grid.Style{Fg: p.Str("color")}
	var pairs [][2]string
	for _, b := range p.List("bindings") {
		k, a, _ := strings.Cut(b, ":")
		pairs = append(pairs, [2]string{k, a})
	}
	if p.Bool("expanded") {
		keyW := 0
		for _, pr := range pairs {
			keyW = max(keyW, len([]rune(pr[0])))
		}
		for i, pr := range pairs {
			if i >= g.H {
				break
			}
			g.Text(0, i, pr[0], key, keyW)
			g.Text(keyW+2, i, pr[1], p.Dim(), g.W-keyW-2)
		}
		return
	}
	x := 0
	for i, pr := range pairs {
		if i > 0 {
			x += g.Text(x, 0, " • ", p.Dim(), g.W-x)
		}
		x += g.Text(x, 0, pr[0], key, g.W-x)
		x += g.Text(x, 0, " "+pr[1], p.Dim(), g.W-x)
	}
}
