package stage

import "github.com/meta-ui-labs/cuppa/libs/document/drawlayer"

// SetPreview shows cells over the canvas: what a drawing tool would draw, or,
// with erase, what it would take away. Pass nil to clear it.
func (m *Model) SetPreview(cells []drawlayer.Cell, erase bool) {
	m.preview, m.previewErase = cells, erase
}
