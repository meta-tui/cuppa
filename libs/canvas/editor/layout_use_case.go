package editor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/meta-ui-labs/cuppa/libs/document/design"
	"github.com/meta-ui-labs/cuppa/libs/layout/expr"
)

// Axes accepted by SetLayout.
const (
	AxisX = "x"
	AxisY = "y"
	AxisW = "w"
	AxisH = "h"
)

// SetLayout gives one axis of the node a size expression ("50%",
// "100% - 10"), as one undo step. An empty value makes the axis fixed again at
// its current cells, and a plain number of cells sets the axis to that. The
// node is resolved at the current canvas size at once.
func (e *Editor) SetLayout(id design.NodeID, axis, value string) error {
	n, ok := e.doc.Get(id)
	if !ok {
		return fmt.Errorf("editor: no such component")
	}
	if n.Locked {
		return fmt.Errorf("editor: the component is locked")
	}
	value = strings.TrimSpace(value)
	field := layoutField(&n.Layout, axis)
	if field == nil {
		return fmt.Errorf("editor: no layout axis %q", axis)
	}
	fixed, cells := false, 0
	if value != "" {
		x, err := expr.Parse(value)
		if err != nil {
			return fmt.Errorf("editor: %w", err)
		}
		if x.IsFixed() {
			// A number of cells is not an expression: the axis becomes fixed.
			fixed, cells, value = true, x.Resolve(0), ""
		}
	}
	if *field == value && !fixed {
		return nil
	}
	e.apply(func() bool {
		r := n.Rect
		if fixed {
			switch axis {
			case AxisX:
				r.X = cells
			case AxisY:
				r.Y = cells
			case AxisW:
				r.W = cells
			case AxisH:
				r.H = cells
			}
			r = e.clampRect(n, r)
		}
		e.doc.Update(id, func(n *design.Node) {
			*layoutField(&n.Layout, axis) = value
			n.Rect = r
		})
		e.resolveNode(id)
		return true
	})
	return nil
}

// Layout returns the node's size expressions.
func (e *Editor) Layout(id design.NodeID) design.Layout {
	n, _ := e.doc.Get(id)
	return n.Layout
}

// Resolve recomputes the rectangle of every component that has layout
// expressions for the current canvas size. It is not an undo step: it follows
// from the document.
func (e *Editor) Resolve() {
	for _, n := range e.doc.Nodes {
		e.resolveNode(n.ID)
	}
}

func layoutField(l *design.Layout, axis string) *string {
	switch axis {
	case AxisX:
		return &l.X
	case AxisY:
		return &l.Y
	case AxisW:
		return &l.W
	case AxisH:
		return &l.H
	}
	return nil
}

// resolveNode applies the node's expressions to its rectangle, keeping the
// component's minimum size and the canvas bounds.
func (e *Editor) resolveNode(id design.NodeID) {
	n, ok := e.doc.Get(id)
	if !ok || n.Layout.IsZero() {
		return
	}
	before := n
	r := n.Rect
	w, h := e.doc.Width, e.doc.Height
	env := e.layoutEnv()
	pick := func(src string, parent int, cur int) int {
		if src == "" {
			return cur
		}
		x, err := expr.Parse(src)
		if err != nil {
			return cur
		}
		return x.ResolveIn(parent, env)
	}
	r.W = pick(n.Layout.W, w, r.W)
	r.H = pick(n.Layout.H, h, r.H)
	minW, minH := e.minSize(n)
	r.W = min(max(r.W, minW), w)
	r.H = min(max(r.H, minH), h)
	r.X = pick(n.Layout.X, w, r.X)
	r.Y = pick(n.Layout.Y, h, r.Y)
	r = r.MoveInto(e.doc.Bounds())
	e.doc.Update(id, func(n *design.Node) { n.Rect = r })
	// The members of a container follow it when the canvas size moved it.
	e.shiftMembers(before, nil)
}

// layoutEnv is what an expression can read: the inputs of the screen with the
// value they have in the design (the bound property's value, or true for a
// show-if), and the places the components before it have now.
func (e *Editor) layoutEnv() *expr.Env {
	inputs := map[string]float64{}
	bound := map[string]bool{}
	var scan func(nodes []design.Node)
	scan = func(nodes []design.Node) {
		for _, n := range nodes {
			for key, name := range n.Bind {
				bound[name] = true
				value, set := n.Props[key]
				if def, ok := e.cat.Get(n.Component); ok && !set {
					if spec, ok := def.Prop(key); ok {
						value = spec.Default
					}
				}
				switch value {
				case "true":
					inputs[name] = 1
				case "false":
					inputs[name] = 0
				default:
					if f, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil {
						inputs[name] = f
					}
				}
			}
			if n.ShowIf != "" && !expr.IsCondition(n.ShowIf) {
				if _, known := inputs[n.ShowIf]; !known {
					inputs[n.ShowIf] = 1
				}
			}
			scan(n.Children)
		}
	}
	scan(e.doc.Nodes)
	// An input only an expression reads has the value the design gave it to
	// preview with; one a property is bound to keeps that property's value.
	for name, v := range e.doc.Samples {
		if f, ok := design.SampleNumber(v); ok && !bound[name] {
			inputs[name] = f
		}
	}
	return &expr.Env{
		W: e.doc.Width, H: e.doc.Height,
		Input: func(name string) (float64, bool) { v, ok := inputs[name]; return v, ok },
		Rect: func(name, field string) (float64, bool) {
			for _, n := range e.doc.Nodes {
				if n.Name == name {
					return expr.RectField(n.Rect.X, n.Rect.Y, n.Rect.W, n.Rect.H, field), true
				}
			}
			return 0, false
		},
	}
}

// followRect writes a drag or resize back into the node's expressions so each
// axis keeps its unit: a node 50% wide stays a percentage. before is the
// rectangle prior to the change.
func (e *Editor) followRect(id design.NodeID, before design.Rect) {
	n, ok := e.doc.Get(id)
	if !ok || n.Layout.IsZero() {
		return
	}
	w, h := e.doc.Width, e.doc.Height
	shift := func(src string, delta, parent int) string {
		if src == "" || delta == 0 {
			return src
		}
		x, err := expr.Parse(src)
		if err != nil {
			return src
		}
		return x.Shift(delta, parent).String()
	}
	l := n.Layout
	l.X = shift(l.X, n.Rect.X-before.X, w)
	l.Y = shift(l.Y, n.Rect.Y-before.Y, h)
	l.W = shift(l.W, n.Rect.W-before.W, w)
	l.H = shift(l.H, n.Rect.H-before.H, h)
	e.doc.Update(id, func(n *design.Node) { n.Layout = l })
}

// ToggleLayoutUnit switches an axis between fixed cells and a percentage of
// the canvas, keeping its current size and position, as one undo step.
func (e *Editor) ToggleLayoutUnit(id design.NodeID, axis string) error {
	n, ok := e.doc.Get(id)
	if !ok {
		return fmt.Errorf("editor: no such component")
	}
	field := layoutField(&n.Layout, axis)
	if field == nil {
		return fmt.Errorf("editor: no layout axis %q", axis)
	}
	if *field != "" {
		return e.SetLayout(id, axis, "")
	}
	cells, parent := n.Rect.X, e.doc.Width
	switch axis {
	case AxisY:
		cells, parent = n.Rect.Y, e.doc.Height
	case AxisW:
		cells = n.Rect.W
	case AxisH:
		cells, parent = n.Rect.H, e.doc.Height
	}
	return e.SetLayout(id, axis, expr.PercentOf(cells, parent).String())
}
