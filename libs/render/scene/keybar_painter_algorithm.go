package scene

import (
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// keybarHints reads the hints property: "key:label" items separated by commas;
// the key ends at the first colon.
func keybarHints(spec string) [][2]string {
	var out [][2]string
	for _, item := range rowsSplit(spec, ",", true) {
		key, label, ok := strings.Cut(item, ":")
		if !ok {
			key, label = item, ""
		}
		if strings.TrimSpace(key) == "" && strings.TrimSpace(label) == "" {
			continue
		}
		out = append(out, [2]string{key, label})
	}
	return out
}

// paintKeybar draws the key bar: a space, then for each hint the key in bold
// in the key colour, a space and the label in the label colour, the hints
// separated by gap spaces, cut at the width.
func paintKeybar(g *grid.Grid, p Props) {
	keyStyle := grid.Style{Fg: p.Str("keyColor"), Bold: true}
	labelStyle := grid.Style{Fg: p.Str("labelColor")}
	gap := min(max(p.Int("gap", 2), 0), 8)
	x := 1
	for i, h := range keybarHints(p.Str("hints")) {
		if i > 0 {
			x += gap
		}
		if x >= g.W {
			break
		}
		x += g.Text(x, 0, h[0], keyStyle, g.W-x)
		if h[1] != "" && x < g.W {
			x += g.Text(x, 0, " ", grid.Style{}, g.W-x)
			x += g.Text(x, 0, h[1], labelStyle, g.W-x)
		}
	}
}

// paintSlot draws the sample of a slot: the program fills it with the view of
// a model of its own, here only its text is shown ("|" starts a new line).
func paintSlot(g *grid.Grid, p Props) {
	view := p.Str("view")
	lines := strings.Split(view, "\n")
	if len(lines) == 1 {
		lines = pipeLines(view)
	}
	for y, line := range lines {
		if y >= g.H {
			break
		}
		x := 0
		for _, run := range parseSGR(line, grid.Style{}) {
			x += g.Text(x, y, run.text, run.style, g.W-x)
			if x >= g.W {
				break
			}
		}
	}
}
