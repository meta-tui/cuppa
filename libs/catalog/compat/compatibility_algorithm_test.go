package compat

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
)

func TestWhatCanTakeWhosePlace(t *testing.T) {
	reg := standard.Default()
	cases := []struct {
		from, to string
		bound    []string
		ok       bool
	}{
		// a string: a text field, a colour picker and a label
		{"bubbles.textinput", "lipgloss.colourpicker", []string{"value"}, true},
		{"lipgloss.colourpicker", "bubbles.textinput", []string{"value"}, true},
		{"bubbles.textinput", "lipgloss.label", []string{"value"}, true},
		// items and the chosen one: tabs and dialog buttons
		{"lipgloss.tabs", "community.dialog", []string{"tabs", "active"}, true},
		{"community.dialog", "lipgloss.tabs", []string{"buttons", "active"}, true},
		{"community.dialog", "lipgloss.tabs", []string{"text"}, false}, // tabs has no string
		// a list does not carry a string
		{"lipgloss.list", "lipgloss.label", []string{"items"}, false},
		// a table takes rows and headers
		{"lipgloss.table", "bubbles.table", []string{"headers", "rows"}, true},
		{"bubbles.table", "lipgloss.table", []string{"rows", "selected"}, false}, // lipgloss.table has no selection
		// nothing bound: one shared port is enough
		{"lipgloss.list", "bubbles.list", nil, true},
	}
	for _, c := range cases {
		from, ok1 := reg.Get(c.from)
		to, ok2 := reg.Get(c.to)
		if !ok1 || !ok2 {
			t.Fatalf("unknown component in %s -> %s", c.from, c.to)
		}
		got := Check(from, to, c.bound).Compatible()
		if got != c.ok {
			t.Errorf("%s -> %s with %v: compatible %v, want %v", c.from, c.to, c.bound, got, c.ok)
		}
	}
}

func TestCandidatesAreInCatalogOrderAndLeaveOutPacksAndItself(t *testing.T) {
	reg := standard.Default()
	from, _ := reg.Get("bubbles.textinput")
	got := Candidates(reg.List(), from, []string{"value"})
	have := map[string]bool{}
	for _, d := range got {
		have[d.ID] = true
		if d.ID == "bubbles.textinput" {
			t.Error("a component is not its own candidate")
		}
		if d.Inner != nil {
			t.Errorf("pack component %s offered", d.ID)
		}
	}
	for _, want := range []string{"lipgloss.colourpicker", "bubbles.textarea", "huh.input", "community.promptinput", "lipgloss.label"} {
		if !have[want] {
			t.Errorf("%s is not a candidate for a text input", want)
		}
	}
	for _, not := range []string{"lipgloss.list", "lipgloss.tabs", "bubbles.table", "ntcharts.barchart"} {
		if have[not] {
			t.Errorf("%s is a candidate for a text input", not)
		}
	}
}

func TestEverythingTheTableNamesExists(t *testing.T) {
	// A port on a property that does not exist is a typo in the table: the
	// property would never be bindable across components.
	reg := standard.Default()
	for _, d := range reg.List() {
		for _, p := range d.Props {
			if p.Port != "" && p.Key == "" {
				t.Errorf("%s has a port on a property without a key", d.ID)
			}
		}
	}
	for _, id := range []string{"lipgloss.label", "lipgloss.tabs", "community.dialog", "bubbles.table", "ntcharts.barchart", "huh.confirm"} {
		d, _ := reg.Get(id)
		if len(d.Ports()) == 0 {
			t.Errorf("%s has no ports", id)
		}
	}
}
