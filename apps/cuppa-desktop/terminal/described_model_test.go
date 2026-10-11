package terminal

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	ttygo "github.com/meta-tui/treactui/packages/tty-go"
)

// fake is a model whose status changes when it gets a string message.
type fake struct{ status string }

func (f *fake) Init() tea.Cmd { return nil }
func (f *fake) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if s, ok := msg.(string); ok {
		f.status = s
	}
	return f, nil
}
func (f *fake) View() tea.View { return tea.NewView("") }
func (f *fake) Describe() a11y.Snapshot {
	return a11y.Snapshot{Title: "T", Status: f.status, Nodes: []a11y.Node{
		a11y.List("L", a11y.Item("one", true)),
	}}
}

func TestTheDescriptionIsMappedToTheSnapshotShape(t *testing.T) {
	f := &fake{}
	m := Describe(f, f).(ttygo.Accessible)
	snap := m.Accessible()
	if snap.Title != "T" || len(snap.Nodes) != 1 || snap.Nodes[0].Role != "list" ||
		len(snap.Nodes[0].Children) != 1 || !snap.Nodes[0].Children[0].Selected || snap.Nodes[0].Children[0].Label != "one" {
		t.Fatalf("snapshot = %+v", snap)
	}
}

func TestANewStatusIsAnnouncedOnceAndNotRepeated(t *testing.T) {
	f := &fake{}
	m := Describe(f, f)
	_, cmd := m.Update("Saved a.cuppa")
	if cmd == nil {
		t.Fatal("a new status should be announced")
	}
	if msg := run(cmd); !hasAnnounce(msg, "Saved a.cuppa") {
		t.Fatalf("announcement = %#v", msg)
	}
	if _, cmd = m.Update("Saved a.cuppa"); cmd != nil {
		t.Fatal("the same status is not read again")
	}
	if _, cmd = m.Update(""); cmd != nil {
		t.Fatal("a cleared status says nothing")
	}
}

func run(cmd tea.Cmd) tea.Msg { return cmd() }

func hasAnnounce(msg tea.Msg, text string) bool {
	if a, ok := msg.(ttygo.AnnounceMsg); ok {
		return a.Text == text
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c != nil && hasAnnounce(c(), text) {
				return true
			}
		}
	}
	return false
}
