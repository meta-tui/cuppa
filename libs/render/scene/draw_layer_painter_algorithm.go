package scene

import (
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// paintDrawLayer paints the drawing: only the cells that were painted, so the
// components under it show through everywhere else.
func paintDrawLayer(g *grid.Grid, p Props) {
	for _, c := range drawlayer.Decode(p.Str(drawlayer.PropCells)).Cells() {
		g.Set(c.X, c.Y, grid.Cell{Ch: c.Ch, Style: grid.Style{Fg: c.Fg, Bg: c.Bg}})
	}
}
