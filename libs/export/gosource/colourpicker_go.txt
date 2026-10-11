// Package colourpicker is a colour picker for Bubble Tea v2 programs: tabs for
// the 16 and 256 terminal palettes (swatches) and for RGB and HSL (gradient
// sliders), a preview and the chosen value. It is operated with the keyboard
// and the mouse, and has no other dependency than Bubble Tea and Lip Gloss.
//
//	picker := colourpicker.New(colourpicker.WithValue("#ff007f"))
//	// in Update:
//	picker, cmd = picker.Update(msg)
//	// in View:
//	picker.View()
//
// The value is "" (none), a palette number such as "212", or "#rrggbb".
//
// The package is one file on purpose: Cuppa (https://github.com/meta-ui-labs/cuppa)
// copies it into the programs it exports. Its description for tools is in
// cuppa.component.json.
package colourpicker

import (
	"math"
	"sort"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// The tabs a picker can show.
const (
	Tab16  = "16"
	Tab256 = "256"
	TabRGB = "RGB"
	TabHSL = "HSL"
)

// ChangedMsg is returned by Update (as a command) when the keyboard or the
// mouse changed the colour.
type ChangedMsg struct{ Value string }

// Model is a colour picker. It is a value: Update returns the changed copy.
type Model struct {
	value  string
	tabs   []string
	tab    string
	slide  int
	accent string
	ox, oy int
	// hsl is the hue, saturation and lightness the HSL sliders edit; it is
	// kept because a grey or a black has no hue to read back from the RGB.
	hsl    pickerHSL
	hslFor string
	hslOK  bool
}

// Option configures New.
type Option func(*Model)

// WithValue sets the colour: "", a palette number or "#rrggbb" (or "#rgb").
func WithValue(value string) Option { return func(m *Model) { *m = m.SetValue(value) } }

// WithTabs sets which tabs are shown, in order. Unknown names are ignored.
func WithTabs(tabs ...string) Option { return func(m *Model) { m.tabs = pickerTabs(tabs) } }

// WithTab sets the tab that is shown.
func WithTab(tab string) Option { return func(m *Model) { *m = m.SetTab(tab) } }

// WithAccent sets the colour of the selected tab and slider ("212", "#ff5fd7").
func WithAccent(colour string) Option { return func(m *Model) { m.accent = colour } }

// WithOrigin sets where the picker is drawn on the screen, so that mouse
// positions can be read: the column and row of its top-left cell.
func WithOrigin(x, y int) Option { return func(m *Model) { m.ox, m.oy = x, y } }

// New returns a picker showing all four tabs, on RGB, with no colour.
func New(opts ...Option) Model {
	m := Model{tabs: pickerTabs(nil), accent: "212"}
	m.tab = TabRGB
	for _, o := range opts {
		o(&m)
	}
	if !m.hasTab(m.tab) {
		m.tab = m.tabs[0]
	}
	return m
}

// Value is the colour: "", a palette number or "#rrggbb".
func (m Model) Value() string { return m.value }

// Hex is the colour as "#rrggbb" whatever the value is: a palette colour is
// given as the terminal palette (xterm) shows it. It is "" for no colour.
func (m Model) Hex() string {
	if c, _, ok := pickerParse(m.value); ok {
		return c.hex()
	}
	return ""
}

// SetValue sets the colour. A value that is not "", a palette number or hex
// is ignored.
func (m Model) SetValue(value string) Model {
	value = strings.TrimSpace(value)
	if value == "" {
		m.value = ""
		return m
	}
	c, index, ok := pickerParse(value)
	switch {
	case !ok:
	case index >= 0:
		m.value = strconv.Itoa(index)
	default:
		m.value = c.hex()
	}
	return m
}

// Tab is the tab that is shown.
func (m Model) Tab() string { return m.tab }

// SetTab shows a tab. One that is not among the tabs is ignored.
func (m Model) SetTab(tab string) Model {
	if tab = strings.ToUpper(strings.TrimSpace(tab)); m.hasTab(tab) {
		m.tab = tab
	}
	return m
}

// Tabs are the tabs that are shown, in order.
func (m Model) Tabs() []string { return append([]string(nil), m.tabs...) }

// Slide is the slider the keyboard is on (0 to 2) on the RGB and HSL tabs.
func (m Model) Slide() int { return m.slide }

// SetSlide puts the keyboard on a slider (0 to 2).
func (m Model) SetSlide(slide int) Model {
	m.slide = min(max(slide, 0), 2)
	return m
}

// SetOrigin sets where the picker is drawn on the screen (see WithOrigin).
func (m Model) SetOrigin(x, y int) Model {
	m.ox, m.oy = x, y
	return m
}

// SetAccent sets the colour of the selected tab and slider.
func (m Model) SetAccent(colour string) Model {
	m.accent = colour
	return m
}

func (m Model) hasTab(tab string) bool {
	for _, t := range m.tabs {
		if t == tab {
			return true
		}
	}
	return false
}

// Init implements tea.Model's shape; the picker starts nothing.
func (m Model) Init() tea.Cmd { return nil }

// Update handles keys and the mouse. Keys: tab and shift+tab change tab; the
// arrows move over swatches, and on the sliders up and down choose a slider
// and left and right change it (shift: ten at a time). The mouse: click a tab,
// a swatch, [-] or [+], or press and drag on a bar. It returns a command that
// produces ChangedMsg when the colour changed.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	before := m.value
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		m = m.key(msg.String())
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			m = m.pointer(msg.X-m.ox, msg.Y-m.oy, true)
		}
	case tea.MouseMotionMsg:
		if msg.Button == tea.MouseLeft {
			m = m.pointer(msg.X-m.ox, msg.Y-m.oy, false)
		}
	}
	if m.value != before {
		value := m.value
		return m, func() tea.Msg { return ChangedMsg{Value: value} }
	}
	return m, nil
}

// Width is the width of what View draws.
func (m Model) Width() int { return pickerWidth }

// Height is the number of lines View draws.
func (m Model) Height() int {
	_, _, h := m.layout()
	return h
}

// View draws the picker.
func (m Model) View() string {
	cv, _, h := m.layout()
	lines := make([]string, h)
	for y := 0; y < h; y++ {
		var out, run strings.Builder
		var cur pickerStyle
		flush := func() {
			if run.Len() == 0 {
				return
			}
			text := run.String()
			if cur != (pickerStyle{}) {
				st := lipgloss.NewStyle().Bold(cur.bold)
				if cur.fg != "" {
					st = st.Foreground(lipgloss.Color(cur.fg))
				}
				if cur.bg != "" {
					st = st.Background(lipgloss.Color(cur.bg))
				}
				text = st.Render(text)
			}
			out.WriteString(text)
			run.Reset()
		}
		for x := 0; x < pickerWidth; x++ {
			c := cv.at(x, y)
			if c.style != cur {
				flush()
				cur = c.style
			}
			ch := c.ch
			if ch == 0 {
				ch = ' '
			}
			run.WriteRune(ch)
		}
		flush()
		lines[y] = out.String()
	}
	return strings.Join(lines, "\n")
}

// pickerWidth is the width of the picker: fixed, so that a program can place
// it.
const pickerWidth = 46

// pickerBarWidth is how many cells a slider's gradient bar has.
const pickerBarWidth = 28

// Hit kinds.
const (
	pickerHitTab    = iota // arg: the index in the tabs
	pickerHitSwatch        // arg: the palette index
	pickerHitStep          // arg: slider*2, plus 1 for [+]
	pickerHitBar           // arg: the slider
)

type pickerHit struct {
	x, y, w, h int
	kind, arg  int
}

type pickerStyle struct {
	fg, bg string
	bold   bool
}

type pickerCell struct {
	ch    rune
	style pickerStyle
}

type pickerCanvas struct{ cells [pickerWidth * 32]pickerCell }

func (cv *pickerCanvas) at(x, y int) pickerCell {
	if x < 0 || y < 0 || x >= pickerWidth || y >= 32 {
		return pickerCell{}
	}
	return cv.cells[y*pickerWidth+x]
}

func (cv *pickerCanvas) set(x, y int, c pickerCell) {
	if x >= 0 && y >= 0 && x < pickerWidth && y < 32 {
		cv.cells[y*pickerWidth+x] = c
	}
}

func (cv *pickerCanvas) text(x, y int, s string, st pickerStyle) {
	for i, r := range []rune(s) {
		cv.set(x+i, y, pickerCell{ch: r, style: st})
	}
}

func (cv *pickerCanvas) fill(x, y, w, h int, st pickerStyle) {
	for yy := y; yy < y+h; yy++ {
		for xx := x; xx < x+w; xx++ {
			cv.set(xx, yy, pickerCell{ch: ' ', style: st})
		}
	}
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

func (c pickerRGB) hex() string {
	const digits = "0123456789abcdef"
	b := []byte{'#', 0, 0, 0, 0, 0, 0}
	for i, v := range []int{c.r, c.g, c.b} {
		v = min(max(v, 0), 255)
		b[1+2*i], b[2+2*i] = digits[v>>4], digits[v&15]
	}
	return string(b)
}

func (c pickerRGB) luma() float64 {
	return 0.299*float64(c.r) + 0.587*float64(c.g) + 0.114*float64(c.b)
}

// pickerParse reads a palette number or #rgb / #rrggbb. index is -1 for hex.
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

// pickerTabs keeps the known names, once each; none given means all four.
func pickerTabs(names []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, name := range names {
		n := strings.ToUpper(strings.TrimSpace(name))
		if (n == Tab16 || n == Tab256 || n == TabRGB || n == TabHSL) && !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		out = []string{Tab16, Tab256, TabRGB, TabHSL}
	}
	return out
}

func pickerContrast(c pickerRGB) string {
	if c.luma() > 140 {
		return "#000000"
	}
	return "#ffffff"
}

// current is the colour the sliders start from: the value, or a pink when
// there is none.
func (m Model) current() pickerRGB {
	if c, _, ok := pickerParse(m.value); ok {
		return c
	}
	return pickerRGB{255, 135, 215}
}

// hslNow is the hue, saturation and lightness of the colour; the remembered
// one while it still describes the value.
func (m Model) hslNow() pickerHSL {
	if m.hslOK && m.hslFor == m.value {
		return m.hsl
	}
	return m.current().hsl()
}

// sliderSet is the channel values, maxima, and units of the sliders of a tab.
func (m Model) sliders() (labels [3]string, max, value [3]float64, unit [3]string) {
	c := m.current()
	if m.tab == TabHSL {
		h := m.hslNow()
		return [3]string{"H", "S", "L"}, [3]float64{359, 100, 100}, [3]float64{h.h, h.s * 100, h.l * 100}, [3]string{"°", "%", "%"}
	}
	return [3]string{"R", "G", "B"}, [3]float64{255, 255, 255}, [3]float64{float64(c.r), float64(c.g), float64(c.b)}, [3]string{}
}

// setSlider sets slider i to v (clamped) and the colour with it.
func (m Model) setSlider(i int, v float64) Model {
	_, maxes, _, _ := m.sliders()
	v = math.Min(math.Max(math.Round(v), 0), maxes[i])
	if m.tab == TabHSL {
		h := m.hslNow()
		switch i {
		case 0:
			h.h = v
		case 1:
			h.s = v / 100
		default:
			h.l = v / 100
		}
		m.value = h.rgb().hex()
		m.hsl, m.hslFor, m.hslOK = h, m.value, true
		return m
	}
	c := m.current()
	switch i {
	case 0:
		c.r = int(v)
	case 1:
		c.g = int(v)
	default:
		c.b = int(v)
	}
	m.value = c.hex()
	return m
}

// layout draws the picker and returns the cells, what each part does and the
// number of lines used.
func (m Model) layout() (*pickerCanvas, []pickerHit, int) {
	cv := &pickerCanvas{}
	var hits []pickerHit
	dim := pickerStyle{fg: "245"}
	c, index, ok := pickerParse(m.value)

	active := 0
	for i, name := range m.tabs {
		if name == m.tab {
			active = i
		}
	}
	x := 1
	for i, name := range m.tabs {
		label := " " + name + " "
		style := dim
		if i == active {
			style = pickerStyle{fg: "0", bg: m.accent, bold: true}
		}
		cv.text(x, 0, label, style)
		hits = append(hits, pickerHit{x: x, y: 0, w: len(label), h: 1, kind: pickerHitTab, arg: i})
		x += len(label) + 1
	}

	swatch := func(x, y, w, h, idx int) {
		bg := strconv.Itoa(idx)
		cv.fill(x, y, w, h, pickerStyle{bg: bg})
		if ok && index == idx {
			cv.set(x+(w-1)/2, y, pickerCell{ch: '✓', style: pickerStyle{fg: pickerContrast(pickerANSI(idx)), bg: bg}})
		}
		hits = append(hits, pickerHit{x: x, y: y, w: w, h: h, kind: pickerHitSwatch, arg: idx})
	}

	y := 2
	switch m.tab {
	case Tab16:
		for blockRow := 0; blockRow < 2; blockRow++ {
			for col := 0; col < 8; col++ {
				swatch(1+col*6, y, 5, 2, blockRow*8+col)
			}
			y += 3
		}
		y--
	case Tab256:
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
		labels, maxes, values, units := m.sliders()
		cur := m.current()
		h := m.hslNow()
		at := func(i int, t float64) pickerRGB {
			if m.tab == TabHSL {
				switch i {
				case 0:
					return pickerHSL{t * 359, 1, 0.5}.rgb()
				case 1:
					return pickerHSL{h.h, t, h.l}.rgb()
				}
				return pickerHSL{h.h, h.s, t}.rgb()
			}
			out, v := cur, int(math.Round(t*255))
			switch i {
			case 0:
				out.r = v
			case 1:
				out.g = v
			default:
				out.b = v
			}
			return out
		}
		for i := 0; i < 3; i++ {
			style := pickerStyle{bold: true}
			if i == m.slide {
				style = pickerStyle{fg: "0", bg: m.accent, bold: true}
			}
			cv.text(1, y, " "+labels[i]+" ", style)
			cv.text(4, y, "[-]", pickerStyle{fg: "0", bg: "250"})
			hits = append(hits, pickerHit{x: 4, y: y, w: 3, h: 1, kind: pickerHitStep, arg: i * 2})
			barX := 8
			knob := int(math.Round(values[i] / maxes[i] * float64(pickerBarWidth-1)))
			for k := 0; k < pickerBarWidth; k++ {
				col := at(i, float64(k)/float64(pickerBarWidth-1))
				cell := pickerCell{ch: ' ', style: pickerStyle{bg: col.hex()}}
				if k == knob {
					cell = pickerCell{ch: '●', style: pickerStyle{fg: pickerContrast(col), bg: col.hex(), bold: true}}
				}
				cv.set(barX+k, y, cell)
			}
			hits = append(hits, pickerHit{x: barX, y: y, w: pickerBarWidth, h: 1, kind: pickerHitBar, arg: i})
			cv.text(barX+pickerBarWidth+1, y, "[+]", pickerStyle{fg: "0", bg: "250"})
			hits = append(hits, pickerHit{x: barX + pickerBarWidth + 1, y: y, w: 3, h: 1, kind: pickerHitStep, arg: i*2 + 1})
			cv.text(barX+pickerBarWidth+5, y, strconv.Itoa(int(math.Round(values[i])))+units[i], pickerStyle{bold: true})
			y += 2
		}
		y--
	}

	y++
	if ok {
		bg := c.hex()
		if index >= 0 {
			bg = strconv.Itoa(index)
		}
		cv.fill(1, y, 10, 1, pickerStyle{bg: bg})
		name := c.hex() + "  true colour"
		if index >= 0 {
			name = "palette " + strconv.Itoa(index) + "  " + c.hex()
		}
		cv.text(12, y, name, dim)
	} else {
		cv.text(1, y, "  (none)  ", dim)
	}
	value := m.value
	if value == "" {
		value = "none"
	}
	cv.text(1, y+1, "Value ", dim)
	cv.text(7, y+1, value, pickerStyle{bold: true})
	return cv, hits, y + 2
}

func (m Model) key(k string) Model {
	switch k {
	case "tab", "shift+tab":
		i := 0
		for j, t := range m.tabs {
			if t == m.tab {
				i = j
			}
		}
		step := 1
		if k == "shift+tab" {
			step = len(m.tabs) - 1
		}
		m.tab = m.tabs[(i+step)%len(m.tabs)]
		return m
	}
	if m.tab == Tab16 || m.tab == Tab256 {
		return m.moveSwatch(k)
	}
	switch k {
	case "up":
		m.slide = max(m.slide-1, 0)
	case "down":
		m.slide = min(m.slide+1, 2)
	case "left", "right", "shift+left", "shift+right":
		step := 1.0
		if strings.HasPrefix(k, "shift+") {
			step = 10
		}
		if strings.HasSuffix(k, "left") {
			step = -step
		}
		_, _, values, _ := m.sliders()
		return m.setSlider(m.slide, values[m.slide]+step)
	}
	return m
}

// moveSwatch moves the choice to the neighbouring swatch in a direction.
func (m Model) moveSwatch(k string) Model {
	_, hits, _ := m.layout()
	var swatches []pickerHit
	for _, h := range hits {
		if h.kind == pickerHitSwatch {
			swatches = append(swatches, h)
		}
	}
	_, index, ok := pickerParse(m.value)
	if !ok || index < 0 {
		if len(swatches) > 0 && (k == "left" || k == "right" || k == "up" || k == "down") {
			m.value = strconv.Itoa(swatches[0].arg)
		}
		return m
	}
	var cur *pickerHit
	for i := range swatches {
		if swatches[i].arg == index {
			cur = &swatches[i]
		}
	}
	if cur == nil {
		return m
	}
	overlapY := func(h pickerHit) bool { return h.y < cur.y+cur.h && h.y+h.h > cur.y }
	overlapX := func(h pickerHit) bool { return h.x < cur.x+cur.w && h.x+h.w > cur.x }
	var cand []pickerHit
	for _, h := range swatches {
		switch {
		case h.arg == cur.arg:
		case k == "left" && h.x+h.w <= cur.x && overlapY(h):
			cand = append(cand, h)
		case k == "right" && h.x >= cur.x+cur.w && overlapY(h):
			cand = append(cand, h)
		case k == "up" && h.y+h.h <= cur.y && overlapX(h):
			cand = append(cand, h)
		case k == "down" && h.y >= cur.y+cur.h && overlapX(h):
			cand = append(cand, h)
		}
	}
	if len(cand) == 0 {
		return m
	}
	abs := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
	// The nearest in the direction of the key, then the best aligned.
	distance := func(h pickerHit) int {
		if k == "left" || k == "right" {
			return abs(h.x-cur.x)*4 + abs(h.y-cur.y)
		}
		return abs(h.y-cur.y)*4 + abs(h.x-cur.x)
	}
	sort.SliceStable(cand, func(i, j int) bool { return distance(cand[i]) < distance(cand[j]) })
	m.value = strconv.Itoa(cand[0].arg)
	return m
}

// pointer applies a press (click true) or a drag at a cell of the picker.
func (m Model) pointer(x, y int, click bool) Model {
	_, hits, _ := m.layout()
	for _, h := range hits {
		if x < h.x || x >= h.x+h.w || y < h.y || y >= h.y+h.h {
			continue
		}
		switch h.kind {
		case pickerHitTab:
			if click {
				m.tab = m.tabs[h.arg]
			}
		case pickerHitSwatch:
			m.value = strconv.Itoa(h.arg)
		case pickerHitStep:
			if click {
				slider := h.arg / 2
				step := -1.0
				if h.arg%2 == 1 {
					step = 1
				}
				_, _, values, _ := m.sliders()
				m = m.setSlider(slider, math.Round(values[slider])+step)
				m.slide = slider
			}
		case pickerHitBar:
			t := float64(min(max(x-h.x, 0), pickerBarWidth-1)) / float64(pickerBarWidth-1)
			_, maxes, _, _ := m.sliders()
			m = m.setSlider(h.arg, math.Round(t*maxes[h.arg]))
			m.slide = h.arg
		}
		return m
	}
	return m
}
