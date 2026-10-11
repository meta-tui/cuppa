package editor

import (
	"strings"
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func optionIDs(e *Editor, id design.NodeID) []string {
	var out []string
	for _, d := range e.ChangeOptions(id) {
		out = append(out, d.ID)
	}
	return out
}

func TestATextInputBecomesAColourPickerAndKeepsItsVariables(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "bubbles.textinput", 2, 2)
	e.Rename(id, "Colour field")
	must(t, e.SetProp(id, "value", "#102030"))
	must(t, e.SetBinding(id, "value", "Colour"))
	must(t, e.SetShowIf(id, "Show colour"))
	must(t, e.SetEvent(id, "Pick colour"))
	before, _ := e.Document().Get(id)

	report, err := e.ChangeComponent(id, "lipgloss.colourpicker", false)
	if err != nil {
		t.Fatal(err)
	}
	if !report.Compatible() || len(report.Moves) != 1 || report.Moves[0].From != "value" || report.Moves[0].To != "value" {
		t.Fatalf("report %+v", report)
	}
	n, _ := e.Document().Get(id)
	if n.Component != "lipgloss.colourpicker" || n.Name != before.Name {
		t.Errorf("component %q name %q", n.Component, n.Name)
	}
	if n.Bind["value"] != "Colour" || n.ShowIf != "Show colour" || n.Event != "Pick colour" {
		t.Errorf("variables: bind %v showif %q event %q", n.Bind, n.ShowIf, n.Event)
	}
	if n.Props["value"] != "#102030" {
		t.Errorf("the colour is kept: %v", n.Props)
	}
	if n.Rect.W < 46 || n.Rect.X != before.Rect.X || n.Rect.Y != before.Rect.Y {
		t.Errorf("rect %+v (was %+v): grows to the picker's smallest size, keeps its place", n.Rect, before.Rect)
	}

	// One undo step puts everything back.
	e.Undo()
	n, _ = e.Document().Get(id)
	if n.Component != "bubbles.textinput" || n.Bind["value"] != "Colour" || n.Props["value"] != "#102030" || n.Rect != before.Rect {
		t.Errorf("undo: %+v", n)
	}

	// The same variables are listed after the change.
	names := func() string {
		var out []string
		for _, v := range e.Variables() {
			out = append(out, string(v.Kind)+":"+v.Name)
		}
		return strings.Join(out, ",")
	}
	want := names()
	if _, err := e.ChangeComponent(id, "lipgloss.colourpicker", false); err != nil {
		t.Fatal(err)
	}
	if got := names(); got != want {
		t.Errorf("variables after the change %q, before %q", got, want)
	}
}

func TestAValueThatIsNotValidForTheNewComponentIsLeftOut(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "bubbles.textinput", 0, 0)
	must(t, e.SetProp(id, "value", "hello"))
	must(t, e.SetBinding(id, "value", "Name"))
	if _, err := e.ChangeComponent(id, "lipgloss.colourpicker", false); err != nil {
		t.Fatal(err)
	}
	n, _ := e.Document().Get(id)
	if _, set := n.Props["value"]; set {
		t.Errorf("hello is not a colour, but %q was carried", n.Props["value"])
	}
	if n.Bind["value"] != "Name" {
		t.Errorf("the binding is kept: %v", n.Bind)
	}
}

func TestTabsAndDialogButtonsTakeEachOthersPlace(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.tabs", 0, 0)
	must(t, e.SetProp(id, "tabs", "Home,Settings,About"))
	must(t, e.SetProp(id, "active", "2"))
	must(t, e.SetBinding(id, "tabs", "Sections"))
	must(t, e.SetBinding(id, "active", "Section"))

	found := false
	for _, c := range optionIDs(e, id) {
		if c == "community.dialog" {
			found = true
		}
	}
	if !found {
		t.Fatalf("dialog is not offered for tabs: %v", optionIDs(e, id))
	}
	if _, err := e.ChangeComponent(id, "community.dialog", false); err != nil {
		t.Fatal(err)
	}
	n, _ := e.Document().Get(id)
	if n.Bind["buttons"] != "Sections" || n.Bind["active"] != "Section" {
		t.Errorf("bind %v", n.Bind)
	}
	if n.Props["buttons"] != "Home,Settings,About" || n.Props["active"] != "2" {
		t.Errorf("props %v", n.Props)
	}
}

func TestAChangeThatWouldLoseAVariableIsRefusedUnlessAllowed(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.list", 0, 0)
	must(t, e.SetBinding(id, "items", "Teas"))
	if _, err := e.ChangeComponent(id, "lipgloss.label", false); err == nil || !strings.Contains(err.Error(), "Teas") {
		t.Fatalf("want a refusal naming the variable, got %v", err)
	}
	n, _ := e.Document().Get(id)
	if n.Component != "lipgloss.list" || n.Bind["items"] != "Teas" {
		t.Errorf("a refused change must change nothing: %+v", n)
	}
	for _, c := range optionIDs(e, id) {
		if c == "lipgloss.label" {
			t.Error("label is offered for a list with a bound list")
		}
	}
	report, err := e.ChangeComponent(id, "lipgloss.label", true)
	if err != nil || report.Compatible() {
		t.Fatalf("allowed: %v %+v", err, report)
	}
	n, _ = e.Document().Get(id)
	if n.Component != "lipgloss.label" || len(n.Bind) != 0 {
		t.Errorf("allowed change: %+v", n)
	}
}

func TestOnlyPlainComponentsCanBeChanged(t *testing.T) {
	e := newEditor()
	id := mustAdd(t, e, "lipgloss.label", 0, 0)
	if _, err := e.ChangeComponent(id, "no.such.component", false); err == nil {
		t.Error("an unknown component was accepted")
	}
	if !e.SetLocked(id, true) {
		t.Fatal("could not lock")
	}
	if got := e.ChangeOptions(id); len(got) != 0 {
		t.Errorf("a locked component has options: %d", len(got))
	}
	if _, err := e.ChangeComponent(id, "bubbles.textinput", false); err == nil {
		t.Error("a locked component was changed")
	}
}
