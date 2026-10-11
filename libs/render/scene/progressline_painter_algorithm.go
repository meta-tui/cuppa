package scene

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// orColour is the colour when it is set, else the fallback.
func orColour(c, fallback string) string {
	if strings.TrimSpace(c) == "" {
		return fallback
	}
	return c
}

// paintProgressLine draws a progress bar followed by a text: barWidth cells of
// the full rune for the share done (rounded down) and of the empty rune for
// the rest, the gap, then the suffix. A barWidth of 0 takes what the suffix
// leaves of the width.
func paintProgressLine(g *grid.Grid, p Props) {
	suffix := p.Str("suffix")
	gap := min(max(p.Int("gap", 3), 0), 16)
	barWidth := p.Int("barWidth", 32)
	if barWidth <= 0 {
		barWidth = g.W - gap - grid.Width(suffix)
	}
	barWidth = min(max(barWidth, 0), g.W)
	value := p.Float("value", 0)
	value = min(max(value, 0), 100)
	done := int(value / 100 * float64(barWidth))
	full := firstRune(p.Str("full"), '█')
	empty := firstRune(p.Str("empty"), '░')
	fullStyle := grid.Style{Fg: p.Str("fullColor")}
	emptyStyle := grid.Style{Fg: p.Str("emptyColor")}
	for i := 0; i < barWidth; i++ {
		if i < done {
			g.Set(i, 0, grid.Cell{Ch: full, Style: fullStyle})
		} else {
			g.Set(i, 0, grid.Cell{Ch: empty, Style: emptyStyle})
		}
	}
	x := barWidth + gap
	if x < g.W {
		g.Text(x, 0, suffix, grid.Style{Fg: p.Str("suffixColor")}, g.W-x)
	}
}

func firstRune(s string, fallback rune) rune {
	if r := []rune(s); len(r) > 0 {
		return r[0]
	}
	return fallback
}
