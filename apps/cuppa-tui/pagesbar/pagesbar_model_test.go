package pagesbar

import (
	"slices"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func fixture(t *testing.T) (*Model, *editor.Editor) {
	t.Helper()
	ed := editor.New(standard.Default(), design.NewDocument("App", 80, 24))
	m := New(ed)
	m.Bind(func(title, initial string, then func(string)) { then("Renamed") }, func(title, body string) { t.Errorf("%s: %s", title, body) })
	return m, ed
}

func press(m *Model, line string, text string) bool {
	x := strings.Index(ansi.Strip(line), text)
	return m.HandleBar(pointer.Event{X: len([]rune(ansi.Strip(line)[:x])), Y: 0, Phase: pointer.Down, Left: true})
}

func TestTheBarAddsAPageAndWalksThem(t *testing.T) {
	m, ed := fixture(t)
	if !press(m, m.Line(60, false), "[+ Page]") {
		t.Fatal("the add button did nothing")
	}
	if got := ed.PageNames(); len(got) != 2 || ed.ActivePage() != 1 {
		t.Fatalf("pages %v active %d", got, ed.ActivePage())
	}
	if !press(m, m.Line(60, false), "◂") || ed.ActivePage() != 0 {
		t.Fatalf("◂ did not go back: %d", ed.ActivePage())
	}
	if !press(m, m.Line(60, false), "▸") || ed.ActivePage() != 1 {
		t.Fatalf("▸ did not go on: %d", ed.ActivePage())
	}
	if line := ansi.Strip(m.Line(60, false)); !strings.Contains(line, "Page 2  2/2") {
		t.Errorf("line = %q", line)
	}
}

func TestTheListSwitchesRenamesMovesAndDeletes(t *testing.T) {
	m, ed := fixture(t)
	if err := ed.AddPage("Settings"); err != nil {
		t.Fatal(err)
	}
	ed.SwitchPage(0)
	line := m.Line(60, false)
	if !press(m, line, "1/2") || !m.Open() {
		t.Fatal("the name should open the list")
	}
	x, y, lines := m.Overlay(0, 20)
	if len(lines) == 0 || y >= 20 {
		t.Fatalf("list at %d,%d with %d lines", x, y, len(lines))
	}
	row := func(label string) int {
		for i, l := range lines {
			if strings.Contains(ansi.Strip(l), label) {
				return y + i
			}
		}
		t.Fatalf("no row %q in\n%s", label, ansi.Strip(strings.Join(lines, "\n")))
		return 0
	}
	// Choose Settings from the list.
	m.HandleMenu(pointer.Event{X: x + 4, Y: row("Settings"), Phase: pointer.Down, Left: true})
	if m.Open() || ed.ActivePage() != 1 {
		t.Fatalf("open %v active %d", m.Open(), ed.ActivePage())
	}
	// Rename it through the asker.
	press(m, m.Line(60, false), "2/2")
	x, y, lines = m.Overlay(0, 20)
	m.HandleMenu(pointer.Event{X: x + 4, Y: row("Rename"), Phase: pointer.Down, Left: true})
	if got := ed.PageNames(); !slices.Equal(got, []string{"Page 1", "Renamed"}) {
		t.Fatalf("pages = %v", got)
	}
	// Move it left, then delete it, with the keyboard.
	press(m, m.Line(60, false), "2/2")
	m.Overlay(0, 20)
	for _, k := range []string{"down", "down", "down", "enter"} {
		m.Key(k)
	}
	if got := ed.PageNames(); !slices.Equal(got, []string{"Renamed", "Page 1"}) {
		t.Fatalf("after move left: %v", got)
	}
	press(m, m.Line(60, false), "1/2")
	m.Overlay(0, 20)
	for i := 0; i < 6; i++ {
		m.Key("down")
	}
	m.Key("enter")
	if ed.Document().IsPaged() {
		t.Fatalf("after delete: %v", ed.PageNames())
	}
}

func TestAnIdleBarDoesNothing(t *testing.T) {
	m, ed := fixture(t)
	line := m.Line(60, true)
	if press(m, line, "[+ Page]") || len(ed.PageNames()) != 1 {
		t.Fatal("the bar is idle while previewing")
	}
}

// Pages with a folder are listed under it, in the order the folders first
// appear, and a row still switches to its page.
func TestThePagesAreListedUnderTheirFolders(t *testing.T) {
	m, ed := fixture(t)
	for _, name := range []string{"Colours", "Download", "Songs"} {
		if err := ed.AddPage(name); err != nil {
			t.Fatal(err)
		}
	}
	// Page 1, Colours, Download, Songs
	for i, folder := range map[int]string{1: "setup/colours", 2: "run/download", 3: "setup/songs"} {
		if err := ed.SetPageFolder(i, folder); err != nil {
			t.Fatal(err)
		}
	}
	ed.SwitchPage(0)
	m.Line(60, false)
	m.toggle()
	var rows []string
	for _, it := range m.items {
		if it.separator && it.heading != "" {
			rows = append(rows, it.heading)
		} else if !it.separator && it.label != "" && len(rows) > 0 && len(rows) < 20 {
			rows = append(rows, it.label)
		}
		if it.label == "Add page" {
			break
		}
	}
	want := []string{"pages/", "Page 1", "setup/", "Colours", "Songs", "run/", "Download"}
	if !slices.Equal(rows[:len(want)], want) {
		t.Fatalf("rows = %v", rows)
	}
	for i, it := range m.items {
		if it.label == "Songs" {
			m.choose(i)
		}
	}
	if ed.ActivePage() != 3 {
		t.Fatalf("active page = %d", ed.ActivePage())
	}
}
