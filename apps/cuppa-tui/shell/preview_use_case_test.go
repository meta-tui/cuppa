package shell

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
)

func ctrl(r rune) tea.Msg { return tea.KeyPressMsg{Code: r, Mod: tea.ModCtrl} }

func typed(r rune) tea.Msg { return tea.KeyPressMsg{Code: r, Text: string(r)} }

// feed runs a command (giving up on one that waits for a timer, as the real
// program runs those in the background) and feeds its message back.
func feed(m *Model, cmd tea.Cmd) {
	if cmd == nil {
		return
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				feed(m, c)
			}
			return
		}
		if msg != nil {
			_, next := m.Update(msg)
			feed(m, next)
		}
	case <-time.After(40 * time.Millisecond):
	}
}

func screen(m *Model) string { return ansi.Strip(m.render()) }

func TestCtrlPRunsTheDesignAndEscGoesBackToEditing(t *testing.T) {
	m := newShell(t)
	_, _ = m.ed.Add("bubbles.textinput", 4, 3)
	m.ed.Clear()
	_, cmd := m.Update(ctrl('p'))
	feed(m, cmd)
	if !m.Previewing() {
		t.Fatal("Ctrl+P starts the preview")
	}
	if !strings.Contains(m.Describe().Title, "Untitled") || !strings.Contains(described(m), "Preview running") {
		t.Fatal("a screen reader is told the preview is running")
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if m.Previewing() {
		t.Fatal("Esc leaves the preview")
	}
	if len(m.ed.Document().Nodes) != 1 {
		t.Fatal("previewing never changes the design")
	}
}

func TestInThePreviewARealTextInputTakesTyping(t *testing.T) {
	m := newShell(t)
	id, _ := m.ed.Add("bubbles.textinput", 4, 3)
	if err := m.ed.SetProp(id, "placeholder", "Your name"); err != nil {
		t.Fatal(err)
	}
	m.ed.Clear()
	_, cmd := m.Update(ctrl('p'))
	feed(m, cmd)
	if !strings.Contains(screen(m), "Your name") {
		t.Fatalf("the placeholder shows at first:\n%s", screen(m))
	}
	// Click on the input: stage origin plus the node's cell.
	x, y := m.layout.stage.X+6, m.layout.stage.Y+3
	_, cmd = m.Update(click(x, y))
	feed(m, cmd)
	for _, r := range "Ada" {
		_, cmd = m.Update(typed(r))
		feed(m, cmd)
	}
	out := screen(m)
	if !strings.Contains(out, "Ada") || strings.Contains(out, "Your name") {
		t.Fatalf("typing shows in the text input:\n%s", out)
	}
	if n := len(m.ed.Document().Nodes); n != 1 {
		t.Fatal("typing in a preview does not edit the design")
	}
}

func TestEditingIsOffWhilePreviewing(t *testing.T) {
	m := newShell(t)
	id, _ := m.ed.Add("lipgloss.box", 4, 3)
	m.ed.Clear()
	m.Update(ctrl('p'))
	// Pressing and dragging on the box must not select or move it.
	x, y := m.layout.stage.X+6, m.layout.stage.Y+4
	m.Update(click(x, y))
	m.Update(motion(x+10, y+3))
	m.Update(release(x+10, y+3))
	n, _ := m.ed.Document().Get(id)
	if n.Rect.X != 4 || n.Rect.Y != 3 || m.ed.IsSelected(id) {
		t.Fatalf("the canvas is not editable in a preview: %+v", n.Rect)
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyDelete})
	if len(m.ed.Document().Nodes) != 1 {
		t.Fatal("Delete does nothing in a preview")
	}
}

func TestASpinnerAnimatesInThePreview(t *testing.T) {
	m := newShell(t)
	id, _ := m.ed.Add("bubbles.spinner", 4, 3)
	if err := m.ed.SetProp(id, "style", "line"); err != nil {
		t.Fatal(err)
	}
	m.ed.Clear()
	_, cmd := m.Update(ctrl('p'))
	before := screen(m)
	feed(m, cmd)
	if screen(m) == before {
		t.Fatalf("a tick moves the spinner:\n%s", before)
	}
}

func TestEditingActionsAndOpeningEndThePreview(t *testing.T) {
	m := newShell(t)
	_, _ = m.ed.Add("bubbles.spinner", 4, 3)
	m.ed.MarkSaved() // nothing unsaved, so New does not ask
	m.Update(ctrl('p'))
	m.Update(ctrl('n'))
	if m.Previewing() {
		t.Fatal("starting a new design ends the preview")
	}
	m.Update(ctrl('p'))
	m.perform(menubar.HelpAbout)
	if !m.Previewing() {
		t.Fatal("Help does not end it")
	}
}
