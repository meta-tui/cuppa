package panels

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func text(body string) func(w, h int) string {
	return func(w, h int) string { return body }
}

func sections(open ...int) Model {
	return New(
		WithPages(Page{Title: "A", Body: text("body a")}, Page{Title: "B", Body: text("body b")}, Page{Title: "C", Body: text("body c")}),
		WithOpen(open...),
		WithSize(30, 13),
	)
}

func tabs(active int) Model {
	return New(
		WithMode(Tabs),
		WithPages(Page{Title: "Overview", Body: text("one")}, Page{Title: " Colours ", Body: text("two")}, Page{Title: "Keys", Body: text("three")}),
		WithActive(active),
		WithSize(40, 10),
	)
}

// The same arithmetic as Cuppa's designer, with the same numbers.
func TestOpenSectionsShareTheRowsAfterTheHeaders(t *testing.T) {
	g := sections(0, 2).Layout()
	// 13 rows, 3 headers: 10 rows for 2 open sections.
	if g.Heads[0] != 0 || g.Bodies[0] != (Rect{0, 1, 30, 5}) {
		t.Fatalf("section 0: head %d body %+v", g.Heads[0], g.Bodies[0])
	}
	if g.Heads[1] != 6 || g.Bodies[1].H != 0 || g.Bodies[1].Y != 7 {
		t.Fatalf("closed section 1: head %d body %+v", g.Heads[1], g.Bodies[1])
	}
	if g.Heads[2] != 7 || g.Bodies[2] != (Rect{0, 8, 30, 5}) {
		t.Fatalf("section 2: head %d body %+v", g.Heads[2], g.Bodies[2])
	}
}

func TestTabsHaveABarAndABodyUnderIt(t *testing.T) {
	m := tabs(1)
	g := m.Layout()
	if len(g.Tabs) != 3 || g.Tabs[0] != (Span{0, 10}) || g.Tabs[1].X0 != 11 {
		t.Fatalf("tabs = %+v", g.Tabs)
	}
	if want := (Rect{0, 2, 40, 8}); g.Bodies[0] != want {
		t.Fatalf("body = %+v", g.Bodies[0])
	}
	if !m.Shown(1) || m.Shown(0) {
		t.Error("only the active tab is shown")
	}
	if tabs(9).Active() != 2 || tabs(-3).Active() != 0 {
		t.Error("the active page is kept inside the pages")
	}
}

func TestPressingAHeaderOpensAndClosesAndAnAccordionKeepsOne(t *testing.T) {
	m := sections(0)
	if !m.Press(1) || !m.Open(0) || !m.Open(1) {
		t.Fatalf("open = %v", m.OpenSections())
	}
	if !m.Press(0) || m.Open(0) {
		t.Fatal("pressing an open header closes it")
	}
	a := New(WithPages(Page{Title: "A"}, Page{Title: "B"}), WithExclusive(true), WithOpen(0))
	a.Press(1)
	if a.Open(0) || !a.Open(1) {
		t.Fatalf("accordion open = %v", a.OpenSections())
	}
}

func TestAClickOnATabChangesThePageAndSaysSo(t *testing.T) {
	m := tabs(0)
	m.SetOrigin(5, 3)
	next, cmd := m.Update(tea.MouseClickMsg{X: 5 + 14, Y: 3, Button: tea.MouseLeft})
	if next.Active() != 1 {
		t.Fatalf("active = %d", next.Active())
	}
	if cmd == nil {
		t.Fatal("no ChangedMsg")
	}
	if got, ok := cmd().(ChangedMsg); !ok || got.Active != 1 {
		t.Fatalf("message = %#v", cmd())
	}
	if _, cmd := next.Update(tea.MouseClickMsg{X: 5 + 14, Y: 3, Button: tea.MouseLeft}); cmd != nil {
		t.Error("pressing the shown tab again changes nothing")
	}
}

func TestTheKeyboardWalksSectionsAndTabs(t *testing.T) {
	m := sections()
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if !m.Open(1) || m.Open(0) {
		t.Fatalf("open = %v", m.OpenSections())
	}
	tb := tabs(0)
	tb, _ = tb.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if tb.Active() != 1 {
		t.Fatalf("active = %d", tb.Active())
	}
}

func TestViewDrawsHeadersAndTheBodiesOfShownPagesOnly(t *testing.T) {
	v := ansi.Strip(sections(0, 2).View())
	lines := strings.Split(v, "\n")
	if len(lines) != 13 {
		t.Fatalf("%d lines", len(lines))
	}
	for _, want := range []string{"▾ A", "▸ B", "▾ C", "body a", "body c"} {
		if !strings.Contains(v, want) {
			t.Errorf("view lacks %q:\n%s", want, v)
		}
	}
	if strings.Contains(v, "body b") {
		t.Error("a closed section shows its body")
	}
	for i, l := range lines {
		if ansi.StringWidth(l) != 30 {
			t.Errorf("line %d is %d cells", i, ansi.StringWidth(l))
		}
	}
	tv := ansi.Strip(tabs(1).View())
	if !strings.Contains(tv, "two") || strings.Contains(tv, "one") {
		t.Errorf("tabs view:\n%s", tv)
	}
}

func TestACopyDoesNotShareTheOpenSet(t *testing.T) {
	a := sections(0)
	b, _ := a.Update(tea.MouseClickMsg{X: 1, Y: a.Layout().Heads[1], Button: tea.MouseLeft})
	if a.Open(1) || !b.Open(1) {
		t.Fatalf("a %v b %v", a.OpenSections(), b.OpenSections())
	}
}
