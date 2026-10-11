package scene

import (
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// shadowColour is the ANSI 256 grey of a drop shadow.
const shadowColour = "238"

// castShadow draws a one-cell shadow below and to the right of r, on cells
// nothing has painted yet so a shadow never covers another component.
func castShadow(g *grid.Grid, r design.Rect, bg string) {
	shadow := grid.Cell{Ch: '▒', Style: grid.Style{Fg: shadowColour, Bg: bg}}
	for y := r.Y + 1; y < r.Bottom()+1; y++ {
		for x := r.X + 1; x < r.Right()+1; x++ {
			if g.In(x, y) && g.At(x, y).Ch == 0 {
				g.Set(x, y, shadow)
			}
		}
	}
}

// paintBackground gives the document colour to every painted cell that has no
// background of its own, so spaces inside components match the canvas.
func paintBackground(g *grid.Grid, bg string) {
	if bg == "" {
		return
	}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			g.Restyle(x, y, func(s grid.Style) grid.Style {
				if s.Bg == "" {
					s.Bg = bg
				}
				return s
			})
		}
	}
}

// applyEffects dims cells for the scanlines and vignette effects.
func applyEffects(g *grid.Grid, fx design.Effects) {
	if !fx.Scanlines && !fx.Vignette {
		return
	}
	dim := func(s grid.Style) grid.Style { s.Dim = true; return s }
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			scan := fx.Scanlines && y%2 == 1
			edge := fx.Vignette && (y == 0 || y == g.H-1 || x < 2 || x >= g.W-2)
			if scan || edge {
				g.Restyle(x, y, dim)
			}
		}
	}
}
