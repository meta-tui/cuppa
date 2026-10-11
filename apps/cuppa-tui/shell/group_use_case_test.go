package shell

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
)

func TestCtrlGGroupsAndCtrlUUngroups(t *testing.T) {
	m := newShell(t)
	a, _ := m.ed.Add("lipgloss.box", 5, 5)
	b, _ := m.ed.Add("lipgloss.box", 40, 5)
	m.ed.Select(a, b)
	send(m, tea.KeyPressMsg{Code: 'g', Mod: tea.ModCtrl})
	if n := len(m.ed.Document().Nodes); n != 1 || !m.ed.Document().Nodes[0].IsGroup() {
		t.Fatalf("Ctrl+G should group: %d nodes", n)
	}
	send(m, tea.KeyPressMsg{Code: 'u', Mod: tea.ModCtrl})
	if n := len(m.ed.Document().Nodes); n != 2 {
		t.Fatalf("Ctrl+U should ungroup: %d nodes", n)
	}
}

func TestGroupMenuItemsFollowTheSelection(t *testing.T) {
	m := newShell(t)
	a, _ := m.ed.Add("lipgloss.box", 5, 5)
	b, _ := m.ed.Add("lipgloss.box", 40, 5)
	m.ed.Select(a, b)
	m.perform(menubar.EditGroup)
	if len(m.ed.Document().Nodes) != 1 {
		t.Fatal("the menu groups")
	}
	m.perform(menubar.EditUngroup)
	if len(m.ed.Document().Nodes) != 2 {
		t.Fatal("the menu ungroups")
	}
}
