package inspector

import (
	"strings"
	"testing"
)

// The Change row offers components that can take the place of the selected one
// without losing its variables, and pressing a name changes the component.
func TestChangeRowChangesTheComponentAndKeepsTheVariables(t *testing.T) {
	m, ed, id := setup(t, "bubbles.textinput")
	if err := ed.SetBinding(id, "value", "Colour"); err != nil {
		t.Fatal(err)
	}
	m.Lines()
	plain := stripANSI(strings.Join(m.Lines(), "\n"))
	if !strings.Contains(plain, "Change") || !strings.Contains(plain, "keeps its variables") {
		t.Fatalf("no Change row:\n%s", plain)
	}

	// The list opens under the button and a name in it changes the component.
	clickText(t, m, "Change ", "Replace with")
	if !strings.Contains(stripANSI(strings.Join(m.Lines(), "\n")), "Colour picker") {
		t.Fatalf("the colour picker is not offered for a text input:\n%s", strings.Join(m.Lines(), "\n"))
	}
	clickText(t, m, "Colour picker", "Colour picker")

	n, _ := ed.Document().Get(id)
	if n.Component != "lipgloss.colourpicker" || n.Bind["value"] != "Colour" {
		t.Fatalf("component %q, bind %v (message %q)", n.Component, n.Bind, m.message)
	}
	if !ed.CanUndo() {
		t.Error("the change should be undoable")
	}
	ed.Undo()
	if n, _ = ed.Document().Get(id); n.Component != "bubbles.textinput" || n.Bind["value"] != "Colour" {
		t.Errorf("undo: %q %v", n.Component, n.Bind)
	}
}

func TestChangeRowIsAbsentWhenNothingCanTakeThePlace(t *testing.T) {
	m, _, _ := setup(t, "lipgloss.box")
	if plain := stripANSI(strings.Join(m.Lines(), "\n")); strings.Contains(plain, "keeps its variables") {
		t.Errorf("a box has no compatible component, but:\n%s", plain)
	}
}
