package editor

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func TestColoursAreNamedAndShared(t *testing.T) {
	e := newEditor()
	a := mustAdd(t, e, "lipgloss.label", 0, 0)
	b := mustAdd(t, e, "lipgloss.label", 0, 2)

	// A colour gets a name from its property and the palette holds the value.
	must(t, e.SetColour(a, "color", "#ff0000"))
	na, _ := e.Document().Get(a)
	if na.Props["color"] != "@Foreground" {
		t.Fatalf("a uses %q", na.Props["color"])
	}
	if c, _ := e.Document().Theme.Colour("Foreground"); c != "#ff0000" {
		t.Fatalf("palette = %+v", e.Document().Theme.Palette)
	}

	// The same property on another component gets its own name...
	must(t, e.SetColour(b, "color", "#00ff00"))
	nb, _ := e.Document().Get(b)
	if nb.Props["color"] != "@Foreground 2" {
		t.Fatalf("b uses %q", nb.Props["color"])
	}
	// ...until it is connected to the first one: then they follow each other.
	must(t, e.UseSwatch(b, "color", "Foreground"))
	must(t, e.SetColour(b, "color", "#0000ff"))
	if c, _ := e.Document().Theme.Colour("Foreground"); c != "#0000ff" {
		t.Fatalf("shared colour = %q", c)
	}
	if got := len(e.Swatches()); got != 2 {
		t.Fatalf("palette has %d swatches", got)
	}

	// A swatch in use cannot be removed; an unused one can.
	if err := e.RemoveSwatch("Foreground"); err == nil {
		t.Error("removed a swatch that is in use")
	}
	must(t, e.RemoveSwatch("Foreground 2"))
	if _, ok := e.Document().Theme.Colour("Foreground 2"); ok {
		t.Error("swatch kept")
	}

	// An unknown name is refused, and one undo reverts one change.
	if err := e.UseSwatch(a, "color", "Nope"); err == nil {
		t.Error("unknown swatch accepted")
	}
	e.Undo()
	if _, ok := e.Document().Theme.Colour("Foreground 2"); !ok {
		t.Error("undo did not bring the swatch back")
	}
	_ = design.Node{}
}

func TestEmptyColourLeavesThePaletteAlone(t *testing.T) {
	e := newEditor()
	a := mustAdd(t, e, "lipgloss.label", 0, 0)
	must(t, e.SetColour(a, "color", ""))
	if len(e.Swatches()) != 0 {
		t.Errorf("palette = %+v", e.Swatches())
	}
}
