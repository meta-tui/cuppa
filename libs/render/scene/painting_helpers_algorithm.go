package scene

import (
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// pipeLines splits a value on "|", the line separator used by single-line
// property editors.
func pipeLines(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "|")
	for i := range parts {
		parts[i] = strings.TrimRight(parts[i], " ")
	}
	return parts
}

// pick clamps a property index into [0, n).
func pick(p Props, key string, n int) int {
	if n <= 0 {
		return 0
	}
	return min(max(p.Int(key, 0), 0), n-1)
}

var (
	bold     = grid.Style{Bold: true}
	selected = func(color string) grid.Style { return grid.Style{Fg: "0", Bg: color, Bold: true} }
)

// wrap breaks text into lines of at most width runes, on spaces where possible.
func wrap(text string, width int) []string {
	if width < 1 {
		return nil
	}
	var out []string
	line := ""
	for _, word := range strings.Fields(text) {
		for len([]rune(word)) > width {
			r := []rune(word)
			if line != "" {
				out = append(out, line)
				line = ""
			}
			out = append(out, string(r[:width]))
			word = string(r[width:])
		}
		switch {
		case line == "":
			line = word
		case len([]rune(line))+1+len([]rune(word)) <= width:
			line += " " + word
		default:
			out = append(out, line)
			line = word
		}
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

// sparkLevels are the eight block heights used by compact charts.
var sparkLevels = []rune("▁▂▃▄▅▆▇█")

// numbers parses a comma separated list of numbers, skipping bad entries.
func numbers(p Props, key string) []float64 {
	var out []float64
	for _, item := range p.List(key) {
		if v, err := strconv.ParseFloat(item, 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func extent(vs []float64) (lo, hi float64) {
	if len(vs) == 0 {
		return 0, 1
	}
	lo, hi = vs[0], vs[0]
	for _, v := range vs {
		lo, hi = min(lo, v), max(hi, v)
	}
	if hi == lo {
		hi = lo + 1
	}
	return lo, hi
}

// resample picks n values from vs, stretching or thinning as needed.
func resample(vs []float64, n int) []float64 {
	if len(vs) == 0 || n <= 0 {
		return nil
	}
	out := make([]float64, n)
	for i := range out {
		out[i] = vs[i*len(vs)/n]
	}
	return out
}

// fillColumn draws a vertical bar of the given height in eighths of a cell,
// growing upward from the bottom row of the column rectangle.
func fillColumn(g *grid.Grid, x, bottomY, eighths int, s grid.Style) {
	full, rem := eighths/8, eighths%8
	for i := 0; i < full; i++ {
		g.Set(x, bottomY-i, grid.Cell{Ch: '█', Style: s})
	}
	if rem > 0 {
		g.Set(x, bottomY-full, grid.Cell{Ch: sparkLevels[rem-1], Style: s})
	}
}
