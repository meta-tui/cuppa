package shell

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
)

func TestCtrlCAndCtrlVCopyAndPaste(t *testing.T) {
	m := newShell(t)
	id, _ := m.ed.Add("lipgloss.box", 5, 5)
	m.ed.Select(id)
	send(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if m.flow.Modal() != nil || m.flow.Quitting() {
		t.Fatal("Ctrl+C copies; it no longer quits")
	}
	send(m, tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	send(m, tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	if n := len(m.ed.Document().Nodes); n != 3 {
		t.Fatalf("two pastes after a copy: %d nodes", n)
	}
}

func TestCtrlShiftZRedoesLikeCtrlY(t *testing.T) {
	m := newShell(t)
	_, _ = m.ed.Add("lipgloss.box", 5, 5)
	m.ed.Undo()
	send(m, tea.KeyPressMsg{Code: 'z', Mod: tea.ModCtrl | tea.ModShift})
	if len(m.ed.Document().Nodes) != 1 {
		t.Fatal("Ctrl+Shift+Z redoes")
	}
	m.ed.Undo()
	send(m, tea.KeyPressMsg{Code: 'y', Mod: tea.ModCtrl})
	if len(m.ed.Document().Nodes) != 1 {
		t.Fatal("Ctrl+Y still redoes")
	}
}

func TestCopyAndPasteDoNotActWhileTypingInAField(t *testing.T) {
	m := newShell(t)
	id, _ := m.ed.Add("lipgloss.box", 5, 5)
	m.ed.Select(id)
	m.render()
	clickText(t, m, "Box 1") // the first one is the Name field
	if !m.ins.Editing() {
		t.Fatal("clicking the name should start editing it")
	}
	send(m, tea.KeyPressMsg{Code: 'v', Mod: tea.ModCtrl})
	if len(m.ed.Document().Nodes) != 1 {
		t.Fatal("a paste while editing text must not add components")
	}
}

func TestEditMenuCopyAndPasteFollowTheSelection(t *testing.T) {
	m := newShell(t)
	id, _ := m.ed.Add("lipgloss.box", 5, 5)
	m.ed.Select(id)
	m.perform(menubar.EditCopy)
	m.perform(menubar.EditPaste)
	if len(m.ed.Document().Nodes) != 2 {
		t.Fatal("the menu copies and pastes")
	}
}
