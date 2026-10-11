package inspector

import (
	"testing"

	"github.com/meta-tui/cuppa/libs/document/design"
)

func TestScreenContractRowsEditThroughTheEditor(t *testing.T) {
	m, ed, id := setup(t, "lipgloss.label")

	// A property is bound by opening the list under it and typing a new name.
	clickText(t, m, "⇄", "input…")
	clickText(t, m, "New input", "New input")
	for _, r := range "Title" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	n, _ := ed.Document().Get(id)
	if len(n.Bind) != 1 {
		t.Fatalf("bind = %v (message %q)", n.Bind, m.message)
	}
	var key string
	for k, v := range n.Bind {
		key = k
		if v != "Title" {
			t.Errorf("bound to %q", v)
		}
	}

	// Show if and On click are typed in their rows.
	clickText(t, m, "Show if", "(empty)")
	for _, r := range "Visible" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	clickText(t, m, "On click", "(empty)")
	for _, r := range "Open" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	n, _ = ed.Document().Get(id)
	if n.ShowIf != "Visible" || n.Event != "Open" {
		t.Fatalf("show-if %q, event %q", n.ShowIf, n.Event)
	}

	// An empty name unbinds.
	clickText(t, m, "⇄", "Title")
	clickText(t, m, "(none)", "(none)")
	if n, _ = ed.Document().Get(id); len(n.Bind) != 0 {
		t.Errorf("binding for %q kept: %v", key, n.Bind)
	}
}

func TestScreenKeysAreTypedWhenNothingIsSelected(t *testing.T) {
	m, ed, _ := setup(t, "lipgloss.box")
	ed.Select()
	m.Lines()
	clickText(t, m, "(empty)", "(empty)")
	for _, r := range "s=Save:save, esc=Back" {
		m.Key(string(r), false, false, false)
	}
	m.Key("", false, true, false)
	keys := ed.Document().Keys
	if len(keys) != 2 || keys[0] != (design.KeyBinding{Key: "s", Event: "Save", Label: "save"}) {
		t.Fatalf("keys = %+v (message %q)", keys, m.message)
	}
	m.Lines()
	clickText(t, m, "s=Save:save", "s=Save")
	m.Key("x", false, false, false)
	m.Key("", false, true, false)
	// Editing starts from the current keys, so the typed letter is appended.
	if got := ed.Document().Keys[1].Event; got != "Backx" {
		t.Errorf("event = %q", got)
	}
}
