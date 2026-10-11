package inspector

import (
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/theme"
	"github.com/meta-ui-labs/cuppa/libs/document/container"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// layerDrag is a layer being dragged in the list. Pressing a layer does
// nothing yet: moving to another row starts the drag, and releasing without
// moving selects it.
type layerDrag struct {
	id     design.NodeID
	startY int
	// selectIt is what a plain click on the layer does; it runs on release.
	selectIt func()
	moved    bool
	// row is the list row the layer would land on (0 is the front).
	row       int
	overTrash bool
}

// Dragging reports whether a layer is being dragged.
func (m *Model) Dragging() bool { return m.drag != nil && m.drag.moved }

// CancelDrag drops the drag without changing anything.
func (m *Model) CancelDrag() { m.drag = nil }

// layers lists the layers front to back. Each row has an eye (show or hide) and
// a padlock (lock or unlock); the name selects it and can be dragged to a new
// place in the list, or onto the trash button in the panel header.
func (m *Model) layers(b *builder) {
	doc := m.ed.Document()
	n := len(doc.Nodes)
	m.layerCount = n
	if n == 0 {
		b.text(theme.Faded(" (none)")).end()
		return
	}
	// A container is followed by its members, indented, as a tree; with no
	// containers this is the nodes front to back.
	m.layerIDs = m.layerIDs[:0]
	order := container.Order(doc)
	var tree []design.Node
	for k := len(order) - 1; k >= 0; k-- {
		node := doc.Nodes[order[k]]
		if node.In != "" {
			continue
		}
		tree = append(tree, node)
		if node.IsContainer() {
			for j := len(order) - 1; j >= 0; j-- {
				if member := doc.Nodes[order[j]]; member.In == node.ID {
					tree = append(tree, member)
				}
			}
		}
	}
	for row, node := range tree {
		m.layerIDs = append(m.layerIDs, node.ID)
		m.layerRow(b, node, row)
		b.end()
	}
}

// trashButton draws the delete button: click it to delete the selected layers,
// or drop a dragged layer on it.
func (m *Model) trashButton(b *builder) {
	label := "[Trash]"
	var styled string
	switch {
	case m.drag != nil && m.drag.overTrash:
		styled = theme.Selected(label)
	case m.drag != nil && m.drag.moved:
		styled = theme.Title(label)
	default:
		styled = theme.Button(label, len(m.ed.Selected()) > 0)
	}
	m.trashX0 = b.x
	b.add(styled, func() { m.ed.Delete() })
	m.trashX1 = b.x
}

func (m *Model) layerRow(b *builder, node design.Node, row int) {
	id := node.ID
	marker := " "
	if m.drag != nil && m.drag.moved && !m.drag.overTrash && m.drag.row == row {
		marker = theme.Title("▸")
	}
	b.text(marker)
	eye := "◉"
	if node.Hidden {
		eye = theme.Faded("○")
	}
	b.addMouse(eye, func() { m.ed.SetHidden(id, !node.Hidden) }).text(" ")
	lock := theme.Faded("▢")
	if node.Locked {
		lock = theme.Title("▣")
	}
	b.addMouse(lock, func() { m.ed.SetLocked(id, !node.Locked) }).text(" ")

	name := node.Name
	if node.In != "" {
		name = "  " + name
	}
	rest := max(m.content()-b.x, 1)
	var styled string
	switch {
	case m.drag != nil && m.drag.moved && m.drag.id == id:
		styled = theme.Faded(theme.Fit(name, rest))
	case m.ed.IsSelected(id):
		styled = theme.Selected(theme.Fit(name, rest))
	case node.Hidden:
		styled = theme.Faded(name)
	default:
		styled = name
	}
	b.add(styled, func() { m.ed.Select(id) })
	b.regions[len(b.regions)-1].layer = id
}

// pressLayer starts watching for a drag of the layer under the pointer.
func (m *Model) pressLayer(id design.NodeID, y int, selectIt func()) {
	m.drag = &layerDrag{id: id, startY: y, selectIt: selectIt}
}

// layersView is the Layers panel as it was last drawn.
func (m *Model) layersView() (panelView, bool) {
	for _, v := range m.views {
		if v.id == panelLayers && v.open {
			return v, true
		}
	}
	return panelView{}, false
}

// dragTo follows the pointer while a layer is held. The row it is over is
// counted in the panel's own scrolled list, and a pointer above or below the
// panel stays on the first or last row.
func (m *Model) dragTo(e pointer.Event) {
	d := m.drag
	if e.Y != d.startY {
		d.moved = true
	}
	v, ok := m.layersView()
	if !ok {
		return
	}
	d.overTrash = e.Y == v.top && e.X >= m.trashX0 && e.X < m.trashX1
	if m.layerCount > 0 {
		row := e.Y - v.top - 1 + m.state(v.id).scroll
		d.row = min(max(row, 0), m.layerCount-1)
	}
}

// drop ends the drag: onto the trash deletes, elsewhere moves the layer to the
// row it was dropped on. A press that never moved was just a selection.
func (m *Model) drop(e pointer.Event) {
	d := m.drag
	if d == nil {
		return
	}
	if d.moved {
		m.dragTo(e) // where the pointer was let go
	}
	m.drag = nil
	if !d.moved {
		if d.selectIt != nil {
			d.selectIt()
		}
		return
	}
	if d.overTrash {
		if m.ed.IsSelected(d.id) {
			m.ed.Delete()
		} else {
			m.ed.DeleteLayer(d.id)
		}
		return
	}
	to := m.layerCount - 1 - d.row
	if d.row < len(m.layerIDs) {
		// The row is a place in the tree: it is the place of the layer shown there.
		to = m.ed.Document().Index(m.layerIDs[d.row])
	}
	m.ed.MoveLayer(d.id, to)
}
