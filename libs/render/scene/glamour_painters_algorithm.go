package scene

import (
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

func itoa(n int) string { return strconv.Itoa(n) }

// markdownPalette is how one Glamour style colours the parts of a document.
type markdownPalette struct{ text, heading, code, quote, bg string }

var markdownStyles = map[string]markdownPalette{
	"dark":  {text: "252", heading: "212", code: "215", quote: "245"},
	"light": {text: "235", heading: "129", code: "130", quote: "243", bg: "255"},
	"pink":  {text: "218", heading: "205", code: "213", quote: "175"},
	"notty": {},
}

// paintMarkdown draws the markdown the way Glamour does: a margin of two
// cells, a blank row between blocks, headings padded (the first level) or
// marked with hashes, bullets and quotes indented.
func paintMarkdown(g *grid.Grid, p Props) {
	pal, ok := markdownStyles[p.Str("style")]
	if !ok {
		pal = markdownStyles["dark"]
	}
	base := grid.Style{Fg: pal.text, Bg: pal.bg}
	g.Fill(full(g), ' ', base)
	y := 0
	prev := ""
	block := func(kind string, draw func(y int)) {
		if prev != "" && (kind != prev || kind != "item") {
			y++
		}
		if y < g.H {
			draw(y)
		}
		y++
		prev = kind
	}
	for _, line := range pipeLines(p.Str("markdown")) {
		heading := grid.Style{Fg: pal.heading, Bg: pal.bg, Bold: true}
		switch {
		case strings.HasPrefix(line, "# "):
			block("h", func(y int) { g.Text(2, y, " "+line[2:]+" ", heading, g.W-2) })
		case strings.HasPrefix(line, "## "), strings.HasPrefix(line, "### "):
			block("h", func(y int) { g.Text(2, y, line, heading, g.W-2) })
		case strings.HasPrefix(line, "- "), strings.HasPrefix(line, "* "):
			block("item", func(y int) {
				g.Text(2, y, "• ", grid.Style{Fg: pal.heading, Bg: pal.bg}, 2)
				inline(g, 4, y, line[2:], base, pal)
			})
		case strings.HasPrefix(line, "> "):
			block("quote", func(y int) {
				g.Text(2, y, "│ ", grid.Style{Fg: pal.quote, Bg: pal.bg}, 2)
				inline(g, 4, y, line[2:], grid.Style{Fg: pal.quote, Bg: pal.bg}, pal)
			})
		case strings.TrimSpace(line) == "":
		default:
			block("p", func(y int) { inline(g, 2, y, line, base, pal) })
		}
	}
}

// inline writes a line, styling **bold** and `code` spans.
// inline writes one line with **bold** and `code` spans; a code span is padded
// by a cell on each side, as Glamour pads it.
func inline(g *grid.Grid, x, y int, line string, base grid.Style, pal markdownPalette) {
	rs := []rune(line)
	style := base
	code := grid.Style{Fg: pal.code, Bg: pal.bg}
	for i := 0; i < len(rs) && x < g.W; i++ {
		switch {
		case rs[i] == '`':
			if style == code && pal.code != "" {
				g.Set(x, y, grid.Cell{Ch: ' ', Style: code})
				x++
				style = base
			} else {
				style = code
				g.Set(x, y, grid.Cell{Ch: ' ', Style: code})
				x++
			}
		case rs[i] == '*' && i+1 < len(rs) && rs[i+1] == '*':
			style.Bold = !style.Bold
			i++
		default:
			g.Set(x, y, grid.Cell{Ch: rs[i], Style: style})
			x++
		}
	}
}
