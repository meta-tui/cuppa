package shell

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/a11y"
	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/menubar"
)

// flatten lists "role: label = value" for every node, depth first.
func flatten(nodes []a11y.Node) []string {
	var out []string
	for _, n := range nodes {
		line := n.Role + ": " + n.Label
		if n.Value != "" {
			line += " = " + n.Value
		}
		if n.Selected {
			line += " [selected]"
		}
		out = append(out, line)
		out = append(out, flatten(n.Children)...)
	}
	return out
}

func described(m *Model) string { return strings.Join(flatten(m.Describe().Nodes), "\n") }

func TestTheScreenIsDescribedInReadingOrder(t *testing.T) {
	m := newShell(t)
	text := described(m)
	order := []string{"list: Menus", "heading: Components", "listitem: Box, category", "heading: Canvas", "heading: Details", "textbox: Width = 120", "button: Shadows = off"}
	last := -1
	for _, want := range order {
		i := strings.Index(text, want)
		if i < 0 || i < last {
			t.Fatalf("%q missing or out of order:\n%s", want, text)
		}
		last = i
	}
	if title := m.Describe().Title; !strings.Contains(title, "Untitled") {
		t.Fatalf("title = %q", title)
	}
}

func TestTheSelectionAndItsSettingsAreDescribed(t *testing.T) {
	m := newShell(t)
	id, _ := m.ed.Add("lipgloss.label", 5, 6)
	if err := m.ed.SetProp(id, "text", "Matcha"); err != nil {
		t.Fatal(err)
	}
	text := described(m)
	for _, want := range []string{"textbox: Name = Label 1", "textbox: X = 5", "textbox: Y = 6", "textbox: Text = Matcha", "[selected]", "Layers, front to back", "Label 1, lipgloss.label, at 5,6"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q:\n%s", want, text)
		}
	}
	if !strings.Contains(m.Describe().Title, "unsaved") {
		t.Fatal("unsaved work is part of the title")
	}
}

func TestHiddenLockedAndGroupedStateIsSaid(t *testing.T) {
	m := newShell(t)
	a, _ := m.ed.Add("lipgloss.box", 1, 1)
	b, _ := m.ed.Add("lipgloss.box", 40, 1)
	m.ed.SetHidden(a, true)
	m.ed.SetLocked(b, true)
	m.ed.Clear()
	text := described(m)
	if !strings.Contains(text, ", hidden") || !strings.Contains(text, ", locked") {
		t.Fatalf("layer state should be spoken:\n%s", text)
	}
	m.ed.Select(b)
	if !strings.Contains(described(m), "Locked: unlock it in Layers to edit") {
		t.Fatal("a locked selection says why edits do nothing")
	}
}

func TestAnOpenMenuAndAnOpenDialogTakeOverTheDescription(t *testing.T) {
	m := newShell(t)
	send(m, click(14, 0)) // the File label
	if !strings.Contains(described(m), "File menu items") || !strings.Contains(described(m), "Save, Ctrl+S") {
		t.Fatalf("an open menu lists its items:\n%s", described(m))
	}
	m.bar.Close()

	m.perform(menubar.EditPacks)
	text := described(m)
	if !strings.Contains(text, "heading: Component packs") || strings.Contains(text, "heading: Components\n") || strings.Contains(text, "Details") {
		t.Fatalf("a dialog is the whole screen:\n%s", text)
	}
	if !strings.Contains(text, "Lip Gloss, on") {
		t.Fatalf("each pack says whether it is on:\n%s", text)
	}
}

func TestStatusFeedbackIsExposedForAnnouncing(t *testing.T) {
	m := newShell(t)
	m.flow.Notice("Hello", "x")
	m.flow.Modal().Key("", false, true, false)
	m.flow.Resolve()
	m.flow.NewDesign()
	if m.Describe().Status == "" {
		t.Fatal("New design leaves a status to announce")
	}
}
