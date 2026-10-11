package scene

import (
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// paintTabs draws the Lip Gloss "layout" tabs: a row of tabs whose active tab
// opens into the bordered window below it.
//
//	╭─────╮╭───────╮
//	│ One ││ Two   │
//	├─────┘└───────┴──────╮
//	│                     │
//	╰─────────────────────╯
func paintTabs(g *grid.Grid, p Props) {
	if g.H < 4 || g.W < 8 {
		paintGeneric(g, "Tabs")
		return
	}
	names := p.List("tabs")
	if len(names) == 0 {
		names = []string{"Tab"}
	}
	active := min(max(p.Int("active", 1), 1), len(names)) - 1
	edge := fg(p.Str("color"))
	bold := grid.Style{Fg: p.Str("color"), Bold: true}
	put := func(x, y int, ch rune) { g.Set(x, y, grid.Cell{Ch: ch, Style: edge}) }

	// The window: its top edge is the bottom row of the tab strip (row 2).
	for x := 1; x < g.W-1; x++ {
		put(x, 2, '─')
		put(x, g.H-1, '─')
	}
	for y := 3; y < g.H-1; y++ {
		put(0, y, '│')
		put(g.W-1, y, '│')
	}
	put(0, g.H-1, '╰')
	put(g.W-1, g.H-1, '╯')
	put(g.W-1, 2, '╮')
	put(0, 2, '├')

	x := 0
	for i, name := range names {
		label := []rune(" " + name + " ")
		w := len(label) // cells inside the tab
		if x+w+2 > g.W-1 {
			break // the rest do not fit
		}
		for dx := 1; dx <= w; dx++ {
			put(x+dx, 0, '─')
		}
		put(x, 0, '╭')
		put(x+w+1, 0, '╮')
		put(x, 1, '│')
		put(x+w+1, 1, '│')
		style := p.Dim()
		if i == active {
			style = bold
		}
		g.Text(x+1, 1, string(label), style, w)
		if i == active {
			// Open the tab into the window: ┘ spaces └.
			put(x, 2, '┘')
			if x == 0 {
				put(x, 2, '│')
			}
			for dx := 1; dx <= w; dx++ {
				g.Set(x+dx, 2, grid.Cell{Ch: ' '})
			}
			put(x+w+1, 2, '└')
		} else {
			for dx := 1; dx <= w; dx++ {
				put(x+dx, 2, '─')
			}
			put(x+w+1, 2, '┴')
			if x == 0 {
				put(x, 2, '├')
			}
		}
		x += w + 2
	}
}
