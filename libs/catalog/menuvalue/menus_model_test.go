package menuvalue

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func names(m Menus, menu int) []string {
	var out []string
	for _, e := range m[menu].Entries {
		if e.Separator {
			out = append(out, "-")
		} else {
			out = append(out, e.Name)
		}
	}
	return out
}

func sample() Menus {
	m, err := Parse(`[{"name":"File","entries":[{"name":"New","shortcut":"Ctrl+N"},{"separator":true},{"name":"Quit"}]},{"name":"Edit","entries":[{"name":"Undo"}]},{"name":"Help","entries":[]}]`)
	if err != nil {
		panic(err)
	}
	return m
}

func TestParseAndEncodeRoundTrip(t *testing.T) {
	m := sample()
	again, err := Parse(m.Encode())
	if err != nil || !reflect.DeepEqual(m, again) {
		t.Fatalf("%v %v", again, err)
	}
	if !strings.Contains(m.Encode(), `"entries":[]`) {
		t.Errorf("a menu with no entries is written as []: %s", m.Encode())
	}
	if got, _ := Normalise(""); got != "[]" {
		t.Errorf("empty = %q", got)
	}
}

func TestParseRefusesNamelessMenusAndEntries(t *testing.T) {
	for _, text := range []string{`[{"name":" ","entries":[]}]`, `[{"name":"A","entries":[{"shortcut":"x"}]}]`, `{"a":1}`, `nope`} {
		if _, err := Parse(text); err == nil {
			t.Errorf("Parse(%s) accepted it", text)
		}
	}
}

func TestEditsDoNotChangeTheirReceiver(t *testing.T) {
	m := sample()
	before := m.Encode()
	m.AddMenu("View")
	m.RemoveMenu(0)
	m.MoveMenu(0, 2)
	m.AddSeparator(1)
	m.RemoveEntry(0, 0)
	m.MoveEntry(0, 0, 1, 0)
	if m.Encode() != before {
		t.Fatalf("the value changed:\n%s\n%s", before, m.Encode())
	}
}

func TestMenusAreAddedRenamedMovedAndRemoved(t *testing.T) {
	m := sample().AddMenu("  ")
	if len(m) != 4 || m[3].Name != "Menu" {
		t.Fatalf("menus = %v", m)
	}
	m, err := m.RenameMenu(3, "View")
	if err != nil || m[3].Name != "View" {
		t.Fatalf("rename: %v %v", m, err)
	}
	if _, err := m.RenameMenu(3, "  "); !errors.Is(err, ErrNoName) {
		t.Errorf("an empty name was accepted: %v", err)
	}
	m = m.MoveMenu(3, 0)
	if m[0].Name != "View" || m[1].Name != "File" {
		t.Errorf("order after move = %s %s", m[0].Name, m[1].Name)
	}
	m = m.RemoveMenu(0)
	if len(m) != 3 || m[0].Name != "File" {
		t.Errorf("after remove: %v", m)
	}
}

func TestEntriesAreAddedEditedAndRemoved(t *testing.T) {
	m := sample().AddEntry(2, Entry{Name: "About"}).AddSeparator(2)
	if got := names(m, 2); !reflect.DeepEqual(got, []string{"About", "-"}) {
		t.Fatalf("entries = %v", got)
	}
	m, err := m.SetEntry(2, 0, " About Cuppa ", " F1 ")
	if err != nil || m[2].Entries[0] != (Entry{Name: "About Cuppa", Shortcut: "F1"}) {
		t.Fatalf("set: %+v %v", m[2].Entries[0], err)
	}
	if _, err := m.SetEntry(2, 0, "", "F1"); !errors.Is(err, ErrNoName) {
		t.Errorf("an entry with no name was accepted")
	}
	same, err := m.SetEntry(2, 1, "x", "y") // a separator has no name or shortcut
	if err != nil || !reflect.DeepEqual(same, m) {
		t.Errorf("a separator was edited")
	}
	m = m.RemoveEntry(2, 1)
	if got := names(m, 2); !reflect.DeepEqual(got, []string{"About Cuppa"}) {
		t.Errorf("after remove = %v", got)
	}
}

func TestAnEntryMovesWithinAMenuAndToAnother(t *testing.T) {
	m := sample().MoveEntry(0, 2, 0, 0) // Quit to the top of File
	if got := names(m, 0); !reflect.DeepEqual(got, []string{"Quit", "New", "-"}) {
		t.Fatalf("File = %v", got)
	}
	m = m.MoveEntry(0, 0, 1, 1) // Quit to the end of Edit
	if got := names(m, 1); !reflect.DeepEqual(got, []string{"Undo", "Quit"}) {
		t.Fatalf("Edit = %v", got)
	}
	m = m.MoveEntry(1, 0, 2, 5) // past the end of the empty Help menu
	if got := names(m, 2); !reflect.DeepEqual(got, []string{"Undo"}) {
		t.Fatalf("Help = %v", got)
	}
}

func TestNudgeCrossesIntoTheNeighbouringMenu(t *testing.T) {
	m := sample()
	m2, menu, at := m.NudgeEntry(0, 0, -1)
	if menu != 0 || at != 0 || !reflect.DeepEqual(m2, m) {
		t.Errorf("the first entry of the first menu has nowhere to go: %d %d", menu, at)
	}
	m2, menu, at = m.NudgeEntry(0, 2, 1) // Quit down: first entry of Edit
	if menu != 1 || at != 0 || !reflect.DeepEqual(names(m2, 1), []string{"Quit", "Undo"}) {
		t.Errorf("down: %d %d %v", menu, at, names(m2, 1))
	}
	m2, menu, at = m2.NudgeEntry(1, 0, -1) // and back up: the end of File
	if menu != 0 || at != 2 || !reflect.DeepEqual(names(m2, 0), []string{"New", "-", "Quit"}) {
		t.Errorf("up: %d %d %v", menu, at, names(m2, 0))
	}
	m2, menu, at = m.NudgeEntry(0, 0, 1)
	if menu != 0 || at != 1 || !reflect.DeepEqual(names(m2, 0), []string{"-", "New", "Quit"}) {
		t.Errorf("within: %d %d %v", menu, at, names(m2, 0))
	}
}
