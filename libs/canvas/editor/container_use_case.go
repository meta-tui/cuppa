package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/meta-tui/cuppa/libs/document/container"
	"github.com/meta-tui/cuppa/libs/document/design"
)

// shiftMembers moves the members of a container by how far the body of their
// page moved between before (the container as it was) and now: the container
// was moved or resized, or a section opened or closed. Members in skip were
// already moved by the change itself (they were selected with it).
func (e *Editor) shiftMembers(before design.Node, skip map[design.NodeID]bool) {
	if !before.IsContainer() {
		return
	}
	now, ok := e.doc.Get(before.ID)
	if !ok {
		return
	}
	moves := container.Moves(before, now)
	for i := range e.doc.Nodes {
		m := &e.doc.Nodes[i]
		if m.In != before.ID || skip[m.ID] || m.Page < 0 || m.Page >= len(moves) {
			continue
		}
		d := moves[m.Page]
		m.Rect = m.Rect.Translate(d[0], d[1])
	}
}

// containerAt is the container whose shown page body holds the cell, and the
// page.
func (e *Editor) containerAt(x, y int) (design.Node, int, bool) {
	order := container.Order(e.doc)
	for k := len(order) - 1; k >= 0; k-- {
		n := e.doc.Nodes[order[k]]
		if !n.IsContainer() || n.Hidden || n.In != "" {
			continue
		}
		if page, ok := container.PageAt(n, x, y); ok {
			return n, page, true
		}
	}
	return design.Node{}, 0, false
}

// joinContainer makes r the rectangle of a new member when its top-left is on
// the shown page of a container, keeping it inside the body.
func (e *Editor) joinContainer(n *design.Node) {
	if n.IsContainer() {
		return
	}
	c, page, ok := e.containerAt(n.Rect.X, n.Rect.Y)
	if !ok {
		return
	}
	body := container.Layout(c).Bodies[page]
	n.In, n.Page = c.ID, page
	n.Rect.W, n.Rect.H = min(n.Rect.W, body.W), min(n.Rect.H, body.H)
	n.Rect = n.Rect.MoveInto(body)
}

// Members are the components that belong to the container, back to front.
func (e *Editor) Members(id design.NodeID) []design.Node {
	var out []design.Node
	for _, n := range e.doc.Nodes {
		if n.In == id {
			out = append(out, n)
		}
	}
	return out
}

// PressContainerHeader does what pressing the tab or header of a page does:
// the tab becomes the shown one, a section opens or closes. It is one undo
// step, and a selected member whose page goes out of sight is deselected.
func (e *Editor) PressContainerHeader(id design.NodeID, page int) bool {
	c, ok := e.doc.Get(id)
	if !ok || !c.IsContainer() || page < 0 || page >= len(container.Pages(c)) {
		return false
	}
	changed := e.apply(func() bool {
		return e.setContainerProps(c, container.Pressed(c, page))
	})
	if changed {
		e.sel = e.keepSelected(func(id design.NodeID) bool {
			n, ok := e.doc.Get(id)
			return ok && container.Visible(e.doc, n)
		})
	}
	return changed
}

// setContainerProps writes properties of a container and moves its members
// with their pages. It reports whether anything changed.
func (e *Editor) setContainerProps(c design.Node, props map[string]string) bool {
	// c comes from the document and shares its property map, which is about to
	// be written: keep the state it had.
	c = c.Clone()
	changed := false
	for k, v := range props {
		if c.Props[k] != v {
			changed = true
		}
	}
	if !changed {
		return false
	}
	e.doc.Update(c.ID, func(n *design.Node) {
		if n.Props == nil {
			n.Props = map[string]string{}
		}
		for k, v := range props {
			n.Props[k] = v
		}
	})
	e.shiftMembers(c, nil)
	e.clampMemberPages(c.ID)
	return true
}

// clampMemberPages keeps the page of every member inside the container's pages.
func (e *Editor) clampMemberPages(id design.NodeID) {
	c, ok := e.doc.Get(id)
	if !ok {
		return
	}
	last := len(container.Pages(c)) - 1
	for i := range e.doc.Nodes {
		if m := &e.doc.Nodes[i]; m.In == id && m.Page > last {
			m.Page = last
		}
	}
}

// revealMembers shows the page of each selected member, as selecting it
// should: the tab is chosen or the section opened. It is not an undo step,
// because it only follows what the person picked.
func (e *Editor) revealMembers() {
	for _, id := range e.sel {
		m, ok := e.doc.Get(id)
		if !ok || m.In == "" {
			continue
		}
		c, ok := e.doc.Get(m.In)
		if !ok || !c.IsContainer() {
			continue
		}
		if props := container.Revealed(c, m.Page); props != nil {
			e.setContainerProps(c, props)
		}
	}
}

// AddPage appends a page (a tab, a section) to a container as one undo step.
func (e *Editor) AddPage(id design.NodeID, name string) error {
	c, ok := e.doc.Get(id)
	if !ok || !c.IsContainer() {
		return fmt.Errorf("editor: that is not a container")
	}
	if c.Locked {
		return fmt.Errorf("editor: %s is locked", c.Name)
	}
	name = strings.NewReplacer(",", " ").Replace(strings.TrimSpace(name))
	pages := container.Pages(c)
	if name == "" {
		name = fmt.Sprintf("Page %d", len(pages)+1)
	}
	e.apply(func() bool {
		return e.setContainerProps(c, map[string]string{container.PropPages: strings.Join(append(pages, name), ",")})
	})
	return nil
}

// RemovePage deletes a page of a container with the components on it, as one
// undo step. A container keeps at least one page.
func (e *Editor) RemovePage(id design.NodeID, page int) error {
	c, ok := e.doc.Get(id)
	if !ok || !c.IsContainer() {
		return fmt.Errorf("editor: that is not a container")
	}
	pages := container.Pages(c)
	if len(pages) < 2 || page < 0 || page >= len(pages) {
		return fmt.Errorf("editor: a container keeps at least one page")
	}
	if c.Locked {
		return fmt.Errorf("editor: %s is locked", c.Name)
	}
	c = c.Clone()
	e.apply(func() bool {
		var rest []design.Node
		for _, n := range e.doc.Nodes {
			if n.In == id && n.Page == page {
				continue
			}
			if n.In == id && n.Page > page {
				n.Page--
			}
			rest = append(rest, n)
		}
		e.doc.Nodes = rest
		props := map[string]string{
			container.PropPages: strings.Join(append(append([]string{}, pages[:page]...), pages[page+1:]...), ","),
		}
		switch a := container.Active(c); {
		case a > page, a == page && page == len(pages)-1:
			props[container.PropActive] = strconv.Itoa(max(a-1, 0))
		}
		var open []string
		for i := range pages {
			if i == page || !container.OpenSet(c)[i] {
				continue
			}
			open = append(open, strconv.Itoa(map[bool]int{true: i - 1, false: i}[i > page]))
		}
		props[container.PropOpen] = strings.Join(open, ",")
		e.doc.Update(id, func(n *design.Node) {
			for k, v := range props {
				n.Props[k] = v
			}
		})
		e.shiftMembers(c, nil)
		e.sel = e.keepSelected(func(id design.NodeID) bool { return e.doc.Index(id) >= 0 })
		return true
	})
	return nil
}

// removeMembersOf takes the members of the given containers out of the
// document.
func (e *Editor) removeMembersOf(removed map[design.NodeID]bool) {
	if len(removed) == 0 {
		return
	}
	kept := e.doc.Nodes[:0]
	for _, n := range e.doc.Nodes {
		if n.In != "" && removed[n.In] {
			continue
		}
		kept = append(kept, n)
	}
	e.doc.Nodes = kept
}

// duplicateMembers copies the members of container from to the new container
// to, at the same offsets.
func (e *Editor) duplicateMembers(from, to design.NodeID, dx, dy int) {
	for _, m := range e.Members(from) {
		c := m.Clone()
		c.ID = ""
		c.In = to
		c.Locked = false
		c.Rect = m.Rect.Translate(dx, dy)
		c.Name = e.uniqueName(baseName(m.Name))
		e.doc.Add(c)
	}
}
