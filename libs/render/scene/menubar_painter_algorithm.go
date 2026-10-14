package scene

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/catalog/menuvalue"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// menuPadX is the space at each side of an entry, inside the box.
const menuPadX = 2

// paintMenuBar draws the menu bar: the title and the menu names on the first
// row and, when a menu is shown open, its box under its name, as the menubar
// component draws it.
//
//	 File  Edit  Help
//	╭───────────────────╮
//	│  New       Ctrl+N │
//	│  Open…     Ctrl+O │
//	│───────────────────│
//	│  Quit      Ctrl+Q │
//	╰───────────────────╯
func paintMenuBar(g *grid.Grid, p Props) {
	menus, err := menuvalue.Parse(p.Str("menus"))
	if err != nil || g.W < 1 || g.H < 1 {
		paintGeneric(g, "Menu bar")
		return
	}
	accent := grid.Style{Fg: p.Str("color")}
	selected := grid.Style{Fg: p.Str("color"), Reverse: true}
	muted := grid.Style{Fg: p.Str("mutedColor")}
	rule := grid.Style{Fg: p.Str("ruleColor")}
	open := p.Int("open", -1)

	x := 1
	if title := p.Str("title"); title != "" {
		accent.Bold = true
		x += g.Text(x, 0, title, accent, g.W-x)
		accent.Bold = false
		x += g.Text(x, 0, " ", grid.Style{}, g.W-x)
		x += g.Text(x, 0, "│", rule, g.W-x)
		x += g.Text(x, 0, " ", grid.Style{}, g.W-x)
	}
	starts := make([]int, len(menus))
	for i, m := range menus {
		starts[i] = x
		style := muted
		if i == open {
			style = selected
		}
		x += g.Text(x, 0, " "+m.Name+" ", style, g.W-x) + 1
	}
	if open < 0 || open >= len(menus) || g.H < 2 {
		return
	}

	entries := menus[open].Entries
	inner := 0
	for _, e := range entries {
		w := grid.Width(e.Name)
		if e.Shortcut != "" {
			w += 3 + grid.Width(e.Shortcut)
		}
		inner = max(inner, w)
	}
	inner += 2 * menuPadX
	left := min(starts[open], max(g.W-inner-2, 0))
	line := func(y int, text string, style grid.Style) {
		if y < g.H {
			g.Text(left, y, text, style, g.W-left)
		}
	}
	line(1, "╭"+strings.Repeat("─", inner)+"╮", rule)
	for j, e := range entries {
		y := 2 + j
		if e.Separator {
			line(y, "│"+strings.Repeat("─", inner)+"│", rule)
			continue
		}
		gap := max(inner-2*menuPadX-grid.Width(e.Name)-grid.Width(e.Shortcut), 1)
		pad := strings.Repeat(" ", menuPadX)
		line(y, "│", rule)
		if y >= g.H {
			continue
		}
		used := g.Text(left+1, y, pad+e.Name+strings.Repeat(" ", gap), grid.Style{}, g.W-left-1)
		if e.Shortcut != "" {
			used += g.Text(left+1+used, y, e.Shortcut, muted, g.W-left-1-used)
		}
		g.Text(left+1+used, y, pad, grid.Style{}, g.W-left-1-used)
		line2 := left + 1 + inner
		if line2 < g.W {
			g.Text(line2, y, "│", rule, 1)
		}
	}
	line(2+len(entries), "╰"+strings.Repeat("─", inner)+"╯", rule)
}
