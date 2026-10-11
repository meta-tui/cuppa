package inspector

import (
	"slices"
	"strings"
	"testing"

	"github.com/meta-tui/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-tui/cuppa/libs/catalog/standard"
	"github.com/meta-tui/cuppa/libs/canvas/editor"
	"github.com/meta-tui/cuppa/libs/document/design"
)

func TestPanelHeightsGiveTheLastOpenPanelWhatIsLeft(t *testing.T) {
	for _, c := range []struct {
		name        string
		avail       int
		wants, mins []int
		open        []bool
		want        []int
	}{
		{"roomy", 30, []int{8, 10, 5}, []int{3, 3, 3}, []bool{true, true, true}, []int{8, 10, 12}},
		{"a closed panel has no body", 30, []int{8, 10, 5}, []int{3, 3, 3}, []bool{true, false, true}, []int{8, 0, 22}},
		{"short bar takes from the bottom up", 14, []int{8, 10, 5}, []int{3, 3, 3}, []bool{true, true, true}, []int{8, 3, 3}},
		{"very short bar goes under the minimums", 4, []int{8, 10, 5}, []int{3, 3, 3}, []bool{true, true, true}, []int{3, 1, 0}},
		{"all closed", 10, []int{8}, []int{3}, []bool{false}, []int{0}},
	} {
		if got := panelHeights(c.avail, c.wants, c.mins, c.open); !slices.Equal(got, c.want) {
			t.Errorf("%s: %v, want %v", c.name, got, c.want)
		}
	}
}

func plainLines(m *Model) []string {
	out := m.Lines()
	for i := range out {
		out[i] = stripANSI(out[i])
	}
	return out
}

func rowOf(lines []string, text string) int {
	for y, l := range lines {
		if strings.Contains(l, text) {
			return y
		}
	}
	return -1
}

// A panel keeps its place when the selection has more or fewer properties:
// the layer list does not jump.
func TestPanelsStayPutWhenTheSelectionChanges(t *testing.T) {
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 100, 60))
	box, _ := ed.Add("lipgloss.box", 2, 2)
	picker, _ := ed.Add("huh.filepicker", 20, 2)
	m := New(ed, cat)
	m.SetSize(32, 40)

	ed.Select(box)
	before := rowOf(plainLines(m), "▾ Layers")
	ed.Select(picker)
	if after := rowOf(plainLines(m), "▾ Layers"); after != before || before < 0 {
		t.Fatalf("the Layers header moved from row %d to %d", before, after)
	}
}

func TestClickingAHeaderClosesAndOpensItsPanel(t *testing.T) {
	m, _, _ := setup(t, "lipgloss.box")
	if rowOf(plainLines(m), "[Duplicate]") < 0 {
		t.Fatal("Arrange starts open")
	}
	clickText(t, m, "▾ Arrange", "Arrange")
	m.Handle(pointer.Event{Phase: pointer.Up, Left: true})
	lines := plainLines(m)
	if rowOf(lines, "[Duplicate]") >= 0 || rowOf(lines, "▸ Arrange") < 0 {
		t.Fatalf("Arrange should be closed:\n%s", strings.Join(lines, "\n"))
	}
	clickText(t, m, "▸ Arrange", "Arrange")
	m.Handle(pointer.Event{Phase: pointer.Up, Left: true})
	if rowOf(plainLines(m), "[Duplicate]") < 0 {
		t.Fatal("clicking again opens it")
	}
}

func TestDraggingAHeaderResizesThePanelAboveIt(t *testing.T) {
	m, _, _ := setup(t, "lipgloss.box")
	m.SetSize(32, 60)
	for id := range panelDefaults {
		m.state(id).want = panelDefaults[id].want
	}
	lines := plainLines(m)
	props, arrange := rowOf(lines, "▾ Properties"), rowOf(lines, "▾ Arrange")
	before := arrange - props - 1
	m.Handle(pointer.Event{X: 3, Y: arrange, Phase: pointer.Down, Left: true})
	m.Handle(pointer.Event{X: 3, Y: arrange + 3, Phase: pointer.Move, Held: true})
	m.Handle(pointer.Event{X: 3, Y: arrange + 3, Phase: pointer.Up, Left: true})
	lines = plainLines(m)
	props, arrange = rowOf(lines, "▾ Properties"), rowOf(lines, "▾ Arrange")
	if got := arrange - props - 1; got != before+3 {
		t.Fatalf("Properties is %d rows tall, want %d", got, before+3)
	}
	if rowOf(lines, "▾ Properties") < 0 {
		t.Fatal("dragging a header must not close the panel")
	}
}

// The wheel scrolls the panel under the pointer and nothing else.
func TestEachPanelScrollsInsideItself(t *testing.T) {
	m, _, _ := setup(t, "lipgloss.box")
	m.SetSize(32, 40)
	st := m.state(panelProperties)
	st.want, st.closed = 4, false
	m.state(panelDetails).want = 8
	lines := plainLines(m)
	props := rowOf(lines, "▾ Properties")
	first := lines[props+1]
	m.Handle(pointer.Event{X: 5, Y: props + 2, Phase: pointer.Wheel, WheelY: 1})
	lines = plainLines(m)
	if lines[props+1] == first {
		t.Fatalf("the Properties body did not scroll:\n%s", strings.Join(lines, "\n"))
	}
	if got := rowOf(lines, "▾ Layers"); got < 0 {
		t.Fatal("the other panels stay where they were")
	}
	if m.state(panelDetails).scroll != 0 {
		t.Error("only the panel under the pointer scrolls")
	}
}

func TestAChoiceOpensADropdownTheKeyboardCanUse(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.box")
	clickText(t, m, "rounded", "rounded")
	if !m.Editing() || m.menu == nil {
		t.Fatal("pressing a choice opens its list")
	}
	m.Step(-1) // Down
	m.Key("", false, true, false)
	if n, _ := ed.Document().Get(id); n.Props["border"] != "normal" {
		t.Fatalf("border = %q", n.Props["border"])
	}
	if m.menu != nil {
		t.Fatal("choosing closes the list")
	}
	clickText(t, m, "normal", "normal")
	m.Key("", false, false, true)
	if m.menu != nil {
		t.Fatal("Esc closes the list")
	}
	if n, _ := ed.Document().Get(id); n.Props["border"] != "normal" {
		t.Fatal("Esc must not change the value")
	}
}

func TestAClickOutsideTheDropdownClosesItWithoutChoosing(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.box")
	clickText(t, m, "rounded", "rounded")
	m.Lines()
	m.Handle(pointer.Event{X: 30, Y: 0, Phase: pointer.Down, Left: true})
	if m.menu != nil {
		t.Fatal("the list should close")
	}
	if n, _ := ed.Document().Get(id); n.Props["border"] == "normal" {
		t.Fatal("nothing is chosen")
	}
}
