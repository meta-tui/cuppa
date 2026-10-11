package themepicker

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/color/scheme"
)

func open() *Model {
	m := New("Theme from a colour scheme", "Dracula")
	m.Place(120, 40)
	m.Lines()
	return m
}

func press(m *Model, x, y int) {
	m.Handle(pointer.Event{X: m.rect.X + x, Y: m.rect.Y + y, Phase: pointer.Down, Left: true})
	m.Lines()
}

func TestOpensOnTheNamedSchemeAndShowsItsFiveColours(t *testing.T) {
	m := open()
	if got := m.current().Name; got != "Dracula" {
		t.Fatalf("scheme = %q", got)
	}
	text := ansi.Strip(strings.Join(m.Lines(), "\n"))
	for _, want := range []string{"Dracula", "Background", "Text", "Muted", "Border", "Secondary", "[ Apply ]", "Sample"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q:\n%s", want, text)
		}
	}
	lines := m.Lines()
	if len(lines) != m.Rect().H {
		t.Fatalf("%d lines in a rect %d tall", len(lines), m.Rect().H)
	}
	for i, l := range lines {
		if w := ansi.StringWidth(l); w != m.Rect().W {
			t.Errorf("line %d is %d wide, box %d", i, w, m.Rect().W)
		}
	}
}

func TestStepJumpAndWheelChangeTheScheme(t *testing.T) {
	m := open()
	start := m.index
	m.step(1)
	if m.index != start+1 {
		t.Fatalf("step: %d", m.index)
	}
	m.jump('N')
	if name := m.current().Name; !strings.HasPrefix(strings.ToUpper(name), "N") {
		t.Fatalf("jump N landed on %q", name)
	}
	before := m.index
	m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: 1, X: m.rect.X + 5, Y: m.rect.Y + 5})
	if m.index != before+1 {
		t.Fatalf("wheel: %d", m.index)
	}
	// The buttons on the name row are clickable.
	for _, g := range m.regions {
		if g.y == rowName {
			before = m.index
			press(m, g.x0, g.y)
			if m.index == before {
				t.Error("a step button should change the scheme")
			}
			break
		}
	}
}

func TestApplyCarriesTheBackgroundAndTheFourThemeColours(t *testing.T) {
	m := open()
	m.Key("", false, true, false)
	o, done := m.Outcome()
	if !done || o.Canceled {
		t.Fatal("Enter applies")
	}
	background, th, ok := Parse(o.Value)
	if !ok {
		t.Fatalf("value %q", o.Value)
	}
	want := scheme.All()[m.index].Theme()
	if background != want.Background.Hex() || th.Text != want.Text.Hex() || th.Border != want.Border.Hex() ||
		th.Muted != want.Muted.Hex() || th.Secondary != want.Secondary.Hex() {
		t.Fatalf("applied %q %+v, want %+v", background, th, want)
	}
}

func TestEscCancels(t *testing.T) {
	m := open()
	m.Key("", false, false, true)
	if o, done := m.Outcome(); !done || !o.Canceled {
		t.Fatal("Esc cancels")
	}
}

func TestParseRefusesAWrongValue(t *testing.T) {
	if _, _, ok := Parse("#000000|#111111"); ok {
		t.Fatal("five colours are needed")
	}
}
