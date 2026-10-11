package scene

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// huhHead draws the title and the description shared by every form field and
// returns the first free row for the field's own content.
func huhHead(g *grid.Grid, p Props) int {
	y := 0
	if t := p.Str("title"); t != "" {
		g.Text(2, y, t, grid.Style{Fg: p.Str("color"), Bold: true}, g.W-2)
		y++
	}
	if d := p.Str("description"); d != "" && y < g.H {
		g.Text(2, y, d, p.Dim(), g.W-2)
		y++
	}
	return y
}

// huhBars draws the focus bar down the first rows of the field.
func huhBars(g *grid.Grid, p Props, rows int) {
	accent := grid.Style{Fg: p.Str("color")}
	for y := 0; y < rows && y < g.H; y++ {
		g.Set(0, y, grid.Cell{Ch: '┃', Style: accent})
	}
}

// huhFrame is huhHead with the bar down every row, for the fields that fill
// their box.
func huhFrame(g *grid.Grid, p Props) int {
	huhBars(g, p, g.H)
	return huhHead(g, p)
}

func paintHuhInput(g *grid.Grid, p Props) {
	y := huhHead(g, p)
	huhBars(g, p, y+1)
	n := g.Text(2, y, "> ", fg(p.Str("color")), g.W-2)
	if v := p.Str("value"); v != "" {
		g.Text(2+n, y, v, grid.Style{}, g.W-2-n)
		return
	}
	g.Text(2+n, y, p.Str("placeholder"), p.Dim(), g.W-2-n)
}

func paintHuhText(g *grid.Grid, p Props) {
	y := huhHead(g, p)
	huhBars(g, p, y+max(g.H-3, 1))
	lines := pipeLines(p.Str("value"))
	if len(lines) == 0 {
		g.Text(2, y, p.Str("placeholder"), p.Dim(), g.W-2)
		return
	}
	for i, l := range lines {
		if y+i >= g.H {
			break
		}
		g.Text(2, y+i, l, grid.Style{}, g.W-2)
	}
}

func paintHuhSelect(g *grid.Grid, p Props) {
	y := huhHead(g, p)
	opts := p.List("options")
	huhBars(g, p, y+len(opts)+1)
	sel := pick(p, "selected", len(opts))
	for i, o := range opts {
		if y+i >= g.H {
			break
		}
		if i == sel {
			g.Text(2, y+i, "> "+o, grid.Style{Fg: p.Str("color"), Bold: true}, g.W-2)
		} else {
			g.Text(2, y+i, "  "+o, grid.Style{}, g.W-2)
		}
	}
}

func paintHuhMultiSelect(g *grid.Grid, p Props) {
	y := huhHead(g, p)
	opts := p.List("options")
	huhBars(g, p, y+len(opts)+1)
	checked := map[string]bool{}
	for _, c := range p.List("checked") {
		checked[c] = true
	}
	cursor := pick(p, "cursor", len(opts))
	for i, o := range opts {
		if y+i >= g.H {
			break
		}
		mark := "• "
		if checked[strings.TrimSpace(itoa(i))] {
			mark = "✓ "
		}
		prefix, style := "  ", grid.Style{}
		if i == cursor {
			prefix, style = "> ", grid.Style{Fg: p.Str("color"), Bold: true}
		}
		g.Text(2, y+i, prefix+mark+o, style, g.W-2)
	}
}

func paintHuhConfirm(g *grid.Grid, p Props) {
	y := huhHead(g, p)
	if p.Str("description") == "" {
		y++ // the description's row stays, empty
	}
	huhBars(g, p, y+1)
	yes, no := "  "+p.Str("affirmative")+"  ", "  "+p.Str("negative")+"  "
	on, off := selected(p.Str("color")), grid.Style{Fg: "250", Bg: "238"}
	ys, ns := off, on
	if p.Bool("value") {
		ys, ns = on, off
	}
	n := g.Text(2, y, yes, ys, g.W-2)
	g.Text(2+n+1, y, no, ns, g.W-3-n)
}

func paintHuhNote(g *grid.Grid, p Props) {
	lines := wrap(p.Str("body"), g.W-2)
	y := 0
	if t := p.Str("title"); t != "" {
		g.Text(2, y, t, grid.Style{Fg: p.Str("color"), Bold: true}, g.W-2)
		y += 2 // the title, then a blank row
	}
	huhBars(g, p, y+len(lines)+1)
	for i, l := range lines {
		if y+i >= g.H {
			break
		}
		g.Text(2, y+i, l, grid.Style{}, g.W-2)
	}
}

func paintHuhFilePicker(g *grid.Grid, p Props) {
	y := huhFrame(g, p)
	entries := p.List("entries")
	sel := pick(p, "selected", len(entries))
	for i, e := range entries {
		if y+i >= g.H {
			break
		}
		style := grid.Style{}
		if strings.HasSuffix(e, "/") {
			style = grid.Style{Fg: "39"}
		}
		prefix := "  "
		if i == sel {
			prefix, style.Bold = "> ", true
		}
		g.Text(2, y+i, prefix+e, style, g.W-2)
	}
}

func paintHuhForm(g *grid.Grid, p Props) {
	color := p.Str("color")
	g.Box(full(g), grid.BorderNamed("rounded"), fg(color))
	if t := p.Str("title"); t != "" && g.W > 4 {
		g.Text(2, 0, " "+t+" ", grid.Style{Fg: color, Bold: true}, g.W-4)
	}
	rows := (g.H - 3) / 2
	for i := 0; i < rows; i++ {
		y := 1 + i*2
		g.Set(2, y, grid.Cell{Ch: '┃', Style: fg(color)})
		g.Text(4, y, strings.Repeat("▁", max(g.W-8, 0)), p.Dim(), g.W-6)
	}
	if f := p.Str("footer"); f != "" && g.H > 2 {
		g.Text(2, g.H-2, f, p.Dim(), g.W-4)
	}
}

// paintHuhSpinner draws Huh's spinner: one frame and its title on a single line.
func paintHuhSpinner(g *grid.Grid, p Props) {
	frame := spinnerFrames[p.Str("style")]
	if frame == "" {
		frame = "⣾"
	}
	g.Text(0, 0, frame, fg(p.Str("color")), 1)
	g.Text(2, 0, p.Str("title"), grid.Style{}, g.W-2)
}
