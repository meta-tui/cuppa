package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestEnterOnTheCanvasFocusesTheDetailsAndRightStepsANumber(t *testing.T) {
	m := newShell(t)
	n := place(t, m, 5)
	send(m, key(tea.KeyEnter, 0))
	if m.focus != inInspector {
		t.Fatalf("focus = %v", m.focus)
	}
	m.View() // lay the bar out so its stops are known
	// Stops: Undo, Redo, (name is plain text), X, Y, W, H ... walk to X.
	for i := 0; i < 8; i++ {
		m.render()
		got, _ := m.Editor().Primary()
		before := got.Rect.X
		send(m, key(tea.KeyRight, 0))
		m.render()
		after, _ := m.Editor().Primary()
		if after.Rect.X == before+1 {
			send(m, key(tea.KeyLeft, tea.ModShift))
			back, _ := m.Editor().Primary()
			if back.Rect.X != n.Rect.X+1-10 && back.Rect.X != 0 {
				t.Fatalf("Shift+Left moved x to %d", back.Rect.X)
			}
			return
		}
	}
	t.Fatal("no stop changed the x position with Right")
}

func TestTabWalksTheContextBarAndEnterPressesAButton(t *testing.T) {
	m := newShell(t)
	place(t, m, 5)
	send(m, key('4', tea.ModAlt))
	m.render()
	// The first stop is the name in the bar above the canvas: Enter types it.
	send(m, key(tea.KeyEnter, 0))
	if !m.ins.Editing() {
		t.Fatal("Enter on the name should start editing it")
	}
	send(m, key(tea.KeyEscape, 0))
	// The [Delete] button of the bar removes the placed component.
	for i := 0; i < 40 && len(m.Editor().Document().Nodes) > 0; i++ {
		send(m, key(tea.KeyTab, 0))
		m.render()
		if strings.Contains(m.ins.Current(), "Delete") {
			send(m, key(tea.KeyEnter, 0))
		}
	}
	if n := len(m.Editor().Document().Nodes); n != 0 {
		t.Fatalf("Tab and Enter on [Delete] left %d components", n)
	}
}

func TestLayerKeysHideLockAndDelete(t *testing.T) {
	m := newShell(t)
	n := place(t, m, 5)
	send(m, key('4', tea.ModAlt))
	m.render()
	send(m, key(tea.KeyTab, tea.ModShift)) // wraps to the last stop: the layer
	m.render()
	send(m, typed('h'))
	got, _ := m.Editor().Document().Get(n.ID)
	if !got.Hidden {
		t.Fatal("h should hide the layer")
	}
	send(m, typed('l'))
	got, _ = m.Editor().Document().Get(n.ID)
	if !got.Locked {
		t.Fatal("l should lock the layer")
	}
	send(m, typed('l')) // a locked layer refuses deletion, as it does for the mouse
	send(m, key(tea.KeyDelete, 0))
	if _, ok := m.Editor().Document().Get(n.ID); ok {
		t.Fatal("Delete should remove the layer")
	}
}
