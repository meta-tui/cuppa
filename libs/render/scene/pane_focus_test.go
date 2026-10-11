package scene

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

// A pane without focus shows its selected row as accent text, not inverted,
// and its box in the blurred colour.
func TestAPaneWithoutFocusShowsAccentTextAndTheBlurredBorder(t *testing.T) {
	props := map[string]string{
		"columns": "Name:8,Value",
		"rows":    "Accent,#ff007f;Focus,#00f0ff",
		"styles":  "selected,normal",
		"color":   "212",
	}
	on := rowsGrid(30, 2, props)
	if c := on.At(0, 0); c.Bg != "212" {
		t.Errorf("focused: selected background %q, want the accent", c.Bg)
	}
	props["focused"] = "false"
	off := rowsGrid(30, 2, props)
	if c := off.At(0, 0); c.Bg != "" || c.Fg != "212" {
		t.Errorf("without focus: fg %q bg %q, want accent text on no background", c.Fg, c.Bg)
	}
	if c := off.At(29, 0); c.Bg != "" {
		t.Errorf("without focus the row is still filled: %q", c.Bg)
	}

	box := design.Node{Component: "lipgloss.box", Rect: design.Rect{W: 10, H: 3},
		Props: map[string]string{"color": "212", "blurColor": "240", "focused": "false"}}
	if c := RenderNode(box, standard.Default()).At(0, 0); c.Fg != "240" {
		t.Errorf("blurred border %q, want 240", c.Fg)
	}
	box.Props["focused"] = "true"
	if c := RenderNode(box, standard.Default()).At(0, 0); c.Fg != "212" {
		t.Errorf("focused border %q, want 212", c.Fg)
	}
}

// Rows without focus can draw the selected row's text in a colour of their
// own, apart from the one the focused row is filled with.
func TestRowsWithoutFocusUseTheirBlurColour(t *testing.T) {
	props := map[string]string{
		"columns": "Name:8,Value",
		"rows":    "Accent,#ff007f;Focus,#00f0ff",
		"styles":  "selected,normal",
		"color":   "212",
		"focused": "false",
	}
	if c := rowsGrid(30, 2, props).At(0, 0); c.Fg != "212" {
		t.Errorf("no blur colour: fg %q, want the colour", c.Fg)
	}
	props["blurColor"] = "51"
	g := rowsGrid(30, 2, props)
	if c := g.At(0, 0); c.Fg != "51" || c.Bg != "" {
		t.Errorf("blur colour: fg %q bg %q", c.Fg, c.Bg)
	}
	props["focused"] = "true"
	if c := rowsGrid(30, 2, props).At(0, 0); c.Bg != "212" {
		t.Errorf("focused: bg %q, want the colour", c.Bg)
	}
}
