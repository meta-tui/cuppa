package shell

import (
	"os"
	"path/filepath"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/meta-ui-labs/cuppa/libs/cuppafile/disk"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func ctrlR() tea.Msg { return tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl} }

// savedDesign opens a design on disk that says how to run an application.
func savedDesign(t *testing.T, m *Model, run design.RunSettings) {
	t.Helper()
	doc := design.NewDocument("App", 40, 10)
	doc.Run = run
	path := filepath.Join(t.TempDir(), "app.cuppa")
	if _, err := disk.Save(path, doc); err != nil {
		t.Fatal(err)
	}
	if err := m.flow.OpenPath(path); err != nil {
		t.Fatal(err)
	}
}

func TestRunningADesignThatSaysNothingToRunTellsWhatToDo(t *testing.T) {
	m := newShell(t)
	savedDesign(t, m, design.RunSettings{})
	send(m, ctrlR())
	if m.Hosting() {
		t.Fatal("a design with nothing to run is hosting")
	}
	if m.flow.Modal() == nil {
		t.Fatal("the person was not told what is missing")
	}
}

func TestCtrlRShowsTheApplicationAndGoesBackKeepingItRunning(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	m := newShell(t)
	t.Cleanup(m.stopHost)
	savedDesign(t, m, design.RunSettings{Run: []string{exe, "-test.run=NONE"}})
	send(m, ctrlR())
	if !m.Hosting() || m.host == nil {
		t.Fatal("Ctrl+R did not show the application")
	}
	send(m, ctrlR())
	if m.Hosting() {
		t.Fatal("Ctrl+R again did not go back to editing")
	}
	if m.host == nil {
		t.Fatal("going back ended the application")
	}
	send(m, ctrlR())
	if !m.Hosting() {
		t.Fatal("the application was not shown again")
	}
}

func TestLeavingForAnotherDesignEndsTheApplication(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	m := newShell(t)
	t.Cleanup(m.stopHost)
	savedDesign(t, m, design.RunSettings{Run: []string{exe, "-test.run=NONE"}})
	send(m, ctrlR())
	send(m, tea.KeyPressMsg{Code: 'n', Mod: tea.ModCtrl})
	if m.Hosting() || m.host != nil {
		t.Fatal("a new design kept the old application")
	}
}
