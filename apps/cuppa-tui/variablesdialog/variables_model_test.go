package variablesdialog

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
)

func sample() []editor.Variable {
	return []editor.Variable{
		{Kind: editor.VarColour, Name: "Accent", Type: "colour", Value: "#ff007f", Uses: []editor.Use{
			{Node: "n1", NodeName: "Title", Where: "Foreground", Prop: "color", Role: editor.UseProp},
			{Node: "n2", NodeName: "Cursor", Where: "Foreground", Prop: "color", Role: editor.UseProp},
		}},
		{Kind: editor.VarInput, Name: "Status", Type: "text", Value: "saved", Uses: []editor.Use{
			{Node: "n3", NodeName: "Status", Where: "Text", Prop: "text", Role: editor.UseProp},
		}},
		{Kind: editor.VarEvent, Name: "Back", Type: "event", Uses: []editor.Use{
			{NodeName: "Screen keys", Where: "key esc", Prop: "esc", Role: editor.UseKey},
		}},
	}
}

func open(t *testing.T) *Model {
	t.Helper()
	m := New(sample())
	m.Place(100, 40)
	return m
}

func text(m *Model) string { return ansi.Strip(strings.Join(m.Lines(), "\n")) }

func TestListsEveryVariableWithItsUses(t *testing.T) {
	m := open(t)
	got := text(m)
	for _, want := range []string{"Variables", "Accent", "#ff007f", "2 uses", "Status", "1 use", "Back", "[ Go to ]", "[ Rename… ]", "[ Done ]"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestGoToAVariableUsedOnceSelectsItsComponent(t *testing.T) {
	m := open(t)
	m.Nav("down")
	m.Key("", false, true, false)
	o, done := m.Outcome()
	if !done || o.Button != GoToButton || o.Value != "n3" {
		t.Fatalf("outcome = %+v done=%v", o, done)
	}
}

func TestAVariableUsedSeveralTimesListsItsUsesToChooseOne(t *testing.T) {
	m := open(t)
	m.Key("", false, true, false) // Accent has two uses
	if _, done := m.Outcome(); done {
		t.Fatal("several uses must be listed first")
	}
	got := text(m)
	if !strings.Contains(got, "is used in 2 places") || !strings.Contains(got, "Title") || !strings.Contains(got, "Cursor") {
		t.Fatalf("uses not listed:\n%s", got)
	}
	m.Nav("down")
	m.Key("", false, true, false)
	o, done := m.Outcome()
	if !done || o.Value != "n2" {
		t.Fatalf("outcome = %+v", o)
	}
}

func TestEscGoesBackThenCloses(t *testing.T) {
	m := open(t)
	m.Key("", false, true, false)
	m.Key("", false, false, true)
	if _, done := m.Outcome(); done {
		t.Fatal("Esc in the uses list goes back to the variables")
	}
	m.Key("", false, false, true)
	if o, done := m.Outcome(); !done || !o.Canceled {
		t.Fatalf("outcome = %+v", o)
	}
}

func TestRenameReportsTheKindAndName(t *testing.T) {
	m := open(t)
	m.Nav("end")
	m.Nav("r")
	o, done := m.Outcome()
	if !done || o.Button != RenameButton || o.Value != "event:Back" {
		t.Fatalf("outcome = %+v", o)
	}
}

func TestTheMouseSelectsAndPressesButtons(t *testing.T) {
	m := open(t)
	r := m.Rect()
	_ = text(m) // lays the rows out
	// Click the second variable, then Go to.
	m.Handle(pointer.Event{X: r.X + 10, Y: r.Y + m.rowY + 1, Phase: pointer.Down, Left: true})
	if m.cursor != 1 {
		t.Fatalf("cursor = %d", m.cursor)
	}
	press := func(x int) bool {
		m.Handle(pointer.Event{X: r.X + x, Y: r.Y + r.H - 2, Phase: pointer.Down, Left: true})
		_, done := m.Outcome()
		return done
	}
	for x := 2; x < r.W-2; x++ {
		if press(x) {
			break
		}
	}
	if o, done := m.Outcome(); !done || o.Button == "" {
		t.Fatalf("no button pressed: %+v", o)
	}
}
