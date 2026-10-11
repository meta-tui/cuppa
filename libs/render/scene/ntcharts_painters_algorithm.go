package scene

import (
	"strconv"
	"math"
	"fmt"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// paintBarChart draws ntcharts' bar chart: bars of equal width one cell apart,
// the top of each in eighths, an axis line under them and the labels below,
// each starting at its bar.
func paintBarChart(g *grid.Grid, p Props) {
	vs := numbers(p, "values")
	if len(vs) == 0 || g.H < 3 {
		paintGeneric(g, "Bar chart")
		return
	}
	labels := p.List("labels")
	style := fg(p.Str("color"))
	_, hi := extent(append([]float64{0}, vs...))
	plotH := g.H - 2
	barW := max((g.W-(len(vs)-1))/len(vs), 1)
	for x := 0; x < g.W; x++ {
		g.Set(x, g.H-2, grid.Cell{Ch: '─', Style: p.Dim()})
	}
	for i, v := range vs {
		x0 := i * (barW + 1)
		if x0 >= g.W {
			break
		}
		eighths := int(v/hi*float64(plotH*8) + 0.5)
		for dx := 0; dx < barW && x0+dx < g.W; dx++ {
			fillColumn(g, x0+dx, plotH-1, eighths, style)
		}
		if i < len(labels) {
			g.Text(x0, g.H-1, labels[i], p.Dim(), barW)
		}
	}
}

// paintLineChart draws ntcharts' line chart: a braille line over value labels
// on the left and index labels along the bottom.
// paintLineChart draws ntcharts' line chart: a braille line over value labels
// on the left and index labels along the bottom.
func paintLineChart(g *grid.Grid, p Props) {
	vs := numbers(p, "values")
	if len(vs) == 0 {
		paintGeneric(g, "Line chart")
		return
	}
	plotBraille(g, vs, false, p.Bool("axes"), fg(p.Str("color")), p.Dim(), func(left, plotW int) []axisLabel {
		var out []axisLabel
		for i := range vs {
			// Labels sit on every second column, one in from the axis.
			x := left - 1
			if len(vs) > 1 {
				x += 2 * int(float64(i)*float64(plotW-1)/float64(len(vs)-1)/2)
			}
			out = append(out, axisLabel{x, strconv.Itoa(i)})
		}
		return out
	})
}

// paintSparkline draws ntcharts' sparkline: one block per value, heights in
// eighths of the largest value, drawn from the right edge.
func paintSparkline(g *grid.Grid, p Props) {
	vs := numbers(p, "values")
	if len(vs) > g.W {
		vs = vs[len(vs)-g.W:]
	}
	_, hi := extent(append([]float64{0}, vs...))
	style := fg(p.Str("color"))
	for i, v := range vs {
		level := int(v/hi*8 + 0.5)
		if level < 1 {
			continue
		}
		g.Set(g.W-len(vs)+i, 0, grid.Cell{Ch: sparkLevels[min(level, 8)-1], Style: style})
	}
}

func paintStreamline(g *grid.Grid, p Props) {
	vs := resample(numbers(p, "values"), g.W)
	_, hi := extent(append([]float64{0}, vs...))
	style := fg(p.Str("color"))
	for x, v := range vs {
		fillColumn(g, x, g.H-1, int(v/hi*float64(g.H*8)), style)
	}
}

// paintTimeSeries draws ntcharts' time series line chart: the same line, with
// clock times along the bottom.
// paintTimeSeries draws ntcharts' time series line chart: the value axis starts
// at zero, and clock times (hours, minutes, seconds) label every twelfth column.
func paintTimeSeries(g *grid.Grid, p Props) {
	vs := numbers(p, "values")
	if len(vs) == 0 {
		paintGeneric(g, "Time series")
		return
	}
	span := float64(len(vs)-1) * 3600 // one value an hour
	plotBraille(g, vs, true, p.Bool("axes"), fg(p.Str("color")), p.Dim(), func(left, plotW int) []axisLabel {
		var out []axisLabel
		for x := left - 1; x+8 <= left-1+plotW; x += 12 {
			t := int(span * float64(x-(left-1)) / float64(plotW))
			out = append(out, axisLabel{x, fmt.Sprintf("%02d:%02d:%02d", t/3600, t/60%60, t%60)})
		}
		return out
	})
}

// heatColors run from cold to hot.
var heatColors = []string{"17", "19", "21", "33", "45", "190", "214", "196"}

// paintHeatMap fills every cell with a background colour, the way ntcharts'
// heat map does.
func paintHeatMap(g *grid.Grid, p Props) {
	seed := p.Int("seed", 0)
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			v := (x*7 + y*13 + seed*31 + (x*y)%5*3) % len(heatColors)
			g.Set(x, y, grid.Cell{Ch: ' ', Style: grid.Style{Bg: heatColors[v]}})
		}
	}
}

// paintChartCanvas draws ntcharts' canvas: an empty field with the title at the
// top right.
func paintChartCanvas(g *grid.Grid, p Props) {
	if t := strings.TrimSpace(p.Str("title")); t != "" && g.W > 6 {
		g.Text(g.W-len([]rune(t))-1, 0, t, fg(p.Str("color")), g.W)
	}
}

// brailleBits maps a dot of a cell (column 0 or 1, row 0 to 3) to its bit in
// the braille character.
var brailleBits = [2][4]rune{{0x01, 0x02, 0x04, 0x40}, {0x08, 0x10, 0x20, 0x80}}

// plotBraille draws values as a line in braille dots (two by four to a cell).
// With axes it reserves a label column on the left, the axis, and a row of
// labels below, placed along the line at the points they belong to.
// axisLabel is a label under the axis, at a column of the chart.
type axisLabel struct {
	x    int
	text string
}

// plotBraille draws values as a line in braille dots (two by four to a cell).
// With axes it reserves a label column on the left, the axis, and a row of
// labels below. The value axis spans the data, and starts at zero when asked.
func plotBraille(g *grid.Grid, vs []float64, fromZero, axes bool, style, muted grid.Style, labelsAt func(left, plotW int) []axisLabel) {
	lo, hi := extent(vs)
	if fromZero {
		lo = min(lo, 0)
	}
	left, plotW, plotH := 0, g.W, g.H
	if axes && g.W >= 8 && g.H >= 4 {
		width := 1
		for _, v := range []float64{lo, hi} {
			width = max(width, len(strconv.Itoa(int(math.RoundToEven(v)))))
		}
		left, plotW, plotH = width+1, g.W-width-1, g.H-2
		for y := 0; y < plotH; y++ {
			g.Set(left-1, y, grid.Cell{Ch: '│', Style: muted})
			if y%2 == 0 {
				v := hi - (hi-lo)*float64(y)/float64(plotH)
				text := strconv.Itoa(int(math.RoundToEven(v)))
				g.Text(left-1-len(text), y, text, muted, width)
			}
		}
		g.Set(left-1, plotH, grid.Cell{Ch: '└', Style: muted})
		for x := left; x < g.W; x++ {
			g.Set(x, plotH, grid.Cell{Ch: '─', Style: muted})
		}
		bottom := strconv.Itoa(int(math.RoundToEven(lo)))
		g.Text(left-1-len(bottom), plotH, bottom, muted, width)
		last := -1
		for _, l := range labelsAt(left, plotW) {
			if l.x <= last || l.x+len(l.text) > g.W {
				continue
			}
			g.Text(l.x, plotH+1, l.text, muted, g.W-l.x)
			last = l.x + len(l.text)
		}
	}
	pw, ph := plotW*2, plotH*4
	px := func(i int) int {
		if len(vs) < 2 {
			return 0
		}
		return int(math.Round(float64(i) * float64(pw-1) / float64(len(vs)-1)))
	}
	py := func(v float64) int { return int(math.Round((hi - v) / (hi - lo) * float64(ph-1))) }
	bits := map[[2]int]rune{}
	dot := func(x, y int) {
		if x >= 0 && y >= 0 && x < pw && y < ph {
			bits[[2]int{x / 2, y / 4}] |= brailleBits[x%2][y%4]
		}
	}
	for i := range vs {
		x0, y0 := px(i), py(vs[i])
		if i == 0 {
			dot(x0, y0)
			continue
		}
		x1, y1 := px(i-1), py(vs[i-1])
		dx, dy := x0-x1, y0-y1
		steps := max(abs(dx), abs(dy))
		for st := 0; st <= steps; st++ {
			if steps == 0 {
				dot(x0, y0)
				break
			}
			dot(x1+int(math.Round(float64(dx)*float64(st)/float64(steps))), y1+int(math.Round(float64(dy)*float64(st)/float64(steps))))
		}
	}
	for cell, mask := range bits {
		g.Set(left+cell[0], cell[1], grid.Cell{Ch: 0x2800 + mask, Style: style})
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
