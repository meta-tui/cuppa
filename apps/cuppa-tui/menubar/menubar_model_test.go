package menubar

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
)

func down(x, y int) pointer.Event {
	return pointer.Event{X: x, Y: y, Phase: pointer.Down, Left: true}
}

func newBar() *Model {
	m := New()
	m.SetWidth(100)
	m.Line("")
	return m
}

func TestClickLabelOpensAndClickItemChooses(t *testing.T) {
	m := newBar()
	if a, used := m.Handle(down(m.labelX[0]+1, 0)); a != "" || !used || !m.Open() {
		t.Fatalf("open: %q %v %v", a, used, m.Open())
	}
	x, lines := m.Dropdown()
	if len(lines) == 0 {
		t.Fatal("no dropdown")
	}
	if !strings.Contains(ansi.Strip(strings.Join(lines, "\n")), "Save As") {
		t.Fatalf("File menu lacks Save As:\n%s", ansi.Strip(strings.Join(lines, "\n")))
	}
	// Items start on screen row 2: New, Open, separator, Save, Save As...
	a, used := m.Handle(down(x+3, 2+4))
	if a != FileSaveAs || !used || m.Open() {
		t.Fatalf("chose %q used=%v open=%v", a, used, m.Open())
	}
}

func TestClickOutsideClosesWithoutActing(t *testing.T) {
	m := newBar()
	m.Handle(down(m.labelX[0]+1, 0))
	if a, used := m.Handle(down(80, 20)); a != "" || !used || m.Open() {
		t.Fatalf("got %q used=%v open=%v", a, used, m.Open())
	}
	// With nothing open, a stray click belongs to the panes.
	if _, used := m.Handle(down(80, 20)); used {
		t.Fatal("closed bar swallowed a click")
	}
}

func TestHoverSwitchesBetweenMenus(t *testing.T) {
	m := newBar()
	m.Handle(down(m.labelX[0]+1, 0))
	m.Handle(pointer.Event{X: m.labelX[2] + 1, Y: 0, Phase: pointer.Move})
	if m.open != 2 {
		t.Fatalf("open = %d", m.open)
	}
}

func TestDisabledAndSeparatorsDoNothing(t *testing.T) {
	m := newBar()
	m.SetEnabled(EditUndo, false)
	m.Handle(down(m.labelX[1]+1, 0))
	x, _ := m.Dropdown()
	if a, used := m.Handle(down(x+3, 2)); a != "" || !used || !m.Open() {
		t.Fatalf("disabled undo: %q %v open=%v", a, used, m.Open())
	}
	if a, _ := m.Handle(down(x+3, 4)); a != "" || !m.Open() {
		t.Fatalf("separator: %q", a)
	}
}

func TestClickingOpenLabelAgainCloses(t *testing.T) {
	m := newBar()
	m.Handle(down(m.labelX[3]+1, 0))
	m.Handle(down(m.labelX[3]+1, 0))
	if m.Open() {
		t.Fatal("still open")
	}
}

func TestLineIsExactlyWide(t *testing.T) {
	m := New()
	m.SetWidth(60)
	if w := ansi.StringWidth(m.Line("demo.cuppa •")); w != 60 {
		t.Fatalf("width %d", w)
	}
}

func TestDropdownStaysOnScreen(t *testing.T) {
	m := New()
	m.SetWidth(30)
	m.Line("")
	m.Handle(down(m.labelX[3]+1, 0))
	x, lines := m.Dropdown()
	if x+ansi.StringWidth(lines[0]) > 30 {
		t.Fatalf("dropdown spills: x=%d w=%d", x, ansi.StringWidth(lines[0]))
	}
}

func TestIconFontBrandKeepsTheBarsWidthAndLayout(t *testing.T) {
	plain := New()
	plain.SetWidth(80)
	icon := New()
	icon.SetWidth(80)
	icon.SetIconFont(true)
	a, b := plain.Line("x"), icon.Line("x")
	if ansi.StringWidth(a) != ansi.StringWidth(b) {
		t.Fatalf("widths differ: %d and %d", ansi.StringWidth(a), ansi.StringWidth(b))
	}
	if !strings.Contains(b, "") || strings.Contains(b, "☕") {
		t.Fatalf("the icon bar should hold the two logo characters: %q", ansi.Strip(b))
	}
	if strings.Contains(a, "") {
		t.Fatal("a plain terminal must not get private-use characters")
	}
	plain.layout()
	icon.layout()
	for i := range plain.labelX {
		if plain.labelX[i] != icon.labelX[i] {
			t.Fatalf("menu %d moved from column %d to %d", i, plain.labelX[i], icon.labelX[i])
		}
	}
}
