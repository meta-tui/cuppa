package preview

import (
	tea "charm.land/bubbletea/v2"
	"github.com/meta-tui/cuppa/libs/canvas/editor"
	"github.com/meta-tui/cuppa/libs/catalog/definition"
	"github.com/meta-tui/cuppa/libs/document/container"
	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/render/grid"
	"github.com/meta-tui/cuppa/libs/render/scene"
)

// Catalog is what the preview needs to know about components.
type Catalog interface {
	Get(id string) (definition.Definition, bool)
}

// tagged is a message that a running component asked for (a timer tick, a
// cursor blink), labelled with whose it is so it finds its way back.
type tagged struct {
	id  design.NodeID
	msg tea.Msg
}

// Session is a design being run. It owns the running components; the design
// itself cannot change while it runs.
type Session struct {
	doc     design.Document
	cat     scene.Catalog
	live    map[design.NodeID]*live
	focus   design.NodeID
	stopped bool
	// ed applies a press on a tab or a section header the way the designer
	// does, so the page changes and the members move with their bodies.
	ed *editor.Editor
}

// Start runs doc. The returned command starts whatever needs time (spinners,
// timers, blinking cursors); the caller runs it.
func Start(doc design.Document, cat scene.Catalog) (*Session, tea.Cmd) {
	s := &Session{doc: doc.Clone(), cat: cat, live: map[design.NodeID]*live{}, ed: editor.New(cat, doc)}
	return s, tea.Batch(s.startAll()...)
}

// startAll builds the running component of every node of the page shown and
// returns what each needs to start.
func (s *Session) startAll() []tea.Cmd {
	s.live = map[design.NodeID]*live{}
	var starts []tea.Cmd
	for _, n := range s.doc.Nodes {
		build, ok := makers[n.Component]
		if !ok || n.Hidden {
			continue
		}
		l := build(propsOf(n, s.cat, s.doc), n.Rect.W, n.Rect.H)
		s.live[n.ID] = l
		if l.start != nil {
			starts = append(starts, tag(n.ID, l.start))
		}
	}
	return starts
}

// propsOf is the node's properties over the catalog defaults.
func propsOf(n design.Node, cat scene.Catalog, doc design.Document) props {
	p := props{}
	if def, ok := cat.Get(n.Component); ok {
		for k, v := range def.Effective(n.Props, doc.Theme, doc.Background) {
			p[k] = v
		}
		return p
	}
	for k, v := range n.Props {
		p[k] = v
	}
	return p
}

// tag makes the message cmd produces come back labelled with id.
func tag(id design.NodeID, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	return func() tea.Msg {
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			tagged := make(tea.BatchMsg, 0, len(batch))
			for _, c := range batch {
				tagged = append(tagged, tag(id, c))
			}
			return tagged
		}
		if msg == nil {
			return nil
		}
		return tagged{id: id, msg: msg}
	}
}

// Stop ends the run: late messages are ignored.
func (s *Session) Stop() { s.stopped = true }

// Count is how many components are running as real models.
func (s *Session) Count() int { return len(s.live) }

// Update takes a message that came back from a running component. It reports
// whether the message was one of theirs.
func (s *Session) Update(msg tea.Msg) (bool, tea.Cmd) {
	t, ok := msg.(tagged)
	if !ok {
		return false, nil
	}
	l := s.live[t.id]
	if s.stopped || l == nil {
		return true, nil
	}
	return true, tag(t.id, l.update(t.msg))
}

// Key gives a key press to the focused component. It reports whether there was one.
func (s *Session) Key(msg tea.KeyPressMsg) (bool, tea.Cmd) {
	if msg.String() == "tab" || msg.String() == "shift+tab" {
		return true, s.cycle(msg.String() == "shift+tab")
	}
	l := s.live[s.focus]
	if l == nil {
		return false, nil
	}
	return true, tag(s.focus, l.update(msg))
}

// cycle moves the keyboard to the next (or previous) component that reads it.
func (s *Session) cycle(back bool) tea.Cmd {
	var order []design.NodeID
	for _, n := range s.doc.Nodes {
		if l := s.live[n.ID]; l != nil && l.keys && container.Visible(s.doc, n) {
			order = append(order, n.ID)
		}
	}
	if len(order) == 0 {
		return nil
	}
	at := -1
	for i, id := range order {
		if id == s.focus {
			at = i
		}
	}
	step := 1
	if back {
		step = len(order) - 1
	}
	return s.focusOn(order[(at+step+len(order))%len(order)])
}

// Press is a click at canvas cell (x, y): the top-most running component there
// takes the keyboard and gets the click. Clicking anything else clears focus.
func (s *Session) Press(x, y int) tea.Cmd {
	if top, found := s.topAt(x, y); found && top.IsContainer() {
		// A press on a tab or a header shows that page.
		if page, hit := container.HeaderAt(top, x, y); hit {
			s.ed.PressContainerHeader(top.ID, page)
			s.doc = s.ed.Document()
			return s.focusOn("")
		}
	}
	if top, found := s.topAt(x, y); found && top.Goto != "" {
		// A component that goes to a page takes the person there.
		if i := s.doc.PageIndex(top.Goto); i >= 0 && s.ed.SwitchPage(i) {
			s.doc = s.ed.Document()
			s.focus = ""
			return tea.Batch(s.startAll()...)
		}
	}
	n, ok := s.at(x, y)
	if !ok {
		return s.focusOn("")
	}
	cmds := []tea.Cmd{s.focusOn(n.ID)}
	if l := s.live[n.ID]; l != nil {
		cmds = append(cmds, tag(n.ID, l.update(tea.MouseClickMsg{X: x - n.Rect.X, Y: y - n.Rect.Y, Button: tea.MouseLeft})))
	}
	return tea.Batch(cmds...)
}

// Wheel scrolls the running component under the pointer: dy < 0 is up.
func (s *Session) Wheel(x, y, dy int) tea.Cmd {
	n, ok := s.at(x, y)
	if !ok || s.live[n.ID] == nil {
		return nil
	}
	code := tea.KeyDown
	if dy < 0 {
		code = tea.KeyUp
	}
	return tag(n.ID, s.live[n.ID].update(tea.KeyPressMsg{Code: code}))
}

// topAt is the top-most visible node under the cell, running or not.
func (s *Session) topAt(x, y int) (design.Node, bool) {
	order := container.Order(s.doc)
	for k := len(order) - 1; k >= 0; k-- {
		n := s.doc.Nodes[order[k]]
		if !n.Hidden && container.Visible(s.doc, n) && n.Rect.Contains(x, y) {
			return n, true
		}
	}
	return design.Node{}, false
}

// at is the top-most visible running component under the cell.
func (s *Session) at(x, y int) (design.Node, bool) {
	order := container.Order(s.doc)
	for k := len(order) - 1; k >= 0; k-- {
		n := s.doc.Nodes[order[k]]
		if !n.Hidden && container.Visible(s.doc, n) && n.Rect.Contains(x, y) {
			return n, s.live[n.ID] != nil
		}
	}
	return design.Node{}, false
}

func (s *Session) focusOn(id design.NodeID) tea.Cmd {
	if id != "" && (s.live[id] == nil || !s.live[id].keys) {
		id = ""
	}
	if id == s.focus {
		return nil
	}
	var cmds []tea.Cmd
	if old := s.live[s.focus]; old != nil && old.focus != nil {
		cmds = append(cmds, tag(s.focus, old.focus(false)))
	}
	s.focus = id
	if now := s.live[id]; now != nil && now.focus != nil {
		cmds = append(cmds, tag(id, now.focus(true)))
	}
	return tea.Batch(cmds...)
}

// Focused is the component that has the keyboard, if any.
func (s *Session) Focused() (design.Node, bool) {
	n, ok := s.doc.Get(s.focus)
	return n, ok && s.focus != ""
}

// Render draws the design with the running components in place of their previews.
func (s *Session) Render() *grid.Grid {
	return scene.RenderWith(s.doc, s.cat, func(n design.Node) *grid.Grid {
		l := s.live[n.ID]
		if l == nil {
			return nil
		}
		return gridOf(l.view(), n.Rect.W, n.Rect.H)
	})
}
