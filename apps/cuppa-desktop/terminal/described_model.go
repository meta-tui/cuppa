package terminal

import (
	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	ttygo "github.com/meta-tui/treactui/packages/tty-go"
)

// Describer is what the Cuppa shell offers: a description of its screen.
type Describer interface {
	Describe() a11y.Snapshot
}

// Describe wraps a model so the page gets a screen description after every
// update (tty-go sends it when it changes) and hears the app's feedback, such
// as "Saved design.cuppa", read out as an announcement.
func Describe(inner tea.Model, source Describer) tea.Model {
	return &described{inner: inner, source: source, spoken: source.Describe().Status}
}

type described struct {
	inner  tea.Model
	source Describer
	// spoken is the last status that was announced.
	spoken string
}

func (m *described) Init() tea.Cmd { return m.inner.Init() }

func (m *described) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.inner.Update(msg)
	m.inner = next
	if status := m.source.Describe().Status; status != m.spoken {
		m.spoken = status
		if status != "" {
			announce := func() tea.Msg { return ttygo.AnnounceMsg{Text: status} }
			cmd = tea.Batch(cmd, announce)
		}
	}
	return m, cmd
}

func (m *described) View() tea.View { return m.inner.View() }

// Accessible implements ttygo.Accessible.
func (m *described) Accessible() ttygo.Snapshot { return snapshotOf(m.source.Describe()) }

func snapshotOf(s a11y.Snapshot) ttygo.Snapshot {
	return ttygo.Snapshot{Title: s.Title, Nodes: nodesOf(s.Nodes)}
}

func nodesOf(in []a11y.Node) []ttygo.A11yNode {
	out := make([]ttygo.A11yNode, len(in))
	for i, n := range in {
		out[i] = ttygo.A11yNode{
			Role: n.Role, Label: n.Label, Value: n.Value, Selected: n.Selected, Focused: n.Focused,
			Children: nodesOf(n.Children),
		}
	}
	return out
}
