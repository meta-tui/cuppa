// Package hittest answers "what is under this cell" for the canvas.
package hittest

import (
	"github.com/meta-tui/cuppa/libs/document/container"
	"github.com/meta-tui/cuppa/libs/document/design"
	"github.com/meta-tui/cuppa/libs/document/drawlayer"
)

// Node returns the topmost visible node covering the cell (x, y). Hidden
// nodes cannot be hit, and neither can a member of a page that is not shown.
func Node(doc design.Document, x, y int) (design.NodeID, bool) {
	order := container.Order(doc)
	for k := len(order) - 1; k >= 0; k-- {
		n := doc.Nodes[order[k]]
		if hittable(doc, n) && n.Rect.Contains(x, y) {
			return n.ID, true
		}
	}
	return "", false
}

// NodesIn returns the visible nodes whose rectangle touches r, back to front.
func NodesIn(doc design.Document, r design.Rect) []design.NodeID {
	var out []design.NodeID
	for _, i := range container.Order(doc) {
		if n := doc.Nodes[i]; hittable(doc, n) && n.Rect.Intersects(r) {
			out = append(out, n.ID)
		}
	}
	return out
}

func hittable(doc design.Document, n design.Node) bool {
	return !n.Hidden && n.Component != drawlayer.Component && container.Visible(doc, n)
}
