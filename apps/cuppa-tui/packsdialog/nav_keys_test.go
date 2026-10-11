package packsdialog

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/definition"
)

func TestSpaceTogglesThePackUnderTheCursorAndRRemovesAnInstalledOne(t *testing.T) {
	entries := []Entry{
		{Pack: definition.Pack{ID: "a", Name: "A"}, Count: 1},
		{Pack: definition.Pack{ID: "b", Name: "B", Source: "b.cupp"}, Count: 2},
	}
	var toggled []definition.Family
	m := New(entries, func(definition.Family) bool { return true }, func(f definition.Family) { toggled = append(toggled, f) })
	m.Place(80, 24)
	m.Nav("space")
	m.Nav("down")
	m.Nav("space")
	if len(toggled) != 2 || toggled[0] != "a" || toggled[1] != "b" {
		t.Fatalf("toggled = %v, want a then b", toggled)
	}
	m.Nav("r")
	out, done := m.Outcome()
	if !done || out.Button != RemoveButton || out.Value != "b" {
		t.Fatalf("outcome %+v done=%v, want a removal of b", out, done)
	}
}
