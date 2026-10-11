package editor

import (
	"fmt"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// How far a duplicate sits from its original, in cells.
const (
	duplicateDX = 2
	duplicateDY = 1
)

// Add places a new node of the given component with its top-left at (x, y),
// shifted to stay on the canvas, and selects it.
func (e *Editor) Add(componentID string, x, y int) (design.NodeID, error) {
	def, ok := e.cat.Get(componentID)
	if !ok {
		return "", fmt.Errorf("editor: unknown component %q", componentID)
	}
	var added design.Node
	e.apply(func() bool {
		r := design.Rect{X: x, Y: y, W: min(def.DefaultSize.W, e.doc.Width), H: min(def.DefaultSize.H, e.doc.Height)}
		n := design.Node{
			Component: componentID,
			Name:      e.uniqueName(def.Name),
			Rect:      r.MoveInto(e.doc.Bounds()),
		}
		// A component dropped on the shown page of a container belongs to it.
		e.joinContainer(&n)
		added = e.doc.Add(n)
		return true
	})
	e.Select(added.ID)
	return added.ID, nil
}

// Delete removes the selected nodes, except locked ones.
func (e *Editor) Delete() bool {
	ids := e.unlockedSelection()
	changed := e.apply(func() bool {
		removed := map[design.NodeID]bool{}
		for _, id := range ids {
			if n, ok := e.doc.Get(id); ok && n.IsContainer() {
				removed[id] = true
			}
			e.doc.Remove(id)
		}
		e.removeMembersOf(removed)
		return len(ids) > 0
	})
	e.sel = e.keepSelected(func(id design.NodeID) bool { return e.doc.Index(id) >= 0 })
	return changed
}

// Duplicate copies the selected nodes, offset a little, and selects the copies.
func (e *Editor) Duplicate() bool {
	ids := e.selectedInOrder()
	var copies []design.NodeID
	changed := e.apply(func() bool {
		for _, id := range ids {
			n, _ := e.doc.Get(id)
			c := n.Clone()
			c.ID = ""
			c.Locked, c.Hidden = false, false
			c.Rect = n.Rect.Translate(duplicateDX, duplicateDY).MoveInto(e.doc.Bounds())
			c.Name = e.uniqueName(baseName(n.Name))
			added := e.doc.Add(c)
			copies = append(copies, added.ID)
			if n.IsContainer() {
				e.duplicateMembers(n.ID, added.ID, duplicateDX, duplicateDY)
			}
		}
		return len(ids) > 0
	})
	if changed {
		e.Select(copies...)
	}
	return changed
}

// uniqueName returns "<base> N" with the smallest N not used by any node.
func (e *Editor) uniqueName(base string) string {
	used := make(map[string]bool, len(e.doc.Nodes))
	for _, n := range e.doc.Nodes {
		used[n.Name] = true
	}
	for i := 1; ; i++ {
		if name := fmt.Sprintf("%s %d", base, i); !used[name] {
			return name
		}
	}
}

// baseName strips a trailing " N" from a generated name.
func baseName(name string) string {
	for i := len(name) - 1; i > 0; i-- {
		if name[i] == ' ' && i < len(name)-1 {
			return name[:i]
		}
		if name[i] < '0' || name[i] > '9' {
			break
		}
	}
	return name
}
