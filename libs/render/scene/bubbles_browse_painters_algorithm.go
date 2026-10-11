package scene

import (
	"fmt"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// paintTextArea draws the text area: the prompt bar down its left edge, the
// line numbers when asked for, and the text or the placeholder.
func paintTextArea(g *grid.Grid, p Props) {
	accent := fg(p.Str("color"))
	lines := pipeLines(p.Str("value"))
	showNumbers := p.Bool("line_numbers")
	x0 := 2
	for y := 0; y < g.H; y++ {
		g.Set(0, y, grid.Cell{Ch: '┃', Style: accent})
	}
	content := lines
	style := grid.Style{}
	if len(content) == 0 {
		content, style = []string{p.Str("placeholder")}, p.Dim()
	}
	for i, l := range content {
		if i >= g.H {
			break
		}
		x := x0
		if showNumbers {
			x += g.Text(x, i, fmt.Sprintf("%3d ", i+1), p.Dim(), g.W-x)
		}
		g.Text(x, i, l, style, g.W-x)
	}
}

// paintBubbleList draws the Bubbles list as it lays itself out: the title, a
// blank row, the status line, a blank row, a page of items, the page dots and
// the key help. A page holds what is left after those rows.
func paintBubbleList(g *grid.Grid, p Props) {
	items := p.List("items")
	accent := p.Str("color")
	y := 0
	if t := p.Str("title"); t != "" {
		g.Text(2, y, " "+t+" ", grid.Style{Fg: "0", Bg: accent, Bold: true}, g.W-2)
		y += 2
	}
	if p.Bool("show_status") {
		g.Text(2, y, fmt.Sprintf("%d items", len(items)), p.Dim(), g.W-2)
		y += 2
	}
	room := max(g.H-y-4, 1)
	sel := pick(p, "selected", len(items))
	start := sel / room * room
	for i := 0; i < room && start+i < len(items); i++ {
		if start+i == sel {
			g.Text(0, y+i, "│ "+items[start+i], grid.Style{Fg: accent, Bold: true}, g.W)
		} else {
			g.Text(2, y+i, items[start+i], grid.Style{}, g.W-2)
		}
	}
	if len(items) > room && g.H >= 3 {
		pages := (len(items) + room - 1) / room
		for i := 0; i < pages && 2+i < g.W; i++ {
			style := p.Dim()
			if i == start/room {
				style = grid.Style{Fg: accent}
			}
			g.Set(2+i, g.H-3, grid.Cell{Ch: '•', Style: style})
		}
	}
	if g.H >= 2 {
		g.Text(2, g.H-1, "↑/k up • ↓/j down • / filter • q quit • ? more", p.Dim(), g.W-2)
	}
}

// paintBubbleTable draws the Bubbles table: equal columns, a space of padding
// each side, the header in bold and the selected row highlighted.
func paintBubbleTable(g *grid.Grid, p Props) {
	cols := p.List("columns")
	var rows [][]string
	for _, r := range splitRaw(p.Str("rows"), ";") {
		rows = append(rows, splitList(r, ","))
	}
	n := len(cols)
	for _, r := range rows {
		n = max(n, len(r))
	}
	if n == 0 {
		paintGeneric(g, "Table")
		return
	}
	each := max((g.W-2*n)/n, 3)
	draw := func(y int, cells []string, s grid.Style) {
		for i := 0; i < n && i < len(cells); i++ {
			g.Text(i*(each+2)+1, y, cells[i], s, each)
		}
	}
	sel := pick(p, "selected", len(rows))
	y := 0
	if len(cols) > 0 {
		draw(y, cols, bold)
		y++
	}
	for i, r := range rows {
		if y >= g.H {
			break
		}
		if i == sel {
			g.Fill(designRow(g, y), ' ', selected(p.Str("color")))
			draw(y, r, selected(p.Str("color")))
		} else {
			draw(y, r, grid.Style{})
		}
		y++
	}
}

func paintViewport(g *grid.Grid, p Props) {
	lines := pipeLines(p.Str("content"))
	room := g.H
	offset := 0
	if len(lines) > room {
		offset = p.Int("percent", 0) * (len(lines) - room) / 100
	}
	textW := g.W
	if p.Bool("show_scrollbar") && g.W > 3 {
		textW = g.W - 2
	}
	for i := 0; i < room && offset+i < len(lines); i++ {
		g.Text(0, i, lines[offset+i], grid.Style{}, textW)
	}
	if textW < g.W {
		track := fg(p.Str("color"))
		thumbH := max(g.H*room/max(len(lines), room), 1)
		thumbY := 0
		if len(lines) > room {
			thumbY = offset * (g.H - thumbH) / (len(lines) - room)
		}
		for y := 0; y < g.H; y++ {
			ch := '│'
			if y >= thumbY && y < thumbY+thumbH {
				ch = '█'
			}
			g.Set(g.W-1, y, grid.Cell{Ch: ch, Style: track})
		}
	}
}

// paintPaginator draws the page dots (a filled dot for this page, hollow for
// the others, with no gaps) or "page/pages".
func paintPaginator(g *grid.Grid, p Props) {
	total := max(p.Int("total", 1), 1)
	page := min(max(p.Int("page", 1), 1), total)
	accent := fg(p.Str("color"))
	if p.Str("style") == "arabic" {
		g.Text(0, 0, fmt.Sprintf("%d/%d", page, total), accent, g.W)
		return
	}
	for i := 1; i <= total && i-1 < g.W; i++ {
		if i == page {
			g.Set(i-1, 0, grid.Cell{Ch: '•', Style: accent})
		} else {
			g.Set(i-1, 0, grid.Cell{Ch: '○', Style: p.Dim()})
		}
	}
}

func paintFilePicker(g *grid.Grid, p Props) {
	accent := grid.Style{Fg: p.Str("color"), Bold: true}
	g.Text(0, 0, p.Str("path"), accent, g.W)
	entries := p.List("entries")
	sel := pick(p, "selected", len(entries))
	for i, e := range entries {
		y := i + 1
		if y >= g.H {
			break
		}
		style := grid.Style{}
		if strings.HasSuffix(e, "/") {
			style = grid.Style{Fg: "39"}
		}
		if i == sel {
			g.Text(0, y, "> ", accent, 2)
			style.Bold = true
		} else {
			g.Text(0, y, "  ", grid.Style{}, 2)
		}
		g.Text(2, y, e, style, g.W-2)
	}
}

// paintBubbleTree draws the Bubbles tree as the component does: a root with
// Lip Gloss connectors, an open marker before every branch, the cursor arrow on
// the selected row, and the key help under the tree.
func paintBubbleTree(g *grid.Grid, p Props) {
	type item struct {
		name   string
		depth  int
		branch bool
		last   bool
	}
	lines := pipeLines(p.Str("items"))
	var items []item
	for _, line := range lines {
		name := strings.TrimLeft(line, " ")
		if name == "" {
			continue
		}
		items = append(items, item{name: name, depth: (len(line) - len(name)) / 2})
	}
	for i := range items {
		items[i].last = true
		for j := i + 1; j < len(items); j++ {
			if items[j].depth < items[i].depth {
				break
			}
			if items[j].depth == items[i].depth {
				items[i].last = false
				break
			}
		}
		items[i].branch = i+1 < len(items) && items[i+1].depth > items[i].depth
	}
	rowsShown := len(items) + 1
	sel := pick(p, "selected", rowsShown)
	accent := p.Str("color")
	cursor := grid.Style{Fg: accent, Bold: true}
	row := func(y int, selectedRow bool, text string, style grid.Style) {
		if y >= g.H {
			return
		}
		if selectedRow {
			g.Text(0, y, "→ ", cursor, g.W)
			style = grid.Style{Fg: accent, Bold: true}
		}
		g.Text(2, y, text, style, g.W-2)
	}
	row(0, sel == 0, "▼ "+p.Str("root"), grid.Style{Fg: "#ee6ff8"})
	var ancestors []bool // whether each ancestor level was the last child
	for i, it := range items {
		ancestors = ancestors[:min(it.depth, len(ancestors))]
		prefix := ""
		for _, last := range ancestors {
			if last {
				prefix += "   "
			} else {
				prefix += "│  "
			}
		}
		connector := "├──"
		if it.last {
			connector = "└──"
		}
		text := prefix + connector
		style := grid.Style{Fg: "#b0b0b0"}
		if it.branch {
			text += "▼ "
			style = grid.Style{Fg: "99"}
		}
		row(i+1, sel == i+1, text+it.name, style)
		ancestors = append(ancestors, it.last)
	}
	if p.Bool("show_help") && g.H >= rowsShown+2 {
		g.Text(0, g.H-1, "↓/j down • ↑/k up • ⏎ toggle • ? more", p.Dim(), g.W)
	}
}
