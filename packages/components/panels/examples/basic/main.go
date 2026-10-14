// Command basic shows a panel stack and a tab bar side by side. Press a header
// or use the arrows; q quits.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	panels "github.com/meta-ui-labs/cuppa/packages/components/panels"
)

type model struct {
	stack, tabs panels.Model
	last        string
}

func lines(prefix string, n int) func(w, h int) string {
	return func(w, h int) string {
		var out []string
		for i := 1; i <= min(n, h); i++ {
			out = append(out, fmt.Sprintf(" %s %d", prefix, i))
		}
		return strings.Join(out, "\n")
	}
}

func newModel() model {
	return model{
		stack: panels.New(
			panels.WithPages(
				panels.Page{Title: "Details", Body: lines("detail", 3)},
				panels.Page{Title: "Properties", Body: lines("property", 6)},
				panels.Page{Title: "Layers", Body: lines("layer", 8)},
			),
			panels.WithOpen(0, 2),
			panels.WithSize(30, 16),
			panels.WithOrigin(1, 1),
		),
		tabs: panels.New(
			panels.WithMode(panels.Tabs),
			panels.WithPages(
				panels.Page{Title: "Overview", Body: lines("overview", 3)},
				panels.Page{Title: "Colours", Body: lines("colour", 5)},
				panels.Page{Title: "Keys", Body: lines("key", 4)},
			),
			panels.WithSize(36, 10),
			panels.WithOrigin(33, 1),
		),
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case panels.ChangedMsg:
		m.last = fmt.Sprintf("changed: active %d, open %v", msg.Active, msg.Open)
		return m, nil
	}
	var c1, c2 tea.Cmd
	m.stack, c1 = m.stack.Update(msg)
	m.tabs, c2 = m.tabs.Update(msg)
	return m, tea.Batch(c1, c2)
}

func (m model) View() tea.View {
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.stack.View(), " ", m.tabs.View())
	v := tea.NewView("\n " + strings.ReplaceAll(body, "\n", "\n ") + "\n\n " + m.last)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
