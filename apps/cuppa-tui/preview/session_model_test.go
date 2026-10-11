package preview

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/meta-tui/cuppa/libs/catalog/standard"
	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/render/grid"
)

func text(g *grid.Grid) string {
	var b strings.Builder
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			ch := g.At(x, y).Ch
			if ch == 0 {
				ch = ' '
			}
			b.WriteRune(ch)
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// look is the text with the colours: a selection is often only a colour.
func look(g *grid.Grid) string {
	var b strings.Builder
	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			c := g.At(x, y)
			b.WriteString(string(c.Ch) + c.Fg + "/" + c.Bg + ";")
		}
	}
	return b.String()
}

func newSession(t *testing.T, nodes ...design.Node) (*Session, tea.Cmd) {
	t.Helper()
	doc := design.NewDocument("t", 80, 20)
	for _, n := range nodes {
		doc.Add(n)
	}
	return Start(doc, standard.Default())
}

func node(component, name string, r design.Rect, props map[string]string) design.Node {
	return design.Node{Component: component, Name: name, Rect: r, Props: props}
}

// settle runs a command and feeds what comes back to the session, a few rounds.
func settle(s *Session, cmd tea.Cmd) {
	for range 5 {
		if cmd == nil {
			return
		}
		msg := runWithin(cmd)
		cmd = nil
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				if c != nil {
					settle(s, c)
				}
			}
			return
		}
		if _, next := s.Update(msg); next != nil {
			cmd = next
		}
	}
}

// runWithin runs a command but gives up on one that waits on a timer (a
// cursor blink): in the real program those run in the background.
func runWithin(cmd tea.Cmd) tea.Msg {
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		return msg
	case <-time.After(40 * time.Millisecond):
		return nil
	}
}

func key(s string) tea.KeyPressMsg {
	switch s {
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	}
	r := []rune(s)
	return tea.KeyPressMsg{Code: r[0], Text: s}
}

func TestTypingIntoARealTextInputChangesWhatIsDrawn(t *testing.T) {
	s, _ := newSession(t, node("bubbles.textinput", "Name", design.Rect{X: 2, Y: 2, W: 30, H: 1}, map[string]string{"placeholder": "Your name"}))
	if !strings.Contains(text(s.Render()), "Your name") {
		t.Fatalf("the placeholder shows until you type:\n%s", text(s.Render()))
	}
	settle(s, s.Press(5, 2)) // click to focus
	if _, ok := s.Focused(); !ok {
		t.Fatal("clicking a text input gives it the keyboard")
	}
	for _, r := range "Ada" {
		handled, cmd := s.Key(key(string(r)))
		if !handled {
			t.Fatal("the focused input takes the key")
		}
		settle(s, cmd)
	}
	out := text(s.Render())
	if !strings.Contains(out, "Ada") || strings.Contains(out, "Your name") {
		t.Fatalf("what was typed replaces the placeholder:\n%s", out)
	}
}

func TestAListMovesItsSelectionWithTheKeyboard(t *testing.T) {
	s, _ := newSession(t, node("bubbles.list", "Pantry", design.Rect{X: 1, Y: 1, W: 30, H: 10}, map[string]string{"show_filter": "false", "show_status": "false"}))
	before := look(s.Render())
	settle(s, s.Press(5, 3))
	_, cmd := s.Key(key("down"))
	settle(s, cmd)
	if look(s.Render()) == before {
		t.Fatal("moving down changes which item is highlighted")
	}
}

func TestASpinnerAnimatesWhenItsTicksAreFedBack(t *testing.T) {
	s, start := newSession(t, node("bubbles.spinner", "Busy", design.Rect{X: 1, Y: 1, W: 20, H: 1}, map[string]string{"style": "line", "label": "Brewing"}))
	first := text(s.Render())
	if start == nil {
		t.Fatal("a spinner asks to be ticked")
	}
	settle(s, start)
	if text(s.Render()) == first {
		t.Fatalf("a tick advances the frame:\n%s", first)
	}
	if !strings.Contains(text(s.Render()), "Brewing") {
		t.Fatal("the label stays")
	}
}

func TestProgressAndPaginatorRespondToArrowKeys(t *testing.T) {
	s, start := newSession(t,
		node("bubbles.progress", "Bar", design.Rect{X: 1, Y: 1, W: 30, H: 1}, map[string]string{"percent": "50"}),
		node("bubbles.paginator", "Pages", design.Rect{X: 1, Y: 3, W: 20, H: 1}, map[string]string{"total": "5", "page": "1", "style": "arabic"}),
	)
	settle(s, start)
	if !strings.Contains(text(s.Render()), "1/5") {
		t.Fatalf("starts on page 1:\n%s", text(s.Render()))
	}
	settle(s, s.Press(2, 3)) // focus the paginator
	_, cmd := s.Key(key("right"))
	settle(s, cmd)
	if !strings.Contains(text(s.Render()), "2/5") {
		t.Fatalf("right goes to page 2:\n%s", text(s.Render()))
	}
}

func TestTabMovesTheKeyboardBetweenComponentsAndClickingElsewhereClearsIt(t *testing.T) {
	s, _ := newSession(t,
		node("bubbles.textinput", "A", design.Rect{X: 1, Y: 1, W: 20, H: 1}, nil),
		node("bubbles.textarea", "B", design.Rect{X: 1, Y: 4, W: 20, H: 4}, nil),
		node("lipgloss.box", "Static", design.Rect{X: 40, Y: 1, W: 10, H: 3}, nil),
	)
	if s.Count() != 2 {
		t.Fatalf("two components run as real models, the box does not: %d", s.Count())
	}
	if _, ok := s.Focused(); ok {
		t.Fatal("nothing is focused at the start")
	}
	_, cmd := s.Key(key("tab"))
	settle(s, cmd)
	if n, _ := s.Focused(); n.Name != "A" {
		t.Fatalf("Tab focuses the first: %q", n.Name)
	}
	_, cmd = s.Key(key("tab"))
	settle(s, cmd)
	if n, _ := s.Focused(); n.Name != "B" {
		t.Fatalf("Tab again focuses the next: %q", n.Name)
	}
	settle(s, s.Press(45, 2)) // the static box
	if _, ok := s.Focused(); ok {
		t.Fatal("clicking something that does not take the keyboard clears focus")
	}
	if handled, _ := s.Key(key("x")); handled {
		t.Fatal("with nothing focused a key is not handled")
	}
}

func TestStoppedSessionsIgnoreLateMessages(t *testing.T) {
	s, start := newSession(t, node("bubbles.spinner", "Busy", design.Rect{X: 1, Y: 1, W: 20, H: 1}, nil))
	before := text(s.Render())
	s.Stop()
	settle(s, start)
	if text(s.Render()) != before {
		t.Fatal("after Stop nothing moves")
	}
}

func TestComponentsWithoutARealModelKeepTheirPreview(t *testing.T) {
	s, _ := newSession(t, node("lipgloss.box", "Box", design.Rect{X: 1, Y: 1, W: 10, H: 3}, nil))
	if s.Count() != 0 || !strings.ContainsAny(text(s.Render()), "╭┌") {
		t.Fatal("a box is drawn by its painter")
	}
}

func TestEveryRealModelStartsAtItsDefaultAndMinimumSize(t *testing.T) {
	cat := standard.Default()
	for id := range makers {
		def, ok := cat.Get(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		for _, size := range []definition_size{{def.DefaultSize.W, def.DefaultSize.H}, {def.MinSize.W, def.MinSize.H}} {
			doc := design.NewDocument("t", 80, 30)
			doc.Add(design.Node{Component: id, Name: "X", Rect: design.Rect{W: size.w, H: size.h}})
			s, start := Start(doc, cat)
			settle(s, start)
			g := s.Render() // must not panic at any size
			if g.W != 80 {
				t.Errorf("%s: odd grid", id)
			}
		}
	}
}

type definition_size struct{ w, h int }

// In the preview, pressing a tab shows its page and hides the others; the
// design itself is not changed.
func TestPressingATabInThePreviewChangesThePage(t *testing.T) {
	cat := standard.Default()
	doc := design.NewDocument("t", 60, 12)
	tabs := doc.Add(design.Node{Component: "cuppa.tabs", Name: "Tabs", Rect: design.Rect{X: 2, Y: 1, W: 40, H: 8},
		Props: map[string]string{"pages": "Alpha,Beta", "active": "0"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "A", In: tabs.ID, Page: 0, Rect: design.Rect{X: 4, Y: 4, W: 12, H: 1}, Props: map[string]string{"text": "ALPHA PAGE"}})
	doc.Add(design.Node{Component: "lipgloss.label", Name: "B", In: tabs.ID, Page: 1, Rect: design.Rect{X: 4, Y: 4, W: 12, H: 1}, Props: map[string]string{"text": "BETA PAGE"}})
	s, _ := Start(doc, cat)
	if v := text(s.Render()); !strings.Contains(v, "ALPHA PAGE") || strings.Contains(v, "BETA PAGE") {
		t.Fatalf("at the start:\n%s", v)
	}
	s.Press(14, 1) // the Beta tab
	if v := text(s.Render()); strings.Contains(v, "ALPHA PAGE") || !strings.Contains(v, "BETA PAGE") {
		t.Fatalf("after pressing Beta:\n%s", v)
	}
	if n, _ := doc.Get(tabs.ID); n.Props["active"] != "0" {
		t.Error("the preview changed the design")
	}
}

// In the preview, a component that goes to a page takes the person there.
func TestAComponentThatGoesToAPageChangesThePageInThePreview(t *testing.T) {
	cat := standard.Default()
	doc := design.NewDocument("t", 40, 8)
	doc.Add(design.Node{Component: "lipgloss.label", Name: "Go", Rect: design.Rect{X: 1, Y: 1, W: 12, H: 1},
		Props: map[string]string{"text": "HOME PAGE"}, Goto: "Two"})
	doc.Pages = []design.Page{{Name: "One"}, {Name: "Two", Nodes: []design.Node{
		{ID: "n9", Component: "lipgloss.label", Name: "Other", Rect: design.Rect{X: 1, Y: 1, W: 12, H: 1}, Props: map[string]string{"text": "SECOND PAGE"}},
	}}}
	s, _ := Start(doc, cat)
	if v := text(s.Render()); !strings.Contains(v, "HOME PAGE") {
		t.Fatalf("at the start:\n%s", v)
	}
	s.Press(2, 1)
	if v := text(s.Render()); !strings.Contains(v, "SECOND PAGE") || strings.Contains(v, "HOME PAGE") {
		t.Fatalf("after the click:\n%s", v)
	}
}
