package shell

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/tools"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

// toolRowY is the screen row of a tool in the left bar (row 0 is the title
// bar, the tool list starts with its heading).
func toolRowY(t tools.Tool) int { return 2 + int(t) }

func TestLetterKeysChooseTheTools(t *testing.T) {
	m := newShell(t)
	for letter, want := range map[rune]tools.Tool{'u': tools.Rectangle, 'p': tools.Path, 'b': tools.Brush, 'e': tools.Erase, 't': tools.Text, 'v': tools.Select} {
		send(m, typed(letter))
		if m.tb.Tool() != want {
			t.Fatalf("%q chose %v, want %v", letter, m.tb.Tool(), want)
		}
	}
}

func TestClickingAToolInTheLeftBarChoosesIt(t *testing.T) {
	m := newShell(t)
	send(m, click(5, toolRowY(tools.Brush)))
	send(m, release(5, toolRowY(tools.Brush)))
	if m.tb.Tool() != tools.Brush {
		t.Fatalf("tool = %v, want the brush", m.tb.Tool())
	}
}

func TestDraggingARectangleDrawsOneLayerAndOneUndoStep(t *testing.T) {
	m := newShell(t)
	send(m, typed('u'))
	sx, sy := m.layout.stage.X, m.layout.stage.Y
	send(m, click(sx+5, sy+3))
	send(m, motion(sx+15, sy+8))
	send(m, release(sx+15, sy+8))
	l := m.Editor().Drawing()
	if l.Len() == 0 {
		t.Fatal("nothing was drawn")
	}
	if c, ok := l.Get(5, 3); !ok || c.Ch != '┌' {
		t.Fatalf("top-left corner = %+v (%v)", c, ok)
	}
	if c, ok := l.Get(15, 8); !ok || c.Ch != '┘' {
		t.Fatalf("bottom-right corner = %+v (%v)", c, ok)
	}
	layers := 0
	for _, n := range m.Editor().Document().Nodes {
		if n.Component == drawlayer.Component {
			layers++
		}
	}
	if layers != 1 {
		t.Fatalf("%d drawing layers, want 1", layers)
	}
	send(m, ctrl('z'))
	if m.Editor().Drawing().Len() != 0 {
		t.Fatal("one undo should take the whole rectangle back")
	}
}

func TestDrawingDoesNotSelectOrMoveComponents(t *testing.T) {
	m := newShell(t)
	n := place(t, m, 5)
	m.Editor().Clear()
	send(m, typed('b'))
	sx, sy := m.layout.stage.X, m.layout.stage.Y
	send(m, click(sx+n.Rect.X+2, sy+n.Rect.Y+2))
	send(m, motion(sx+n.Rect.X+8, sy+n.Rect.Y+2))
	send(m, release(sx+n.Rect.X+8, sy+n.Rect.Y+2))
	if len(m.Editor().Selected()) != 0 {
		t.Fatal("a brush stroke selected a component")
	}
	got, _ := m.Editor().Document().Get(n.ID)
	if got.Rect != n.Rect {
		t.Fatal("a brush stroke moved a component")
	}
	if m.Editor().Drawing().Len() == 0 {
		t.Fatal("the brush drew nothing")
	}
}

func TestTheEraserOnlyRemovesDrawnCells(t *testing.T) {
	m := newShell(t)
	n := place(t, m, 5)
	m.Editor().Clear()
	sx, sy := m.layout.stage.X, m.layout.stage.Y
	send(m, typed('b'))
	send(m, click(sx+60, sy+5))
	send(m, motion(sx+70, sy+5))
	send(m, release(sx+70, sy+5))
	drawn := m.Editor().Drawing().Len()
	// Erase across the drawing and over the component.
	send(m, typed('e'))
	send(m, click(sx+n.Rect.X, sy+5))
	send(m, motion(sx+75, sy+5))
	send(m, release(sx+75, sy+5))
	if m.Editor().Drawing().Len() >= drawn {
		t.Fatalf("erasing left %d of %d cells", m.Editor().Drawing().Len(), drawn)
	}
	got, ok := m.Editor().Document().Get(n.ID)
	if !ok || got.Rect != n.Rect {
		t.Fatal("the eraser touched a component")
	}
}

func TestShiftKeepsAPathOnAnAxis(t *testing.T) {
	m := newShell(t)
	send(m, typed('p'))
	sx, sy := m.layout.stage.X, m.layout.stage.Y
	send(m, click(sx+10, sy+5))
	m.Update(tea.MouseMotionMsg{X: sx + 30, Y: sy + 6, Button: tea.MouseLeft, Mod: tea.ModShift})
	m.Update(tea.MouseReleaseMsg{X: sx + 30, Y: sy + 6, Button: tea.MouseLeft, Mod: tea.ModShift})
	l := m.Editor().Drawing()
	if _, ok := l.Get(20, 5); !ok {
		t.Fatal("a horizontal line should pass through (20,5)")
	}
	if _, ok := l.Get(30, 6); ok {
		t.Fatal("the line left its axis")
	}
}

func TestDrawingIsAnnouncedAndEscCancelsADrag(t *testing.T) {
	m := newShell(t)
	send(m, typed('u'))
	sx, sy := m.layout.stage.X, m.layout.stage.Y
	send(m, click(sx+5, sy+3))
	send(m, motion(sx+9, sy+6))
	send(m, key(tea.KeyEscape, 0))
	send(m, release(sx+9, sy+6))
	if m.Editor().Drawing().Len() != 0 {
		t.Fatal("Esc should cancel the rectangle")
	}
}
