package scene

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

func TestTheDrawingLeavesWhatIsUnderItVisible(t *testing.T) {
	doc := design.NewDocument("t", 40, 10)
	doc.Add(design.Node{Component: "lipgloss.label", Name: "L", Rect: design.Rect{X: 2, Y: 2, W: 10, H: 1}, Props: map[string]string{"text": "hello"}})
	l := drawlayer.New()
	l.Paint(drawlayer.Cell{X: 20, Y: 5, Ch: '#', Fg: "212"})
	doc.Add(design.Node{Component: drawlayer.Component, Name: "Drawing", Rect: design.Rect{W: 40, H: 10},
		Props: map[string]string{drawlayer.PropCells: l.Encode()}})
	g := Render(doc, standard.Default())
	if c := g.At(2, 2); c.Ch != 'h' {
		t.Fatalf("the label under the drawing is covered: %q", c.Ch)
	}
	if c := g.At(20, 5); c.Ch != '#' || c.Fg != "212" {
		t.Fatalf("the drawn cell = %+v", c)
	}
}
