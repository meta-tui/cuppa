package inspector

import (
	"reflect"
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/canvas/editor"
	"github.com/meta-ui-labs/cuppa/libs/catalog/menuvalue"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// menuNames reads the stored menus as "Menu: entry, entry" lines, a separator
// as "-".
func menuNames(t *testing.T, ed *editor.Editor, id design.NodeID) []string {
	t.Helper()
	n, _ := ed.Document().Get(id)
	value, ok := n.Props["menus"]
	if !ok {
		def, _ := standardMenus()
		value = def
	}
	menus, err := menuvalue.Parse(value)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, mn := range menus {
		var entries []string
		for _, e := range mn.Entries {
			switch {
			case e.Separator:
				entries = append(entries, "-")
			case e.Shortcut != "":
				entries = append(entries, e.Name+"("+e.Shortcut+")")
			default:
				entries = append(entries, e.Name)
			}
		}
		out = append(out, mn.Name+": "+strings.Join(entries, ", "))
	}
	return out
}

func standardMenus() (string, bool) {
	m, ok := standard.Default().Get("cuppa.menubar")
	if !ok {
		return "", false
	}
	p, ok := m.Prop("menus")
	return p.Default, ok
}

// gripOf is the position of the grip on the row that shows text.
func gripOf(t *testing.T, m *Model, text string) (x, y int) {
	t.Helper()
	for row, l := range m.Lines() {
		plain := stripANSI(l)
		if strings.Contains(plain, text) {
			if i := strings.Index(plain, "≡"); i >= 0 {
				return len([]rune(plain[:i])), row
			}
		}
	}
	t.Fatalf("no row with a grip shows %q:\n%s", text, strings.Join(m.Lines(), "\n"))
	return 0, 0
}

func typeText(m *Model, text string) {
	for _, r := range text {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	m.Lines()
}

func TestMenuPropertyIsListedWithItsEntriesAndAddButtons(t *testing.T) {
	m, _, _ := setup(t, "cuppa.menubar")
	view := stripANSI(strings.Join(m.Lines(), "\n"))
	for _, want := range []string{"File", "Edit", "Help", "New", "Ctrl+N", "Quit", "[+ Entry]", "[+ Menu]", "≡"} {
		if !strings.Contains(view, want) {
			t.Errorf("the list lacks %q:\n%s", want, view)
		}
	}
}

func TestAddingAMenuAsksForItsName(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	clickText(t, m, "[+ Menu]", "[+ Menu]")
	if !m.Editing() {
		t.Fatal("a new menu should be waiting for its name")
	}
	typeText(m, "View")
	got := menuNames(t, ed, id)
	if len(got) != 4 || got[3] != "View: " {
		t.Errorf("menus = %q", got)
	}
}

func TestAddingAnEntryAndASeparator(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	// The first [+ Entry] belongs to File.
	clickText(t, m, "[+ Entry]", "[+ Entry]")
	typeText(m, "Close")
	clickText(t, m, "[+ Entry]", "[+ Separator]")
	got := menuNames(t, ed, id)[0]
	if want := "File: New(Ctrl+N), Open…(Ctrl+O), -, Quit(Ctrl+Q), Close, -"; got != want {
		t.Errorf("File = %q, want %q", got, want)
	}
}

func TestAnEntryNeedsAName(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	before := menuNames(t, ed, id)
	x, y := find(t, m, "Undo")
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Down, Left: true})
	m.Lines()
	for range "Undo" {
		m.Key("", true, false, false)
	}
	m.Key("", false, true, false)
	m.Lines()
	if !reflect.DeepEqual(menuNames(t, ed, id), before) {
		t.Errorf("an entry lost its name: %q", menuNames(t, ed, id))
	}
	if !strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "A name is needed") {
		t.Error("the person is not told a name is needed")
	}
}

func TestAShortcutIsOptionalAndEditable(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	x, y := find(t, m, "Ctrl+Q")
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Down, Left: true})
	m.Lines()
	for range "Ctrl+Q" {
		m.Key("", true, false, false)
	}
	typeText(m, "Ctrl+W")
	if got := menuNames(t, ed, id)[0]; !strings.Contains(got, "Quit(Ctrl+W)") {
		t.Errorf("File = %q", got)
	}
	x, y = find(t, m, "Ctrl+W")
	m.Handle(pointer.Event{X: x, Y: y, Phase: pointer.Down, Left: true})
	m.Lines()
	for range "Ctrl+W" {
		m.Key("", true, false, false)
	}
	m.Key("", false, true, false)
	if got := menuNames(t, ed, id)[0]; !strings.HasSuffix(got, "Quit") {
		t.Errorf("an emptied shortcut is removed: %q", got)
	}
}

func TestDraggingAnEntryToAnotherMenu(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	gx, gy := gripOf(t, m, "Quit")
	_, ty := find(t, m, "Undo")
	press(m, gx, gy)
	dragTo(m, gx, ty)
	if !m.Dragging() {
		t.Fatal("the grip was not taken")
	}
	if !strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "▸") {
		t.Error("the row it would land on is not marked")
	}
	release(m, gx, ty)
	got := menuNames(t, ed, id)
	if got[0] != "File: New(Ctrl+N), Open…(Ctrl+O), -" || !strings.HasPrefix(got[1], "Edit: Quit(Ctrl+Q), Undo") {
		t.Errorf("menus = %q", got)
	}
}

func TestDraggingAMenuReordersTheMenus(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	gx, gy := gripOf(t, m, "Help")
	_, ty := find(t, m, "File")
	press(m, gx, gy)
	dragTo(m, gx, ty)
	release(m, gx, ty)
	got := menuNames(t, ed, id)
	if !strings.HasPrefix(got[0], "Help") || !strings.HasPrefix(got[1], "File") || !strings.HasPrefix(got[2], "Edit") {
		t.Errorf("order = %q", got)
	}
}

func TestAPressOnAGripThatDoesNotMoveChangesNothing(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	before := menuNames(t, ed, id)
	gx, gy := gripOf(t, m, "Quit")
	press(m, gx, gy)
	release(m, gx, gy)
	if !reflect.DeepEqual(menuNames(t, ed, id), before) || m.Dragging() {
		t.Error("a click on a grip changed the menus")
	}
}

func TestTheKeysMoveAnEntryAcrossTheEdgeOfItsMenu(t *testing.T) {
	m, ed, id := setup(t, "cuppa.menubar")
	ref := menuRef{node: id, key: "menus", menu: 0, entry: 3} // Quit, the last of File
	if !m.menuKey("alt+down", ref) {
		t.Fatal("the key was not used")
	}
	got := menuNames(t, ed, id)
	if got[0] != "File: New(Ctrl+N), Open…(Ctrl+O), -" || !strings.HasPrefix(got[1], "Edit: Quit(Ctrl+Q), Undo") {
		t.Errorf("menus = %q", got)
	}
	m.menuKey("delete", menuRef{node: id, key: "menus", menu: 1, entry: 0})
	if got = menuNames(t, ed, id); strings.Contains(got[1], "Quit") {
		t.Errorf("delete did not remove the entry: %q", got)
	}
}

func TestThePaneStaysTheSizeItWasGiven(t *testing.T) {
	m, _, _ := setup(t, "cuppa.menubar")
	lines := m.Lines()
	if len(lines) != 200 {
		t.Fatalf("%d lines", len(lines))
	}
	for i, l := range lines {
		if w := len([]rune(stripANSI(l))); w != 32 {
			t.Errorf("line %d is %d cells wide: %q", i, w, stripANSI(l))
			break
		}
	}
}
