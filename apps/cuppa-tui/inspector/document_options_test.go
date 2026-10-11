package inspector

import (
	"strings"
	"testing"

	"github.com/meta-tui/cuppa/libs/canvas/editor"
	"github.com/meta-tui/cuppa/libs/catalog/standard"
	"github.com/meta-tui/cuppa/libs/document/design"
)

func emptyFixture(t *testing.T) (*Model, *editor.Editor) {
	t.Helper()
	cat := standard.Default()
	ed := editor.New(cat, design.NewDocument("t", 100, 40))
	m := New(ed, cat)
	m.SetSize(34, 200)
	expandAll(m)
	m.Lines()
	return m, ed
}

func click(m *Model, x, y int) {
	press(m, x, y)
	release(m, x, y)
}

func TestNothingSelectedShowsTheDocumentOptions(t *testing.T) {
	m, _ := emptyFixture(t)
	screen := stripANSI(strings.Join(m.Lines(), "\n"))
	for _, want := range []string{"Canvas", "Background", "terminal default", "Display", "[x] Grid dots", "[ ] Shadows", "[ ] Scanlines", "[ ] Vignette"} {
		if !strings.Contains(screen, want) {
			t.Errorf("missing %q:\n%s", want, screen)
		}
	}
}

func TestEffectCheckboxesToggleTheDocument(t *testing.T) {
	m, ed := emptyFixture(t)
	for _, c := range []struct {
		label string
		on    func() bool
	}{
		{"Shadows", func() bool { return ed.Document().Effects.Shadow }},
		{"Scanlines", func() bool { return ed.Document().Effects.Scanlines }},
		{"Vignette", func() bool { return ed.Document().Effects.Vignette }},
		{"Grid dots", func() bool { return ed.Document().HideGrid }},
	} {
		x, y := find(t, m, c.label)
		click(m, x, y)
		if !c.on() {
			t.Fatalf("clicking %s changes the document", c.label)
		}
	}
}

func TestCanvasSizeCanBeNudgedAndTyped(t *testing.T) {
	m, ed := emptyFixture(t)
	x, y := find(t, m, "[+]") // the width's plus
	click(m, x, y)
	if ed.Document().Width != 101 {
		t.Fatalf("width = %d", ed.Document().Width)
	}
	x, y = find(t, m, " 101")
	click(m, x+1, y)
	for range 3 {
		m.Key("", true, false, false)
	}
	m.Key("90", false, false, false)
	m.Key("", false, true, false)
	if ed.Document().Width != 90 {
		t.Fatalf("typed width = %d", ed.Document().Width)
	}
	if m.message != "" {
		t.Fatalf("unexpected message %q", m.message)
	}
}

func TestBackgroundOpensThePickerAndAppliesTheChoice(t *testing.T) {
	m, ed := emptyFixture(t)
	var title, current string
	var apply func(string)
	m.BindColorPicker(func(t, c string, a func(string)) { title, current, apply = t, c, a })
	x, y := find(t, m, "terminal default")
	click(m, x, y)
	if apply == nil || title != "Background" || current != "" {
		t.Fatalf("picker not opened properly: %q %q", title, current)
	}
	apply("#112233")
	if ed.Document().Background != "#112233" {
		t.Fatalf("background = %q", ed.Document().Background)
	}
	if !strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "#112233") {
		t.Fatal("the value is shown")
	}
}

func TestTerminalPreviewAndColourProfileCanBeSwitched(t *testing.T) {
	m, ed := emptyFixture(t)
	x, y := find(t, m, "[Light]")
	click(m, x, y)
	if !ed.Document().Light {
		t.Fatal("clicking Light previews a light terminal")
	}
	x, y = find(t, m, "[Dark]")
	click(m, x, y)
	if ed.Document().Light {
		t.Fatal("clicking Dark goes back")
	}
	if !strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "true colour") {
		t.Fatal("the profile is shown")
	}
	// The profile is chosen from a list that opens under it.
	clickText(t, m, "Colours", "true colour")
	clickText(t, m, "256 colours", "256 colours")
	if ed.Document().Profile != design.Profile256 {
		t.Fatalf("the list chooses 256 colours: %q", ed.Document().Profile)
	}
	clickText(t, m, "Colours", "256 colours")
	clickText(t, m, "no colour", "no colour")
	if ed.Document().Profile != design.ProfileNone {
		t.Fatalf("the list chooses no colour: %q", ed.Document().Profile)
	}
}
