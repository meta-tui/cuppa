package scene

import (
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// The painters in this file draw components of libraries Cuppa does not link
// against, from what each library's README shows. The exported program carries
// the same drawings (libs/export/gosource/community_widgets_go.txt, kept in
// step by TestWidgetsDrawTheSameInTheDesignerAndTheProgram).

// paintDropdown draws bubble-dropdown's "[ Label ▼ ]" trigger and, when open,
// its panel below: a rounded box with the options, the chosen one inverted.
func paintDropdown(g *grid.Grid, p Props) {
	accent := p.Str("color")
	options := p.List("options")
	sel := p.Int("selected", -1)
	label, labelStyle := p.Str("placeholder"), p.Dim()
	if sel >= 0 && sel < len(options) {
		label, labelStyle = options[sel], grid.Style{}
	}
	x := g.Text(0, 0, "[ ", grid.Style{}, g.W)
	x += g.Text(x, 0, label, labelStyle, g.W-x)
	x += g.Text(x, 0, " ", grid.Style{}, g.W-x)
	x += g.Text(x, 0, "▼", grid.Style{Fg: accent, Bold: true}, g.W-x)
	g.Text(x, 0, " ]", grid.Style{}, g.W-x)
	if !p.Bool("open") || g.H < 3 || len(options) == 0 {
		return
	}
	longest := 0
	for _, o := range options {
		longest = max(longest, len([]rune(o)))
	}
	triggerW := len([]rune(label)) + 6
	w := min(g.W, max(triggerW, longest+4))
	rows := min(len(options), g.H-3)
	if rows < 1 {
		return
	}
	start := 0
	if sel >= rows {
		start = min(sel-rows+1, len(options)-rows)
	}
	g.Box(design.Rect{X: 0, Y: 1, W: w, H: rows + 2}, grid.BorderNamed("rounded"), grid.Style{Fg: accent})
	for i := 0; i < rows; i++ {
		style := grid.Style{}
		if start+i == sel {
			style = selected(accent)
		}
		g.Fill(design.Rect{X: 1, Y: 2 + i, W: w - 2, H: 1}, ' ', style)
		g.Text(2, 2+i, options[start+i], style, w-4)
	}
}

// paintPromptInput draws promptkit's text input: the prompt, the value (or the
// placeholder) and the cursor, and under it the validation error.
func paintPromptInput(g *grid.Grid, p Props) {
	accent := p.Str("color")
	x := g.Text(0, 0, p.Str("prompt"), grid.Style{Fg: accent, Bold: true}, g.W)
	x += g.Text(x, 0, " ", grid.Style{}, g.W-x)
	cursor := grid.Style{Fg: accent}
	value := p.Str("value")
	switch {
	case value == "":
		x += g.Text(x, 0, "█", cursor, g.W-x)
		g.Text(x, 0, p.Str("placeholder"), p.Dim(), g.W-x)
	case p.Bool("hidden"):
		x += g.Text(x, 0, strings.Repeat("•", len([]rune(value))), grid.Style{}, g.W-x)
		g.Text(x, 0, "█", cursor, g.W-x)
	default:
		x += g.Text(x, 0, value, grid.Style{}, g.W-x)
		g.Text(x, 0, "█", cursor, g.W-x)
	}
	if e := p.Str("error"); e != "" && g.H >= 2 {
		g.Text(0, 1, "Error: "+e, grid.Style{Fg: "196"}, g.W)
	}
}

// paintPromptSelect draws promptkit's selection prompt: the prompt with the
// filter after it, then the choices that match, a cursor on the chosen one and
// arrows where more choices are out of view.
func paintPromptSelect(g *grid.Grid, p Props) {
	accent := p.Str("color")
	x := g.Text(0, 0, p.Str("prompt"), grid.Style{Fg: accent, Bold: true}, g.W)
	filter := p.Str("filter")
	if filter != "" {
		x += g.Text(x, 0, " "+filter, grid.Style{}, g.W-x)
		g.Text(x, 0, "█", grid.Style{Fg: accent}, g.W-x)
	}
	var choices []string
	for _, c := range p.List("choices") {
		if filter == "" || strings.Contains(strings.ToLower(c), strings.ToLower(filter)) {
			choices = append(choices, c)
		}
	}
	rows := g.H - 1
	if rows < 1 || len(choices) == 0 {
		return
	}
	sel := min(max(p.Int("selected", 0), 0), len(choices)-1)
	start := 0
	if len(choices) > rows {
		start = min(max(sel-rows/2, 0), len(choices)-rows)
	}
	end := min(start+rows, len(choices))
	for i := 0; start+i < end; i++ {
		idx := start + i
		prefix, style := "  ", grid.Style{}
		switch {
		case idx == sel:
			prefix, style = "▸ ", grid.Style{Fg: accent, Bold: true}
		case i == 0 && start > 0:
			prefix = "⇡ "
		case start+i == end-1 && end < len(choices):
			prefix = "⇣ "
		}
		cx := g.Text(0, 1+i, prefix, style, g.W)
		g.Text(cx, 1+i, choices[idx], style, g.W-cx)
	}
}

// paintDataTree draws bubble-data-tree: a data structure as a tree of
// "key: value" lines, nested by two spaces of indentation.
func paintDataTree(g *grid.Grid, p Props) {
	accent := p.Str("color")
	lines := pipeLines(p.Str("data"))
	depths := make([]int, len(lines))
	texts := make([]string, len(lines))
	for i, l := range lines {
		trimmed := strings.TrimLeft(l, " ")
		depths[i], texts[i] = (len(l)-len(trimmed))/2, trimmed
	}
	hasLater := func(i int) bool {
		for j := i + 1; j < len(lines); j++ {
			if depths[j] == depths[i] {
				return true
			}
			if depths[j] < depths[i] {
				return false
			}
		}
		return false
	}
	for i := range lines {
		if i >= g.H {
			break
		}
		x := 0
		for level := 1; level < depths[i]; level++ {
			ancestor := -1
			for k := i - 1; k >= 0; k-- {
				if depths[k] <= level {
					if depths[k] == level {
						ancestor = k
					}
					break
				}
			}
			mark := "   "
			if ancestor >= 0 && hasLater(ancestor) {
				mark = "│  "
			}
			x += g.Text(x, i, mark, p.Dim(), g.W-x)
		}
		if depths[i] > 0 {
			branch := "├─ "
			if !hasLater(i) {
				branch = "└─ "
			}
			x += g.Text(x, i, branch, p.Dim(), g.W-x)
		}
		key, value, found := strings.Cut(texts[i], ": ")
		if !found {
			g.Text(x, i, texts[i], grid.Style{Fg: accent, Bold: true}, g.W-x)
			continue
		}
		x += g.Text(x, i, key, grid.Style{Fg: accent}, g.W-x)
		x += g.Text(x, i, ": ", p.Dim(), g.W-x)
		g.Text(x, i, value, grid.Style{}, g.W-x)
	}
}

// paintPDFView draws ntcharts-pdf: a light page with the text of the document
// (text mode) or the same lines greeked into bars (image mode), and a status
// line with the file, the page and the mode.
func paintPDFView(g *grid.Grid, p Props) {
	if g.H < 2 {
		g.Text(0, 0, p.Str("path"), p.Dim(), g.W)
		return
	}
	page := grid.Style{Fg: "234", Bg: "255"}
	g.Fill(design.Rect{W: g.W, H: g.H - 1}, ' ', page)
	image := p.Str("mode") == "image"
	for i, l := range pipeLines(p.Str("text")) {
		if i >= g.H-1 {
			break
		}
		if image {
			l = strings.Map(func(r rune) rune {
				if r == ' ' {
					return ' '
				}
				return '▬'
			}, l)
		}
		g.Text(2, i, l, page, g.W-3)
	}
	status := " " + p.Str("path") + " · page " + strconv.Itoa(p.Int("page", 1)) + "/" + strconv.Itoa(p.Int("pages", 1)) + " · " + p.Str("mode") + " mode"
	g.Text(0, g.H-1, status, p.Dim(), g.W)
}

// wave3D is the height pattern the 3D chart's series follow.
var wave3D = [...]int{2, 3, 4, 6, 5, 4, 3, 2}

// paintChart3D draws ntcharts3d: three axes (z up, x across, y into the
// screen), the series of the chosen kind in the plot, a title and a legend.
func paintChart3D(g *grid.Grid, p Props) {
	accent := p.Str("color")
	top := 0
	if t := p.Str("title"); t != "" && g.H >= 6 {
		g.Text(0, 0, t, grid.Style{Bold: true}, g.W)
		top = 1
	}
	kind := p.Str("kind")
	if p.Bool("legend") && g.W >= 24 {
		label := "● " + kind
		g.Text(g.W-len([]rune(label)), top, label, grid.Style{Fg: accent}, g.W)
	}
	floor := g.H - 3
	if floor <= top+1 || g.W < 8 {
		g.Text(0, 0, "3D chart", p.Dim(), g.W)
		return
	}
	axis := p.Dim()
	for y := top + 1; y < floor; y++ {
		g.Set(2, y, grid.Cell{Ch: '│', Style: axis})
	}
	g.Set(2, floor, grid.Cell{Ch: '└', Style: axis})
	for x := 3; x < g.W-1; x++ {
		g.Set(x, floor, grid.Cell{Ch: '─', Style: axis})
	}
	g.Set(1, floor+1, grid.Cell{Ch: '╱', Style: axis})
	g.Set(0, floor+2, grid.Cell{Ch: 'y', Style: axis})
	g.Set(3, top, grid.Cell{Ch: 'z', Style: axis})
	g.Set(g.W-1, floor, grid.Cell{Ch: 'x', Style: axis})

	left, right := 4, g.W-2
	height := floor - top - 1
	series := grid.Style{Fg: accent}
	column := func(x int) int { return wave3D[(x/2)%len(wave3D)] * height / 8 }
	switch kind {
	case "scatter":
		for i := 0; i < (right-left)*height/10+3; i++ {
			x := left + (i*37)%(right-left)
			y := floor - 1 - (i*53)%height
			ch := '●'
			if i%3 == 0 {
				ch = '·'
			}
			g.Set(x, y, grid.Cell{Ch: ch, Style: series})
		}
	case "bar":
		for x := left; x+2 < right; x += 4 {
			h := max(wave3D[(x/4)%len(wave3D)]*height/7, 1)
			for k := 0; k < h; k++ {
				g.Set(x, floor-1-k, grid.Cell{Ch: '█', Style: series})
				g.Set(x+1, floor-1-k, grid.Cell{Ch: '█', Style: series})
				g.Set(x+2, floor-1-k, grid.Cell{Ch: '▒', Style: series})
			}
		}
	case "line":
		prev := -1
		for x := left; x < right; x++ {
			y := floor - 1 - column(x)
			if prev >= 0 {
				for yy := min(prev, y); yy <= max(prev, y); yy++ {
					g.Set(x, yy, grid.Cell{Ch: '│', Style: series})
				}
			}
			g.Set(x, y, grid.Cell{Ch: '●', Style: series})
			prev = y
		}
	case "vector":
		arrows := []rune("→↗↑↖←↙↓↘")
		for y := top + 1; y < floor; y += 2 {
			for x := left; x < right; x += 4 {
				g.Set(x, y, grid.Cell{Ch: arrows[(x*3+y*5)%len(arrows)], Style: series})
			}
		}
	default: // surface
		for x := left; x < right; x++ {
			h := max(column(x), 1)
			for k := 0; k < h; k++ {
				ch := '░'
				if k == h-1 {
					ch = '▓'
				}
				g.Set(x, floor-1-k, grid.Cell{Ch: ch, Style: series})
			}
		}
	}
}
