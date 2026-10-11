package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func key(code rune, mod tea.KeyMod) tea.Msg { return tea.KeyPressMsg{Code: code, Mod: mod} }

func TestF6WalksTheAreasAndShiftF6WalksBack(t *testing.T) {
	m := newShell(t)
	if m.focus != inStage {
		t.Fatalf("starts on %v, want the canvas", m.focus)
	}
	var seen []pane
	for range areas {
		send(m, key(tea.KeyF6, 0))
		seen = append(seen, m.focus)
	}
	want := []pane{inInspector, inMenu, inPalette, inStage}
	for i := range want {
		if seen[i] != want[i] {
			t.Fatalf("F6 order = %v, want %v", seen, want)
		}
	}
	send(m, key(tea.KeyF6, tea.ModShift))
	if m.focus != inPalette {
		t.Fatalf("Shift+F6 went to %v, want the palette", m.focus)
	}
}

func TestAltNumberJumpsToAnArea(t *testing.T) {
	m := newShell(t)
	send(m, key('2', tea.ModAlt))
	if m.focus != inPalette {
		t.Fatalf("Alt+2 = %v, want the palette", m.focus)
	}
	send(m, key('4', tea.ModAlt))
	if m.focus != inInspector {
		t.Fatalf("Alt+4 = %v, want the details bar", m.focus)
	}
}

func TestEscLeavesAnAreaForTheCanvasAndKeepsTheSelection(t *testing.T) {
	m := newShell(t)
	place(t, m, 5)
	send(m, key(tea.KeyF6, 0)) // details bar
	send(m, key(tea.KeyEscape, 0))
	if m.focus != inStage {
		t.Fatalf("Esc left focus on %v", m.focus)
	}
	if len(m.Editor().Selected()) != 1 {
		t.Fatal("Esc from an area must not clear the selection")
	}
	send(m, key(tea.KeyEscape, 0))
	if len(m.Editor().Selected()) != 0 {
		t.Fatal("Esc on the canvas clears the selection, as before")
	}
}

func TestClickingAnAreaFocusesIt(t *testing.T) {
	m := newShell(t)
	send(m, click(5, firstComponentY))
	send(m, release(5, 4))
	if m.focus != inPalette {
		t.Fatalf("after clicking the palette focus = %v", m.focus)
	}
}

func TestF10OpensTheFirstMenuAndArrowsMoveThroughIt(t *testing.T) {
	m := newShell(t)
	send(m, key(tea.KeyF10, 0))
	if !m.bar.Open() || m.focus != inMenu {
		t.Fatal("F10 should open the first menu")
	}
	send(m, key(tea.KeyRight, 0))
	if _, items := m.bar.Dropdown(); len(items) == 0 || !strings.Contains(strings.Join(items, "\n"), "Undo") {
		t.Fatal("Right should open the Edit menu")
	}
	send(m, key(tea.KeyEscape, 0))
	if m.bar.Open() {
		t.Fatal("Esc closes the dropdown")
	}
	send(m, key(tea.KeyEscape, 0))
	if m.focus != inStage {
		t.Fatalf("a second Esc gives the keyboard back, focus = %v", m.focus)
	}
}

func TestAltLetterOpensItsMenuAndEnterRunsTheItem(t *testing.T) {
	m := newShell(t)
	place(t, m, 5)
	send(m, key('e', tea.ModAlt)) // Edit: Undo is first and enabled after placing
	if !m.bar.Open() {
		t.Fatal("Alt+E should open the Edit menu")
	}
	send(m, key(tea.KeyEnter, 0)) // Undo
	if n := len(m.Editor().Document().Nodes); n != 0 {
		t.Fatalf("Enter on Undo left %d nodes", n)
	}
	if m.bar.Open() || m.focus != inStage {
		t.Fatal("running an item closes the menu and returns to the canvas")
	}
}

func TestMenuKeysSkipDisabledItems(t *testing.T) {
	m := newShell(t)
	send(m, key('e', tea.ModAlt)) // nothing to undo or redo: first usable is Paste? Copy is disabled too
	send(m, key(tea.KeyEnter, 0))
	// Nothing in the design: Undo, Redo and Copy are disabled, so Enter must not
	// have run a disabled item; the document is untouched.
	if m.Editor().Dirty() {
		t.Fatal("a disabled item ran")
	}
}

func TestFocusedAreaIsFlaggedForTheScreenReader(t *testing.T) {
	m := newShell(t)
	send(m, key('2', tea.ModAlt))
	flagged := 0
	for _, n := range m.Describe().Nodes {
		if n.Focused {
			flagged++
		}
	}
	if flagged != 1 {
		t.Fatalf("%d focused nodes, want 1", flagged)
	}
}

func TestStatusBarListsTheKeysOfTheFocusedArea(t *testing.T) {
	m := newShell(t)
	var _ = design.Rect{}
	if !strings.Contains(m.statusBar(), "Tab layer") {
		t.Fatalf("canvas hints: %q", m.statusBar())
	}
	send(m, key('2', tea.ModAlt))
	if !strings.Contains(m.statusBar(), "Palette") {
		t.Fatalf("palette hints: %q", m.statusBar())
	}
}
