package editor

import (
	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/document/drawlayer"
)

// drawingNode is the drawing layer's node, if the design has one.
func (e *Editor) drawingNode() (design.Node, bool) {
	for _, n := range e.doc.Nodes {
		if n.Component == drawlayer.Component {
			return n, true
		}
	}
	return design.Node{}, false
}

// Drawing returns the painted cells (an empty layer when nothing is drawn).
func (e *Editor) Drawing() *drawlayer.Layer {
	if n, ok := e.drawingNode(); ok {
		return drawlayer.Decode(n.Props[drawlayer.PropCells])
	}
	return drawlayer.New()
}

// ApplyDrawing lets change paint into or erase from the drawing, as one undo
// step. The first stroke creates the drawing layer, covering the canvas, in
// front of everything; there is only ever one. It reports false when the
// layer is locked or the change did nothing.
func (e *Editor) ApplyDrawing(change func(l *drawlayer.Layer)) bool {
	existing, has := e.drawingNode()
	if has && existing.Locked {
		return false
	}
	layer := e.Drawing()
	before := layer.Encode()
	change(layer)
	after := layer.Encode()
	if after == before {
		return false
	}
	return e.apply(func() bool {
		bounds := e.doc.Bounds()
		if !has {
			e.doc.Add(design.Node{
				Component: drawlayer.Component,
				Name:      "Drawing",
				Rect:      bounds,
				Props:     map[string]string{drawlayer.PropCells: after},
			})
			return true
		}
		e.doc.Update(existing.ID, func(n *design.Node) {
			n.Rect = bounds
			if n.Props == nil {
				n.Props = map[string]string{}
			}
			n.Props[drawlayer.PropCells] = after
		})
		return true
	})
}
