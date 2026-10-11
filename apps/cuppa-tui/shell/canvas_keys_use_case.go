package shell

import (
	"fmt"
	"github.com/meta-ui-labs/cuppa/libs/document/container"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// paletteKey handles a key while the palette has the keyboard. It reports
// whether the key was one it uses. Enter on a component places it at the
// canvas cursor.
func (m *Model) paletteKey(text string) bool {
	switch text {
	case "up", "down", "left", "right", "home", "end", "enter", "/":
	default:
		return false
	}
	if id := m.pal.NavKey(text); id != "" {
		m.placeAtCursor(id)
	} else {
		m.say(m.pal.Current())
	}
	return true
}

// placeAtCursor puts a component on the canvas at the cursor and selects it.
// The cursor then moves below it, so the next one does not land on top.
func (m *Model) placeAtCursor(id string) {
	nid, err := m.ed.Add(id, m.curX, m.curY)
	if err != nil {
		return
	}
	n, ok := m.ed.Document().Get(nid)
	if !ok {
		return
	}
	m.layerCur = nid
	doc := m.ed.Document()
	m.say(fmt.Sprintf("Placed %s at %d, %d", n.Name, n.Rect.X, n.Rect.Y))
	m.curX, m.curY = n.Rect.X, n.Rect.Bottom()+1
	if m.curY >= doc.Height {
		m.curX, m.curY = min(n.Rect.Right()+1, doc.Width-1), 0
	}
}

// canvasKey handles a key while the canvas has the keyboard and it is not
// one of the shortcuts. It reports whether the key was one it uses.
func (m *Model) canvasKey(k tea.Key, text string) bool {
	selected := m.ed.Selected()
	switch {
	case text == "tab" || text == "shift+tab":
		dir := 1
		if text == "shift+tab" {
			dir = -1
		}
		m.stepLayer(dir)
		m.saySelection()
	case text == "space":
		m.extendSelection()
		m.saySelection()
	case text == "ctrl+a":
		var ids []design.NodeID
		doc := m.ed.Document()
		for _, n := range doc.Nodes {
			if !n.Hidden && container.Visible(doc, n) {
				ids = append(ids, n.ID)
			}
		}
		m.ed.Select(ids...)
		m.saySelection()
	case text == "enter" && len(selected) > 0:
		m.ins.Focus()
		m.setFocus(inInspector)
		m.sayFocus()
	case isArrow(k.Code) && k.Mod&tea.ModAlt != 0:
		if len(selected) != 1 {
			return false
		}
		m.resizeSelection(k)
	case isArrow(k.Code) && len(selected) == 0:
		m.moveCursor(k)
		m.sayCursor()
	default:
		return false
	}
	return true
}

// layers are the ids of the layers Tab walks through: back to front, hidden
// ones left out.
func (m *Model) layers() []design.NodeID {
	var ids []design.NodeID
	doc := m.ed.Document()
	for _, n := range doc.Nodes {
		if !n.Hidden && container.Visible(doc, n) {
			ids = append(ids, n.ID)
		}
	}
	return ids
}

// layerIndex is where the layer cursor is: on the selected layer when exactly
// one is selected, otherwise where Tab last stopped. It is -1 when nowhere.
func (m *Model) layerIndex(ids []design.NodeID) int {
	at := m.layerCur
	if sel := m.ed.Selected(); len(sel) == 1 {
		at = sel[0]
	}
	for i, id := range ids {
		if id == at {
			return i
		}
	}
	return -1
}

// stepLayer moves the layer cursor and, while the selection is one layer or
// nothing, the selection with it.
func (m *Model) stepLayer(dir int) {
	ids := m.layers()
	if len(ids) == 0 {
		return
	}
	at := m.layerIndex(ids)
	var next int
	switch {
	case at < 0 && dir > 0:
		next = 0
	case at < 0:
		next = len(ids) - 1
	default:
		next = ((at+dir)%len(ids) + len(ids)) % len(ids)
	}
	m.layerCur = ids[next]
	if len(m.ed.Selected()) <= 1 {
		m.ed.Select(m.layerCur)
	}
}

// extendSelection adds the layer after the layer cursor to the selection and
// moves the cursor onto it, so repeated presses select a run of layers.
func (m *Model) extendSelection() {
	ids := m.layers()
	if len(ids) == 0 {
		return
	}
	next := (m.layerIndex(ids) + 1) % len(ids)
	m.layerCur = ids[next]
	if !m.ed.IsSelected(m.layerCur) {
		m.ed.Toggle(m.layerCur)
	}
}

// moveCursor moves the canvas cursor with an arrow key: 1 cell, or 10 with
// Shift, kept on the canvas.
func (m *Model) moveCursor(k tea.Key) {
	step := stepSize(k)
	doc := m.ed.Document()
	switch k.Code {
	case tea.KeyLeft:
		m.curX -= step
	case tea.KeyRight:
		m.curX += step
	case tea.KeyUp:
		m.curY -= step
	case tea.KeyDown:
		m.curY += step
	}
	m.curX = min(max(m.curX, 0), max(doc.Width-1, 0))
	m.curY = min(max(m.curY, 0), max(doc.Height-1, 0))
}

// resizeSelection changes the width or height of the selected component with
// Alt and an arrow: right and down grow it, left and up shrink it.
func (m *Model) resizeSelection(k tea.Key) {
	n, ok := m.ed.Primary()
	if !ok {
		return
	}
	step := stepSize(k)
	r := n.Rect
	defer func() {
		if got, ok := m.ed.Primary(); ok && got.Rect != n.Rect {
			m.say(fmt.Sprintf("Resized to %d by %d", got.Rect.W, got.Rect.H))
		} else {
			m.sayRefusal("resize")
		}
	}()
	switch k.Code {
	case tea.KeyLeft:
		r.W -= step
	case tea.KeyRight:
		r.W += step
	case tea.KeyUp:
		r.H -= step
	case tea.KeyDown:
		r.H += step
	}
	m.ed.SetRect(n.ID, r, true)
}
