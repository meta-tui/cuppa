package shell

import (
	"github.com/meta-tui/cuppa/apps/cuppa-tui/inspector"
	"github.com/meta-tui/cuppa/apps/cuppa-tui/tools"
	"github.com/meta-tui/cuppa/libs/document/design"
)

// layout places the three panes below the title bar and above the status bar.
// A one-cell separator column sits between neighbouring panes.
type layout struct {
	// tools is the tool list at the top of the left bar, palette the component
	// list under it, options the contextual bar above the canvas, and stage
	// the canvas below that.
	tools, palette, options, stage, inspector design.Rect
}

const (
	preferredPaletteW   = 26
	preferredInspectorW = 32
	minPaletteW         = 16
	minInspectorW       = 20
	minStageW           = 20
	chromeRows          = 2 // title bar + status bar
)

// computeLayout sizes the panes for a terminal of w by h cells. palW and insW
// are the side bars' wanted widths (0 for the default). The side bars give way
// first when the terminal is narrow, so the stage always keeps room.
func computeLayout(w, h, palW, insW int) layout {
	pal, ins := palW, insW
	if pal <= 0 {
		pal = preferredPaletteW
	}
	if ins <= 0 {
		ins = preferredInspectorW
	}
	pal, ins = max(pal, minPaletteW), max(ins, minInspectorW)
	for w-pal-ins-2 < minStageW && (pal > minPaletteW || ins > minInspectorW) {
		if pal > minPaletteW {
			pal--
		}
		if ins > minInspectorW {
			ins--
		}
	}
	stageW := max(w-pal-ins-2, 1)
	body := max(h-chromeRows, 1)
	toolRows := min(tools.Rows, max(body-1, 0))
	return layout{
		tools:     design.Rect{X: 0, Y: 1, W: pal, H: toolRows},
		palette:   design.Rect{X: 0, Y: 1 + toolRows, W: pal, H: max(body-toolRows, 0)},
		options:   design.Rect{X: pal + 1, Y: 1, W: stageW, H: inspector.ContextRows},
		stage:     design.Rect{X: pal + 1, Y: 1 + inspector.ContextRows, W: stageW, H: max(body-inspector.ContextRows, 1)},
		inspector: design.Rect{X: pal + 1 + stageW + 1, Y: 1, W: ins, H: body},
	}
}

// divider is one of the two columns between the panes that can be dragged.
type divider int

const (
	noDivider divider = iota
	leftDivider
	rightDivider
)

// dividerAt returns the divider under a screen cell, if any.
func (l layout) dividerAt(x, y int) divider {
	if y < l.inspector.Y || y >= l.inspector.Y+l.inspector.H {
		return noDivider
	}
	switch x {
	case l.palette.X + l.palette.W:
		return leftDivider
	case l.inspector.X - 1:
		return rightDivider
	}
	return noDivider
}

// widthsFor returns the side bar widths that put the divider at screen column
// x, within the limits (stage at least minStageW wide).
func widthsFor(w, pal, ins int, d divider, x int) (int, int) {
	switch d {
	case leftDivider:
		pal = min(max(x, minPaletteW), max(w-ins-2-minStageW, minPaletteW))
	case rightDivider:
		ins = min(max(w-x-1, minInspectorW), max(w-pal-2-minStageW, minInspectorW))
	}
	return pal, ins
}

// paneAt returns the pane under a screen cell.
func (l layout) paneAt(x, y int) pane {
	switch {
	case l.tools.Contains(x, y):
		return inTools
	case l.options.Contains(x, y):
		return inOptions
	case l.palette.Contains(x, y):
		return inPalette
	case l.stage.Contains(x, y):
		return inStage
	case l.inspector.Contains(x, y):
		return inInspector
	}
	return nowhere
}
