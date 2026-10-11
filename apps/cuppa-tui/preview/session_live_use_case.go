package preview

import (
	"encoding/json"

	tea "charm.land/bubbletea/v2"
)

// SetLive gives the running design the values an application published (ADR
// 0016): what is drawn and what runs follows them, and the keyboard stays where
// it was when the component is still there.
func (s *Session) SetLive(page string, inputs map[string]json.RawMessage) tea.Cmd {
	s.ed.SetLive(page, inputs)
	focus := s.focus
	s.doc = s.ed.RenderDocument()
	cmds := s.startAll()
	if _, ok := s.live[focus]; ok {
		s.focus = focus
	}
	return tea.Batch(cmds...)
}

// ClearLive goes back to the values of the design.
func (s *Session) ClearLive() tea.Cmd {
	s.ed.ClearLive()
	s.doc = s.ed.RenderDocument()
	return tea.Batch(s.startAll()...)
}

// PageName is the page being run, as an application names it.
func (s *Session) PageName() string {
	names := s.ed.PageNames()
	return names[min(max(s.ed.ActivePage(), 0), len(names)-1)]
}

// EventAt is the event a click at canvas cell (x, y) raises: the name the
// design gave the top-most component there, and the cell inside it.
func (s *Session) EventAt(x, y int) (name string, lx, ly int, ok bool) {
	n, found := s.topAt(x, y)
	if !found || n.Event == "" {
		return "", 0, 0, false
	}
	return n.Event, x - n.Rect.X, y - n.Rect.Y, true
}

// KeyEvent is the event a key raises: the screen key of the page bound to it.
func (s *Session) KeyEvent(key string) (string, bool) {
	for _, k := range s.doc.Keys {
		if k.Key == key && k.Event != "" {
			return k.Event, true
		}
	}
	return "", false
}
