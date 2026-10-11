package inspector

import (
	"strings"

	"github.com/meta-tui/cuppa/libs/document/design"
)

// focusStop is the stop to draw as focused: none while the bar does not have
// the keyboard.
func (m *Model) focusStop() int {
	if !m.focused {
		return -1
	}
	return m.stop
}

// stopRegions are the regions the keyboard stops on, in reading order.
func (m *Model) stopRegions() []region {
	var out []region
	for _, r := range m.regions {
		if !r.mouseOnly {
			out = append(out, r)
		}
	}
	return out
}

// Current is the stop under the keyboard, as it is read out.
func (m *Model) Current() string {
	stops := m.stopRegions()
	if m.stop < 0 || m.stop >= len(stops) {
		return ""
	}
	return stops[m.stop].label
}

// Focus puts the keyboard on the bar, on its first stop.
func (m *Model) Focus() { m.stop = 0 }

// NavKey takes a key while the bar has the keyboard and reports whether it
// used it. Tab, Down and Right go to the next stop, Shift+Tab, Up and Left to
// the previous; on a number or a choice Left and Right (and - and +) change
// the value by one, with Shift by ten; Enter and Space press the stop. On a
// layer, h hides or shows it, l locks or unlocks it, Alt+Up and Alt+Down move
// it and Delete removes it.
func (m *Model) NavKey(name string) bool {
	stops := m.stopRegions()
	if len(stops) == 0 {
		return false
	}
	m.stop = min(max(m.stop, 0), len(stops)-1)
	cur := stops[m.stop]
	shift := strings.HasPrefix(name, "shift+")
	base := strings.TrimPrefix(name, "shift+")
	amount := 1
	if shift {
		amount = 10
	}
	move := func(d int) { m.stop = ((m.stop+d)%len(stops) + len(stops)) % len(stops) }
	switch {
	case name == "tab" || name == "down":
		move(1)
	case name == "shift+tab" || name == "up":
		move(-1)
	case base == "left" || base == "right":
		d := 1
		if base == "left" {
			d = -1
		}
		if cur.step == nil {
			move(d)
			break
		}
		cur.step(d * amount)
	case name == "-" || name == "+" || name == "=":
		if cur.step == nil {
			return false
		}
		d := 1
		if name == "-" {
			d = -1
		}
		cur.step(d)
	case name == "enter" || name == "space":
		cur.act()
	case cur.layer != "" && m.layerKey(name, cur.layer):
	default:
		return false
	}
	if now := m.stopRegions(); m.stop < len(now) {
		m.reveal(now[m.stop].y)
	}
	return true
}

// layerKey handles the keys of the layers list for the layer under the stop.
func (m *Model) layerKey(name string, id design.NodeID) bool {
	doc := m.ed.Document()
	n, ok := doc.Get(id)
	if !ok {
		return false
	}
	switch name {
	case "h":
		m.ed.SetHidden(id, !n.Hidden)
	case "l":
		m.ed.SetLocked(id, !n.Locked)
	case "alt+up":
		m.ed.MoveLayer(id, doc.Index(id)+1)
	case "alt+down":
		m.ed.MoveLayer(id, doc.Index(id)-1)
	case "delete":
		if m.ed.IsSelected(id) {
			m.ed.Delete()
		} else {
			m.ed.DeleteLayer(id)
		}
	default:
		return false
	}
	return true
}
