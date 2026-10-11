package scene

import (
	"fmt"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// paintBubbleTableCommunity draws Evertras' bubble table: a heavy frame, equal
// columns with right-aligned cells, the header, and the footer in its own row.
func paintBubbleTableCommunity(g *grid.Grid, p Props) {
	cols := p.List("columns")
	var rows [][]string
	for _, r := range splitRaw(p.Str("rows"), ";") {
		rows = append(rows, splitList(r, ","))
	}
	n := len(cols)
	if n == 0 || g.W < n+2 {
		paintGeneric(g, "Table")
		return
	}
	each := max((g.W-n-1)/n-2, 3)
	style := fg(p.Str("color"))
	rule := func(y int, l, m, r string) {
		x := 0
		put := func(s string) {
			for _, ch := range s {
				g.Set(x, y, grid.Cell{Ch: ch, Style: style})
				x++
			}
		}
		put(l)
		for i := 0; i < n; i++ {
			if i > 0 {
				put(m)
			}
			put(strings.Repeat("━", each))
		}
		put(r)
	}
	row := func(y int, cells []string, s grid.Style) {
		x := 0
		g.Set(x, y, grid.Cell{Ch: '┃', Style: style})
		x++
		for i := 0; i < n; i++ {
			text := ""
			if i < len(cells) {
				text = cells[i]
			}
			r := []rune(text)
			if len(r) > each {
				r = r[:each]
			}
			g.Text(x+each-len(r), y, string(r), s, each)
			x += each
			g.Set(x, y, grid.Cell{Ch: '┃', Style: style})
			x++
		}
	}
	y := 0
	rule(y, "┏", "┳", "┓")
	y++
	row(y, cols, bold)
	y++
	rule(y, "┣", "╋", "┫")
	y++
	for _, r := range rows {
		if y >= g.H-1 {
			break
		}
		row(y, r, grid.Style{})
		y++
	}
	footer := p.Str("footer")
	if footer == "" {
		if y < g.H {
			rule(y, "┗", "┻", "┛")
		}
		return
	}
	if y < g.H {
		rule(y, "┣", "┻", "┫")
		y++
	}
	if y < g.H {
		width := n*each + n - 1
		row2 := []rune(footer)
		if len(row2) > width {
			row2 = row2[:width]
		}
		g.Set(0, y, grid.Cell{Ch: '┃', Style: style})
		g.Text(1+width-len(row2), y, string(row2), grid.Style{}, width)
		g.Set(width+1, y, grid.Cell{Ch: '┃', Style: style})
		y++
	}
	if y < g.H {
		x := 0
		g.Set(x, y, grid.Cell{Ch: '┗', Style: style})
		for i := 1; i <= n*each+n-1; i++ {
			g.Set(i, y, grid.Cell{Ch: '━', Style: style})
		}
		g.Set(n*each+n, y, grid.Cell{Ch: '┛', Style: style})
	}
}

// cells splits length into n slots separated by one-cell gaps and returns the
// start and size of slot i.
func slot(length, n, i int) (start, size int) {
	size = max((length-(n-1))/n, 1)
	return i * (size + 1), size
}

func paintFlexBox(g *grid.Grid, p Props) {
	rows, cols := max(p.Int("rows", 1), 1), max(p.Int("columns", 1), 1)
	style := fg(p.Str("color"))
	for r := 0; r < rows; r++ {
		y, h := slot(g.H, rows, r)
		for c := 0; c < cols; c++ {
			x, w := slot(g.W, cols, c)
			g.Box(design.Rect{X: x, Y: y, W: w, H: h}, grid.BorderNamed("rounded"), style)
			if h >= 3 && w >= 5 {
				g.Text(x+2, y+h/2, fmt.Sprintf("%d,%d", r+1, c+1), p.Dim(), w-3)
			}
		}
	}
}

func paintBoxer(g *grid.Grid, p Props) {
	n := max(p.Int("children", 2), 1)
	style := fg(p.Str("color"))
	vertical := p.Str("orientation") == "vertical"
	for i := 0; i < n; i++ {
		var r design.Rect
		if vertical {
			y, h := slot(g.H, n, i)
			r = design.Rect{Y: y, W: g.W, H: h}
		} else {
			x, w := slot(g.W, n, i)
			r = design.Rect{X: x, W: w, H: g.H}
		}
		g.Box(r, grid.BorderNamed("normal"), style)
		if r.W >= 9 && r.H >= 3 {
			g.Text(r.X+2, r.Y+r.H/2, fmt.Sprintf("model %d", i+1), p.Dim(), r.W-3)
		}
	}
}

func paintDatePicker(g *grid.Grid, p Props) {
	accent := grid.Style{Fg: p.Str("color"), Bold: true}
	month := p.Str("month")
	g.Text(0, 0, "◂", accent, 1)
	g.Text((g.W-len([]rune(month)))/2, 0, month, bold, g.W)
	g.Text(g.W-1, 0, "▸", accent, 1)
	g.Text(0, 1, "Su Mo Tu We Th Fr Sa", p.Dim(), g.W)
	const firstWeekday = 4 // 1 October 2026 is a Thursday.
	day := min(max(p.Int("day", 1), 1), 31)
	for d := 1; d <= 31; d++ {
		slotN := firstWeekday + d - 1
		x, y := (slotN%7)*3, 2+slotN/7
		s := grid.Style{}
		if d == day {
			s = selected(p.Str("color"))
		}
		g.Text(x, y, fmt.Sprintf("%2d", d), s, 2)
	}
}

func paintOverlay(g *grid.Grid, p Props) {
	color := p.Str("color")
	g.Box(full(g), grid.BorderNamed("thick"), fg(color))
	if g.W > 4 {
		g.Text(2, 0, " "+p.Str("title")+" ", grid.Style{Fg: color, Bold: true}, g.W-4)
	}
	for i, l := range wrap(p.Str("body"), g.W-4) {
		if 2+i >= g.H-2 {
			break
		}
		g.Text(2, 1+i, l, grid.Style{}, g.W-4)
	}
	if g.H >= 5 {
		n := g.Text(2, g.H-2, " OK ", selected(color), g.W-4)
		g.Text(3+n, g.H-2, " Cancel ", grid.Style{Fg: "250", Bg: "238"}, g.W-5-n)
	}
}

func paintStatusBar(g *grid.Grid, p Props) {
	g.Fill(full(g), ' ', grid.Style{Bg: "236"})
	x := g.Text(0, 0, " "+p.Str("left")+" ", selected(p.Str("color")), g.W)
	x += g.Text(x, 0, " "+p.Str("middle")+" ", grid.Style{Fg: "252", Bg: "240"}, g.W-x)
	end := " " + p.Str("end") + " "
	right := " " + p.Str("right") + " "
	ew, rw := len([]rune(end)), len([]rune(right))
	if g.W-ew >= x {
		g.Text(g.W-ew, 0, end, grid.Style{Fg: "0", Bg: "250", Bold: true}, ew)
	}
	if g.W-ew-rw >= x {
		g.Text(g.W-ew-rw, 0, right, grid.Style{Fg: "252", Bg: "238"}, rw)
	}
}

func paintFileTree(g *grid.Grid, p Props) {
	entries := p.List("entries")
	sel := pick(p, "selected", len(entries))
	for i, e := range entries {
		if i >= g.H {
			break
		}
		name, style := "  "+e, grid.Style{}
		if strings.HasSuffix(e, "/") {
			name, style = "▸ "+e, grid.Style{Fg: "39"}
		}
		if i == sel {
			style = selected(p.Str("color"))
			g.Fill(designRow(g, i), ' ', style)
		}
		g.Text(0, i, name, style, g.W)
	}
}

// paintTitledFrame draws clambin's frame: the title sits in the top border,
// padded by one cell each side, at the left, centre or right.
func paintTitledFrame(g *grid.Grid, p Props) {
	style := fg(p.Str("color"))
	g.Box(full(g), grid.BorderNamed(p.Str("border")), style)
	title := " " + p.Str("title") + " "
	room := g.W - 2
	if p.Str("title") != "" && len([]rune(title)) <= room {
		x := 1
		switch p.Str("position") {
		case "center":
			x = 1 + (room-len([]rune(title)))/2
		case "right":
			x = g.W - 1 - len([]rune(title))
		}
		g.Text(x, 0, title, grid.Style{Fg: p.Str("color"), Bold: true}, len([]rune(title)))
	}
	for i, l := range wrap(p.Str("content"), g.W-2) {
		if 1+i >= g.H-1 {
			break
		}
		g.Text(1, 1+i, l, grid.Style{}, g.W-2)
	}
}

// paintDialog draws clambin's dialog: a rounded box, the text centred, and a
// row of buttons below it with the active one inverted.
func paintDialog(g *grid.Grid, p Props) {
	g.Box(full(g), grid.BorderNamed("rounded"), fg(p.Str("color")))
	inner := g.W - 2
	buttons := p.List("buttons")
	if g.H >= 5 {
		row := g.H - 3
		active := pick(p, "active", len(buttons))
		total := 0
		for _, b := range buttons {
			total += len([]rune(b)) + 4
		}
		x := 1 + max((inner-total)/2, 0)
		for i, b := range buttons {
			label := "  " + b + "  "
			style := grid.Style{Fg: "255", Bg: "238"}
			if i == active {
				style = grid.Style{Fg: "0", Bg: "255"}
			}
			x += g.Text(x, row, label, style, g.W-1-x)
		}
	}
	for i, l := range wrap(p.Str("text"), inner) {
		y := 2 + i
		if y >= g.H-3 {
			break
		}
		g.Text(1+max((inner-len([]rune(l))+1)/2, 0), y, l, grid.Style{Fg: "255"}, inner)
	}
}

// paintStatusMessage draws clambin's status bar: one line of text styled by
// its level, followed by a spinner frame when asked.
func paintStatusMessage(g *grid.Grid, p Props) {
	style := grid.Style{Fg: "252", Bg: "236"}
	switch p.Str("level") {
	case "warning":
		style = grid.Style{Fg: "0", Bg: "220"}
	case "error":
		style = grid.Style{Fg: "255", Bg: "160"}
	}
	g.Fill(full(g), ' ', style)
	text := " " + p.Str("text")
	if p.Bool("spinner") {
		text += " ⣾"
	}
	g.Text(0, 0, text, style, g.W)
}

// toastKinds are BubbleUp's built-in alert kinds: colour and the two prefix sets.
var toastKinds = map[string]struct{ color, unicode, ascii string }{
	"info":  {"#00ff00", "ⓘ", "(i)"},
	"warn":  {"#ffff00", "⚠", "(!)"},
	"error": {"#ff0000", "✘", "[!!]"},
	"debug": {"#ff00ff", "?", "(?)"},
}

// paintToast draws a BubbleUp alert: a rounded border and the symbol and
// message in the colour of the kind.
func paintToast(g *grid.Grid, p Props) {
	kind, ok := toastKinds[p.Str("kind")]
	if !ok {
		kind = toastKinds["info"]
	}
	prefix := kind.unicode
	if p.Str("symbols") == "ascii" {
		prefix = kind.ascii
	}
	style := fg(kind.color)
	g.Box(full(g), grid.BorderNamed("rounded"), style)
	g.Text(2, g.H/2, prefix+" "+p.Str("message"), style, g.W-4)
}
