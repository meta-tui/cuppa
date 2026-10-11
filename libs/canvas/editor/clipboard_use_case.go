package editor

import "github.com/meta-ui-labs/cuppa/libs/document/design"

// Copy remembers the selected components. The clipboard belongs to the editor,
// not the document, so it survives New and Open: a design can be pasted into
// another. It reports whether anything was copied.
func (e *Editor) Copy() bool {
	ids := e.selectedInOrder()
	if len(ids) == 0 {
		return false
	}
	e.clip = e.clip[:0]
	for _, id := range ids {
		n, _ := e.doc.Get(id)
		e.clip = append(e.clip, n.Clone())
	}
	e.pastes = 0
	return true
}

// CanPaste reports whether there is something to paste.
func (e *Editor) CanPaste() bool { return len(e.clip) > 0 }

// Paste adds a copy of what was copied, each time a little further down and to
// the right so copies do not hide each other, as one undo step. The copies are
// visible, unlocked, in the same front-to-back order, and end up selected.
func (e *Editor) Paste() bool {
	if len(e.clip) == 0 {
		return false
	}
	e.pastes++
	var pasted []design.NodeID
	changed := e.apply(func() bool {
		for _, n := range e.clip {
			c := n.Clone()
			c.ID = e.doc.NewID()
			c.Locked, c.Hidden = false, false
			c.Rect = n.Rect.Translate(e.pastes*duplicateDX, e.pastes*duplicateDY).MoveInto(e.doc.Bounds())
			c.Name = e.uniqueName(baseName(n.Name))
			pasted = append(pasted, e.doc.Add(c).ID)
		}
		return true
	})
	if changed {
		e.Select(pasted...)
	}
	return changed
}
