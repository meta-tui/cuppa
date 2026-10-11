package scene

import (
	"strings"

	"github.com/meta-tui/cuppa/libs/document/container"
	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/render/grid"
)

// containerNode is the container a painter was given: a node the size of the
// grid, which is what container.Layout reads.
func containerNode(component string, g *grid.Grid, p Props) design.Node {
	return design.Node{Component: component, Rect: design.Rect{W: g.W, H: g.H}, Props: p}
}

// paintTabPages draws the bar of a tab container: each page's name, the shown one
// in the accent, and a rule under them. The body is left empty; the members
// of the shown page are drawn over it.
func paintTabPages(g *grid.Grid, p Props) {
	n := containerNode(design.TabsComponent, g, p)
	geo := container.Layout(n)
	names := container.Pages(n)
	active := container.Active(n)
	rule := grid.Style{Fg: p.Str("ruleColor")}
	for x := 0; x < g.W; x++ {
		g.Set(x, 1, grid.Cell{Ch: '─', Style: rule})
	}
	for i, t := range geo.Tabs {
		style := grid.Style{Fg: p.Str("mutedColor")}
		if i == active {
			style = grid.Style{Fg: p.Str("color"), Bold: true}
		}
		g.Text(t.X0, 0, " "+names[t.Page]+" ", style, t.X1-t.X0)
		if i == active {
			for x := t.X0; x < t.X1; x++ {
				g.Set(x, 1, grid.Cell{Ch: '━', Style: grid.Style{Fg: p.Str("color")}})
			}
		}
		if i < len(geo.Tabs)-1 && t.X1 < g.W {
			g.Set(t.X1, 0, grid.Cell{Ch: '│', Style: rule})
		}
	}
}

// paintPanelStack draws the headers of a panel stack: an arrow that says whether
// the section is open, its name and a rule to the right.
func paintPanelStack(g *grid.Grid, p Props) {
	n := containerNode(design.PanelsComponent, g, p)
	names := container.Pages(n)
	for _, h := range container.Layout(n).Heads {
		open := container.Shown(n, h.Page)
		arrow, style := "▸ ", grid.Style{Fg: p.Str("mutedColor")}
		if open {
			arrow, style = "▾ ", grid.Style{Fg: p.Str("color"), Bold: true}
		}
		used := g.Text(0, h.Y, arrow+names[h.Page], style, g.W)
		if used < g.W {
			g.Text(used, h.Y, " "+strings.Repeat("─", max(g.W-used-1, 0)), grid.Style{Fg: p.Str("ruleColor")}, g.W-used)
		}
	}
}
