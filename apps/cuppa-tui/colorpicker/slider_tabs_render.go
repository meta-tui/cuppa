package colorpicker

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
)

// barWidth is how many cells a slider's gradient bar has.
const barWidth = 28

// slider describes one channel: its label, range, current value, how to show
// the colour at a position along the bar, and how to apply a new value.
type slider struct {
	label string
	max   float64
	value float64
	unit  string
	at    func(t float64) space.RGB // colour at 0..1 along the bar
	set   func(v float64)
}

func (m *Model) rgbLines(y0 int) []string {
	c := m.rgb
	chan8 := func(name string, get uint8, with func(space.RGB, uint8) space.RGB) slider {
		return slider{
			label: name, max: 255, value: float64(get),
			at:  func(t float64) space.RGB { return with(c, uint8(math.Round(t*255))) },
			set: func(v float64) { m.setRGB(with(m.rgb, uint8(math.Round(v)))) },
		}
	}
	return m.sliderLines(y0, []slider{
		chan8("R", c.R, func(c space.RGB, v uint8) space.RGB { c.R = v; return c }),
		chan8("G", c.G, func(c space.RGB, v uint8) space.RGB { c.G = v; return c }),
		chan8("B", c.B, func(c space.RGB, v uint8) space.RGB { c.B = v; return c }),
	})
}

func (m *Model) hslLines(y0 int) []string {
	h := m.hsl
	return m.sliderLines(y0, []slider{
		{label: "H", max: 359, value: h.H, unit: "°",
			at:  func(t float64) space.RGB { return space.HSL{H: t * 359, S: 1, L: 0.5}.RGB() },
			set: func(v float64) { m.setHSL(space.HSL{H: v, S: m.hsl.S, L: m.hsl.L}) }},
		{label: "S", max: 100, value: h.S * 100, unit: "%",
			at:  func(t float64) space.RGB { return space.HSL{H: h.H, S: t, L: h.L}.RGB() },
			set: func(v float64) { m.setHSL(space.HSL{H: m.hsl.H, S: v / 100, L: m.hsl.L}) }},
		{label: "L", max: 100, value: h.L * 100, unit: "%",
			at:  func(t float64) space.RGB { return space.HSL{H: h.H, S: h.S, L: t}.RGB() },
			set: func(v float64) { m.setHSL(space.HSL{H: m.hsl.H, S: m.hsl.S, L: v / 100}) }},
	})
}

// sliderLines draws one bar per slider with [-] and [+] for single steps and
// the number; pressing or dragging on the bar sets the value.
func (m *Model) sliderLines(y0 int, sliders []slider) []string {
	var lines []string
	for i, s := range sliders {
		s := s
		label := theme.Bold(s.label)
		if i == m.slide {
			label = theme.Selected(" " + s.label + " ")
		} else {
			label = " " + label + " "
		}
		step := func(d float64) func(int) {
			return func(int) { s.set(math.Min(math.Max(math.Round(s.value)+d, 0), s.max)) }
		}
		fromX := func(x int) {
			t := float64(min(max(x, 0), barWidth-1)) / float64(barWidth-1)
			s.set(math.Round(t * s.max))
		}
		r := m.newRow(y0 + len(lines))
		r.text(label)
		r.span(theme.Button("[-]", true), step(-1), nil).text(" ")
		r.span(m.bar(s), fromX, fromX)
		r.text(" ").span(theme.Button("[+]", true), step(1), nil)
		r.text(theme.Bold(fmt.Sprintf(" %3d", int(math.Round(s.value)))) + s.unit)
		lines = append(lines, r.String(), "")
	}
	return lines
}

// bar renders the gradient with the knob at the current value.
func (m *Model) bar(s slider) string {
	knob := int(math.Round(s.value / s.max * float64(barWidth-1)))
	var b strings.Builder
	for i := 0; i < barWidth; i++ {
		c := s.at(float64(i) / float64(barWidth-1))
		st := lipgloss.NewStyle().Background(lipgloss.Color(c.Hex()))
		if i == knob {
			b.WriteString(st.Foreground(contrast(c)).Bold(true).Render("●"))
		} else {
			b.WriteString(st.Render(" "))
		}
	}
	return b.String()
}
