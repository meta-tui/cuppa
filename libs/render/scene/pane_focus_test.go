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
	if c := on.At(0, 0); c.Style.Bg != "212" {
		t.Errorf("focused: selected background %q, want the accent", c.Style.Bg)
	}
	props["focused"] = "false"
	off := rowsGrid(30, 2, props)
	if c := off.At(0, 0); c.Style.Bg != "" || c.Style.Fg != "212" {
		t.Errorf("without focus: fg %q bg %q, want accent text on no background", c.Style.Fg, c.Style.Bg)
	}
	if c := off.At(29, 0); c.Style.Bg != "" {
		t.Errorf("without focus the row is still filled: %q", c.Style.Bg)
	}

	box := design.Node{Component: "lipgloss.box", Rect: design.Rect{W: 10, H: 3},
		Props: map[string]string{"color": "212", "blurColor": "240", "focused": "false"}}
	if c := RenderNode(box, standard.Default()).At(0, 0); c.Style.Fg != "240" {
		t.Errorf("blurred border %q, want 240", c.Style.Fg)
	}
	box.Props["focused"] = "true"
	if c := RenderNode(box, standard.Default()).At(0, 0); c.Style.Fg != "212" {
		t.Errorf("focused border %q, want 212", c.Style.Fg)
	}
}
