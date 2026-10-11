package scene

import (
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// rowsColumn is one column of the Rows component: its name, its width (0 for
// the rest of the row) and how its cells are drawn.
type rowsColumn struct {
	name  string
	width int
	// colour makes the cell a block of that colour (a background); glyph makes
	// it full blocks drawn in that colour.
	colour, glyph bool
	// ellipsis cuts a cell that is too long with "…".
	ellipsis bool
	// fg and bg are the colour of the column's text and its background; selBg,
	// selFg and selBold are what the column takes on a selected row. They are
	// colours or @Name of the palette.
	fg, bg, selBg, selFg string
	selBold              bool
	// fit limits the selected background to the cells the text uses, not the
	// column's whole width.
	fit bool
	// fill takes what the other columns leave (shared by the fill columns), and
	// auto is as wide as the text of its own cell in that row: columns after a
	// fill column are therefore flush with the right end, however wide an
	// auto cell is.
	fill, auto bool
	// hasFit sizes the column from its longest cell: at least fitMin, at most
	// fitMax cells, or fitPct percent of the width the fixed columns leave but
	// never below floor (fitMin when 0). hideEmpty gives it no width and no
	// gap when every cell is empty; minWidth is the least a fill column takes.
	hasFit         bool
	fitMin, fitMax int
	fitPct, floor  int
	hideEmpty      bool
	minWidth       int
}

// rowsColumns reads "Name[:width][:token]..." items separated by commas. A
// token is colour, glyph, ellipsis, fg=<colour>, bg=<colour>,
// selbg=<colour>, selfg=<colour>, selbold, fit, fill or auto. A column without a width takes twelve cells, except the
// last, which takes what is left of the row.
func rowsColumns(spec string) []rowsColumn {
	var out []rowsColumn
	for _, item := range rowsSplit(spec, ",", true) {
		parts := strings.Split(strings.TrimSpace(item), ":")
		col := rowsColumn{name: strings.TrimSpace(parts[0])}
		for _, extra := range parts[1:] {
			extra = strings.TrimSpace(extra)
			lower := strings.ToLower(extra)
			if n, err := strconv.Atoi(extra); err == nil {
				col.width = max(n, 0)
				continue
			}
			switch {
			case lower == "colour" || lower == "color":
				col.colour = true
			case lower == "glyph":
				col.glyph = true
			case lower == "ellipsis":
				col.ellipsis = true
			case lower == "fit":
				col.fit = true
			case strings.HasPrefix(lower, "fit="):
				col.hasFit = true
				spec := strings.TrimSpace(extra[4:])
				lo, hi, _ := strings.Cut(spec, "..")
				col.fitMin, _ = strconv.Atoi(strings.TrimSpace(lo))
				hi = strings.TrimSpace(hi)
				if pct, ok := strings.CutSuffix(hi, "%"); ok {
					col.fitPct, _ = strconv.Atoi(strings.TrimSpace(pct))
				} else if hi != "" {
					col.fitMax, _ = strconv.Atoi(hi)
				}
			case strings.HasPrefix(lower, "floor="):
				col.floor, _ = strconv.Atoi(strings.TrimSpace(extra[6:]))
			case strings.HasPrefix(lower, "min="):
				col.minWidth, _ = strconv.Atoi(strings.TrimSpace(extra[4:]))
			case lower == "hide-empty":
				col.hideEmpty = true
			case lower == "fill":
				col.fill = true
			case lower == "auto":
				col.auto = true
			case lower == "selbold":
				col.selBold = true
			case strings.HasPrefix(lower, "selfg="):
				col.selFg = strings.TrimSpace(extra[6:])
			case strings.HasPrefix(lower, "fg="):
				col.fg = strings.TrimSpace(extra[3:])
			case strings.HasPrefix(lower, "bg="):
				col.bg = strings.TrimSpace(extra[3:])
			case strings.HasPrefix(lower, "selbg="):
				col.selBg = strings.TrimSpace(extra[6:])
			}
		}
		out = append(out, col)
	}
	for i := range out {
		if out[i].width == 0 && i < len(out)-1 && !out[i].fill && !out[i].auto && !out[i].hasFit {
			out[i].width = 12
		}
	}
	return out
}

// rowsSplit cuts s at sep. Unlike the list splitting it keeps empty pieces and
// whitespace, because a cell may be empty or start with spaces. A backslash
// before a comma, a semicolon or another backslash is that character itself
// (unescape false keeps the backslash, for the first cut of the rows).
func rowsSplit(s, sep string, unescape bool) []string {
	var out []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\\' && i+1 < len(s) && (s[i+1] == ',' || s[i+1] == ';' || s[i+1] == '\\') {
			if !unescape {
				cur.WriteByte(c)
			}
			cur.WriteByte(s[i+1])
			i++
			continue
		}
		if strings.HasPrefix(s[i:], sep) {
			out = append(out, cur.String())
			cur.Reset()
			i += len(sep) - 1
			continue
		}
		cur.WriteByte(c)
	}
	return append(out, cur.String())
}

// rowsCells reads the rows: cells separated by commas, rows by semicolons.
func rowsCells(s string) [][]string {
	var rows [][]string
	for _, raw := range rowsSplit(s, ";", false) {
		if raw == "" {
			continue
		}
		rows = append(rows, rowsSplit(raw, ",", true))
	}
	return rows
}

// rowsPlan decides the widths of the columns of all the rows: a fixed column
// as it is, a fit column from its longest cell (within its bounds, a share of
// what the fixed columns leave), a column that hide-empty hides when it has no
// text, an auto column as wide as the cell of each row, and a fill column what
// is left.
type rowsPlan struct {
	columns []rowsColumn
	visible []bool
	// shared is the width of the fixed and fit columns, -1 for the others.
	shared []int
	// used is the total of shared, gaps the cells between visible columns.
	used, gaps, total int
	gap               int
}

func newRowsPlan(columns []rowsColumn, rows [][]string, total, gap int) rowsPlan {
	p := rowsPlan{columns: columns, total: total, gap: gap}
	p.visible = make([]bool, len(columns))
	p.shared = make([]int, len(columns))
	longest := make([]int, len(columns))
	shown := 0
	for i, c := range columns {
		empty := true
		for _, r := range rows {
			if i < len(r) && r[i] != "" {
				empty = false
				longest[i] = max(longest[i], runesWidth(parseSGR(r[i], grid.Style{})))
			}
		}
		p.visible[i] = !c.hideEmpty || !empty
		if p.visible[i] {
			shown++
		}
	}
	p.gaps = gap * max(shown-1, 0)
	for i, c := range columns {
		switch {
		case !p.visible[i]:
			p.shared[i] = 0
		case c.hasFit:
			p.shared[i] = -2
		case c.auto || c.fill || c.width == 0 && i == len(columns)-1:
			p.shared[i] = -1
		default:
			p.shared[i] = c.width
			p.used += c.width
		}
	}
	rest := max(total-p.used-p.gaps, 0)
	for i, c := range columns {
		if p.shared[i] != -2 {
			continue
		}
		upper := longest[i]
		switch {
		case c.fitPct > 0:
			upper = max(max(c.floor, c.fitMin), rest*c.fitPct/100)
		case c.fitMax > 0:
			upper = c.fitMax
		}
		w := min(max(longest[i], c.fitMin), max(upper, c.fitMin))
		p.shared[i] = w
		p.used += w
	}
	return p
}

// row is the width of each column in one row.
func (p rowsPlan) row(row []string) []int {
	widths := make([]int, len(p.columns))
	used, fills := p.used, 0
	for i, c := range p.columns {
		switch {
		case !p.visible[i]:
		case p.shared[i] >= 0:
			widths[i] = p.shared[i]
		case c.auto:
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			widths[i] = runesWidth(parseSGR(cell, grid.Style{}))
			used += widths[i]
		default:
			fills++
		}
	}
	left := max(p.total-used-p.gaps, 0)
	for i, c := range p.columns {
		if !p.visible[i] || p.shared[i] >= 0 || c.auto {
			continue
		}
		share := left / fills
		if fills == 1 {
			share = left
		}
		widths[i] = max(share, c.minWidth)
		left = max(left-widths[i], 0)
		fills--
	}
	return widths
}

// paletteColour resolves a colour of a column or a cell: a name of the
// design's palette (@Name) through the palette the screen hands its parts, any
// other text as it is.
func paletteColour(p Props, c string) string {
	c = strings.TrimSpace(c)
	name, isName := strings.CutPrefix(c, "@")
	if !isName {
		return c
	}
	for _, entry := range strings.Split(p.Str("theme.palette"), ";") {
		if key, value, ok := strings.Cut(entry, "="); ok && key == name {
			return value
		}
	}
	return ""
}

// cellStyleOf reads the style a program gave one cell: "fg|bg|flags" with the
// flags b (bold), d (dim) and r (reverse).
func cellStyleOf(p Props, code string, base grid.Style) grid.Style {
	if code == "" || code == "-" {
		return base
	}
	parts := strings.Split(code, "|")
	if len(parts) > 2 && strings.Contains(parts[2], "p") {
		// Plain: the terminal's own colours, whatever the column says.
		base.Fg, base.Bg = "", ""
	}
	if len(parts) > 0 && parts[0] != "" {
		base.Fg = paletteColour(p, parts[0])
	}
	if len(parts) > 1 && parts[1] != "" {
		base.Bg = paletteColour(p, parts[1])
	}
	if len(parts) > 2 {
		base.Bold = base.Bold || strings.Contains(parts[2], "b")
		base.Dim = base.Dim || strings.Contains(parts[2], "d")
		base.Reverse = base.Reverse || strings.Contains(parts[2], "r")
	}
	return base
}

// paintRows draws the Rows component: one line per row, the cells laid out in
// the columns the design names, and each row in the style it was given:
// "selected" inverts the row in the accent colour (or, when a column has a
// selbg, only gives those columns their background), "dim" mutes it and
// "accent" colours its text. A column can set its own colours, cut with an
// ellipsis, or draw a colour as a block or as glyphs; a program can restyle a
// cell, and the text of a cell may carry SGR colours.
func paintRows(g *grid.Grid, p Props) {
	accent := p.Str("color")
	columns := rowsColumns(p.Str("columns"))
	styles := p.List("styles")
	cellStyles := rowsCells(p.Str("cellstyles"))
	anySelBg := false
	for _, c := range columns {
		anySelBg = anySelBg || c.selBg != "" || c.selFg != "" || c.selBold
	}
	allRows := rowsCells(p.Str("rows"))
	// A pane without focus shows its selected row as accent text, not inverted.
	blurred := p.Str("focused") == "false"
	blurAccent := orColour(p.Str("blurColor"), accent)
	gap := min(max(p.Int("gap", 0), 0), 8)
	plan := newRowsPlan(columns, allRows, g.W, gap)
	lastVisible := -1
	for i := range columns {
		if plan.visible[i] {
			lastVisible = i
		}
	}
	for y, row := range allRows {
		if y >= g.H {
			break
		}
		base := grid.Style{}
		isSelected := false
		if y < len(styles) {
			switch styles[y] {
			case "selected":
				isSelected = true
				if blurred {
					base = grid.Style{Fg: blurAccent}
				} else if !anySelBg {
					base = selected(accent)
					g.Fill(design.Rect{X: 0, Y: y, W: g.W, H: 1}, ' ', base)
				}
			case "dim":
				base = p.Dim()
			case "accent":
				base = grid.Style{Fg: accent}
			}
		}
		x := 0
		widths := plan.row(row)
		for i, col := range columns {
			if !plan.visible[i] {
				continue
			}
			w := min(widths[i], max(g.W-x, 0))
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			style := base
			if col.fg != "" {
				style.Fg = paletteColour(p, col.fg)
			}
			if col.bg != "" {
				style.Bg = paletteColour(p, col.bg)
			}
			if isSelected && blurred {
				style.Fg = blurAccent
			} else {
				if isSelected && col.selBg != "" {
					style.Bg = paletteColour(p, col.selBg)
				}
				if isSelected && col.selFg != "" {
					style.Fg = paletteColour(p, col.selFg)
				}
				if isSelected && col.selBold {
					style.Bold = true
				}
			}
			if y < len(cellStyles) && i < len(cellStyles[y]) {
				style = cellStyleOf(p, cellStyles[y][i], style)
			}
			switch {
			case w <= 0:
			case col.colour && cell != "":
				g.Fill(design.Rect{X: x, Y: y, W: w, H: 1}, ' ', grid.Style{Bg: cell})
			case col.glyph && cell != "":
				g.Text(x, y, strings.Repeat("█", w), grid.Style{Fg: cell}, w)
			case col.colour || col.glyph:
				g.Text(x, y, "·", p.Dim(), w)
			default:
				if style.Bg != "" && (!isSelected || !col.fit || col.selBg == "") {
					g.Fill(design.Rect{X: x, Y: y, W: w, H: 1}, ' ', grid.Style{Bg: style.Bg})
				}
				at := x
				for _, run := range cutRuns(parseSGR(cell, style), w, col.ellipsis) {
					at += g.Text(at, y, run.text, run.style, 0)
				}
			}
			x += w
			if i < lastVisible {
				x += gap
			}
		}
	}
}
