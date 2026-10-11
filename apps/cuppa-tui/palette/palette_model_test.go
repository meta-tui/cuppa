package palette

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/apps/cuppa-tui/pointer"
	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
)

func newPalette() *Model {
	m := New(standard.Default())
	m.SetSize(26, 10)
	return m
}

func click(m *Model, y int) string {
	return m.Handle(pointer.Event{X: 3, Y: y, Phase: pointer.Down, Left: true})
}

func TestPaletteSizesItsOutputExactly(t *testing.T) {
	m := newPalette()
	if got := len(m.Lines()); got != 10 {
		t.Fatalf("lines = %d", got)
	}
}

func TestClickingAnItemStartsADrag(t *testing.T) {
	m := newPalette()
	// Row 2 is the first family header, row 3 its first item.
	if id := click(m, 3); id == "" {
		t.Fatal("clicking an item should start a drag")
	}
	if id := click(m, 2); id != "" {
		t.Fatal("clicking a header must not start a drag")
	}
}

func TestHeaderCollapsesFamily(t *testing.T) {
	m := newPalette()
	before := len(m.rows)
	click(m, 2)
	if len(m.rows) >= before {
		t.Fatalf("collapse did not hide rows: %d -> %d", before, len(m.rows))
	}
	click(m, 2)
	if len(m.rows) != before {
		t.Fatal("expanding did not restore rows")
	}
}

func TestWheelScrollsAndClamps(t *testing.T) {
	m := newPalette()
	m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: 1})
	if m.scroll != 3 {
		t.Fatalf("scroll = %d", m.scroll)
	}
	for i := 0; i < 50; i++ {
		m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: 1})
	}
	if m.scroll != len(m.rows)-m.listHeight() {
		t.Fatalf("scroll not clamped: %d", m.scroll)
	}
	m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: -1})
	m.Handle(pointer.Event{Phase: pointer.Wheel, WheelY: -1})
}

func TestSearchFiltersComponents(t *testing.T) {
	m := newPalette()
	click(m, 1)
	if !m.Searching() {
		t.Fatal("clicking the search row should focus it")
	}
	for _, r := range "spin" {
		m.Key(string(r), false, false, false)
	}
	var ids []string
	for _, r := range m.rows {
		ids = append(ids, r.def.ID)
	}
	if len(ids) < 2 || ids[0] != "bubbles.spinner" || ids[1] != "huh.spinner" {
		t.Fatalf("rows = %v", ids)
	}
	m.Key("", false, false, true)
	if m.query != "" || len(m.rows) < 5 {
		t.Fatal("escape should clear the search")
	}
}
