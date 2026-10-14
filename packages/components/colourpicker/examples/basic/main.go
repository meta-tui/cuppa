// Command basic shows the colour picker in a program: pick a colour with the
// keyboard or the mouse; the line below the picker is drawn in it. q quits.
package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	colourpicker "github.com/meta-ui-labs/cuppa/packages/components/colourpicker"
)

type model struct {
	picker colourpicker.Model
}

func (m model) Init() tea.Cmd { return m.picker.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok && (k.String() == "q" || k.String() == "ctrl+c") {
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.picker, cmd = m.picker.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	title := lipgloss.NewStyle().Bold(true).Render("Pick a colour")
	value := m.picker.Value()
	sample := "(none)"
	if value != "" {
		sample = lipgloss.NewStyle().Foreground(lipgloss.Color(value)).Bold(true).Render("The quick brown fox " + value)
	}
	v := tea.NewView(title + "\n\n" + m.picker.View() + "\n\n" + sample + "\n\ntab: tab  arrows: move  q: quit")
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	// The picker starts on line 3 (a title and a blank line above it).
	m := model{picker: colourpicker.New(colourpicker.WithValue("#ff007f"), colourpicker.WithOrigin(0, 2))}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
