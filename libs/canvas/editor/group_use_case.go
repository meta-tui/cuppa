package editor

import "github.com/meta-ui-labs/cuppa/libs/document/design"

// CanGroup reports whether Group would do something: two or more selected
// components, none of them locked.
func (e *Editor) CanGroup() bool {
	ids := e.selectedInOrder()
	for _, id := range ids {
		// A container and its members are kept as they are: a group would
		// take them out of their pages.
		if n, ok := e.doc.Get(id); ok && (n.IsContainer() || n.In != "") {
			return false
		}
	}
	return len(ids) >= 2 && len(e.unlockedSelection()) == len(ids)
}

// CanUngroup reports whether the selection holds an unlocked group.
func (e *Editor) CanUngroup() bool { return len(e.ungroupable()) > 0 }

// Group turns the selected components into one group that moves, resizes and
// layers as a single component, as one undo step. The group takes the place of
// its front-most member in z-order and ends up selected.
func (e *Editor) Group() bool {
	if !e.CanGroup() {
		return false
	}
	members := make(map[design.NodeID]bool)
	for _, id := range e.selectedInOrder() {
		members[id] = true
	}
	var group design.NodeID
	changed := e.apply(func() bool {
		var rest, inside []design.Node
		var box design.Rect
		at := 0 // where the group goes among the nodes that stay
		for _, n := range e.doc.Nodes {
			if !members[n.ID] {
				rest = append(rest, n)
				continue
			}
			if len(inside) == 0 {
				box = n.Rect
			} else {
				box = box.Union(n.Rect)
			}
			inside = append(inside, n)
			at = len(rest)
		}
		for i := range inside {
			inside[i].Rect = inside[i].Rect.Translate(-box.X, -box.Y)
		}
		g := design.Node{
			ID:        e.doc.NewID(),
			Component: design.GroupComponent,
			Rect:      box,
			Children:  inside,
			BaseW:     box.W,
			BaseH:     box.H,
		}
		e.doc.Nodes = append(rest[:at:at], append([]design.Node{g}, rest[at:]...)...)
		e.doc.Update(g.ID, func(n *design.Node) { n.Name = e.uniqueName("Group") })
		group = g.ID
		return true
	})
	if changed {
		e.Select(group)
	}
	return changed
}

// ungroupable lists the selected groups that are not locked.
func (e *Editor) ungroupable() []design.NodeID {
	var out []design.NodeID
	for _, id := range e.selectedInOrder() {
		if n, ok := e.doc.Get(id); ok && n.IsGroup() && !n.Locked {
			out = append(out, id)
		}
	}
	return out
}

// Ungroup replaces each selected group by the components it holds, at the size
// and place the group has now, as one undo step. The released components end
// up selected.
func (e *Editor) Ungroup() bool {
	groups := e.ungroupable()
	if len(groups) == 0 {
		return false
	}
	var released []design.NodeID
	changed := e.apply(func() bool {
		for _, id := range groups {
			i := e.doc.Index(id)
			g := e.doc.Nodes[i]
			parts := make([]design.Node, 0, len(g.Children))
			for _, c := range g.Children {
				c = c.Clone()
				c.Rect = c.Rect.Scale(g.BaseW, g.BaseH, g.Rect.W, g.Rect.H).Translate(g.Rect.X, g.Rect.Y)
				c.Hidden = c.Hidden || g.Hidden
				c.ID = e.doc.NewID()
				released = append(released, c.ID)
				parts = append(parts, c)
			}
			nodes := append([]design.Node(nil), e.doc.Nodes[:i]...)
			nodes = append(nodes, parts...)
			e.doc.Nodes = append(nodes, e.doc.Nodes[i+1:]...)
		}
		return true
	})
	if changed {
		e.Select(released...)
	}
	return changed
}
