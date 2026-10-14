// Command basic shows a menu bar with File, Edit and Help menus. Click a
// heading, press F10 or Alt and a heading's first letter, then use the arrows
// and Enter; a shortcut works at any time. q quits.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	menubar "github.com/meta-ui-labs/cuppa/packages/components/menubar"
)

type model struct {
	bar  menubar.Model
	last string
}

func newModel() model {
	return model{bar: menubar.New(
		menubar.WithTitle("Demo"),
		menubar.WithMenus(
			menubar.Menu{Name: "File", Entries: []menubar.Entry{
				{Name: "New", Shortcut: "Ctrl+N"},
				{Name: "Open…", Shortcut: "Ctrl+O"},
				{Separator: true},
				{Name: "Quit", Shortcut: "Ctrl+Q"},
			}},
			menubar.Menu{Name: "Edit", Entries: []menubar.Entry{
				{Name: "Undo", Shortcut: "Ctrl+Z"},
				{Name: "Redo", Shortcut: "Ctrl+Y"},
				{Separator: true},
				{Name: "Copy", Shortcut: "Ctrl+C"},
				{Name: "Paste", Shortcut: "Ctrl+V"},
			}},
			menubar.Menu{Name: "Help", Entries: []menubar.Entry{{Name: "About"}}},
		),
		menubar.WithSize(50, 9),
	)}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case menubar.SelectedMsg:
		m.last = fmt.Sprintf("chose %s > %s", msg.MenuName, msg.Name)
		if msg.Name == "Quit" {
			return m, tea.Quit
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.bar, cmd = m.bar.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	v := tea.NewView(m.bar.View() + "\n\n " + m.last)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeAllMotion
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
