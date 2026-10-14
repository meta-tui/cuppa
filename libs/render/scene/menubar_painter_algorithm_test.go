package scene

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/render/grid"
)

func menubarGrid(t *testing.T, open string, w, h int) []string {
	t.Helper()
	def, ok := standard.Default().Get("cuppa.menubar")
	if !ok {
		t.Fatal("no menu bar in the catalog")
	}
	props := Props{}
	for _, p := range def.Props {
		props[p.Key] = p.Default
	}
	props["open"] = open
	g := grid.New(w, h)
	paintMenuBar(g, props)
	lines := g.Lines()
	for i := range lines {
		lines[i] = ansi.Strip(lines[i])
	}
	return lines
}

func TestMenuBarDrawsTheNamesOnTheFirstRow(t *testing.T) {
	lines := menubarGrid(t, "-1", 48, 10)
	for _, want := range []string{"File", "Edit", "Help"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("bar lacks %q: %q", want, lines[0])
		}
	}
	if strings.Contains(strings.Join(lines[1:], ""), "New") {
		t.Error("a closed bar drew a box")
	}
}

func TestMenuBarDrawsTheOpenMenusBox(t *testing.T) {
	lines := menubarGrid(t, "0", 48, 10)
	box := strings.Join(lines[1:], "\n")
	for _, want := range []string{"╭", "New", "Ctrl+N", "Open…", "Quit", "Ctrl+Q", "╰"} {
		if !strings.Contains(box, want) {
			t.Errorf("box lacks %q:\n%s", want, box)
		}
	}
	// New, Open, a separator, Quit: the separator is a rule inside the box.
	if !strings.Contains(lines[4], "─") || !strings.HasPrefix(strings.TrimSpace(lines[4]), "│") {
		t.Errorf("row 4 is not a separator: %q", lines[4])
	}
}

func TestMenuBarKeepsTheBoxInsideAndSurvivesABadValue(t *testing.T) {
	lines := menubarGrid(t, "2", 20, 4) // Help is last: the box is pulled left to fit
	if !strings.Contains(strings.Join(lines, "\n"), "About") {
		t.Errorf("box missing:\n%s", strings.Join(lines, "\n"))
	}
	g := grid.New(30, 3)
	paintMenuBar(g, Props{"menus": "not json"})
	if !strings.Contains(g.String(), "Menu bar") {
		t.Errorf("a value that does not parse is drawn as a labelled frame:\n%s", g.String())
	}
}
