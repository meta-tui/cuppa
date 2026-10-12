package preview

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/charmbracelet/x/ansi"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

func fmtSscan(s string, n *int) (int, error) { return fmt.Sscan(strings.TrimSpace(s), n) }

// gridOf paints a styled string (what a Bubbles View returns) onto a grid of
// the given size, keeping colours and bold or faint text. Every cell is
// painted, so the component covers whatever is below it.
func gridOf(view string, w, h int) *grid.Grid {
	buf := uv.NewScreenBuffer(w, h)
	uv.NewStyledString(view).Draw(buf, uv.Rect(0, 0, w, h))
	g := grid.New(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			ch := ' '
			var st grid.Style
			if c := buf.CellAt(x, y); c != nil {
				if r := []rune(c.Content); len(r) > 0 {
					ch = r[0]
				}
				st = styleOf(c.Style)
			}
			g.Set(x, y, grid.Cell{Ch: ch, Style: st})
		}
	}
	return g
}

func styleOf(s uv.Style) grid.Style {
	return grid.Style{
		Fg: colourOf(s.Fg), Bg: colourOf(s.Bg),
		Bold: s.Attrs&uv.AttrBold != 0, Dim: s.Attrs&uv.AttrFaint != 0,
	}
}

// colourOf is the colour as the grid stores it: a palette number or #rrggbb.
func colourOf(c color.Color) string {
	switch v := c.(type) {
	case nil:
		return ""
	case ansi.BasicColor:
		return fmt.Sprint(int(v))
	case ansi.IndexedColor:
		return fmt.Sprint(int(v))
	}
	r, g, b, a := c.RGBA()
	if a == 0 {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", r>>8, g>>8, b>>8)
}
