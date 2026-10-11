package editor

import "github.com/meta-ui-labs/cuppa/libs/document/design"

// Order is a z-order operation.
type Order int

// Z-order operations.
const (
	BringToFront Order = iota
	BringForward
	SendBackward
	SendToBack
)

// Reorder applies the operation to the selection, keeping the selected nodes'
// relative order.
func (e *Editor) Reorder(op Order) bool {
	ids := e.selectedInOrder() // back to front
	return e.apply(func() bool {
		changed := false
		switch op {
		case BringToFront:
			for _, id := range ids {
				changed = e.doc.ToFront(id) || changed
			}
		case SendToBack:
			for i := len(ids) - 1; i >= 0; i-- {
				changed = e.doc.ToBack(ids[i]) || changed
			}
		case BringForward:
			for i := len(ids) - 1; i >= 0; i-- {
				changed = e.stepUp(ids[i]) || changed
			}
		case SendBackward:
			for _, id := range ids {
				changed = e.stepDown(id) || changed
			}
		}
		return changed
	})
}

// stepUp raises the node one place unless the node above it is also selected
// (which then moves as a block).
func (e *Editor) stepUp(id design.NodeID) bool {
	i := e.doc.Index(id)
	if i < 0 || i+1 >= len(e.doc.Nodes) || e.IsSelected(e.doc.Nodes[i+1].ID) {
		return false
	}
	return e.doc.Raise(id)
}

func (e *Editor) stepDown(id design.NodeID) bool {
	i := e.doc.Index(id)
	if i <= 0 || e.IsSelected(e.doc.Nodes[i-1].ID) {
		return false
	}
	return e.doc.Lower(id)
}
