package editor

import (
	"testing"

	"github.com/meta-ui-labs/cuppa/libs/catalog/standard"
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

func newDrawingEditor() *Editor {
	return New(standard.Default(), design.NewDocument("t", 40, 20))
}

func TestTheFirstStrokeMakesOneDrawingLayerAndLaterOnesReuseIt(t *testing.T) {
	e := newDrawingEditor()
	e.ApplyDrawing(func(l *drawlayer.Layer) { l.Paint(drawlayer.Cell{X: 1, Y: 1, Ch: '#'}) })
	e.ApplyDrawing(func(l *drawlayer.Layer) { l.Paint(drawlayer.Cell{X: 2, Y: 1, Ch: '#'}) })
	n := 0
	for _, node := range e.Document().Nodes {
		if node.Component == drawlayer.Component {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d drawing layers, want 1", n)
	}
	if e.Drawing().Len() != 2 {
		t.Fatalf("%d cells, want 2", e.Drawing().Len())
	}
}

func TestEachStrokeIsOneUndoStep(t *testing.T) {
	e := newDrawingEditor()
	e.ApplyDrawing(func(l *drawlayer.Layer) { l.Paint(drawlayer.Cell{X: 1, Y: 1, Ch: '#'}) })
	e.ApplyDrawing(func(l *drawlayer.Layer) { l.Paint(drawlayer.Cell{X: 2, Y: 1, Ch: '#'}, drawlayer.Cell{X: 3, Y: 1, Ch: '#'}) })
	e.Undo()
	if e.Drawing().Len() != 1 {
		t.Fatalf("after one undo %d cells, want 1", e.Drawing().Len())
	}
	e.Undo()
	if len(e.Document().Nodes) != 0 {
		t.Fatal("undoing the first stroke should remove the layer")
	}
}

func TestErasingOnlyTouchesTheDrawing(t *testing.T) {
	e := newDrawingEditor()
	if _, err := e.Add("lipgloss.box", 0, 0); err != nil {
		t.Fatal(err)
	}
	e.ApplyDrawing(func(l *drawlayer.Layer) { l.Paint(drawlayer.Cell{X: 0, Y: 0, Ch: '#'}) })
	e.ApplyDrawing(func(l *drawlayer.Layer) { l.Erase(drawlayer.Cell{X: 0, Y: 0}) })
	if e.Drawing().Len() != 0 {
		t.Fatal("the drawing was not erased")
	}
	boxes := 0
	for _, n := range e.Document().Nodes {
		if n.Component == "lipgloss.box" {
			boxes++
		}
	}
	if boxes != 1 {
		t.Fatal("erasing removed a component")
	}
}

func TestALockedDrawingRefusesStrokes(t *testing.T) {
	e := newDrawingEditor()
	e.ApplyDrawing(func(l *drawlayer.Layer) { l.Paint(drawlayer.Cell{X: 1, Y: 1, Ch: '#'}) })
	id := e.Document().Nodes[0].ID
	e.SetLocked(id, true)
	if e.ApplyDrawing(func(l *drawlayer.Layer) { l.Paint(drawlayer.Cell{X: 5, Y: 5, Ch: '#'}) }) {
		t.Fatal("a locked drawing accepted a stroke")
	}
}
