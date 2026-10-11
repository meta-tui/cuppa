package shell

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/modal"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/variablesdialog"
)

func TestVariablesDialogListsNamesAndGoesToTheComponent(t *testing.T) {
	m := newShell(t)
	id, err := m.ed.Add("lipgloss.label", 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.ed.SetColour(id, "color", "#ff0000"); err != nil {
		t.Fatal(err)
	}
	m.ed.Select()
	m.perform(menubar.ViewVariables)
	dlg := m.flow.Modal()
	if dlg == nil {
		t.Fatal("the Variables dialog should open")
	}
	got := ansi.Strip(strings.Join(dlg.Lines(), "\n"))
	if !strings.Contains(got, "Foreground") || !strings.Contains(got, "#ff0000") {
		t.Fatalf("dialog:\n%s", got)
	}
	m.afterVariables(modal.Outcome{Button: variablesdialog.GoToButton, Value: string(id)})
	if sel := m.ed.Selected(); len(sel) != 1 || sel[0] != id {
		t.Fatalf("selection = %v", sel)
	}

	m.afterVariables(modal.Outcome{Button: variablesdialog.RenameButton, Value: "colour:Foreground"})
	if m.flow.Modal() == nil {
		t.Fatal("renaming asks for the new name")
	}
}
