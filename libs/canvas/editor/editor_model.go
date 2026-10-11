// Package editor is the headless design editor. It owns a document, the
// selection and the undo history, and exposes intent-based commands that any
// front end (terminal, desktop, web) can drive.
package editor

import (
	"reflect"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// Catalog is what the editor needs to know about components.
type Catalog interface {
	Get(id string) (definition.Definition, bool)
}

// Editor edits one document. It is not safe for concurrent use.
type Editor struct {
	cat   Catalog
	doc   design.Document
	sel   []design.NodeID
	undo  []snapshot
	redo  []snapshot
	saved design.Document
	// clip holds what Copy remembered, and pastes counts the pastes since, so
	// each lands a step further on.
	clip   []design.Node
	pastes int
}

// New returns an editor for doc.
func New(cat Catalog, doc design.Document) *Editor {
	e := &Editor{cat: cat, doc: doc.Clone()}
	e.Resolve()
	e.saved = e.doc.Clone()
	return e
}

// Document returns a copy of the current document.
func (e *Editor) Document() design.Document { return e.doc.Clone() }

// Load replaces the document, clearing selection and history. The loaded
// document counts as saved.
func (e *Editor) Load(doc design.Document) {
	e.doc = doc.Clone()
	e.Resolve()
	e.saved = e.doc.Clone()
	e.sel, e.undo, e.redo = nil, nil, nil
}

// Dirty reports whether the document differs from the last loaded or saved one.
func (e *Editor) Dirty() bool { return !reflect.DeepEqual(e.doc, e.saved) }

// MarkSaved records the current document as saved.
func (e *Editor) MarkSaved() { e.saved = e.doc.Clone() }

// Selected returns the selected ids in selection order.
func (e *Editor) Selected() []design.NodeID { return append([]design.NodeID(nil), e.sel...) }

// Primary is the most recently selected node.
func (e *Editor) Primary() (design.Node, bool) {
	if len(e.sel) == 0 {
		return design.Node{}, false
	}
	return e.doc.Get(e.sel[len(e.sel)-1])
}

// IsSelected reports whether the node is selected.
func (e *Editor) IsSelected(id design.NodeID) bool {
	for _, s := range e.sel {
		if s == id {
			return true
		}
	}
	return false
}

// Select makes ids the whole selection. Unknown ids are ignored.
func (e *Editor) Select(ids ...design.NodeID) {
	e.sel = nil
	for _, id := range ids {
		e.addToSelection(id)
	}
	e.revealMembers()
}

// Toggle adds the node to the selection, or removes it if already selected.
func (e *Editor) Toggle(id design.NodeID) {
	if !e.IsSelected(id) {
		e.addToSelection(id)
		e.revealMembers()
		return
	}
	kept := e.sel[:0]
	for _, s := range e.sel {
		if s != id {
			kept = append(kept, s)
		}
	}
	e.sel = kept
}

// Clear deselects everything.
func (e *Editor) Clear() { e.sel = nil }

func (e *Editor) addToSelection(id design.NodeID) {
	if e.doc.Index(id) >= 0 && !e.IsSelected(id) {
		e.sel = append(e.sel, id)
	}
}

// selectedInOrder returns the selected ids sorted back to front.
func (e *Editor) selectedInOrder() []design.NodeID {
	var out []design.NodeID
	for _, n := range e.doc.Nodes {
		if e.IsSelected(n.ID) {
			out = append(out, n.ID)
		}
	}
	return out
}

// minSize is the smallest size the node's component allows.
func (e *Editor) minSize(n design.Node) (int, int) {
	if def, ok := e.cat.Get(n.Component); ok {
		return max(def.MinSize.W, 1), max(def.MinSize.H, 1)
	}
	return 1, 1
}
