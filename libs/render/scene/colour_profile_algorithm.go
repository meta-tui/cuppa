package scene

import (
	"strconv"

	"github.com/meta-ui-labs/cuppa/libs/color/space"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// applyProfile reduces every colour in g to what the target terminal can show:
// 256 colours, the 16 system colours, or none (bold and faint text only).
func applyProfile(g *grid.Grid, profile string) {
	if profile == "" {
		return
	}
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			g.Restyle(x, y, func(s grid.Style) grid.Style {
				s.Fg, s.Bg = reduce(s.Fg, profile), reduce(s.Bg, profile)
				return s
			})
		}
	}
}

// reduce maps one stored colour ("", "0" to "255" or "#rrggbb") to the profile.
func reduce(c, profile string) string {
	if c == "" || profile == design.ProfileNone {
		return ""
	}
	rgb, ok := space.Resolve(c)
	if !ok {
		return ""
	}
	index, isIndex := -1, false
	if n, err := strconv.Atoi(c); err == nil {
		index, isIndex = n, true
	}
	switch profile {
	case design.Profile256:
		if isIndex {
			return c
		}
		return strconv.Itoa(space.Nearest(rgb))
	case design.Profile16:
		if isIndex && index < 16 {
			return c
		}
		return strconv.Itoa(space.Nearest16(rgb))
	}
	return c
}
