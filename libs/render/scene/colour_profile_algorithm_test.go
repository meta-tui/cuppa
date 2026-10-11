package scene

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
)

func coloured(profile string) design.Document {
	doc := design.NewDocument("t", 20, 6)
	doc.Add(design.Node{Component: "lipgloss.label", Name: "L", Rect: design.Rect{X: 1, Y: 1, W: 10, H: 1},
		Props: map[string]string{"text": "Hi", "color": "#ff5fd7", "background": "21", "bold": "true"}})
	doc.Profile = profile
	return doc
}

func TestTrueColourKeepsEveryColour(t *testing.T) {
	c := Render(coloured(""), standard.Default()).At(1, 1)
	if c.Fg != "#ff5fd7" || c.Bg != "21" {
		t.Fatalf("cell = %+v", c)
	}
}

func Test256TurnsHexIntoThePalette(t *testing.T) {
	c := Render(coloured(design.Profile256), standard.Default()).At(1, 1)
	if c.Fg != "206" || c.Bg != "21" || !c.Bold {
		t.Fatalf("hex becomes the nearest palette colour, indexes stay: %+v", c)
	}
}

func Test16UsesOnlyTheSystemColours(t *testing.T) {
	c := Render(coloured(design.Profile16), standard.Default()).At(1, 1)
	for _, v := range []string{c.Fg, c.Bg} {
		if v == "" || len(v) > 2 {
			t.Fatalf("a system colour index is expected, got %q (%+v)", v, c)
		}
	}
	if c.Fg != "13" { // #ff5fd7 is closest to bright magenta
		t.Errorf("fg = %q", c.Fg)
	}
}

func TestNoneRemovesColourButKeepsEmphasis(t *testing.T) {
	c := Render(coloured(design.ProfileNone), standard.Default()).At(1, 1)
	if c.Fg != "" || c.Bg != "" || !c.Bold || c.Ch != 'H' {
		t.Fatalf("text and bold stay, colour goes: %+v", c)
	}
}
