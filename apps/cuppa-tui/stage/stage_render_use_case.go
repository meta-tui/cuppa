package stage

import (
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
	"github.com/meta-ui-labs/cuppa/libs/render/scene"
)

// render draws the visible part of the canvas, the selection and the drag
// feedback onto a grid the size of the pane.
func (m *Model) render() *grid.Grid {
	// The master's components are drawn below the page's own; they are not in
	// the document the person selects from.
	doc := m.ed.RenderDocument()
	rendered := scene.Render(doc, m.cat)
	if m.renderer != nil {
		rendered = m.renderer()
	}
	view := grid.New(m.w, m.h)
	canvasBg := doc.Background
	dotFg := theme.Faint
	if canvasBg == "" {
		canvasBg = theme.Canvas
		if doc.Light {
			canvasBg, dotFg = lightCanvas, lightDots
		}
	}
	for vy := 0; vy < m.h; vy++ {
		for vx := 0; vx < m.w; vx++ {
			cx, cy := m.Canvas(vx, vy)
			if cx >= doc.Width || cy >= doc.Height {
				continue
			}
			c := rendered.At(cx, cy)
			switch {
			case c.Ch != 0:
				if doc.Light && doc.Background == "" {
					c.Style = onLight(c.Style)
				}
				view.Set(vx, vy, c)
			case !doc.HideGrid && cx%4 == 0 && cy%2 == 0:
				view.Set(vx, vy, grid.Cell{Ch: '·', Style: grid.Style{Fg: dotFg, Bg: canvasBg, Dim: c.Dim}})
			default:
				view.Set(vx, vy, grid.Cell{Ch: ' ', Style: grid.Style{Bg: canvasBg, Dim: c.Dim}})
			}
		}
	}
	if m.renderer != nil {
		return view
	}
	selected := m.ed.Selected()
	var group design.Rect
	for i, id := range selected {
		if n, ok := doc.Get(id); ok {
			m.drawSelection(view, n.Rect, len(selected) == 1 && !n.Locked)
			if i == 0 {
				group = n.Rect
			} else {
				group = group.Union(n.Rect)
			}
		}
	}
	if len(selected) > 1 {
		m.drawGroupFrame(view, group)
	}
	m.drawGuides(view)
	if m.mode == marquee {
		m.outline(view, m.marqueeRect(), '·', grid.Style{Fg: theme.Accent})
	}
	if m.ghost != nil {
		m.outline(view, *m.ghost, '░', grid.Style{Fg: theme.Accent})
	}
	for _, c := range m.preview {
		vx, vy := c.X-m.offX, c.Y-m.offY
		if !view.In(vx, vy) {
			continue
		}
		if m.previewErase {
			view.Set(vx, vy, grid.Cell{Ch: '░', Style: grid.Style{Fg: theme.Accent}})
			continue
		}
		view.Set(vx, vy, grid.Cell{Ch: c.Ch, Style: grid.Style{Fg: c.Fg, Bg: c.Bg}})
	}
	if m.showCursor {
		vx, vy := m.cursor[0]-m.offX, m.cursor[1]-m.offY
		if view.In(vx, vy) {
			view.Set(vx, vy, grid.Cell{Ch: '┼', Style: grid.Style{Fg: theme.Accent, Bold: true}})
		}
	}
	return view
}

// toView converts a canvas rectangle to pane coordinates.
func (m *Model) toView(r design.Rect) design.Rect { return r.Translate(-m.offX, -m.offY) }

// drawSelection recolours the outline of a selected node and marks its corners:
// solid squares are resize handles (single selection), hollow ones only show
// that the node is part of a multiple selection. Components without a border
// (lists, pickers) would otherwise show no sign of being selected.
func (m *Model) drawSelection(view *grid.Grid, canvas design.Rect, handles bool) {
	r := m.toView(canvas)
	accent := func(s grid.Style) grid.Style { s.Fg, s.Bold = theme.Accent, true; return s }
	for x := r.X; x < r.Right(); x++ {
		view.Restyle(x, r.Y, accent)
		view.Restyle(x, r.Bottom()-1, accent)
	}
	for y := r.Y; y < r.Bottom(); y++ {
		view.Restyle(r.X, y, accent)
		view.Restyle(r.Right()-1, y, accent)
	}
	glyph := '□'
	if handles {
		glyph = '■'
	}
	mark := func(x, y int) {
		view.Set(x, y, grid.Cell{Ch: glyph, Style: grid.Style{Fg: theme.Accent, Bold: true}})
	}
	switch {
	case r.H == 1 && r.W >= 3:
		mark(r.X, r.Y)
		mark(r.Right()-1, r.Y)
	case r.H >= 2 && r.W >= 2:
		mark(r.X, r.Y)
		mark(r.Right()-1, r.Y)
		mark(r.X, r.Bottom()-1)
		mark(r.Right()-1, r.Bottom()-1)
	}
}

// drawGroupFrame draws a dashed frame around everything selected, over empty
// canvas cells only so it never hides a component.
func (m *Model) drawGroupFrame(view *grid.Grid, canvas design.Rect) {
	r := m.toView(canvas)
	r = design.Rect{X: r.X - 1, Y: r.Y - 1, W: r.W + 2, H: r.H + 2}
	style := grid.Style{Fg: theme.Accent}
	put := func(x, y int, ch rune) {
		if !view.In(x, y) {
			return
		}
		if c := view.At(x, y); c.Ch == 0 || c.Ch == ' ' || c.Ch == '·' {
			view.Set(x, y, grid.Cell{Ch: ch, Style: style})
		}
	}
	for x := r.X; x < r.Right(); x++ {
		put(x, r.Y, '╌')
		put(x, r.Bottom()-1, '╌')
	}
	for y := r.Y; y < r.Bottom(); y++ {
		put(r.X, y, '╎')
		put(r.Right()-1, y, '╎')
	}
}

// outline draws a one-cell frame of ch around the canvas rectangle.
func (m *Model) outline(view *grid.Grid, canvas design.Rect, ch rune, s grid.Style) {
	r := m.toView(canvas)
	for x := r.X; x < r.Right(); x++ {
		view.Set(x, r.Y, grid.Cell{Ch: ch, Style: s})
		view.Set(x, r.Bottom()-1, grid.Cell{Ch: ch, Style: s})
	}
	for y := r.Y; y < r.Bottom(); y++ {
		view.Set(r.X, y, grid.Cell{Ch: ch, Style: s})
		view.Set(r.Right()-1, y, grid.Cell{Ch: ch, Style: s})
	}
}

// drawGuides draws the snap guides over empty canvas cells only, so they never
// hide a component.
func (m *Model) drawGuides(view *grid.Grid) {
	guide := grid.Style{Fg: theme.Accent, Dim: true}
	blank := func(x, y int) bool {
		c := view.At(x, y)
		return c.Ch == 0 || c.Ch == ' ' || c.Ch == '·'
	}
	for _, gd := range m.guides {
		if gd.Vertical {
			x := gd.Pos - m.offX
			for y := 0; y < view.H; y++ {
				if blank(x, y) {
					view.Set(x, y, grid.Cell{Ch: '┆', Style: guide})
				}
			}
			continue
		}
		y := gd.Pos - m.offY
		for x := 0; x < view.W; x++ {
			if blank(x, y) {
				view.Set(x, y, grid.Cell{Ch: '┄', Style: guide})
			}
		}
	}
}
