package stage

import "github.com/meta-ui-labs/cuppa/libs/render/grid"

// A light terminal is previewed with a pale canvas and dark default text: what
// a program that sets no colours of its own looks like there.
const (
	lightCanvas = "255"
	lightDots   = "250"
	lightText   = "16"
)

// onLight gives the cells that rely on the terminal's own colours the colours a
// light terminal would use. Colours the design chose are left alone.
func onLight(s grid.Style) grid.Style {
	if s.Fg == "" {
		s.Fg = lightText
	}
	if s.Bg == "" {
		s.Bg = lightCanvas
	}
	return s
}
