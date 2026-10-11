package colorpicker

import (
	"image/color"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
)

// swatchOf draws one colour as a block w cells wide. Palette colours are sent
// as their number so the terminal shows them its own way, exactly as the
// design will; the marker shows the current choice.
func swatchOf(index int, w int, marked bool) string {
	text := strings.Repeat(" ", w)
	rgb := space.ANSI(index)
	if marked {
		mark := "✓"
		pad := (w - 1) / 2
		text = strings.Repeat(" ", pad) + mark + strings.Repeat(" ", w-pad-1)
	}
	st := lipgloss.NewStyle().Background(lipgloss.Color(strconv.Itoa(index))).Foreground(contrast(rgb))
	return st.Render(text)
}

// contrast picks black or white text for readability on c.
func contrast(c space.RGB) color.Color {
	if c.Luma() > 140 {
		return lipgloss.Color("#000000")
	}
	return lipgloss.Color("#ffffff")
}

// system16Lines are the 16 system colours as big blocks, two rows of eight.
func (m *Model) system16Lines(y0 int) []string {
	var lines []string
	for blockRow := 0; blockRow < 2; blockRow++ {
		for h := 0; h < 2; h++ { // each swatch is two terminal rows tall
			y := y0 + len(lines)
			r := m.newRow(y)
			r.text(" ")
			for col := 0; col < 8; col++ {
				idx := blockRow*8 + col
				marked := m.index == idx && !m.empty
				r.span(swatchOf(idx, 5, marked && h == 0), func(int) { m.setIndex(idx) }, nil).text(" ")
			}
			lines = append(lines, r.String())
		}
		lines = append(lines, "")
	}
	return lines[:len(lines)-1]
}

// palette256Lines lays the 256 colours out by their structure: the 16 system
// colours, the 6x6x6 cube as six blocks (one per red level, green down, blue
// across), then the 24 greys.
func (m *Model) palette256Lines(y0 int) []string {
	var lines []string
	cell := func(r *row, idx int) {
		r.span(swatchOf(idx, 2, m.index == idx && !m.empty), func(int) { m.setIndex(idx) }, nil)
	}
	next := func() *row { return m.newRow(y0 + len(lines)).text(" ") }

	r := next()
	for i := 0; i < 16; i++ {
		cell(r, i)
	}
	lines = append(lines, r.String(), "")
	for blockRow := 0; blockRow < 2; blockRow++ {
		for g := 0; g < 6; g++ {
			r := next()
			for block := 0; block < 3; block++ {
				red := blockRow*3 + block
				for b := 0; b < 6; b++ {
					cell(r, 16+36*red+6*g+b)
				}
				r.text(" ")
			}
			lines = append(lines, r.String())
		}
		if blockRow == 0 {
			lines = append(lines, "")
		}
	}
	lines = append(lines, "")
	for half := 0; half < 2; half++ {
		r := next()
		for i := 0; i < 12; i++ {
			cell(r, 232+half*12+i)
		}
		lines = append(lines, r.String())
	}
	return lines
}
