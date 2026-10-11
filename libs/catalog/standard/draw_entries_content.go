package standard

import (
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

// drawEntries is the drawing layer. The drawing tools make it; it is not in
// the palette.
func drawEntries() []definition.Definition {
	return []definition.Definition{{
		ID: drawlayer.Component, Name: "Drawing", Hidden: true,
		Description: "What the drawing tools paint: rectangles, paths, brush strokes and text.",
		DefaultSize: size(120, 40), MinSize: size(1, 1), Import: lipglossImport, Status: definition.StatusSupported,
		Props: []definition.PropSpec{textProp(drawlayer.PropCells, "Cells", "")},
	}}
}
