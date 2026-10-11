package editor

import "github.com/meta-tui/cuppa/libs/document/design"

// SetHidden shows or hides a layer, as one undo step. A hidden layer is not
// drawn, exported or hit by the pointer, and it leaves the selection.
func (e *Editor) SetHidden(id design.NodeID, hidden bool) bool {
	n, ok := e.doc.Get(id)
	if !ok || n.Hidden == hidden {
		return false
	}
	changed := e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.Hidden = hidden })
	})
	if changed && hidden {
		e.sel = e.keepSelected(func(other design.NodeID) bool { return other != id })
	}
	return changed
}

// SetLocked locks or unlocks a layer, as one undo step. A locked layer is still
// drawn and selectable but cannot be moved, resized, deleted or edited.
func (e *Editor) SetLocked(id design.NodeID, locked bool) bool {
	n, ok := e.doc.Get(id)
	if !ok || n.Locked == locked {
		return false
	}
	return e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.Locked = locked })
	})
}

// MoveLayer puts a layer at position to in z-order (0 is the back, the last
// index the front), as one undo step. If the layer is part of a multiple
// selection, the selected layers move together and keep their relative order.
func (e *Editor) MoveLayer(id design.NodeID, to int) bool {
	if e.doc.Index(id) < 0 {
		return false
	}
	ids := []design.NodeID{id}
	if e.IsSelected(id) && len(e.sel) > 1 {
		ids = e.selectedInOrder() // back to front
	}
	return e.apply(func() bool {
		before := e.order()
		moved := make(map[design.NodeID]bool, len(ids))
		for _, m := range ids {
			moved[m] = true
		}
		// Take the moved layers out, then insert them as a block so that the
		// first of them ends up at index `to` (clamped).
		var rest []design.Node
		var block []design.Node
		for _, n := range e.doc.Nodes {
			if moved[n.ID] {
				block = append(block, n)
			} else {
				rest = append(rest, n)
			}
		}
		at := min(max(to, 0), len(rest))
		out := make([]design.Node, 0, len(e.doc.Nodes))
		out = append(out, rest[:at]...)
		out = append(out, block...)
		out = append(out, rest[at:]...)
		e.doc.Nodes = out
		return !sameOrder(before, e.order())
	})
}

// DeleteLayer removes one layer whether or not it is selected (the trash
// button and dropping a layer on it), unless it is locked. One undo step.
func (e *Editor) DeleteLayer(id design.NodeID) bool {
	n, ok := e.doc.Get(id)
	if !ok || n.Locked {
		return false
	}
	changed := e.apply(func() bool {
		if n.IsContainer() {
			e.removeMembersOf(map[design.NodeID]bool{id: true})
		}
		return e.doc.Remove(id)
	})
	if changed {
		e.sel = e.keepSelected(func(other design.NodeID) bool { return other != id })
	}
	return changed
}

// unlockedSelection is the selection back to front without locked layers.
func (e *Editor) unlockedSelection() []design.NodeID {
	var out []design.NodeID
	for _, id := range e.selectedInOrder() {
		if n, ok := e.doc.Get(id); ok && !n.Locked {
			out = append(out, id)
		}
	}
	return out
}

// keepSelected returns the selected ids for which keep is true.
func (e *Editor) keepSelected(keep func(design.NodeID) bool) []design.NodeID {
	var out []design.NodeID
	for _, id := range e.sel {
		if keep(id) {
			out = append(out, id)
		}
	}
	return out
}

func (e *Editor) order() []design.NodeID {
	ids := make([]design.NodeID, len(e.doc.Nodes))
	for i, n := range e.doc.Nodes {
		ids[i] = n.ID
	}
	return ids
}

func sameOrder(a, b []design.NodeID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// SetDraggable lets the person using the exported program move the component
// with the mouse (or not), as one undo step. A locked layer refuses.
func (e *Editor) SetDraggable(id design.NodeID, on bool) bool {
	n, ok := e.doc.Get(id)
	if !ok || n.Locked || n.Draggable == on {
		return false
	}
	return e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.Draggable = on })
	})
}

// SetResizable lets the person using the exported program resize the component
// by dragging its bottom-right corner (or not), as one undo step. A locked
// layer refuses.
func (e *Editor) SetResizable(id design.NodeID, on bool) bool {
	n, ok := e.doc.Get(id)
	if !ok || n.Locked || n.Resizable == on {
		return false
	}
	return e.apply(func() bool {
		return e.doc.Update(id, func(n *design.Node) { n.Resizable = on })
	})
}
