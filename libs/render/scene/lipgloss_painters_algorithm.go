package scene

import (
	"fmt"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

func full(g *grid.Grid) design.Rect { return design.Rect{W: g.W, H: g.H} }

func fg(color string) grid.Style { return grid.Style{Fg: color} }

var dim = grid.Style{Fg: "240"}

// blurredBox is the box's props as drawn: a box without focus (the "focused"
// prop is false) takes its blurred colour and loses its gradient.
func blurredBox(p Props) Props {
	if p.Str("focused") != "false" {
		return p
	}
	q := make(Props, len(p)+1)
	for k, v := range p {
		q[k] = v
	}
	q["color"] = orColour(p.Str("blurColor"), "240")
	q["gradient"] = ""
	return q
}

func paintBox(g *grid.Grid, p Props) {
	p = blurredBox(p)
	style := fg(p.Str("color"))
	g.Box(full(g), grid.BorderNamed(p.Str("border")), style)
	blendBorder(g, p.Str("color"), p.Str("gradient"))
	if g.W <= 4 {
		return
	}
	// The title sits in the top edge after one border cell; a right label sits
	// before the last one. The title is cut with an ellipsis to leave the
	// label its room.
	title, right := p.Str("title"), p.Str("right")
	maxBlock, rightBlock := g.W-4, ""
	if right != "" {
		if room := g.W - len([]rune(right)) - 7; room >= 4 {
			maxBlock, rightBlock = room, " "+right+" "
		}
	}
	if title != "" && maxBlock >= 4 {
		block := []rune(" " + title + " ")
		if len(block) > maxBlock {
			block = append([]rune(" "), []rune(title)[:maxBlock-3]...)
			block = append(block, '…', ' ')
		}
		g.Text(2, 0, string(block), grid.Style{Fg: orColour(p.Str("titleColor"), p.Str("color")), Bold: p.Bool("titleBold")}, 0)
	}
	if rightBlock != "" {
		g.Text(g.W-2-len([]rune(rightBlock)), 0, rightBlock, grid.Style{Fg: orColour(p.Str("rightColor"), "240")}, 0)
	}
}

func paintLabel(g *grid.Grid, p Props) {
	s := grid.Style{Fg: p.Str("color"), Bg: p.Str("background"), Bold: p.Bool("bold")}
	g.Fill(full(g), ' ', s)
	text := []rune(p.Str("text"))
	if len(text) > g.W {
		text = text[:g.W]
	}
	x := 0
	switch p.Str("align") {
	case "center":
		x = (g.W - len(text)) / 2
	case "right":
		x = g.W - len(text)
	}
	g.Text(x, 0, string(text), s, 0)
}

func paintList(g *grid.Grid, p Props) {
	s := fg(p.Str("color"))
	for i, item := range p.List("items") {
		if i >= g.H {
			break
		}
		g.Text(0, i, marker(p.Str("enumerator"), i)+item, s, g.W)
	}
}

func marker(kind string, i int) string {
	switch kind {
	case "arabic":
		return fmt.Sprintf("%d. ", i+1)
	case "alphabet":
		return fmt.Sprintf("%c. ", 'A'+rune(i%26))
	case "dash":
		return "- "
	case "none":
		return ""
	}
	return "• "
}

func paintTree(g *grid.Grid, p Props) {
	s := fg(p.Str("color"))
	g.Text(0, 0, p.Str("root"), grid.Style{Fg: p.Str("color"), Bold: true}, g.W)
	items := p.List("items")
	for i, item := range items {
		if i+1 >= g.H {
			break
		}
		branch := "├── "
		if i == len(items)-1 {
			branch = "└── "
		}
		g.Text(0, i+1, branch+item, s, g.W)
	}
}

func paintTable(g *grid.Grid, p Props) {
	headers := p.List("headers")
	var rows [][]string
	for _, r := range splitRaw(p.Str("rows"), ";") {
		rows = append(rows, splitList(r, ","))
	}
	cols := len(headers)
	for _, r := range rows {
		cols = max(cols, len(r))
	}
	if cols == 0 {
		paintGeneric(g, "Table")
		return
	}
	widths := make([]int, cols)
	measure := func(cells []string) {
		for i, c := range cells {
			widths[i] = max(widths[i], len([]rune(c)))
		}
	}
	measure(headers)
	for _, r := range rows {
		measure(r)
	}
	bs := fg(p.Str("color"))
	b := grid.BorderNamed(p.Str("border"))
	rule := func(y int, l, m, r rune) {
		x := 0
		g.Set(x, y, grid.Cell{Ch: l, Style: bs})
		x++
		for i, w := range widths {
			for k := 0; k < w+2; k++ {
				g.Set(x, y, grid.Cell{Ch: b.H, Style: bs})
				x++
			}
			if i < cols-1 {
				g.Set(x, y, grid.Cell{Ch: m, Style: bs})
				x++
			}
		}
		g.Set(x, y, grid.Cell{Ch: r, Style: bs})
	}
	line := func(y int, cells []string, s grid.Style) {
		x := 0
		g.Set(x, y, grid.Cell{Ch: b.V, Style: bs})
		x++
		for i, w := range widths {
			text := ""
			if i < len(cells) {
				text = cells[i]
			}
			g.Text(x+1, y, text, s, w)
			x += w + 2
			g.Set(x, y, grid.Cell{Ch: b.V, Style: bs})
			x++
		}
	}
	y := 0
	rule(y, b.TL, '┬', b.TR)
	y++
	if len(headers) > 0 {
		line(y, headers, grid.Style{Bold: true})
		y++
		rule(y, '├', '┼', '┤')
		y++
	}
	for _, r := range rows {
		line(y, r, grid.Style{})
		y++
	}
	rule(y, b.BL, '┴', b.BR)
}

func paintJoinH(g *grid.Grid, p Props) {
	n := max(p.Int("columns", 2), 1)
	g.Box(full(g), grid.BorderNamed("normal"), p.Dim())
	for i := 1; i < n; i++ {
		x := g.W * i / n
		for y := 1; y < g.H-1; y++ {
			g.Set(x, y, grid.Cell{Ch: '┆', Style: p.Dim()})
		}
	}
	g.Text(2, 0, " JoinHorizontal ", p.Dim(), g.W-3)
}

func paintJoinV(g *grid.Grid, p Props) {
	n := max(p.Int("rows", 2), 1)
	g.Box(full(g), grid.BorderNamed("normal"), p.Dim())
	for i := 1; i < n; i++ {
		y := g.H * i / n
		for x := 1; x < g.W-1; x++ {
			g.Set(x, y, grid.Cell{Ch: '┄', Style: p.Dim()})
		}
	}
	g.Text(2, 0, " JoinVertical ", p.Dim(), g.W-3)
}

func paintPlace(g *grid.Grid, p Props) {
	g.Box(full(g), grid.BorderNamed("normal"), p.Dim())
	text := p.Str("text")
	w := len([]rune(text))
	x := 1
	switch p.Str("horizontal") {
	case "center":
		x = (g.W - w) / 2
	case "right":
		x = g.W - w - 1
	}
	y := 1
	switch p.Str("vertical") {
	case "center":
		y = g.H / 2
	case "bottom":
		y = g.H - 2
	}
	g.Text(x, y, text, grid.Style{}, g.W-2)
	g.Text(2, 0, " Place ", p.Dim(), g.W-3)
}

// paintGeneric is the fallback: a p.Dim() box carrying the component name.
func paintGeneric(g *grid.Grid, name string) {
	g.Box(full(g), grid.BorderNamed("normal"), dim)
	if g.H >= 3 {
		g.Text(max((g.W-len([]rune(name)))/2, 1), g.H/2, strings.TrimSpace(name), grid.Style{}, g.W-2)
	} else {
		g.Text(1, 0, name, grid.Style{}, g.W-2)
	}
}
