package scene

import (
	"math"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

// The colour picker component draws the colour dialog of Cuppa: tabs for the
// 16 and 256 palettes (swatches) and for RGB and HSL (sliders), a preview and
// the value. Its drawing is also its hit map: layoutPicker returns what each
// cell does, which the exported program uses to follow the mouse.

// Hit kinds of the colour picker.
const (
	pickerHitTab    = iota // arg: the index of the tab in the visible tabs
	pickerHitSwatch        // arg: the palette index
	pickerHitStep          // arg: slider*2, plus 1 for the [+] button
	pickerHitBar           // arg: the slider; x is the cell inside the bar
)

// pickerBarWidth is how many cells a slider's gradient bar has.
const pickerBarWidth = 28

// pickerHit is a rectangle of the picker that reacts to a click.
type pickerHit struct {
	x, y, w, h int
	kind, arg  int
}

// pickerState is what the picker draws: the colour, the tab and the tab
// names shown, the slider the keyboard is on and the accent colour.
type pickerState struct {
	value  string
	tab    string
	tabs   []string
	slide  int
	accent string
}

// pickerRGB is a colour as three 0..255 channels.
type pickerRGB struct{ r, g, b int }

// pickerSystem are the 16 system colours as xterm shows them.
var pickerSystem = [16]pickerRGB{
	{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0}, {0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
	{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0}, {92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

// pickerANSI is palette colour i (0 to 255) as RGB.
func pickerANSI(i int) pickerRGB {
	switch {
	case i < 0:
		return pickerRGB{}
	case i < 16:
		return pickerSystem[i]
	case i < 232:
		level := func(n int) int {
			if n == 0 {
				return 0
			}
			return 55 + 40*n
		}
		i -= 16
		return pickerRGB{level(i / 36), level(i / 6 % 6), level(i % 6)}
	}
	g := 8 + 10*(min(i, 255)-232)
	return pickerRGB{g, g, g}
}

// pickerHex is the colour as #rrggbb.
func (c pickerRGB) hex() string {
	const digits = "0123456789abcdef"
	b := []byte{'#', 0, 0, 0, 0, 0, 0}
	for i, v := range []int{c.r, c.g, c.b} {
		v = min(max(v, 0), 255)
		b[1+2*i], b[2+2*i] = digits[v>>4], digits[v&15]
	}
	return string(b)
}

// pickerParse reads "" (none), a palette number or #rgb / #rrggbb. index is -1
// for a hex colour.
func pickerParse(s string) (c pickerRGB, index int, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return pickerRGB{}, -1, false
	}
	if n, err := strconv.Atoi(s); err == nil {
		if n < 0 || n > 255 {
			return pickerRGB{}, -1, false
		}
		return pickerANSI(n), n, true
	}
	h := strings.TrimPrefix(s, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 {
		return pickerRGB{}, -1, false
	}
	v, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return pickerRGB{}, -1, false
	}
	return pickerRGB{int(v >> 16), int(v >> 8 & 255), int(v & 255)}, -1, true
}

// pickerLuma is how light a colour looks, 0 to 255.
func (c pickerRGB) luma() float64 {
	return 0.299*float64(c.r) + 0.587*float64(c.g) + 0.114*float64(c.b)
}

// pickerHSL is hue 0..359, saturation and lightness 0..1.
type pickerHSL struct{ h, s, l float64 }

func (c pickerRGB) hsl() pickerHSL {
	r, g, b := float64(c.r)/255, float64(c.g)/255, float64(c.b)/255
	hi, lo := math.Max(r, math.Max(g, b)), math.Min(r, math.Min(g, b))
	l := (hi + lo) / 2
	if hi == lo {
		return pickerHSL{0, 0, l}
	}
	d := hi - lo
	s := d / (1 - math.Abs(2*l-1))
	var h float64
	switch hi {
	case r:
		h = math.Mod((g-b)/d, 6)
	case g:
		h = (b-r)/d + 2
	default:
		h = (r-g)/d + 4
	}
	h *= 60
	if h < 0 {
		h += 360
	}
	return pickerHSL{h, s, l}
}

func (c pickerHSL) rgb() pickerRGB {
	h := math.Mod(c.h, 360)
	if h < 0 {
		h += 360
	}
	s, l := math.Min(math.Max(c.s, 0), 1), math.Min(math.Max(c.l, 0), 1)
	k := (1 - math.Abs(2*l-1)) * s
	x := k * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - k/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = k, x, 0
	case h < 120:
		r, g, b = x, k, 0
	case h < 180:
		r, g, b = 0, k, x
	case h < 240:
		r, g, b = 0, x, k
	case h < 300:
		r, g, b = x, 0, k
	default:
		r, g, b = k, 0, x
	}
	to := func(v float64) int { return int(math.Round((v + m) * 255)) }
	return pickerRGB{to(r), to(g), to(b)}
}

// pickerTabs reads the tabs property: names among 16, 256, RGB and HSL.
func pickerTabs(spec string) []string {
	var out []string
	seen := map[string]bool{}
	for _, name := range splitList(spec, ",") {
		n := strings.ToUpper(strings.TrimSpace(name))
		if (n == "16" || n == "256" || n == "RGB" || n == "HSL") && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		out = []string{"16", "256", "RGB", "HSL"}
	}
	return out
}

// pickerContrast is black or white, whichever reads on c.
func pickerContrast(c pickerRGB) string {
	if c.luma() > 140 {
		return "#000000"
	}
	return "#ffffff"
}

// layoutPicker draws the picker onto g and returns what each part does. The
// layout does not depend on the size of g: parts that do not fit are cut.
func layoutPicker(g *grid.Grid, st pickerState) []pickerHit {
	var hits []pickerHit
	dimStyle := grid.Style{Fg: "245"}
	c, index, ok := pickerParse(st.value)
	// With no colour the sliders start from a pink, as the library's do.
	slideFrom := c
	if !ok {
		slideFrom = pickerRGB{255, 135, 215}
	}

	// The tab bar.
	active := 0
	for i, name := range st.tabs {
		if name == st.tab {
			active = i
		}
	}
	x := 1
	for i, name := range st.tabs {
		label := " " + name + " "
		style := dimStyle
		if i == active {
			style = grid.Style{Fg: "0", Bg: st.accent, Bold: true}
		}
		g.Text(x, 0, label, style, 0)
		hits = append(hits, pickerHit{x: x, y: 0, w: len(label), h: 1, kind: pickerHitTab, arg: i})
		x += len(label) + 1
	}

	swatch := func(x, y, w, h, idx int) {
		bg := strconv.Itoa(idx)
		g.Fill(design.Rect{X: x, Y: y, W: w, H: h}, ' ', grid.Style{Bg: bg})
		if ok && index == idx {
			mark := grid.Cell{Ch: '✓', Style: grid.Style{Fg: pickerContrast(pickerANSI(idx)), Bg: bg}}
			g.Set(x+(w-1)/2, y, mark)
		}
		hits = append(hits, pickerHit{x: x, y: y, w: w, h: h, kind: pickerHitSwatch, arg: idx})
	}

	y := 2
	switch active := pickerTabAt(st, active); active {
	case "16":
		for blockRow := 0; blockRow < 2; blockRow++ {
			for col := 0; col < 8; col++ {
				swatch(1+col*6, y, 5, 2, blockRow*8+col)
			}
			y += 3
		}
		y--
	case "256":
		for i := 0; i < 16; i++ {
			swatch(1+i*2, y, 2, 1, i)
		}
		y += 2
		for blockRow := 0; blockRow < 2; blockRow++ {
			for gr := 0; gr < 6; gr++ {
				for block := 0; block < 3; block++ {
					red := blockRow*3 + block
					for b := 0; b < 6; b++ {
						swatch(1+block*13+b*2, y, 2, 1, 16+36*red+6*gr+b)
					}
				}
				y++
			}
			if blockRow == 0 {
				y++
			}
		}
		y++
		for half := 0; half < 2; half++ {
			for i := 0; i < 12; i++ {
				swatch(1+i*2, y, 2, 1, 232+half*12+i)
			}
			y++
		}
	default: // RGB and HSL: three sliders
		type slider struct {
			label string
			max   float64
			value float64
			unit  string
			at    func(t float64) pickerRGB
		}
		var sliders []slider
		h := slideFrom.hsl()
		if active == "RGB" {
			with := func(ch int) func(t float64) pickerRGB {
				return func(t float64) pickerRGB {
					v := int(math.Round(t * 255))
					out := slideFrom
					switch ch {
					case 0:
						out.r = v
					case 1:
						out.g = v
					default:
						out.b = v
					}
					return out
				}
			}
			sliders = []slider{
				{"R", 255, float64(slideFrom.r), "", with(0)},
				{"G", 255, float64(slideFrom.g), "", with(1)},
				{"B", 255, float64(slideFrom.b), "", with(2)},
			}
		} else {
			sliders = []slider{
				{"H", 359, h.h, "°", func(t float64) pickerRGB { return pickerHSL{t * 359, 1, 0.5}.rgb() }},
				{"S", 100, h.s * 100, "%", func(t float64) pickerRGB { return pickerHSL{h.h, t, h.l}.rgb() }},
				{"L", 100, h.l * 100, "%", func(t float64) pickerRGB { return pickerHSL{h.h, h.s, t}.rgb() }},
			}
		}
		for i, s := range sliders {
			label := " " + s.label + " "
			style := grid.Style{Bold: true}
			if i == st.slide {
				style = grid.Style{Fg: "0", Bg: st.accent, Bold: true}
			}
			g.Text(1, y, label, style, 0)
			g.Text(4, y, "[-]", grid.Style{Fg: "0", Bg: "250"}, 0)
			hits = append(hits, pickerHit{x: 4, y: y, w: 3, h: 1, kind: pickerHitStep, arg: i * 2})
			barX := 8
			knob := int(math.Round(s.value / s.max * float64(pickerBarWidth-1)))
			for k := 0; k < pickerBarWidth; k++ {
				col := s.at(float64(k) / float64(pickerBarWidth-1))
				cell := grid.Cell{Ch: ' ', Style: grid.Style{Bg: col.hex()}}
				if k == knob {
					cell = grid.Cell{Ch: '●', Style: grid.Style{Fg: pickerContrast(col), Bg: col.hex(), Bold: true}}
				}
				g.Set(barX+k, y, cell)
			}
			hits = append(hits, pickerHit{x: barX, y: y, w: pickerBarWidth, h: 1, kind: pickerHitBar, arg: i})
			g.Text(barX+pickerBarWidth+1, y, "[+]", grid.Style{Fg: "0", Bg: "250"}, 0)
			hits = append(hits, pickerHit{x: barX + pickerBarWidth + 1, y: y, w: 3, h: 1, kind: pickerHitStep, arg: i*2 + 1})
			g.Text(barX+pickerBarWidth+5, y, strconv.Itoa(int(math.Round(s.value)))+s.unit, grid.Style{Bold: true}, 0)
			y += 2
		}
		y--
	}

	// The preview and the value.
	y++
	if ok {
		bg := c.hex()
		if index >= 0 {
			bg = strconv.Itoa(index)
		}
		g.Fill(design.Rect{X: 1, Y: y, W: 10, H: 1}, ' ', grid.Style{Bg: bg})
		name := c.hex() + "  true colour"
		if index >= 0 {
			name = "palette " + strconv.Itoa(index) + "  " + c.hex()
		}
		g.Text(12, y, name, dimStyle, 0)
	} else {
		g.Text(1, y, "  (none)  ", dimStyle, 0)
	}
	value := strings.TrimSpace(st.value)
	if value == "" {
		value = "none"
	}
	g.Text(1, y+1, "Value ", dimStyle, 0)
	g.Text(7, y+1, value, grid.Style{Bold: true}, 0)
	return hits
}

// pickerTabAt is the name of the active tab.
func pickerTabAt(st pickerState, active int) string {
	if active >= 0 && active < len(st.tabs) {
		return st.tabs[active]
	}
	return "RGB"
}

// paintColourPicker draws the colour picker of the design.
func paintColourPicker(g *grid.Grid, p Props) {
	layoutPicker(g, pickerState{
		value:  p.Str("value"),
		tab:    strings.ToUpper(p.Str("tab")),
		tabs:   pickerTabs(p.Str("tabs")),
		slide:  min(max(p.Int("slide", 0), 0), 2),
		accent: p.Str("color"),
	})
}
