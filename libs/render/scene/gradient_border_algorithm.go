package scene

import (
	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// blendBorder recolours the outline of g from the colour from at the left edge
// to the colour to at the right edge, the way Lip Gloss blends a border. It
// does nothing when either colour is empty or not a colour.
func blendBorder(g *grid.Grid, from, to string) {
	a, okA := space.Resolve(from)
	b, okB := space.Resolve(to)
	if !okA || !okB || g.W < 2 || g.H < 2 {
		return
	}
	for x := 0; x < g.W; x++ {
		t := float64(x) / float64(g.W-1)
		hex := mix(a, b, t).Hex()
		recolour := func(s grid.Style) grid.Style { s.Fg = hex; return s }
		if x == 0 || x == g.W-1 {
			for y := 0; y < g.H; y++ {
				g.Restyle(x, y, recolour)
			}
			continue
		}
		g.Restyle(x, 0, recolour)
		g.Restyle(x, g.H-1, recolour)
	}
}

func mix(a, b space.RGB, t float64) space.RGB {
	lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return space.RGB{R: lerp(a.R, b.R), G: lerp(a.G, b.G), B: lerp(a.B, b.B)}
}
