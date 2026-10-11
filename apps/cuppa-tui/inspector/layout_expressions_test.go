package inspector

import (
	"strings"
	"testing"

	"github.com/meta-tui/cuppa/libs/canvas/editor"
)

func typeValue(m *Model, s string) {
	for range 12 {
		m.Key("", true, false, false)
	}
	for _, r := range s {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	m.Lines()
}

func TestTypingAnExpressionMakesTheAxisFollowTheCanvas(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.box")
	barClick(t, m, "W", "value")
	typeValue(m, "100% - 20")
	if got := ed.Layout(id).W; got != "100% - 20" {
		t.Fatalf("layout = %q (message %q)", got, m.message)
	}
	if got := rect(ed, id).W; got != 80 {
		t.Fatalf("w = %d", got)
	}
	if !strings.Contains(stripANSI(strings.Join(m.Context(120), "\n")), "100% - 20") {
		t.Fatalf("the expression is not shown: %q", stripANSI(strings.Join(m.Context(120), "|")))
	}
	if err := ed.SetCanvasSize(120, 40); err != nil {
		t.Fatal(err)
	}
	if got := rect(ed, id).W; got != 100 {
		t.Fatalf("w at 120 = %d", got)
	}
}

func TestTypingANumberOverAnExpressionFixesTheAxis(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.box")
	if err := ed.SetLayout(id, editor.AxisW, "50%"); err != nil {
		t.Fatal(err)
	}
	m.Lines()
	barClick(t, m, "W", "value")
	typeValue(m, "30")
	if got := ed.Layout(id).W; got != "" {
		t.Fatalf("layout = %q", got)
	}
	if got := rect(ed, id).W; got != 30 {
		t.Fatalf("w = %d", got)
	}
}

func TestBadExpressionIsReportedAndNothingChanges(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.box")
	barClick(t, m, "W", "value")
	typeValue(m, "50% +")
	if m.message == "" {
		t.Fatal("no message")
	}
	if ed.Layout(id).W != "" || rect(ed, id).W != 24 {
		t.Fatal("the node changed")
	}
}

func TestUnitButtonSwitchesBetweenCellsAndPercent(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.box")
	barClick(t, m, "W", "unit")
	if got := ed.Layout(id).W; got != "24%" {
		t.Fatalf("layout = %q", got)
	}
	barClick(t, m, "W", "unit")
	if ed.Layout(id).W != "" {
		t.Fatal("not fixed again")
	}
}

func TestSizeButtonsSetTheCanvas(t *testing.T) {
	m, ed, _ := setup(t, "lipgloss.box")
	ed.Clear()
	m.Lines()
	clickText(t, m, "[80×24]", "[80×24]")
	if d := ed.Document(); d.Width != 80 || d.Height != 24 {
		t.Fatalf("canvas = %dx%d", d.Width, d.Height)
	}
}

func TestBehaviourCheckboxesSetDraggableAndResizable(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.box")
	clickText(t, m, "Draggable", "[ ] Draggable")
	clickText(t, m, "Resizable", "[ ] Resizable")
	n, _ := ed.Document().Get(id)
	if !n.Draggable || !n.Resizable {
		t.Fatalf("flags = %+v", n)
	}
	clickText(t, m, "Draggable", "[x] Draggable")
	if n, _ := ed.Document().Get(id); n.Draggable {
		t.Fatal("not unticked")
	}
}
