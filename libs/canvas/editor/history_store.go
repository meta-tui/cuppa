package editor

import (
	"reflect"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

const maxHistory = 200

// snapshot is one undo step: the document and selection before a change.
type snapshot struct {
	doc design.Document
	sel []design.NodeID
}

func (e *Editor) snapshot() snapshot {
	return snapshot{doc: e.doc.Clone(), sel: e.Selected()}
}

func (e *Editor) restore(s snapshot) {
	e.doc = s.doc
	e.sel = nil
	for _, id := range s.sel {
		e.addToSelection(id)
	}
}

// Checkpoint records the current state as an undo step. Front ends call it
// once when a drag starts; one-shot commands call it themselves.
func (e *Editor) Checkpoint() {
	e.undo = append(e.undo, e.snapshot())
	if len(e.undo) > maxHistory {
		e.undo = e.undo[len(e.undo)-maxHistory:]
	}
	e.redo = nil
}

// EndGesture drops the last checkpoint if nothing changed since it, so a click
// that moved nothing leaves no undo step.
func (e *Editor) EndGesture() {
	if n := len(e.undo); n > 0 && reflect.DeepEqual(e.undo[n-1].doc, e.doc) {
		e.undo = e.undo[:n-1]
	}
}

// apply runs a one-shot change as a single undo step. fn reports whether it
// changed anything; unchanged runs leave no step.
func (e *Editor) apply(fn func() bool) bool {
	e.Checkpoint()
	if !fn() {
		e.undo = e.undo[:len(e.undo)-1]
		return false
	}
	return true
}

// CanUndo reports whether there is a step to undo.
func (e *Editor) CanUndo() bool { return len(e.undo) > 0 }

// CanRedo reports whether there is a step to redo.
func (e *Editor) CanRedo() bool { return len(e.redo) > 0 }

// Undo reverts the last step.
func (e *Editor) Undo() bool {
	n := len(e.undo)
	if n == 0 {
		return false
	}
	e.redo = append(e.redo, e.snapshot())
	e.restore(e.undo[n-1])
	e.undo = e.undo[:n-1]
	return true
}

// Redo re-applies the last undone step.
func (e *Editor) Redo() bool {
	n := len(e.redo)
	if n == 0 {
		return false
	}
	e.undo = append(e.undo, e.snapshot())
	e.restore(e.redo[n-1])
	e.redo = e.redo[:n-1]
	return true
}
