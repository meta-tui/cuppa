package packsdialog

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

func fixture() (*Model, map[definition.Family]bool) {
	on := map[definition.Family]bool{"a": true, "b": true}
	m := New([]Entry{
		{Pack: definition.Pack{ID: "a", Name: "Alpha", Description: "first"}, Count: 3},
		{Pack: definition.Pack{ID: "b", Name: "Beta", Description: "second", Source: "/packs/beta.cupp"}, Count: 5},
	}, func(f definition.Family) bool { return on[f] }, func(f definition.Family) { on[f] = !on[f] })
	m.Place(100, 40)
	m.Lines()
	return m, on
}

func click(m *Model, x, y int) {
	m.Handle(pointer.Event{X: m.Rect().X + x, Y: m.Rect().Y + y, Phase: pointer.Down, Left: true})
}

func TestEachPackShowsItsStateAndCount(t *testing.T) {
	m, _ := fixture()
	screen := strings.Join(m.Lines(), "\n")
	for _, want := range []string{"[x] Alpha", "3 components", "[x] Beta", "5 components", "first", "second", "Done"} {
		if !strings.Contains(stripANSI(screen), want) {
			t.Errorf("missing %q:\n%s", want, stripANSI(screen))
		}
	}
}

func TestClickingARowTogglesThatPack(t *testing.T) {
	m, on := fixture()
	click(m, 8, rowTop(1))
	if on["b"] || !on["a"] {
		t.Fatalf("state = %v", on)
	}
	click(m, 8, rowTop(1)+1) // its description line counts too
	if !on["b"] {
		t.Fatal("second click switches it back on")
	}
	if _, done := m.Outcome(); done {
		t.Fatal("toggling does not close the box")
	}
}

func TestDoneAndEscapeCloseIt(t *testing.T) {
	m, _ := fixture()
	click(m, m.Rect().W-6, m.Rect().H-2)
	if o, done := m.Outcome(); !done || o.Canceled {
		t.Fatalf("Done button: %+v %v", o, done)
	}
	m, _ = fixture()
	m.Key("", false, false, true)
	if _, done := m.Outcome(); !done {
		t.Fatal("Esc closes")
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	skip := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			skip = true
		case skip && r == 'm':
			skip = false
		case !skip:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func TestOnlyInstalledPacksCanBeRemoved(t *testing.T) {
	m, on := fixture()
	screen := stripANSI(strings.Join(m.Lines(), "\n"))
	if strings.Count(screen, "[Remove]") != 1 {
		t.Fatalf("one Remove, on the installed pack:\n%s", screen)
	}
	click(m, m.Rect().W-5, rowTop(0)) // the same spot on the built-in pack toggles it
	if on["a"] {
		t.Fatal("a built-in pack row only toggles")
	}
	click(m, m.Rect().W-5, rowTop(1))
	o, done := m.Outcome()
	if !done || o.Button != RemoveButton || o.Value != "b" {
		t.Fatalf("outcome = %+v %v", o, done)
	}
	if !on["b"] {
		t.Fatal("removing is not toggling")
	}
}

func TestAddPackEndsTheDialogWithAnAddOutcome(t *testing.T) {
	m, _ := fixture()
	click(m, m.Rect().W-10-1-13+3, m.Rect().H-2)
	if o, done := m.Outcome(); !done || o.Button != AddButton {
		t.Fatalf("outcome = %+v %v", o, done)
	}
}

func TestOneComponentIsNotPlural(t *testing.T) {
	if countText(1) != "1 component" || countText(0) != "0 components" || countText(7) != "7 components" {
		t.Fatal("counts read naturally")
	}
}
