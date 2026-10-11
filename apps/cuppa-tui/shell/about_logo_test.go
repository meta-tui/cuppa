package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
)

func TestAboutShowsTheLogoAboveTheText(t *testing.T) {
	m := newShell(t)
	m.perform(menubar.HelpAbout)
	dlg := m.flow.Modal()
	if dlg == nil {
		t.Fatal("About should open a dialog")
	}
	lines := dlg.Lines()
	if len(lines) != dlg.Rect().H {
		t.Fatalf("%d lines in a rect %d tall", len(lines), dlg.Rect().H)
	}
	plain := ansi.Strip(strings.Join(lines, "\n"))
	if !strings.ContainsAny(plain, "▘▝▀▖▌▞▛▗▚▐▜▄▙▟█") {
		t.Fatalf("the logo's blocks are missing:\n%s", plain)
	}
	if !strings.Contains(plain, "A designer for Bubble Tea interfaces") {
		t.Fatal("the text should still be there")
	}
	for i, l := range lines {
		if w := ansi.StringWidth(l); w != dlg.Rect().W {
			t.Errorf("line %d is %d wide, the box is %d", i, w, dlg.Rect().W)
		}
	}
}

func TestColourSchemeButtonFillsTheThemeInOneUndoStep(t *testing.T) {
	m := newShell(t)
	m.Update(tea.WindowSizeMsg{Width: 140, Height: 60})
	var x, y = -1, -1
	for row, line := range strings.Split(m.render(), "\n") {
		plain := ansi.Strip(line)
		if i := strings.Index(plain, "[Colour scheme…]"); i >= 0 {
			x, y = ansi.StringWidth(plain[:i])+1, row
		}
	}
	if x < 0 {
		t.Fatal("the Theme section should have a colour scheme button")
	}
	send(m, click(x, y))
	if m.flow.Modal() == nil {
		t.Fatal("the button should open the scheme dialog")
	}
	send(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	doc := m.Editor().Document()
	if doc.Background == "" || doc.Theme.Text == "" || doc.Theme.Muted == "" || doc.Theme.Border == "" || doc.Theme.Secondary == "" {
		t.Fatalf("the five colours should be set: %q %+v", doc.Background, doc.Theme)
	}
	m.Editor().Undo()
	if doc = m.Editor().Document(); doc.Background != "" || doc.Theme.Text != "" {
		t.Fatal("one undo takes the whole preset back")
	}
}
