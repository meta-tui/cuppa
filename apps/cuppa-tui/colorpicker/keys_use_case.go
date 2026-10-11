package colorpicker

import (
	"math"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/color/scheme"
	"github.com/meta-ui-labs/cuppa/libs/color/space"
)

// Nav implements modal.Navigator.
//
// Tab and Shift+Tab change the tab. Arrows move through the swatches of the
// 16, 256 and Themes tabs, or, on the RGB and HSL tabs, Up and Down choose a
// slider and Left and Right change it by one (Shift: ten). On the Themes tab
// PageUp and PageDown change the scheme (Shift: ten). Enter accepts, Esc
// cancels, and typing edits the value field, as before.
func (m *Model) Nav(name string) bool {
	switch name {
	case "tab":
		m.tab = (m.tab + 1) % tab(len(tabNames))
		m.rect = m.layout()
		return true
	case "shift+tab":
		m.tab = (m.tab + tab(len(tabNames)) - 1) % tab(len(tabNames))
		m.rect = m.layout()
		return true
	}
	base := strings.TrimPrefix(name, "shift+")
	amount := 1
	if base != name {
		amount = 10
	}
	switch m.tab {
	case tab16:
		return m.moveSwatch(base, 1, 8, 16, func(i int) { m.setIndex(i) }, m.paletteIndex())
	case tab256:
		return m.moveSwatch(base, 1, 6, 256, func(i int) { m.setIndex(i) }, m.paletteIndex())
	case tabRGB, tabHSL:
		return m.moveSlider(base, amount)
	case tabThemes:
		switch base {
		case "pgup":
			m.stepScheme(-amount)
			return true
		case "pgdown":
			m.stepScheme(amount)
			return true
		}
		s := scheme.All()[m.scheme]
		at := -1
		for i, c := range s.ANSI {
			if !m.empty && m.index < 0 && m.rgb == c {
				at = i
			}
		}
		return m.moveSwatch(base, 1, 8, 16, func(i int) { m.setRGB(s.ANSI[i]) }, at)
	}
	return false
}

// paletteIndex is the palette colour the dialog holds, or -1.
func (m *Model) paletteIndex() int {
	if m.empty {
		return -1
	}
	return m.index
}

// moveSwatch moves from the swatch at (-1 for none yet) by one for Left and
// Right and by row for Up and Down, staying inside count swatches.
func (m *Model) moveSwatch(key string, side, row, count int, pick func(int), at int) bool {
	d := 0
	switch key {
	case "left":
		d = -side
	case "right":
		d = side
	case "up":
		d = -row
	case "down":
		d = row
	default:
		return false
	}
	if at < 0 {
		pick(0)
		return true
	}
	pick(min(max(at+d, 0), count-1))
	return true
}

// moveSlider chooses a slider with Up and Down and changes it with Left and
// Right.
func (m *Model) moveSlider(key string, amount int) bool {
	switch key {
	case "up":
		m.slide = (m.slide + 2) % 3
		return true
	case "down":
		m.slide = (m.slide + 1) % 3
		return true
	case "left", "right":
	default:
		return false
	}
	d := float64(amount)
	if key == "left" {
		d = -d
	}
	clamp := func(v, hi float64) float64 { return math.Min(math.Max(v, 0), hi) }
	if m.tab == tabRGB {
		c := m.rgb
		ch := [3]float64{float64(c.R), float64(c.G), float64(c.B)}
		ch[m.slide] = clamp(ch[m.slide]+d, 255)
		m.setRGB(space.RGB{R: uint8(ch[0]), G: uint8(ch[1]), B: uint8(ch[2])})
		return true
	}
	h := m.hsl
	switch m.slide {
	case 0:
		h.H = clamp(math.Round(h.H)+d, 359)
	case 1:
		h.S = clamp(math.Round(h.S*100)+d, 100) / 100
	default:
		h.L = clamp(math.Round(h.L*100)+d, 100) / 100
	}
	m.setHSL(h)
	return true
}
