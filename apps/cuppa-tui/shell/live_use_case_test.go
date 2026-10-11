package shell

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-ui-labs/cuppa/libs/connector/serve"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// messages runs a command and the commands a batch holds, and returns what
// they produced.
func messages(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	switch msg := cmd().(type) {
	case nil:
		return nil
	case tea.BatchMsg:
		var out []tea.Msg
		for _, c := range msg {
			out = append(out, messages(c)...)
		}
		return out
	default:
		return []tea.Msg{msg}
	}
}

// The designer connects to a running application, draws its values instead of
// the design's, and goes back to the design's when it disconnects; the design
// is never changed by it.
func TestTheDesignerDrawsTheValuesOfARunningApplication(t *testing.T) {
	app, err := serve.Listen("127.0.0.1:0", serve.Options{App: "demo", Pages: []string{"Untitled"}})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = app.Close() }()
	if err := app.Publish("Untitled", map[string]any{"Title": "From the app"}, true); err != nil {
		t.Fatal(err)
	}

	m := newShell(t)
	doc := m.ed.Document()
	doc.Add(design.Node{Component: "lipgloss.box", Name: "Box", Rect: design.Rect{W: 20, H: 3},
		Props: map[string]string{"title": "design"}, Bind: map[string]string{"title": "Title"}})
	m.ed.Load(doc)

	m.Connect(app.Addr())
	_, cmd := m.Update(struct{}{})
	connected := messages(cmd)
	if len(connected) != 1 {
		t.Fatalf("dial gave %d messages", len(connected))
	}
	_, cmd = m.Update(connected[0])
	state := messages(cmd)
	if len(state) != 1 {
		t.Fatalf("wait gave %d messages", len(state))
	}
	m.Update(state[0])

	if m.live == nil || !m.ed.Live() {
		t.Fatal("not live")
	}
	if got := m.ed.RenderDocument().Nodes[0].Props["title"]; got != "From the app" {
		t.Fatalf("title %q", got)
	}
	if !strings.Contains(m.statusBar(), "live: demo") {
		t.Errorf("status: %q", m.statusBar())
	}
	if m.ed.Document().Nodes[0].Props["title"] != "design" || m.ed.Dirty() {
		t.Error("the design changed")
	}

	m.disconnect("")
	if m.live != nil || m.ed.Live() {
		t.Fatal("still live after Disconnect")
	}
	if got := m.ed.RenderDocument().Nodes[0].Props["title"]; got != "design" {
		t.Fatalf("title after disconnect %q", got)
	}
}

func TestAnAddressThatIsNotThisComputerIsRefused(t *testing.T) {
	m := newShell(t)
	m.Connect("example.com:80")
	_, cmd := m.Update(struct{}{})
	got := messages(cmd)
	if len(got) != 1 {
		t.Fatalf("messages %v", got)
	}
	m.Update(got[0])
	if m.live != nil {
		t.Fatal("connected to a remote address")
	}
	if m.flow.Modal() == nil {
		t.Fatal("no notice")
	}
}

